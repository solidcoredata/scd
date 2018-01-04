// Copyright 2018 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

// TODO: Register an HTTP handler or signal handler to print or show
// the current registry state, esp errors as they appear.
// Errors in configuration are pushed to the registry under the service lease
// that found them.
func NewMemoryRegistry() Registry { return nil }
