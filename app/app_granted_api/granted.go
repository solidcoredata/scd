package app_granted_api

import (
	"context"
	"net/http"

	"github.com/gowww/router"
	"github.com/solidcoredata/scdhttp/scdhandler"
)

type Handler struct {
	Session scdhandler.SessionManager

	r *router.Router
}

var _ scdhandler.AppHandler = &Handler{}

func (h *Handler) Init(ctx context.Context) error {
	r := router.New()
	r.Get("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(loginGrantedHTML)
	}))
	r.Post("/api/logout", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rs, found := scdhandler.AuthFromContext(r.Context())
		if !found {
			panic("no auth context")
		}
		c, err := r.Cookie(rs.TokenKey)
		if err != nil {
			// If there is no cookie, user may already be logged out.
			return
		}
		err = h.Session.Logout(r.Context(), c.Value)
		if err != nil {
			http.Error(w, "unable to logout", http.StatusInternalServerError)
			return
		}
		// TODO(kardianos): set exire time, secure=true, strict origin.
		http.SetCookie(w, &http.Cookie{
			Name:   rs.TokenKey,
			Path:   "/",
			MaxAge: -1,
		})
	}))
	// TODO(kardianos): Add in additional API endpoints: proc, ui, delta, query, lookup, error.
	h.r = r
	return nil
}

func (h *Handler) URLPartition() (prefix string, consumeRedirect bool) {
	prefix = "/app1/"
	consumeRedirect = true
	return
}

// ServeHTTP returns an initial page with bootstrap loader.
// Provide API handlers for additional controls and requests.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.r.ServeHTTP(w, r)
}

var loginGrantedHTML = []byte(`<!DOCTYPE html>
<meta charset="UTF-8">

<title>Granted to $APP</title>

<h1>Granted to $APP</h1>

<h2>Hello</h2>
<div id=logout>logout</div>

<script>
var logoutButton = document.querySelector("#logout");

logoutButton.addEventListener("click", function(ev) {
	logout();
});
function logout() {
	var req = new XMLHttpRequest();
	req.onerror = function(ev) {
		alert("Unknown error, application may be down.");
	}
	req.onload = function(ev) {
		if(ev.target.status === 200) {
			location.pathname = "/";
			return;
		}
		// User may be already logged out. This may result in the
		// logout endpoint from being available.
		if(ev.target.status === 404) {
			location.pathname = "/";
			return;
		}
		alert("Unknown error, application may be down.");
	}
	req.open("POST", "api/logout", true);
	req.responseType = "text";
	req.send();
}
</script>
`)
