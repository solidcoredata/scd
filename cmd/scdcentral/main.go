// Copyright 2018 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package scdcentral attempts to use a centralized registry service (such
// as etcd) to simplify the how everything works.
// It also integrates versioning and supports load balancing.
//
// While it is packaged as a monolithic application, it could also be broken
// out into separate services. It is the explicit design intent to still allow
// PROD monolithic applications and PROD distributed applications, or to
// transition from one to the other as time goes on.
package main

import (
	"fmt"
	"os"
)

func main() {
	r1 := NewRouter()
	r2 := NewRouter()
	svc1 := NewServiceA()
	svc2 := NewServiceA()
	auth1 := NewAuthenticator()
	auth2 := NewAuthenticator()
	app1 := NewAppA()
	app2 := NewAppA()

	reg := NewMemoryRegistry()

	sr := NewServiceRegister(reg)

	err := sr.Register(
		r1, r2,
		svc1, svc2,
		auth1, auth2,
		app1, app2,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "service registration: %v\n", err)
		os.Exit(1)
	}
	select {}
}

type UserServiceA struct{}

func (s *UserServiceA) Config() {}

func NewRouter() Configurable        { return &UserServiceA{} }
func NewServiceA() Configurable      { return &UserServiceA{} }
func NewAuthenticator() Configurable { return &UserServiceA{} }
func NewAppA() Configurable          { return &UserServiceA{} }
