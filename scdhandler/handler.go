// Package scdhandler provides various HTTP handlers for a server setup.
//
// The design is that each "application" can be plugged into the system.
// Eventually this will run in a k8s style components that are configured.
// At the moment use code to create servers.
//
// The most important design will be the end application under login_granted
// that will proof out various aspects of the system, including just-in-time
// component handling.
package scdhandler

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

/*
 * Client sends request to server.
 * HTTP Server receives request.
 * HTTP Server takes credential token(s) and send the token(s) to the Authentication Server to establish authentication, roles, and login state.
   The result is attached to the request context.
   - Login state (Logged Out, U2F Login Wait, Must Change Password, Logged In)
   - Elevated state (Normal, Elevated Login)
 * HTTP Server sends entire request with context to application back-end, switching off of URL and Login State.
 * ...
*/

// LoginState is a single value for each possible state a login can be in
// when checked.
type LoginState int64

// LoginRole is given meaning by the application and may be composed with
// each other for additional meaning.
// Multiple LoginRoles may be assigned to a request.
type LoginRole int64

//go:generate stringer -type=LoginState

// Known login states.
const (
	LoginError          LoginState = -1 // Error in system configuration, load, or disconnected backend component.
	LoginNone           LoginState = 0  // No credentials.
	LoginU2F            LoginState = 1  // Require user to identify with second factor.
	LoginChangePassword LoginState = 2  // Require user to change password.
	LoginGranted        LoginState = 3  // User has valid credentials.
)

type HTTPError struct {
	Status int
	Msg    string
	Err    error
}

func (e HTTPError) Error() string {
	msg := e.Msg
	if len(msg) == 0 {
		msg = e.Err.Error()
	}
	return fmt.Sprintf("%d: %s", e.Status, msg)
}

type MountProvide struct {
	// AllowCache should be set to true if the resource may be safely cached
	// while the component is loaded.
	AllowCache bool

	// Mount at this point.
	//  "/" Mount at the root.
	//  "/lib/staic/" mount directory.
	//  "/api/syscall" mount endpoint.
	At string

	// Endpoint name to call. Is this needed?
	// I don't think this is needed.
	Call string
}
type MountConsume struct {
	At string
}

type Request struct {
	// Method specifies the HTTP method (GET, POST, PUT, etc.).
	// For client requests an empty string means GET.
	Method string

	// URL specifies either the URI being requested (for server
	// requests) or the URL to access (for client requests).
	//
	// For server requests the URL is parsed from the URI
	// supplied on the Request-Line as stored in RequestURI.  For
	// most requests, fields other than Path and RawQuery will be
	// empty. (See RFC 2616, Section 5.1.2)
	// request.
	URL *url.URL

	// The protocol version for incoming server requests.
	ProtoMajor int16 // 1
	ProtoMinor int16 // 0

	Header http.Header

	// Body is the request's body.
	Body []byte

	// For server requests Host specifies the host on which the
	// URL is sought. Per RFC 2616, this is either the value of
	// the "Host" header or the host name given in the URL itself.
	// It may be of the form "host:port". For international domain
	// names, Host may be in Punycode or Unicode form. Use
	// golang.org/x/net/idna to convert it to either format if
	// needed.
	//
	// For client requests Host optionally overrides the Host
	// header to send. If empty, the Request.Write method uses
	// the value of URL.Host. Host may contain an international
	// domain name.
	Host string

	ContentType string

	// RemoteAddr allows HTTP servers and other software to record
	// the network address that sent the request, usually for
	// logging. This field is not filled in by ReadRequest and
	// has no defined format. The HTTP server in this package
	// sets RemoteAddr to an "IP:port" address before invoking a
	// handler.
	// This field is ignored by the HTTP client.
	RemoteAddr string

	// TLS allows HTTP servers and other software to record
	// information about the TLS connection on which the request
	// was received. This field is not filled in by ReadRequest.
	// The HTTP server in this package sets the field for
	// TLS-enabled connections before invoking a handler;
	// otherwise it leaves the field nil.
	// This field is ignored by the HTTP client.
	TLS *tls.ConnectionState
}

type Response struct {
	// Content type of the body.
	ContentType string

	// Encoding of the response. Often a compression method like "gzip" or "br".
	Encoding string

	Header http.Header

	// Response body.
	Body []byte
}

// AppHandler provides sufficent information to route incomming application
// requests and partition the URL namespace.
type AppHandler interface {
	// URLPartition returns the URL prefix and if an available redirect
	// should be removed path and if the prefix matches, redirected to.
	// The prefix should start and end with a slash "/".
	URLPartition() (prefix string, consumeRedirect bool)

	// Init is called after the application is loaded.
	Init(context.Context) error

	RequireMounts(ctx context.Context) ([]MountConsume, error)
	OptionalMounts(ctx context.Context) ([]MountConsume, error)
	ProvideMounts(ctx context.Context) ([]MountProvide, error)

	// Request should be routed by the r.URL.Path field.
	Request(ctx context.Context, r *Request) (*Response, error)

	Session() SessionManager
}

// LoginStateRouter links login states with a routable handler.
type LoginStateRouter struct {
	State map[LoginState]AppHandler

	// Authenticator is sent the token as found in the request cookie.
	// It is called after detaching the cookie value and before any routing has happened.
	Authenticator Authenticator
}

// URLRouter links an incomming URL Host with a LoginStateRouter.
type HostRouter map[string]LoginStateRouter

// RouteHandler routes requests after coming in from the edge of the system.
// It ensures requests are authenticated to any sensitive areas. It gives downstream
// components isolation.
type RouteHandler struct {
	// Router maps the request to a URL and login state.
	Router HostRouter

	// Loggerf allows feedback to the system.
	// TODO(kardianos): probably not the final interface.
	Loggerf func(f string, v ...interface{})
}

func (h *RouteHandler) Init(ctx context.Context) error {
	uniqueAuth := make(map[Authenticator]bool, len(h.Router))
	uniqueApp := make(map[AppHandler]bool, len(h.Router)*4)

	// Check each application handler to ensure the path prefix all
	// start and end with a "/".
	//
	// Add each authenticator and app to a unique map before calling Init on each.
	for host, logins := range h.Router {
		for state, login := range logins.State {
			partition, _ := login.URLPartition()
			if len(partition) == 0 || partition[0] != '/' || partition[len(partition)-1] != '/' {
				return fmt.Errorf(`URL Parition must begin and end with a slash "/" for %q for state %s.`, host, state)
			}
			uniqueApp[login] = true
		}
		uniqueAuth[logins.Authenticator] = true
	}
	for auth := range uniqueAuth {
		err := auth.Init(ctx)
		if err != nil {
			return err
		}
	}
	for app := range uniqueApp {
		err := app.Init(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *RouteHandler) logf(f string, v ...interface{}) {
	if h.Loggerf == nil {
		return
	}
	h.Loggerf(f, v...)
}

const redirectQueryKey = "redirect-to"

func (h *RouteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get values of credential tokens from request.
	// Get URL.
	// Send credential tokens to authentcation server.
	// Attach result to request context.
	// Send request to correct application server.

	urlrouter, ok := h.Router[r.Host]
	if !ok {
		h.logf("host 404: host=%q\n", r.URL.String())
		http.Error(w, "host 404", http.StatusNotFound)
		return
	}

	tokenKey := TokenKeyName(r.Host)

	rs := &RequestAuth{}
	if c, err := r.Cookie(tokenKey); err == nil {
		rs, err = urlrouter.Authenticator.RequestAuth(r.Context(), c.Value)
		if err != nil {
			h.logf("scdhandler: unable to check auth %v", err)
			http.Error(w, "failed to check auth", http.StatusInternalServerError)
			return
		}
	}
	rs.TokenKey = tokenKey

	next, ok := urlrouter.State[rs.LoginState]
	if !ok {
		http.Error(w, "login state 404", http.StatusNotFound)
		return
	}

	// Handle redirects to correct URL partition and redirects
	// to protected applications.
	prefix, useRedirect := next.URLPartition()
	switch useRedirect {
	case false:
		// If the redirect query value should not be consumed (like on a login application),
		// then set the redirect query key before redirecting.
		if strings.HasPrefix(r.URL.Path, prefix) == false {
			redirectQuery := url.Values{}
			if strings.Count(r.URL.Path, "/") >= 2 {
				redirectQuery.Set(redirectQueryKey, r.URL.RequestURI())
			}
			nextURL := &url.URL{Path: prefix, RawQuery: redirectQuery.Encode()}
			http.Redirect(w, r, nextURL.String(), http.StatusTemporaryRedirect)
			return
		}
	case true:
		rq := r.URL.Query()
		nextURL := rq.Get(redirectQueryKey)
		if len(nextURL) > 0 {
			// If there is not a prefix match, then remove it from
			// the query string and redirect to the same URL but without the
			// redirect.
			if !strings.HasPrefix(nextURL, prefix) {
				rq.Del(redirectQueryKey)
				r.URL.RawQuery = rq.Encode()
				nextURL = r.URL.String()
			}

			http.Redirect(w, r, nextURL, http.StatusTemporaryRedirect)
			return
		}

		// There is no redirect, but the prefix does not match the URL path.
		// Redirect to the correct prefix.
		if strings.HasPrefix(r.URL.Path, prefix) == false {
			http.Redirect(w, r, prefix, http.StatusTemporaryRedirect)
			return
		}
	}

	ctx := AuthNewContext(r.Context(), rs)
	r.URL.Path = strings.TrimPrefix(r.URL.Path, prefix[:len(prefix)-1])

	const readLimit = 1024 * 1024 * 100 // 100 MB. In the future make this property part of the AppHandler interface or RouteHandler.
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, readLimit))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	appReq := &Request{
		Method:      r.Method,
		URL:         r.URL,
		ProtoMajor:  int16(r.ProtoMajor),
		ProtoMinor:  int16(r.ProtoMinor),
		Body:        body,
		Header:      r.Header,
		ContentType: r.Header.Get("Content-Type"),
		Host:        r.Host,
		RemoteAddr:  r.RemoteAddr,
		TLS:         r.TLS,
	}

	appResp, err := next.Request(ctx, appReq)
	if err != nil {
		if status, ok := err.(HTTPError); ok {
			msg := status.Msg
			if len(msg) == 0 {
				msg = status.Err.Error()
			}
			http.Error(w, msg, status.Status)
			return
		}
		next, ok := urlrouter.State[rs.LoginState]
		if !ok {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		appReq.ContentType = "error"
		appReq.Body = []byte(err.Error())
		appResp, err = next.Request(ctx, appReq)
		if err != nil {
			http.Error(w, "unable to render error page: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if len(appResp.ContentType) > 0 {
		w.Header().Set("Content-Type", appResp.ContentType)
	}
	if len(appResp.Encoding) > 0 {
		w.Header().Set("Content-Encoding", appResp.Encoding)
	}
	for key, values := range appResp.Header {
		for _, v := range values {
			w.Header().Add(key, v)
		}
	}
	w.WriteHeader(http.StatusOK)
	w.Write(appResp.Body)
}
