// Copyright 2017 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

local ref = import "ref.libsonnet";

local sn = "example-1.solidcoredata.org/app";

{
	Name: sn,
	Application: [
		{
			AuthResource: "example-1.solidcoredata.org/app/auth/endpoint",
			Host: ["example1.solidcoredata.local:8301"],
			Login: [
				{LoginState: ref.LoginState.None, Prefix: "/login/", ConsumeRedirect: false, Resource: sn + "/none"},
				{LoginState: ref.LoginState.Granted, Prefix: "/app/", ConsumeRedirect: true, Resource: sn + "/granted"},
			],
		},
	],
	Resource: [
		{
			Name: "none", Include: [
				sn + "/auth/login",
				sn + "/ui/login",
				sn + "/ui/favicon",
			],
		},
		{
			Name: "granted", Include: [
				sn + "/auth/logout",
				sn + "/ui/loader",
				sn + "/ui/fetch-ui",
				sn + "/ui/favicon",
			],
		},
		{Name: "auth/login", Parent: "solidcoredata.org/auth/login", C: ref.C.URL{MapTo: "/api/login"}},
		{Name: "auth/logout", Parent: "solidcoredata.org/auth/logout", C: ref.C.URL{MapTo: "/api/logout"}},
		{Name: "auth/endpoint", Parent: "solidcoredata.org/auth/endpoint", C: ref.C.Auth{Area: "System", Environment: "DEV"}},
		{Name: "ui/login", Parent: "solidcoredata.org/base/login", C: ref.C.URL{MapTo: "/"}},
		{Name: "ui/fetch-ui", Parent: "solidcoredata.org/base/fetch-ui", C: ref.C.URL{MapTo: "/api/fetch-ui"}},
		{Name: "ui/favicon", Parent: "solidcoredata.org/base/favicon", C: ref.C.URL{MapTo: "/ui/favicon"}},
		{Name: "ui/loader", Parent: "solidcoredata.org/base/loader", C: ref.C.URL{MapTo: "/", Config: {Next: sn + "/spa/system-menu"}}, Include: [sn + "/spa/system-menu"]},
		{Name: "ctl/spa/funny", Type: ref.Resource.SPACode},
		{Name: "spa/funny", Parent: sn + "/ctl/spa/funny", C: ref.C.SPA{}},
		{Name: "spa/system-menu", Parent: "solidcoredata.org/base/spa/system-menu", Include: [sn+"/spa/funny"], C: ref.C.SPA{Menu: [{Name: "File", Location: "file"}, {Name: "Edit", Location: "edit"}]}},
	],
	Files: [
		{Name: sn + "/ctl/spa/funny", File: "code/funny.js"},
	],	
}
