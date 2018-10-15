// Copyright 2018 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"time"
)

// TODO: Register an HTTP handler or signal handler to print or show
// the current registry state, esp errors as they appear.
// Errors in configuration are pushed to the registry under the service lease
// that found them.
func NewMemoryRegistry() Registry { return nil }

var (
	_ Registry   = &MemoryRegistry{}
	_ RegistryTx = &MemoryTx{}
)

type MemoryRegistry struct {
}

func (mr *MemoryRegistry) NewLease(ctx context.Context, ttl time.Duration) (lease string, err error) {
	return "", nil
}
func (mr *MemoryRegistry) UpdateLease(ctx context.Context, lease string) error { return nil }
func (mr *MemoryRegistry) DeleteLease(ctx context.Context, lease string) error { return nil }

func (mr *MemoryRegistry) Begin(ctx context.Context) (RegistryTx, error) { return nil, nil }

type MemoryTx struct{}

func (mt *MemoryTx) Commit(ctx context.Context) error { return nil }
func (mt *MemoryTx) Abort()                           { return }

// WatchService blocks until ctx is canceled.
func (mt *MemoryTx) WatchService(ctx context.Context, svcs chan []Service) error { return nil }
func (mt *MemoryTx) WatchApplicationVersion(ctx context.Context, av chan []ApplicationVersion) error {
	return nil
}
func (mt *MemoryTx) WatchApplication(ctx context.Context, av chan []Application) error {
	return nil
}

// Lease required
func (mt *MemoryTx) SetService(lease int64, svc Service) error { return nil }

// Lease optional
func (mt *MemoryTx) SetApplicationVersion(lease int64, appver ApplicationVersion) error { return nil }

// Lease optional
func (mt *MemoryTx) SetApplication(lease int64, app Application) error { return nil }
