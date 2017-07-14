package app_granted_api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/solidcoredata/scdhttp/scdhandler"
)

type handler struct {
	ses scdhandler.SessionManager
}

var _ scdhandler.AppHandler = &handler{}

func NewHandler(session scdhandler.SessionManager) scdhandler.AppHandler {
	return &handler{
		ses: session,
	}
}

func (h *handler) Init(ctx context.Context) error {
	return nil
}

func (h *handler) RequireMounts(ctx context.Context) ([]scdhandler.MountConsume, error) {
	return nil, nil
}
func (h *handler) OptionalMounts(ctx context.Context) ([]scdhandler.MountConsume, error) {
	return nil, nil
}
func (h *handler) ProvideMounts(ctx context.Context) ([]scdhandler.MountProvide, error) {
	return nil, nil
}
func (h *handler) Session() scdhandler.SessionManager {
	return h.ses
}

func (h *handler) Request(ctx context.Context, r *scdhandler.Request) (*scdhandler.Response, error) {
	resp := &scdhandler.Response{}
	switch r.URL.Path {
	case "/":
		resp.Body = loginGrantedHTML
		resp.ContentType = "text/html"
	case "/api/logout":
		rs, found := scdhandler.AuthFromContext(ctx)
		if !found {
			panic("no auth context")
		}
		c, err := r.Cookie(rs.TokenKey)
		if err != nil {
			// If there is no cookie, user may already be logged out.
			return resp, nil
		}
		err = h.ses.Logout(ctx, c.Value)
		if err != nil {
			return nil, fmt.Errorf("unable to logout: %v", err)
		}
		resp.Header = make(map[string][]string, 1)
		// TODO(kardianos): set exire time, secure=true, strict origin.
		c = &http.Cookie{
			Name:   rs.TokenKey,
			Path:   "/",
			MaxAge: -1,
		}
		resp.Header.Add("Set-Cookie", c.String())
	}
	return resp, nil
}

func (h *handler) URLPartition() (prefix string, consumeRedirect bool) {
	prefix = "/app1/"
	consumeRedirect = true
	return
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
