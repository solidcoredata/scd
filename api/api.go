// Copyright 2017 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package api provides common interfaces and headers
package api

import (
	"context"

	proto "github.com/golang/protobuf/proto"
)

//go:generate protoc --go_out=plugins=grpc:../api -I ../proto/ ../proto/auth.proto ../proto/request.proto ../proto/router.proto ../proto/spa.proto
//go:generate go build -i github.com/solidcoredata/scd/api
//go:generate go build github.com/solidcoredata/scd/cmd/...

type requestAuthKey struct{}

// AuthNewContext returns a child context with the RequestAuth as a value.
func AuthNewContext(ctx context.Context, rs *RequestAuthResp) context.Context {
	return context.WithValue(ctx, requestAuthKey{}, rs)
}

// AuthFromContext returns the RequestAuth found in the context values if found.
func AuthFromContext(ctx context.Context) (rs *RequestAuthResp, found bool) {
	rs, found = ctx.Value(requestAuthKey{}).(*RequestAuthResp)
	return rs, found
}

var (
	fetchUIActionMissingBytes = []byte("missing")
	fetchUIActionExecuteBytes = []byte("execute")
	fetchUIActionStoreBytes   = []byte("store")
)

type ResourceType = string

const (
	ResourceNone    ResourceType = ""
	ResourceAuth    ResourceType = "solidcoredata.org/resource/auth"
	ResourceURL     ResourceType = "solidcoredata.org/resource/url"
	ResourceSPACode ResourceType = "solidcoredata.org/resource/spa-code"
	ResourceQuery   ResourceType = "solidcoredata.org/resource/query"
)

func (c *ConfigureURL) Encode() ([]byte, error) {
	return proto.Marshal(c)
}
func (c *ConfigureURL) EncodeMust() []byte {
	b, err := proto.Marshal(c)
	if err != nil {
		panic(err)
	}
	return b
}
func (c *ConfigureURL) Decode(b []byte) error {
	err := proto.Unmarshal(b, c)
	return err
}

func (c *ConfigureAuth) Encode() ([]byte, error) {
	return proto.Marshal(c)
}
func (c *ConfigureAuth) EncodeMust() []byte {
	b, err := proto.Marshal(c)
	if err != nil {
		panic(err)
	}
	return b
}
func (c *ConfigureAuth) Decode(b []byte) error {
	err := proto.Unmarshal(b, c)
	return err
}
