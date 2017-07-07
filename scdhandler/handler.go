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
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gowww/fatal"
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
	LoginNone           LoginState = 0
	LoginU2F            LoginState = 1
	LoginChangePassword LoginState = 2
	LoginGranted        LoginState = 3
)

// AppHandler provides sufficent information to route incomming application
// requests and partition the URL namespace.
type AppHandler interface {
	// Handler for the HTTP requests.
	http.Handler

	// URLPartition returns the URL prefix and if an available redirect
	// should be removed path and if the prefix matches, redirected to.
	// The prefix should start and end with a slash "/".
	URLPartition() (prefix string, consumeRedirect bool)

	// Init is called after the application is loaded.
	Init(context.Context) error
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

	recover *fatal.Options
}

func (h *RouteHandler) Init(ctx context.Context) error {
	h.recover = &fatal.Options{
		RecoverHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			v := fatal.Error(r)
			h.logf("panic: %v", v)
			http.Error(w, "panic", http.StatusInternalServerError)
		}),
	}

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

	urlrouter, ok := h.Router[r.URL.Host]
	if !ok {
		h.logf("host 404: host=%q\n", r.URL.String())
		http.Error(w, "host 404", http.StatusNotFound)
		return
	}

	tokenKey := TokenKeyName(r.URL.Host)

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

	r = r.WithContext(AuthNewContext(r.Context(), rs))
	r.URL.Path = strings.TrimPrefix(r.URL.Path, prefix[:len(prefix)-1])
	next.ServeHTTP(w, r)
}
