// Copyright 2017 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app_granted_ui

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/draw"
	"image/png"

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
		{At: "/ui/favicon"},
	}, nil
}

func (h *handler) Request(ctx context.Context, r *scdhandler.Request) (*scdhandler.Response, error) {
	resp := &scdhandler.Response{}
	switch r.URL.Path {
	case "/":
		resp.Body = loginGrantedHTML
		resp.ContentType = "text/html"
	case "/ui/favicon":
		img := image.NewRGBA(image.Rect(0, 0, 192, 192))
		draw.Draw(img, img.Rect, image.NewUniform(color.RGBA{G: 255, A: 255}), image.ZP, draw.Over)
		buf := &bytes.Buffer{}
		png.Encode(buf, img)
		resp.ContentType = "image/png"
		resp.Body = buf.Bytes()
	}
	return resp, nil
}

var loginGrantedHTML = []byte(`<!DOCTYPE html>
<meta charset="UTF-8">
<link rel="icon" href="ui/favicon">

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
<script src="api/init.js"></script>
`)
