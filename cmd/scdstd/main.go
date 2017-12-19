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
	"html/template"
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
	s := service.New()
	sc := NewServiceConfig(ctx, s)
	s.Setup(ctx, sc)
}

var _ service.Configration = &ServiceConfig{}

func NewServiceConfig(ctx context.Context, s *service.Service) *ServiceConfig {
	sc := &ServiceConfig{
		service:       s,
		loginTemplate: template.New(""),
	}
	return sc
}

type ServiceConfig struct {
	service *service.Service

	mu            sync.RWMutex
	loginTemplate *template.Template
}

func (s *ServiceConfig) HTTPServer() (api.HTTPServer, bool) {
	return s, true
}
func (s *ServiceConfig) AuthServer() (api.AuthServer, bool) {
	return nil, false
}
func (s *ServiceConfig) BundleUpdate(sb *api.ServiceBundle) {
	s.mu.Lock()
	defer s.mu.Unlock()

	grantedName := sb.Name + "/login/granted"
	if res, found := s.service.SPA(grantedName); found {
		var err error
		s.loginTemplate, err = s.loginTemplate.Parse(res.Content)
		if err != nil {
			fmt.Printf("Unable to parse template %q: %v\n", grantedName, err)
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

func (s *ServiceConfig) ServeHTTP(ctx context.Context, r *api.HTTPRequest) (*api.HTTPResponse, error) {
	const serviceName = "solidcoredata.org/base"

	resp := &api.HTTPResponse{}
	switch r.URL.Path {
	default:
		return nil, grpc.Errorf(codes.NotFound, "path %q not found", r.URL.Path)
	case serviceName + "/loader":
		buf := &bytes.Buffer{}
		c := struct {
			Next string
		}{}
		fmt.Println("loader;Config", r.Config.Config)
		err := json.Unmarshal([]byte(r.Config.Config), &c)
		if err != nil {
			return nil, err
		}
		s.mu.RLock()
		err = s.loginTemplate.Execute(buf, c)
		s.mu.RUnlock()

		if err != nil {
			return nil, err
		}
		resp.ContentType = "text/html"
		resp.Body = buf.Bytes()
	case serviceName + "/login":
		resName := serviceName + "/login/none"
		body, found := s.service.SPA(resName)
		if !found {
			return nil, grpc.Errorf(codes.NotFound, "path %q not found", resName)
		}
		resp.ContentType = "text/html"
		resp.Body = []byte(body.Content)
	case serviceName + "/fetch-ui":
		fmt.Printf("fetch-ui: version=%q\n", r.Version)
		setup, foundSetup := s.service.ResConn(r.Version)

		if !foundSetup {
			return nil, fmt.Errorf("unable to find version %s", r.Version)
		}

		remotes := map[*grpc.ClientConn][]*ReturnItem{}
		names := r.URL.Query.Values["name"].Value

		// Lookup names in service registry.
		// Hit all services in parallel and agg all results and respond to client.
		ret := make([]*ReturnItem, 0, len(names))
		for _, n := range names {
			res, resFound := setup[n]
			if resFound {
				fmt.Printf("Resource found for %q with config %q\n", n, string(res.Resource.Configuration))
			} else {
				fmt.Printf("Resource not found %q\n", n)
				for name := range setup {
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
			} else if body, found := s.service.SPA(res.Resource.Name); found {
				ri.Body = body.Content
			} else {
				remotes[res.Conn] = append(remotes[res.Conn], ri)
			}
			ret = append(ret, ri)
		}

		if len(remotes) > 0 {
			g, ctx := errgroup.WithContext(ctx)
			for conn, riList := range remotes {
				riList := riList
				client := api.NewSPAClient(conn)
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
