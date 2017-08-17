// Copyright 2017 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// scdauth is a authorization service for solid core data systems.
package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/solidcoredata/scd/api"
	"github.com/solidcoredata/scd/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func main() {
	ctx := context.TODO()
	service.Setup(ctx, NewServiceConfig())
}

var _ service.ServiceConfigration = &ServiceConfig{}

func NewServiceConfig() *ServiceConfig {
	s := &ServiceConfig{
		bundle: make(chan *api.ServiceBundle, 5),
	}

	s.staticConfig = s.createConfig()
	s.bundle <- s.staticConfig
	s.am = NewAuthenticateMemory(
		&MemoryUser{
			ID:       1,
			Identity: "u1",
			Password: "p1",
		},
		&MemoryUser{
			ID:       2,
			Identity: "u2",
			Password: "p2",
		},
	)
	return s
}

type ServiceConfig struct {
	bundle chan *api.ServiceBundle

	staticConfig *api.ServiceBundle
	am           *AuthenticateMemory
}

func (s *ServiceConfig) createConfig() *api.ServiceBundle {
	c := &api.ServiceBundle{
		Name: "solidcoredata.org/auth",
		Potential: []*api.PotentialResource{
			{Name: "login", Type: api.PotentialResource_ResourceURL},
			{Name: "logout", Type: api.PotentialResource_ResourceURL},
			{Name: "endpoint", Type: api.PotentialResource_ResourceAuth},
		},
	}
	return c
}

func (s *ServiceConfig) ServiceBundle() chan *api.ServiceBundle {
	return s.bundle
}
func (s *ServiceConfig) RequestHanderServer() (api.RequestHanderServer, bool) {
	return s, true
}
func (s *ServiceConfig) AuthServer() (api.AuthServer, bool) {
	return s, true
}

func (s *ServiceConfig) Request(ctx context.Context, r *api.RequestReq) (*api.RequestResp, error) {
	resp := &api.RequestResp{}
	// TODO(kardianos): determine best way to notify client of bad login.
	switch r.URL.Path {
	default:
		return nil, grpc.Errorf(codes.NotFound, "path %q not found", r.URL.Path)
	case "login":
		f, err := r.FormValues()
		if err != nil {
			return nil, grpc.Errorf(codes.Internal, "unable to parse form values %v", err)
		}
		u, p := f.Value.Get("u"), f.Value.Get("p")
		u = strings.TrimSpace(u)
		p = strings.TrimSpace(p)

		token, err := s.am.Login(ctx, u, p)
		if err != nil {
			return nil, grpc.Errorf(codes.PermissionDenied, "bad login: %v", err)
		}
		rs, found := api.AuthFromContext(ctx)
		if !found {
			panic("no auth context")
		}
		resp.Header = &api.KeyValueList{}
		// TODO(kardianos): set exire time, secure=true, strict origin.
		resp.Header.Add("Set-Cookie", (&http.Cookie{
			Name:     rs.TokenKey,
			Value:    token,
			Path:     "/",
			HttpOnly: true,
		}).String())
	case "logout":
		rs, found := api.AuthFromContext(ctx)
		if !found {
			panic("no auth context")
		}
		c, err := r.Cookie(rs.TokenKey)
		if err != nil {
			// If there is no cookie, user may already be logged out.
			return resp, nil
		}
		err = s.am.Logout(ctx, c.Value)
		if err != nil {
			return nil, fmt.Errorf("unable to logout: %v", err)
		}
		resp.Header = &api.KeyValueList{}
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

func (s *ServiceConfig) RequestAuth(ctx context.Context, r *api.RequestAuthReq) (*api.RequestAuthResp, error) {
	return s.am.RequestAuth(ctx, r.Token)
}