// Copyright 2017 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app_none_ui

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/draw"
	"image/png"

	"github.com/solidcoredata/scd/scdhandler"
)

func NewHandler() scdhandler.AppComponentHandler {
	return &handler{}
}

type handler struct {
}

var _ scdhandler.AppComponentHandler = &handler{}

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
		{At: "/ui/favicon"},
	}, nil
}
func (h *handler) Request(ctx context.Context, r *scdhandler.Request) (*scdhandler.Response, error) {
	resp := &scdhandler.Response{}
	switch r.URL.Path {
	case "/":
		resp.ContentType = "text/html"
		resp.Body = loginNoneHTML
	case "/ui/favicon":
		img := image.NewRGBA(image.Rect(0, 0, 192, 192))
		draw.Draw(img, img.Rect, image.NewUniform(color.RGBA{R: 255, A: 255}), image.ZP, draw.Over)
		buf := &bytes.Buffer{}
		png.Encode(buf, img)
		resp.ContentType = "image/png"
		resp.Body = buf.Bytes()
	}
	return resp, nil
}

var loginNoneHTML = []byte(`<!DOCTYPE html>
<meta charset="UTF-8">
<link rel="icon" href="ui/favicon">

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
