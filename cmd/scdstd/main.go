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
	return s, true
}
func (s *ServiceConfig) AuthServer() (api.AuthServer, bool) {
	return nil, false
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

func (s *ServiceConfig) createConfig() *api.ServiceBundle {
	c := &api.ServiceBundle{
		Name: "solidcoredata.org/base",
		Potential: []*api.PotentialResource{
			{Name: "loader", Type: api.PotentialResource_ResourceURL},
			{Name: "login", Type: api.PotentialResource_ResourceURL},
			{Name: "init.js", Type: api.PotentialResource_ResourceURL},
			{Name: "fetch-ui", Type: api.PotentialResource_ResourceURL},
			{Name: "favicon", Type: api.PotentialResource_ResourceURL},
		},
	}
	return c
}

func (s *ServiceConfig) Request(ctx context.Context, r *api.RequestReq) (*api.RequestResp, error) {
	resp := &api.RequestResp{}
	switch r.URL.Path {
	default:
		return nil, grpc.Errorf(codes.NotFound, "path %q not found", r.URL.Path)
	case "loader":
		resp.ContentType = "text/html"
		resp.Body = loginGrantedHTML
	case "login":
		resp.ContentType = "text/html"
		resp.Body = loginNoneHTML
	case "init.js":
		resp.ContentType = "	application/javascript"
		resp.Body = spaInitJS
	case "fetch-ui":
		cats := r.URL.Query.Values["category"].Value
		names := r.URL.Query.Values["name"].Value
		if len(cats) != len(names) {
			return nil, errors.New("fetch-ui: category and name have un-equal lengths")
		}
		ret := make([]*ReturnItem, 0, len(cats)+2)
		for i := range cats {
			c, n := cats[i], names[i]
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
	case "favicon":
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
