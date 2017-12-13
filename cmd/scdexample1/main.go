// Copyright 2017 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// scdexample1 is an example application that runs in the solid core data environment.
package main

import (
	"context"
	"log"

	"github.com/cortesi/moddwatch"
	"github.com/solidcoredata/scd/api"
	"github.com/solidcoredata/scd/service"
)

/*
	1. Push the configuration fetch into the service.Setup.
	2. Use res files for all services.
	3. Add some custom HTTP point to scdexample1.
	4. Add in a file watch.
	5. Push router config update.
	6. Add a hook to the UI to refresh components.
*/

func main() {
	ctx := context.TODO()

	sc, err := NewServiceConfig("res/scdexample1.jsonnet")
	if err != nil {
		log.Fatalf("failed to create service config: %v", err)
	}
	service.Setup(ctx, sc)
}

var _ service.Configration = &ServiceConfig{}

func NewServiceConfig(config string) (*ServiceConfig, error) {
	s := &ServiceConfig{
		bundle: make(chan *api.ServiceBundle, 5),
	}

	_ = moddwatch.Watch

	var err error
	s.staticConfig, s.code, err = service.OpenServiceConfiguration(config)
	if err != nil {
		return nil, err
	}
	s.bundle <- s.staticConfig
	return s, nil
}

type ServiceConfig struct {
	bundle chan *api.ServiceBundle

	staticConfig *api.ServiceBundle

	code map[string]string
}

func (s *ServiceConfig) ServiceBundle() <-chan *api.ServiceBundle {
	return s.bundle
}
func (s *ServiceConfig) HTTPServer() (api.HTTPServer, bool) {
	return nil, false
}
func (s *ServiceConfig) AuthServer() (api.AuthServer, bool) {
	return nil, false
}
func (s *ServiceConfig) SPAServer() (api.SPAServer, bool) {
	return s, true
}

func (s *ServiceConfig) FetchUI(ctx context.Context, req *api.FetchUIRequest) (*api.FetchUIResponse, error) {
	resp := &api.FetchUIResponse{}
	for _, name := range req.List {
		body, found := s.code[name]
		if !found {
			continue
		}
		resp.List = append(resp.List, &api.FetchUIItem{Name: name, Body: body})
	}
	return resp, nil
}

func (s *ServiceConfig) Config() chan<- *api.ServiceConfig {
	return nil
}
