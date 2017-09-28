// Copyright 2017 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// scdstd hosts standard compoenents used in a solid core data application.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"sync"

	"github.com/solidcoredata/scd/api"
	"github.com/solidcoredata/scd/service"

	"golang.org/x/sync/errgroup"
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
	Name    string
	Type    string // Empty for Javascript, present for configs. Name of the type to use it in.
	Require []string
	Body    string // JSON, Javascript
}

func JSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
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

var spaBody = map[string]string{
	serviceName + "/spa/system-menu": widgetMenu,
}

func (s *ServiceConfig) createConfig() *api.ServiceBundle {
	c := &api.ServiceBundle{
		Name: serviceName,
		Resource: []*api.Resource{
			{Name: "loader", Type: api.ResourceURL},
			{Name: "login", Type: api.ResourceURL},
			{Name: "fetch-ui", Type: api.ResourceURL, Consume: api.ResourceSPACode},
			{Name: "favicon", Type: api.ResourceURL},

			{Name: "spa/system-menu", Type: api.ResourceSPACode},
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
	case serviceName + "/fetch-ui":
		fmt.Printf("fetch-ui: version=%q\n", r.Version)
		s.mu.RLock()
		setup, foundSetup := s.setupVersion[r.Version]
		s.mu.RUnlock()

		if !foundSetup {
			return nil, fmt.Errorf("unable to find version %s", r.Version)
		}

		remotes := map[*Connection][]*ReturnItem{}
		names := r.URL.Query.Values["name"].Value

		// Lookup names in service registry.
		// Hit all services in parallel and agg all results and respond to client.
		ret := make([]*ReturnItem, 0, len(names))
		for _, n := range names {
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
			ri := &ReturnItem{
				Name:    res.Resource.Name,
				Type:    res.Resource.Parent,
				Require: res.Resource.Include,
			}
			if len(res.Resource.Configuration) > 0 {
				ri.Body = string(res.Resource.Configuration)
			} else if body, found := spaBody[res.Resource.Name]; found {
				ri.Body = body
			} else {
				remotes[res.Conn] = append(remotes[res.Conn], ri)
			}
			ret = append(ret, ri)
		}

		if len(remotes) > 0 {
			g, ctx := errgroup.WithContext(ctx)
			for conn, riList := range remotes {
				riList := riList
				client := api.NewSPAClient(conn.Conn)
				g.Go(func() error {
					list := make([]string, len(riList))
					for i := 0; i < len(list); i++ {
						list[i] = riList[i].Name
					}
					resp, err := client.FetchUI(ctx, &api.FetchUIRequest{List: list})
					if err != nil {
						return err
					}
					for _, item := range resp.List {
						for _, ri := range riList {
							if item.Name != ri.Name {
								continue
							}
							ri.Body = item.Body
							break
						}
					}
					return nil
				})
			}
			if err := g.Wait(); err != nil {
				return nil, err
			}
		}
		for _, ri := range ret {
			// TODO(kardianos): list all missing names.
			if len(ri.Body) == 0 {
				return nil, fmt.Errorf("missing body for %q", ri.Name)
			}
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
	return s.config
}
