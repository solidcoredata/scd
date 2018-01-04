// Copyright 2017 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// scdexample1 is an example application that runs in the solid core data environment.
package main

import (
	"context"
	"sync"

	"github.com/solidcoredata/scd/api"
	"github.com/solidcoredata/scd/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

/*
	1. [x] Push the configuration fetch into the service.Setup.
	2. [x] Use res files for all services.
	3. [ ] Add some custom HTTP point to scdexample1.
	4. [x] Add in a file watch.
	5. [x] Push router config update.
	6. [ ] Add a hook to the UI to refresh components.
*/

func main() {
	ctx := context.TODO()

	s := service.New()
	sc := &ServiceConfig{service: s}
	s.Setup(ctx, sc)
}

var _ service.Configration = &ServiceConfig{}

type ServiceConfig struct {
	service *service.Service

	mu sync.RWMutex
	sb *api.ServiceBundle
}

func (s *ServiceConfig) HTTPServer() (api.HTTPServer, bool) {
	return s, true
}
func (s *ServiceConfig) AuthServer() (api.AuthServer, bool) {
	return nil, false
}
func (s *ServiceConfig) BundleUpdate(sb *api.ServiceBundle) {
	s.mu.Lock()
	s.sb = sb
	s.mu.Unlock()
}

func (s *ServiceConfig) ServeHTTP(ctx context.Context, r *api.HTTPRequest) (*api.HTTPResponse, error) {
	resp := &api.HTTPResponse{}
	switch r.URL.Path {
	default:
		return nil, grpc.Errorf(codes.NotFound, "path %q not found", r.URL.Path)
	case "proc":
		s.mu.RLock()
		sb := s.sb
		s.mu.RUnlock()

		_ = sb

		setup, found := s.service.ResConn(r.Version)
		if !found {
			return nil, grpc.Errorf(codes.NotFound, "version %q not found", r.Version)
		}

		_ = setup

		// sb.Resource[0].
		// resp.Body
	}
	return resp, nil
}
