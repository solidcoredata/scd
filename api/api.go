// Copyright 2017 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package api provides common interfaces and headers
package api

import (
	"context"
)

//go:generate protoc --go_out=plugins=grpc:../api -I ../proto/ ../proto/auth.proto ../proto/request.proto ../proto/router.proto

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
