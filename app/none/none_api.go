// Copyright 2017 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package none

import (
	"context"
	"net/http"
	"strings"

	"github.com/solidcoredata/scd/scdhandler"
)

func NewSessionHandler(session scdhandler.SessionManager) scdhandler.AppComponentHandler {
	return &sessionHandler{
		ses: session,
	}
}

type sessionHandler struct {
	ses scdhandler.SessionManager
}

var _ scdhandler.AppComponentHandler = &sessionHandler{}

func (h *sessionHandler) Init(ctx context.Context) error {
	return nil
}
func (h *sessionHandler) RequireMounts(ctx context.Context) ([]scdhandler.MountConsume, error) {
	return nil, nil
}
func (h *sessionHandler) OptionalMounts(ctx context.Context) ([]scdhandler.MountConsume, error) {
	return nil, nil
}
func (h *sessionHandler) ProvideMounts(ctx context.Context) ([]scdhandler.MountProvide, error) {
	return []scdhandler.MountProvide{
		{At: "/api/login"},
	}, nil
}
func (h *sessionHandler) Request(ctx context.Context, r *scdhandler.Request) (*scdhandler.Response, error) {
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
