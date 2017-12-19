// Copyright 2017 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// scdexample1 is an example application that runs in the solid core data environment.
package main

import (
	"context"
	"log"

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

	s := service.New()
	sc, err := NewServiceConfig()
	if err != nil {
		log.Fatalf("failed to create service config: %v", err)
	}
	s.Setup(ctx, sc)
}

var _ service.Configration = &ServiceConfig{}

func NewServiceConfig() (*ServiceConfig, error) {
	s := &ServiceConfig{}

	return s, nil
}

type ServiceConfig struct {
	code map[string]*service.ResourceFile
}

func (s *ServiceConfig) HTTPServer() (api.HTTPServer, bool) {
	return nil, false
}
func (s *ServiceConfig) AuthServer() (api.AuthServer, bool) {
	return nil, false
}
func (s *ServiceConfig) BundleUpdate(sb *api.ServiceBundle) {

}
