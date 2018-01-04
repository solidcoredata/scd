// Copyright 2017 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

{
	Resource: {
		Auth: "solidcoredata.org/resource/auth",
		URL: "solidcoredata.org/resource/url",
		SPACode: "solidcoredata.org/resource/spa-code",
		Query: "solidcoredata.org/resource/query",

		Proc: "example-1.solidcoredata.org/proc",
	},
	
	LoginState: {
		None: "None",
		Granted: "Granted",
	},
	
	C: {
		URL: { Kind: "url", MapTo: "" },
		Auth: { Kind: "auth", Area: "System", Environment: "DEV" },
		SPA: { Kind: "spa" },
	},
}
