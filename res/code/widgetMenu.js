// Copyright 2017 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

"use strict";

// Method of navigation:
//
// Tree of data
// Main motivation for any navigation: seamless developer refresh of UI component.
// JS API will take simple arguments and return a URL string that can be provided
// in a link, or set directly to location.href.


function f(config) {
	// Create widget instance based on provided config.
	// Then do something with it.
	this.logoutButton = document.createElement("div");
	this.logoutButton.innerText = "logout";
	
	this.d = document.createElement("div");
	this.d.innerText = "hello world";
	
	this.d.append(this.logoutButton);
	
	this.config = config;
	
	this.logoutButton.addEventListener("click", (ev) => {
		this.logout();
	});
}

f.prototype.ElementRoot = function() {
	return this.d;
};
f.prototype.Open = function() {
	console.log("application started");
};

f.prototype.logout = function() {
	var req = new XMLHttpRequest();
	req.onerror = function(ev) {
		alert("Unknown error, application may be down.");
	}
	req.onload = function(ev) {
		if(ev.target.status === 200) {
			location.pathname = "/";
			return;
		}
		// User may be already logged out. This may result in the
		// logout endpoint from being available.
		if(ev.target.status === 404) {
			location.pathname = "/";
			return;
		}
		alert("Unknown error, application may be down.");
	}
	req.open("POST", "api/logout", true);
	req.responseType = "text";
	req.send();
}

system.init.set("solidcoredata.org/base/spa/system-menu", f);
