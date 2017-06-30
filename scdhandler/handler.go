package scdhandler

import (
	"context"
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

// Known login states.
const (
	LoginNone           LoginState = 0
	LoginU2F            LoginState = 1
	LoginChangePassword LoginState = 2
	LoginGranted        LoginState = 100
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
}

// LoginStateRouter links login states with a routable handler.
type LoginStateRouter map[LoginState]AppHandler

// URLRouter links an incomming URL Host with a LoginStateRouter.
type URLRouter map[string]LoginStateRouter

// RequestAuth holds the authentication state for a request.
type RequestAuth struct {
	LoginState LoginState
	Roles      []LoginRole

	Identity   string // Unique name each application can use to link to the internal user.
	GivenName  string // The given (first) name of the user.
	FamilyName string // The family (last) name of the user.
}

// Authenticator provides the authentication to the request.
type Authenticator func(ctx context.Context, token string) (RequestAuth, error)

// Handler routes requests after coming in from the edge of the system.
// It ensures requests are authenticated to any sensitive areas. It gives downstream
// components isolation.
type Handler struct {
	// TokenKey is the name of the Cookie used to store authentication credentials.
	TokenKey string

	// Router maps the request to a URL and login state.
	Router URLRouter

	// Authenticator is sent the token as found in the request cookie.
	// It is called after detaching the cookie value and before any routing has happened.
	Authenticator Authenticator

	// Loggerf allows feedback to the system.
	// TODO(kardianos): probably not the final interface.
	Loggerf func(f string, v ...interface{})
}

func (h *Handler) logf(f string, v ...interface{}) {
	if h.Loggerf == nil {
		return
	}
	h.Loggerf(f, v...)
}

const redirectQueryKey = "redirect-to"

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get values of credential tokens from request.
	// Get URL.
	// Send credential tokens to authentcation server.
	// Attach result to request context.
	// Send request to correct application server.

	rs := RequestAuth{}
	if c, err := r.Cookie(h.TokenKey); err == nil {
		rs, err = h.Authenticator(r.Context(), c.Value)
		if err != nil {
			h.logf("scdhandler: unable to check auth %v", err)
			http.Error(w, "failed to check auth", http.StatusInternalServerError)
			return
		}
	}
	urlrouter, ok := h.Router[r.URL.Host]
	if !ok {
		http.Error(w, "host 404", http.StatusNotFound)
		return
	}
	next, ok := urlrouter[rs.LoginState]
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
			redirectQuery.Set(redirectQueryKey, r.URL.RequestURI())
			nextURL := &url.URL{Path: prefix, RawQuery: redirectQuery.Encode()}
			http.Redirect(w, r, nextURL, http.StatusTemporaryRedirect)
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

	r = r.WithContext(WithRequestAuth(r.Context(), rs))
	next.ServeHTTP(w, r)
}

type requestAuthKey struct{}

// WithRequestAuth returns a child context with the RequestAuth as a value.
func WithRequestAuth(ctx context.Context, rs RequestAuth) context.Context {
	return context.WithValue(ctx, requestAuthKey{}, rs)
}
