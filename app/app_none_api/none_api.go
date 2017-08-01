// Copyright 2017 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app_none_api

import (
	"context"
	"net/http"
	"strings"

	"github.com/solidcoredata/scd/scdhandler"
)

func NewHandler(session scdhandler.SessionManager) scdhandler.AppComponentHandler {
	return &handler{
		ses: session,
	}
}

type handler struct {
	ses scdhandler.SessionManager
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
		{At: "/api/login"},
	}, nil
}
func (h *handler) Request(ctx context.Context, r *scdhandler.Request) (*scdhandler.Response, error) {
	resp := &scdhandler.Response{}
	switch r.URL.Path {
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
