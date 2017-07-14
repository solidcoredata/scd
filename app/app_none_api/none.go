package app_none_api

import (
	"context"
	"net/http"
	"strings"

	"github.com/solidcoredata/scdhttp/scdhandler"
)

// TODO(kardianos): define an interface that allows returning static assets and can handle dynamic or static page requests.
/*
	Provide a list of static assets.
	Optionally map the static assets to a custom path.
	A custom path may be "/" to serve a static root HTTP page.

	Provide a list of dynamic handlers and their mount points.
	A custom path for a dynamic handler may be "/" to serve a dynamic root HTTP page.

	Use same interface for logged in state as well.

	This handlers methods are uniformally coposed.
	Handler may provide endpoints or provide endpoints.
	Both should be declared up front.

	Each component should declare what it uses. (yes)
*/

func NewHandler(session scdhandler.SessionManager) scdhandler.AppHandler {
	return &handler{
		ses: session,
	}
}

type handler struct {
	ses scdhandler.SessionManager
}

var _ scdhandler.AppHandler = &handler{}

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
		resp.ContentType = "text/html"
		resp.Body = loginNoneHTML
	case "/api/login":
		f, err := r.FormValues()
		if err != nil {
			return nil, scdhandler.HTTPError{Status: http.StatusInternalServerError, Err: err}
		}
		u, p := f.Value.Get("u"), f.Value.Get("p")
		u = strings.TrimSpace(u)
		p = strings.TrimSpace(p)

		token, err := h.ses.Login(ctx, u, p)
		if err != nil {
			return nil, scdhandler.HTTPError{Status: http.StatusForbidden, Msg: "bad login", Err: err}
		}
		rs, found := scdhandler.AuthFromContext(ctx)
		if !found {
			panic("no auth context")
		}
		resp.Header = make(map[string][]string, 1)
		// TODO(kardianos): set exire time, secure=true, strict origin.
		resp.Header.Add("Set-Cookie", (&http.Cookie{
			Name:     rs.TokenKey,
			Value:    token,
			Path:     "/",
			HttpOnly: true,
		}).String())
	}
	return resp, nil
}

func (h *handler) URLPartition() (prefix string, consumeRedirect bool) {
	prefix = "/login/"
	consumeRedirect = false
	return
}

var loginNoneHTML = []byte(`<!DOCTYPE html>
<meta charset="UTF-8">

<title>Login to $APP</title>

<h1>Login to $APP</h1>

<table>
	<tr>
		<td>&nbsp;
		<td><div id=message></div>
	<tr>
		<td><label for=username>Username</label>
		<td><input id=username>
	<tr>
		<td><label for=password>Password</label>
		<td><input id=password type=password>
	<tr>
		<td>&nbsp;
		<td><button id=login>Login</button>

<script>
var usernameInput = document.querySelector("#username");
var passwordInput = document.querySelector("#password");
var loginButton = document.querySelector("#login");
var messageEl = document.querySelector("#message");

loginButton.addEventListener("click", function(ev) {
	message("");
	login();
});
passwordInput.addEventListener("keypress", function(ev) {
	message("");
	if(ev.keyCode !== 13) {
		return;
	}
	login();
});
usernameInput.addEventListener("keypress", function(ev) {
	message("");
	if(ev.keyCode !== 13) {
		return;
	}
	passwordInput.select();
});
usernameInput.select();

function message(text) {
	messageEl.textContent = text;
}
function login() {
	var req = new XMLHttpRequest();
	req.onerror = function(ev) {
		message("Unknown error, application may be down.");
	}
	req.onload = function(ev) {
		if(ev.target.status === 403) {
			message("Incorrect username or password.");
			passwordInput.select();
			return;
		}
		if(ev.target.status === 200) {
			location.reload();
			return;
		}
		message("Unknown error, application may be down.");
	}
	req.open("POST", "api/login", true);
	req.responseType = "text";
	var d = new FormData();
	d.set("u", usernameInput.value);
	d.set("p", passwordInput.value);
	req.send(d);
}
</script>
`)
