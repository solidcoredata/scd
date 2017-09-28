// Copyright 2017 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// scdexample1 is an example application that runs in the solid core data environment.
package main

import (
	"context"
	"encoding/json"

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

func JSON(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

const serviceName = "example-1.solidcoredata.org/app"

func (s *ServiceConfig) createConfig() *api.ServiceBundle {
	c := &api.ServiceBundle{
		Name: serviceName,
		Resource: []*api.Resource{
			{Name: "auth/login", Parent: "solidcoredata.org/auth/login", Configuration: (&api.ConfigureURL{MapTo: "/api/login"}).EncodeMust()},
			{Name: "auth/logout", Parent: "solidcoredata.org/auth/logout", Configuration: (&api.ConfigureURL{MapTo: "/api/logout"}).EncodeMust()},
			{Name: "auth/endpoint", Parent: "solidcoredata.org/auth/endpoint", Configuration: (&api.ConfigureAuth{Area: api.ConfigureAuth_System, Environment: "DEV"}).EncodeMust()},

			{Name: "ui/login", Parent: "solidcoredata.org/base/login", Configuration: (&api.ConfigureURL{MapTo: "/"}).EncodeMust()},
			{Name: "ui/fetch-ui", Parent: "solidcoredata.org/base/fetch-ui", Configuration: (&api.ConfigureURL{MapTo: "/api/fetch-ui"}).EncodeMust()},
			{Name: "ui/favicon", Parent: "solidcoredata.org/base/favicon", Configuration: (&api.ConfigureURL{MapTo: "/ui/favicon"}).EncodeMust()},

			{Name: "ui/loader", Parent: "solidcoredata.org/base/loader", Configuration: (&api.ConfigureURL{MapTo: "/", Config: `{"Next": "example-1.solidcoredata.org/app/spa/system-menu"}`}).EncodeMust()},

			{Name: "spa/funny", Type: api.ResourceSPACode},

			{Name: "spa/system-menu", Parent: "solidcoredata.org/base/spa/system-menu", Include: []string{serviceName + "/spa/funny"}, Configuration: JSON(struct {
				Menu []struct{ Name, Location string }
			}{Menu: []struct{ Name, Location string }{{"File", "file"}, {"Edit", "edit"}}}),
			},

			{
				Name: "none",
				Include: []string{
					"example-1.solidcoredata.org/app/auth/login",
					"example-1.solidcoredata.org/app/ui/login",
					"example-1.solidcoredata.org/app/ui/favicon",
				},
			},
			{
				Name: "granted",
				Include: []string{
					"example-1.solidcoredata.org/app/auth/logout",
					"example-1.solidcoredata.org/app/ui/loader",
					"example-1.solidcoredata.org/app/ui/fetch-ui",
					"example-1.solidcoredata.org/app/ui/favicon",
					"example-1.solidcoredata.org/app/spa/system-menu",
					"example-1.solidcoredata.org/app/spa/funny",
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
						Resource:        "example-1.solidcoredata.org/app/none",
					},
					{
						LoginState:      api.LoginState_Granted,
						Prefix:          "/app/",
						ConsumeRedirect: true,
						Resource:        "example-1.solidcoredata.org/app/granted",
					},
				},

				AuthConfiguredResource: "example-1.solidcoredata.org/app/auth/endpoint",
				Host: []string{"example1.solidcoredata.local:8301"},
			},
		},
	}
	return c
}

var spaBody = map[string]string{
	serviceName + "/spa/funny": "console.log('dancing bears!');",
}

func (s *ServiceConfig) FetchUI(ctx context.Context, req *api.FetchUIRequest) (*api.FetchUIResponse, error) {
	resp := &api.FetchUIResponse{}
	for _, name := range req.List {
		body, found := spaBody[name]
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
