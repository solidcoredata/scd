package scdhandler

import (
	"net/http"
	"strings"

	"github.com/gowww/router"
)

// TODO(kardianos): define an interface that allows returning static assets and can handle dynamic or static page requests.
/*
	Provide a list of static assets.
	Optionally map the static assets to a custom path.
	A custom path may be "/" to serve a static root HTTP page.

	Provide a list of dynamic handlers and their mount points.
	A custom path for a dynamic handler may be "/" to serve a dynamic root HTTP page.

	Use same interface for logged in state as well.
*/

type LoginNoneHandler struct {
	Session SessionManager

	r *router.Router
}

func (h *LoginNoneHandler) Init() *LoginNoneHandler {
	r := router.New()
	// r.Get("/lib/", nil)
	r.Get("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(loginNoneHTML)
	}))
	r.Post("/api/login", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p := r.FormValue("u"), r.FormValue("p")
		u = strings.TrimSpace(u)
		p = strings.TrimSpace(p)

		token, err := h.Session.Login(r.Context(), u, p)
		if err != nil {
			http.Error(w, "bad login", http.StatusForbidden)
			return
		}
		rs, found := AuthFromContext(r.Context())
		if !found {
			http.Error(w, "unable to set cookie", http.StatusInternalServerError)
			return
		}
		// TODO(kardianos): set exire time, secure=true, strict origin.
		http.SetCookie(w, &http.Cookie{
			Name:     rs.TokenKey,
			Value:    token,
			Path:     "/",
			HttpOnly: true,
		})
	}))

	h.r = r
	return h
}

func (h *LoginNoneHandler) URLPartition() (prefix string, consumeRedirect bool) {
	prefix = "/login/"
	consumeRedirect = false
	return
}

// ServeHTTP displays a page or resources on GET or processes login on POST.
func (h *LoginNoneHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.r.ServeHTTP(w, r)
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
