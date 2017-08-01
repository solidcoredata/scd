// Copyright 2017 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app_granted_api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/solidcoredata/scdhttp/scdhandler"
)

type sessionHandler struct {
	ses scdhandler.SessionManager
}

var _ scdhandler.AppComponentHandler = &sessionHandler{}

func NewSessionHandler(session scdhandler.SessionManager) scdhandler.AppComponentHandler {
	return &sessionHandler{
		ses: session,
	}
}

func (h *sessionHandler) Init(ctx context.Context) error {
	return nil
}

func (h *sessionHandler) ProvideMounts(ctx context.Context) ([]scdhandler.MountProvide, error) {
	return []scdhandler.MountProvide{
		{At: "/api/logout"},
	}, nil
}

func (h *sessionHandler) Request(ctx context.Context, r *scdhandler.Request) (*scdhandler.Response, error) {
	resp := &scdhandler.Response{}
	switch r.URL.Path {
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
