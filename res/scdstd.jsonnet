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

local sn = "solidcoredata.org/base";
			
{
	Name: sn,
	Resource: [
		{Name: "loader", Type: ResourceURL},
		{Name: "login", Type: ResourceURL},
		{Name: "fetch-ui", Type: ResourceURL, Consume: ResourceSPACode},
		{Name: "favicon", Type: ResourceURL},
		{Name: "spa/system-menu", Type: ResourceSPACode},
	],
	Files: [
		{Name: sn + "/spa/system-menu", File: "code/widgetMenu.js"},
		{Name: sn + "/login/none", File: "code/login_none.html"},
		{Name: sn + "/login/granted", File: "code/login_granted.html"},
	],
}
