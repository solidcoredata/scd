// Copyright 2017 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// scdexample1 is an example application that runs in the solid core data environment.
package main

import (
	"context"

	"github.com/solidcoredata/scd/api"
	"github.com/solidcoredata/scd/service"
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
	return s
}

type ServiceConfig struct {
	bundle chan *api.ServiceBundle

	staticConfig *api.ServiceBundle
}

func (s *ServiceConfig) ServiceBundle() chan *api.ServiceBundle {
	return s.bundle
}
func (s *ServiceConfig) RequestHanderServer() (api.RequestHanderServer, bool) {
	return nil, false
}
func (s *ServiceConfig) AuthServer() (api.AuthServer, bool) {
	return nil, false
}

func (s *ServiceConfig) createConfig() *api.ServiceBundle {
	c := &api.ServiceBundle{
		Name: "example-1.solidcoredata.org/app",
		Configured: []*api.ConfiguredResource{
			{Name: "auth/login", PotentialResourceName: "solidcoredata.org/auth/login", Configuration: &api.ConfiguredResource_URL{URL: &api.ConfigureURL{MapTo: "/api/login"}}},
			{Name: "auth/logout", PotentialResourceName: "solidcoredata.org/auth/logout", Configuration: &api.ConfiguredResource_URL{URL: &api.ConfigureURL{MapTo: "/api/logout"}}},
			{Name: "auth/endpoint", PotentialResourceName: "solidcoredata.org/auth/endpoint", Configuration: &api.ConfiguredResource_Auth{Auth: &api.ConfigureAuth{Area: api.ConfigureAuth_System, Environment: "DEV"}}},

			{Name: "ui/login", PotentialResourceName: "solidcoredata.org/base/login", Configuration: &api.ConfiguredResource_URL{URL: &api.ConfigureURL{MapTo: "/"}}},
			{Name: "ui/loader", PotentialResourceName: "solidcoredata.org/base/loader", Configuration: &api.ConfiguredResource_URL{URL: &api.ConfigureURL{MapTo: "/"}}},
			{Name: "ui/init.js", PotentialResourceName: "solidcoredata.org/base/init.js", Configuration: &api.ConfiguredResource_URL{URL: &api.ConfigureURL{MapTo: "/api/init.js"}}},
			{Name: "ui/fetch-ui", PotentialResourceName: "solidcoredata.org/base/fetch-ui", Configuration: &api.ConfiguredResource_URL{URL: &api.ConfigureURL{MapTo: "/api/fetch-ui"}}},
			{Name: "ui/favicon", PotentialResourceName: "solidcoredata.org/base/favicon", Configuration: &api.ConfiguredResource_URL{URL: &api.ConfigureURL{MapTo: "/api/favicon"}}},
		},
		Bundle: []*api.Bundle{
			{
				Name: "none",
				Include: []string{
					"example-1.solidcoredata.org/app/auth/login",
					"example-1.solidcoredata.org/app/ui/login",
				},
			},
			{
				Name: "granted",
				Include: []string{
					"example-1.solidcoredata.org/app/auth/logout",
					"example-1.solidcoredata.org/app/ui/loader",
					"example-1.solidcoredata.org/app/ui/init.js",
					"example-1.solidcoredata.org/app/ui/fetch-ui",
					"example-1.solidcoredata.org/app/ui/favicon",
				},
			},
		},
		Application: []*api.ApplicationBundle{
			{
				LoginBundle: []*api.LoginBundle{
					{
						LoginState:      api.LoginState_None,
						Prefix:          "/login/",
						ConsumeRedirect: false,
						Bundle:          "example-1.solidcoredata.org/app/none",
					},
					{
						LoginState:      api.LoginState_Granted,
						Prefix:          "/app/",
						ConsumeRedirect: true,
						Bundle:          "example-1.solidcoredata.org/app/granted",
					},
				},

				AuthConfiguredResource: "example-1.solidcoredata.org/app/auth/endpoint",
				Host: []string{"example1.solidcoredata.local"},
			},
		},
	}
	return c
}
