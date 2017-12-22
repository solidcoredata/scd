// Copyright 2017 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

local ref = import "ref.libsonnet";

local sn = "solidcoredata.org/base";
			
{
	Name: sn,
	Resource: [
		{Name: "loader", Type: ref.Resource.URL},
		{Name: "login", Type: ref.Resource.URL},
		{Name: "fetch-ui", Type: ref.Resource.URL, Consume: ref.Resource.SPACode},
		{Name: "favicon", Type: ref.Resource.URL},
		{Name: "spa/system-menu", Type: ref.Resource.SPACode},
	],
	Files: [
		{Name: sn + "/spa/system-menu", File: "code/widgetMenu.js"},
		{Name: sn + "/login/none", File: "code/login_none.html"},
		{Name: sn + "/login/granted", File: "code/login_granted.html"},
	],
}
