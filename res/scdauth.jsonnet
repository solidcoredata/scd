// Copyright 2017 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

local ResourceAuth = "solidcoredata.org/resource/auth";
local ResourceURL = "solidcoredata.org/resource/url";
local ResourceSPACode = "solidcoredata.org/resource/spa-code";
local ResourceQuery = "solidcoredata.org/resource/query";

local LoginStateNone = "None";
local LoginStateGranted = "Granted";

local URL = {Kind: "url", MapTo: ""};
local Auth = {Kind: "auth", Area: "System", Environment: "DEV"};
local SPA = {Kind: "spa"};

local sn = "solidcoredata.org/auth";

{
	Name: sn,
	Resource: [
		{Name: "login", Type: ResourceURL},
		{Name: "logout", Type: ResourceURL},
		{Name: "endpoint", Type: ResourceAuth},
	],
}
