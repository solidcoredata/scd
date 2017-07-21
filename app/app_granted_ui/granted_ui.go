package app_granted_ui

import (
	"context"

	"github.com/solidcoredata/scdhttp/scdhandler"
)

type handler struct {
}

var _ scdhandler.AppComponentHandler = &handler{}

func NewHandler() scdhandler.AppComponentHandler {
	return &handler{}
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
	return []scdhandler.MountProvide{
		{At: "/"},
	}, nil
}

func (h *handler) Request(ctx context.Context, r *scdhandler.Request) (*scdhandler.Response, error) {
	resp := &scdhandler.Response{}
	switch r.URL.Path {
	case "/":
		resp.Body = loginGrantedHTML
		resp.ContentType = "text/html"
	}
	return resp, nil
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
