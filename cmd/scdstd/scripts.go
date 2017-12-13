// Copyright 2017 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"html/template"
)

var loginNoneHTML = []byte(`<!DOCTYPE html>
<meta charset="UTF-8">
<link rel="icon" href="ui/favicon">

<title>Login to $APP</title>

<h1>Login to $APP</h1>

<table>
	<tr>
		<td>&nbsp;
		<td><div id=message></div>
	<tr>
		<td><label for=username>Username</label>
		<td><input id=username>
	<tr>
		<td><label for=password>Password</label>
		<td><input id=password type=password>
	<tr>
		<td>&nbsp;
		<td><button id=login>Login</button>

<script>
var usernameInput = document.querySelector("#username");
var passwordInput = document.querySelector("#password");
var loginButton = document.querySelector("#login");
var messageEl = document.querySelector("#message");

loginButton.addEventListener("click", function(ev) {
	message("");
	login();
});
passwordInput.addEventListener("keypress", function(ev) {
	message("");
	if(ev.keyCode !== 13) {
		return;
	}
	login();
});
usernameInput.addEventListener("keypress", function(ev) {
	message("");
	if(ev.keyCode !== 13) {
		return;
	}
	passwordInput.select();
});
usernameInput.select();

function message(text) {
	messageEl.textContent = text;
}
function login() {
	var req = new XMLHttpRequest();
	req.onerror = function(ev) {
		message("Unknown error, application may be down.");
	}
	req.onload = function(ev) {
		if(ev.target.status === 403) {
			message("Incorrect username or password.");
			passwordInput.select();
			return;
		}
		if(ev.target.status === 200) {
			location.reload();
			return;
		}
		message("Unknown error, application may be down.");
	}
	req.open("POST", "api/login", true);
	req.responseType = "text";
	var d = new FormData();
	d.set("u", usernameInput.value);
	d.set("p", passwordInput.value);
	req.send(d);
}
</script>
`)

var loginGrantedHTML = template.Must(template.New("").Parse(`<!DOCTYPE html>
<meta charset="UTF-8">
<link rel="icon" href="ui/favicon">

<title>Granted to $APP</title>

Loading...

<script>
"use strict";
window.system = (function() {
	"use strict";
	var init = {};
	var sys = {init: init};
	
	init.errors = [];
	init.onerror = function() {
		for(let i = 0; i < system.init.errors.length; i++) {
			let e = system.init.errors[i];
			let el = document.createElement("div");
			el.style.color = "red";
			el.innerText = [e.Name, e.On, e.Error, e.Input].join(", ");
			document.body.append(el);
			if(console && console.error) {
				console.error(e.Name, e.On, e.Error, e.Input);
			}
		}
	};
	
	function pushHTTPError(list, msg, done) {
		for(let i = 0; i < list.length; i++) {
			let item = list[i];
			init.errors.push({
				Name: item,
				On: "http",
				Error: msg,
				Input: "api/fetch-ui",
			});
		}
		if(typeof init.onerror === "function") {
			init.onerror();
		}
		if(typeof done === "function") {
			done(msg);
		}
	}
	function pushItemError(item, ex) {
		init.errors.push({
			Name: item.Name,
			On: item.Type,
			Error: ex,
			Input: item.Body,
		});
	}
	function processItem(item) /* ok boolean */ {
		if(item.Type.length > 0) { // Configuration.
			let o = null;
			try {
				o = JSON.parse(item.Body);
			} catch (ex) {
				pushItemError(item, ex);
				return false;
			}
			let t = init.get(item.Type);
			if(!t) {
				pushItemError(item, "unable to find type " + item.Type);
				return false;
			}
			try {
				let i = new t(o);
				init.set(item.Name, i);
				return true;
			} catch(ex) {
				pushItemError(item, ex);
				return false;
			}
		}
		let f = new Function(item.Body + "\n//# sourceURL=/system/" + item.Name + ".js");
		try {
			f();
		} catch(ex) {
			pushItemError(item, ex);
			return false;
		}
		return true;
	}
	function finishItem(err, resp, done) {
		let hasError = false;
		if(err != null) {
			hasError = true;
		}
		if(err == null && resp) {
			for(let i = 0; i < resp.length; i++) {
				let item = resp[i];
				if(processItem(item) === false) {
					hasError = true;
				}
			}
		} 
		
		if(hasError && typeof init.onerror === "function") {
			init.onerror();
		}
		
		if(typeof done === "function") {
			done(err);
		}
	};
	init.fetch = function(list, done) {
		let request = new XMLHttpRequest();
		request.responseType = "json";
		request.onerror = function(ev) {
			pushHTTPError(list, "unknown error, application may be down", done);
		}
		request.onload = function(ev) {
			let ok = (ev.target.status === 200);
			let resp = ev.target.response;
			if(!ok) {
				pushHTTPError(list, resp, done);
				return;
			}
			
			let need = [];
			for(let i = 0; i < resp.length; i++) {
				let item = resp[i];
				if(item.Type.length > 0) {
					need.push(item.Type);
				}
				if(item.Require) {
					for(let ri = 0; ri < item.Require.length; ri++) {
						let r = item.Require[ri];
						if(!system.init.has(r.Name)) {
							need.push(r);
						}
					}
				}
			}
			// TODO: de-duplicate needed items.
			if(need.length === 0) {
				finishItem(null, resp, done);
				return;
			}
			init.fetch(need, function(err) {
				finishItem(err, resp, done);
			});
		}
		
		let query = "";
		for(let i = 0; i < list.length; i++) {
			let item = list[i];
			if(i != 0) {
				query += "&"
			}
			query += "name=" + encodeURIComponent(item)
		}
		
		request.open("POST", "api/fetch-ui?" + query, true);
		request.send();
	};

	sys.app = {};
	init.set = function(name, value) {
		sys.app[name] = value;
	};
	init.has = function(name) {
		return !!sys.app[name];
	};
	init.get = function(name) {
		var o = sys.app[name];
		if(!o) {
			return null
		}
		return o;
	};
	return sys;
})();

(function() {
let next = {{$.Next}};
system.init.fetch([next], function(err) {
	if(err != null) {
		console.error("failed to load application", err);
		return;
	}
	let root = system.init.get(next);
	document.body.innerHTML = "";
	root.Open();
	document.body.append(root.ElementRoot());
});
})();
</script>
`))

// Method of navigation:
//
// Tree of data
// Main motivation for any navigation: seamless developer refresh of UI component.
// JS API will take simple arguments and return a URL string that can be provided
// in a link, or set directly to location.href.

var widgetMenu = `"use strict";
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
`
