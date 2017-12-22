// Copyright 2017 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

local ref = import "ref.libsonnet";

local sn = "solidcoredata.org/auth";

{
	Name: sn,
	Resource: [
		{Name: "login", Type: ref.Resource.URL},
		{Name: "logout", Type: ref.Resource.URL},
		{Name: "endpoint", Type: ref.Resource.Auth},
	],
}
