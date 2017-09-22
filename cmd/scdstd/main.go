// Copyright 2017 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// scdstd hosts standard compoenents used in a solid core data application.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"sync"

	"github.com/solidcoredata/scd/api"
	"github.com/solidcoredata/scd/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func main() {
	ctx := context.TODO()
	service.Setup(ctx, NewServiceConfig(ctx))
}

var _ service.ServiceConfigration = &ServiceConfig{}

func NewServiceConfig(ctx context.Context) *ServiceConfig {
	s := &ServiceConfig{
		bundle: make(chan *api.ServiceBundle, 5),
		config: make(chan *api.ServiceConfig, 5),

		setupVersion: make(map[string]*Setup, 5),
		conns:        make(map[string]*Connection, 5),
	}
	go s.run(ctx)

	s.staticConfig = s.createConfig()
	s.bundle <- s.staticConfig
	return s
}

type Connection struct {
	Endpoint string
	Conn     *grpc.ClientConn
}

type Resource struct {
	Conn     *Connection
	Resource *api.Resource
}

type Setup struct {
	lookup map[string]Resource

	// RPC connections speific to this version.
	conns map[string]*Connection
}

type ServiceConfig struct {
	bundle chan *api.ServiceBundle
	config chan *api.ServiceConfig

	staticConfig *api.ServiceBundle

	mu           sync.RWMutex
	setupVersion map[string]*Setup

	// All rpc connections created the server.
	conns map[string]*Connection
}

func (s *ServiceConfig) ServiceBundle() <-chan *api.ServiceBundle {
	return s.bundle
}
func (s *ServiceConfig) HTTPServer() (api.HTTPServer, bool) {
	return s, true
}
func (s *ServiceConfig) AuthServer() (api.AuthServer, bool) {
	return nil, false
}
func (s *ServiceConfig) SPAServer() (api.SPAServer, bool) {
	return s, true
}

func (s *ServiceConfig) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case sc := <-s.config:
			switch sc.Action {
			case api.ServiceConfigAction_Remove:
				// Lookup all endpoint connections.
				// If any are unused then close them.
				s.mu.Lock()
				delete(s.setupVersion, sc.Version)
				closeConns := make([]string, 3)
				for ep, conn := range s.conns {
					found := false
					for _, ver := range s.setupVersion {
						if _, verFound := ver.conns[ep]; verFound {
							found = true
							break
						}
					}
					if !found {
						conn.Conn.Close()
						closeConns = append(closeConns, ep)
					}
				}
				for _, ep := range closeConns {
					delete(s.conns, ep)
				}
				s.mu.Unlock()

			case api.ServiceConfigAction_Add:
				// Lookup all endpoint connections.
				// If any are new then create them and add them.
				setup := &Setup{
					lookup: make(map[string]Resource, len(sc.List)*10),
					conns:  make(map[string]*Connection, len(sc.List)),
				}

				s.mu.Lock()
				for _, sce := range sc.List {
					conn, found := s.conns[sce.Endpoint]
					if !found {
						cc, err := grpc.DialContext(ctx, sce.Endpoint, grpc.WithInsecure())
						if err != nil {
							fmt.Printf("Failed to dial rpc %q: %v\n", sce.Endpoint, err)
							continue
						}
						conn = &Connection{
							Endpoint: sce.Endpoint,
							Conn:     cc,
						}
						s.conns[sce.Endpoint] = conn
					}
					setup.conns[sce.Endpoint] = conn
					for _, res := range sce.Resource {
						setup.lookup[res.Name] = Resource{
							Conn:     conn,
							Resource: res,
						}
					}
				}
				s.setupVersion[sc.Version] = setup
				s.mu.Unlock()
			}
			fmt.Printf("got %v\n", sc.Version)
		}
	}
}

// Return an array of items:
type ReturnItem struct {
	Action   string // store | execute
	Category string // Widget, Field, code, ...
	Name     string // Text, Numeric, SearchListDetail
	Require  []CN
	Body     string // JSON, Javascript
}

func JSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}

type CN struct{ Category, Name string }

func init() {
	for key, value := range requestMap {
		for _, item := range value {
			if len(item.Category) == 0 {
				item.Category = key.Category
			}
			if len(item.Name) == 0 {
				item.Name = key.Name
			}
		}
	}
}

var requestMap = map[CN][]*ReturnItem{
	CN{"base", "setup"}: []*ReturnItem{
		{Action: "store", Category: "base", Name: "config", Body: JSON(struct{ Next CN }{CN{Category: "config", Name: "example1.solidcoredata.org/system-menu"}})},
		{Action: "execute", Category: "base", Name: "loader", Body: baseLoader},
	},
	CN{"config", "example1.solidcoredata.org/system-menu"}: []*ReturnItem{
		{Action: "store", Require: []CN{{"code", "solidcoredata.org/system-menu"}}, Body: JSON(struct {
			Type string
			Menu []struct{ Name, Location string }
		}{Type: "solidcoredata.org/system-menu", Menu: []struct{ Name, Location string }{{"File", "file"}, {"Edit", "edit"}}})},
	},
	CN{"code", "solidcoredata.org/system-menu"}: []*ReturnItem{
		{Action: "execute", Body: widgetMenu},
	},
}

// Attempting to fit the previously defined model of SPA code/config into
// the new distributed model of potential resource that get configured.
//
// Code (JS widgits) are act more like static resources.
// Code actually has two parts:
//   1. A widget instance that gets created from a configuration.
//   2. A widget code that gets sent to the client to allow instance to be created.
//
// Part (1) needs the configuration. Part (2) could/should just be a potential resource.
// Currently Code and Config are switched on in client, but sent down in the same way.
// I like how failure could be isolated and reported on as a module error, but I'd like
// to only use Typescript, Dart, GopherJS, or other compiled languages.
//
// It is also likely that code will have some amount of other resources required:
// images, templates. It would be nice to be able to serve these up
// from the code's own handle or a shared reference handle.
// Potential Resource Types could have certain attributes. SPACode can both be
// configured into an instance with a Configured Resource and downloaded directly.
// Static types will just be downloaded directly, these include images. PDFs,
// or other completly static assets.
//
// This will translate into a client API of:
// * GET /api/fetch-static?example1.solidcoredata.org/ref/image.png
// * GET /api/fetch-static?solidcoredata.org/base/spa/base
// * GET /api/fetch-static?solidcoredata.org/base/spa/menu-system
// * POST /api/fetch-ui?example1.solidcoredata.org/my-menu-system-config
//   - Return []struct{Name string, Type string, Require []string, Config string(any)}
// How do I know if the required resource is a code or configuration? Probably
// have two different Required field, one for config, one for code.

const serviceName = "solidcoredata.org/base"

func (s *ServiceConfig) createConfig() *api.ServiceBundle {
	c := &api.ServiceBundle{
		Name: serviceName,
		Resource: []*api.Resource{
			{Name: "loader", Type: api.ResourceURL},
			{Name: "login", Type: api.ResourceURL},
			{Name: "init.js", Type: api.ResourceURL},
			{Name: "fetch-ui", Type: api.ResourceURL, Consume: api.ResourceSPACode},
			{Name: "favicon", Type: api.ResourceURL},

			{Name: "spa/setup", Type: api.ResourceSPACode}, // Remove?
			{Name: "spa/system-menu", Type: api.ResourceSPACode},
			{Name: "app/system-menu", Parent: "solidcoredata.org/base/spa/system-menu", Configuration: []byte(`{"File":"Quit"}`)},
		},
	}
	return c
}

func (s *ServiceConfig) ServeHTTP(ctx context.Context, r *api.HTTPRequest) (*api.HTTPResponse, error) {
	resp := &api.HTTPResponse{}
	switch r.URL.Path {
	default:
		return nil, grpc.Errorf(codes.NotFound, "path %q not found", r.URL.Path)
	case serviceName + "/loader":
		resp.ContentType = "text/html"
		buf := &bytes.Buffer{}
		c := struct {
			Next string
		}{}
		fmt.Println("loader;Config", r.Config.Config)
		err := json.Unmarshal([]byte(r.Config.Config), &c)
		if err != nil {
			return nil, err
		}
		err = loginGrantedHTML.Execute(buf, c)
		if err != nil {
			return nil, err
		}
		resp.Body = buf.Bytes()
	case serviceName + "/login":
		resp.ContentType = "text/html"
		resp.Body = loginNoneHTML
	case serviceName + "/init.js":
		resp.ContentType = "	application/javascript"
		resp.Body = spaInitJS
	case serviceName + "/fetch-ui":
		fmt.Printf("fetch-ui: version=%q\n", r.Version)
		s.mu.RLock()
		setup, foundSetup := s.setupVersion[r.Version]
		s.mu.RUnlock()
		
		if !foundSetup {
			return nil, fmt.Errorf("unable to find version %s", r.Version)
		}
		
		
		cats := r.URL.Query.Values["category"].Value
		names := r.URL.Query.Values["name"].Value
		if len(cats) != len(names) {
			return nil, errors.New("fetch-ui: category and name have un-equal lengths")
		}
		// Lookup names in service registry.
		// Hit all services in parallel and agg all results and respond to client.
		ret := make([]*ReturnItem, 0, len(cats)+2)
		for i := range cats {
			c, n := cats[i], names[i]
			
			res, resFound := setup.lookup[n]
			if resFound {
				fmt.Printf("Resource found for %q with config %q\n", n, string(res.Resource.Configuration))
			} else {
				fmt.Printf("Resource not found %q\n", n)
				for name := range setup.lookup {
					fmt.Printf("\t%s\n", name)
				}
			}
			// Send config to client.
			// Client look for parent.
			// If parent and include not found fetch.
			// Check to ensure parent is found.
			// Set config-name = new parent-name(config).
			//
			// Set category=ResourceType on client.
			// This will partition off the namespace.
			
			riList, found := requestMap[CN{c, n}]
			if !found {
				return nil, fmt.Errorf("fetch-ui: category=%q name=%q not found", c, n)
			}
			ret = append(ret, riList...)
		}
		var err error
		resp.ContentType = "application/json"
		resp.Body, err = json.Marshal(ret)
		return resp, err
	case serviceName + "/favicon":
		var c color.Color
		switch r.Auth.LoginState {
		default:
			c = color.RGBA{B: 255, A: 255}
		case api.LoginState_Granted:
			c = color.RGBA{G: 255, A: 255}
		case api.LoginState_None:
			c = color.RGBA{R: 255, A: 255}
		}
		img := image.NewRGBA(image.Rect(0, 0, 192, 192))
		draw.Draw(img, img.Rect, image.NewUniform(c), image.ZP, draw.Over)
		buf := &bytes.Buffer{}
		png.Encode(buf, img)
		resp.ContentType = "image/png"
		resp.Body = buf.Bytes()
	}
	return resp, nil
}

func (s *ServiceConfig) FetchUI(ctx context.Context, req *api.FetchUIRequest) (*api.FetchUIResponse, error) {
	ret := make([]*api.FetchUIItem, 0, len(req.List)+2)
	for _, item := range req.List {
		riList, found := requestMap[CN{item.Category, item.Name}]
		if !found {
			return nil, fmt.Errorf("category=%q name=%q not found", item.Category, item.Name)
		}
		for _, ri := range riList {
			action := api.FetchUIAction_ActionMissing
			switch ri.Action {
			case "execute":
				action = api.FetchUIAction_ActionExecute
			case "store":
				action = api.FetchUIAction_ActionStore
			}
			require := make([]*api.FetchUICN, len(ri.Require))
			for i, cn := range ri.Require {
				require[i] = &api.FetchUICN{
					Category: cn.Category,
					Name:     cn.Name,
				}
			}
			ret = append(ret, &api.FetchUIItem{
				Action:   action,
				Category: ri.Category,
				Name:     ri.Name,
				Require:  require,
				Body:     ri.Body,
			})
		}
	}
	return &api.FetchUIResponse{List: ret}, nil
}

func (s *ServiceConfig) Config() chan<- *api.ServiceConfig {
	return s.config
}
