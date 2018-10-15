// Copyright 2018 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"time"
)

// Register service (lease): version, IP, consumers, resources
// Register application version configurations (lease optional).
// Register specific application to version (lease optional).

type NameVersion struct {
	Name    string // example-1.solidcoredata.org
	Version string // abc123024ddsa or 1.2.3
}

type Resource struct {
	Name    string // "proc"
	Kind    string // solidcoredata.org/resource/url
	Consume string // solidcoredata.org/resource/proc

	Parent  string   // Type of resource instance.
	Include []string // Bring in these included resources as well.
	Config  []byte
}

type Service struct {
	NameVersion

	Resources []Resource
}

type ApplicationVersion struct {
	NameVersion

	Uses []NameVersion // Matches the service name with the version to use.

	Resources []Resource
}

type Login struct {
	Percent         float64
	LoginState      string // solidcoredata.org/auth/none or solidcoredata.org/auth/granted
	Prefix          string // "login" or "app"
	ConsumeRedirect bool
	Resource        NameVersion // In the future allow specifying multiple Resources for A/B, blue-green, canary releases.
}

type Application struct {
	Authentication string
	Host           []string
	Login          []Login
}

type Registry interface {
	NewLease(ctx context.Context, ttl time.Duration) (lease string, err error)
	UpdateLease(ctx context.Context, lease string) error
	DeleteLease(ctx context.Context, lease string) error

	Begin(ctx context.Context) (RegistryTx, error)
}

/*
 - solidcoredata.org/consumer/solidcoredata.org/resource/auth
 - solidcoredata.org/consumer/example-1.solidcoredata.org/resource/proc

 - solidcoredata.org/resource/auth/<ip>
 - example-1.solidcoredata.org/resource/proc/<ip>

 - solidcoredata.org/c/menu-system/example-1.solidcoredata.org/i/funny
 - example-1.solidcoredata.org/i/funny-ctl ->
*/

type RegistryTx interface {
	Commit(ctx context.Context) error
	Abort()

	Watch(ctx context.Context, prefix string, changes chan []string) error
	Set(ctx context.Context, key, value string) error
}
