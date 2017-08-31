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

<h1>Granted to $APP</h1>

<h2>Hello</h2>
<div id=logout>logout</div>

<script>
var logoutButton = document.querySelector("#logout");

logoutButton.addEventListener("click", function(ev) {
	logout();
});
function logout() {
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
</script>
<script>
"use strict";
window.system = (function() {
	"use strict";
	var init = {};
	var sys = {init: init};
	
	init.errors = [];
	init.onError = function() {
		if(!console || !console.error) {
			return;
		}
		for(let i = 0; i < system.init.errors.length; i++) {
			let e = system.init.errors[i];
			console.error(e.Category, e.Name, e.On, e.Error, e.Input);
		}
	};
	
	function pushHTTPError(cnList, msg, done) {
		for(let i = 0; i < cnList.length; i++) {
			let cn = cnList[i];
			init.errors.push({
				Category: cn.Category,
				Name: cn.Name,
				On: "http",
				Error: msg,
				Input: "api/fetch-ui",
			});
		}
		if(typeof init.onError === "function") {
			init.onError();
		}
		if(typeof done === "function") {
			done(msg);
		}
	}
	function pushItemError(item, ex) {
		init.errors.push({
			Category: item.Category,
			Name: item.Name,
			On: item.Action,
			Error: ex,
			Input: item.Body,
		});
	}
	function processItem(item) /* ok boolean */ {
		switch(item.Action) {
		case "store":
			let o = null;
			try {
				o = JSON.parse(item.Body);
			} catch (ex) {
				pushItemError(item, ex);
				return false;
			}
			init.set(item.Category, item.Name, o);
			break;
		case "execute":
			let f = new Function(item.Body + "\n//# sourceURL=/system/" + item.Category + "/" + item.Name + ".js");
			try {
				f();
			} catch(ex) {
				pushItemError(item, ex);
				return false;
			}
			break;
		default:
			pushItemError(item, "unknown action: " + item.Action);
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
		
		if(hasError && typeof init.onError === "function") {
			init.onError();
		}
		
		if(typeof done === "function") {
			done(err);
		}
	};
	init.fetch = function(cnList, done) {
		let request = new XMLHttpRequest();
		request.responseType = "json";
		request.onerror = function(ev) {
			pushHTTPError(cnList, "unknown error, application may be down", done);
		}
		request.onload = function(ev) {
			let ok = (ev.target.status === 200);
			let resp = ev.target.response;
			if(!ok) {
				pushHTTPError(cnList, resp, done);
				return;
			}
			
			let need = [];
			for(let i = 0; i < resp.length; i++) {
				let item = resp[i];
				if(item.Require) {
					for(let ri = 0; ri < item.Require.length; ri++) {
						let r = item.Require[ri];
						if(!system.init.has(r.Category, r.Name)) {
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
		for(let i = 0; i < cnList.length; i++) {
			let cn = cnList[i];
			if(i != 0) {
				query += "&"
			}
			query += "category=" + encodeURIComponent(cn.Category) + "&name=" + encodeURIComponent(cn.Name)
		}
		
		request.open("POST", "api/fetch-ui?" + query, true);
		request.send();
	};

	init.set = function(categoryName, name, value) {
		if(!sys[categoryName]) {
			sys[categoryName] = {};
		}
		let category = sys[categoryName];
		category[name] = value;
	};
	init.has = function(categoryName, name) {
		if(!sys[categoryName]) {
			return false;
		}
		return !!sys[categoryName][name];
	};
	init.get = function(categoryName, name) {
		if(!sys[categoryName]) {
			return null;
		}
		var o = sys[categoryName][name];
		if(!o) {
			return null
		}
		return o;
	};
	return sys;
})();

window.Next = {{$.Next}};

system.init.fetch([{Category:"base", Name: window.Next}], function(err) {
	if(err != null) {
		console.error("failed to load application", err);
		return;
	}
});
</script>
`))

// Method of navigation:
//
// Tree of data
// Main motivation for any navigation: seamless developer refresh of UI component.
// JS API will take simple arguments and return a URL string that can be provided
// in a link, or set directly to location.href.

var widgetMenu = `"use strict";
var f = function(config) {
	// Create widget instance based on provided config.
	// Then do something with it.
	this.d = document.createElement("div");
	this.d.innerText = "hello world";
}

f.prototype.ElementRoot = function() {
	return this.d;
};
system.init.set("code","solidcoredata.org/system-menu", f);
`

// baseLoader loads configs and sets up the initial widgets.
// config struct schema:
//  struct {
//		Required []struct{Category, Name string}
//		Load     struct{Cagegory, Name string}
//		Config   any
//	}
//
// Use the history API to navigate to new pages, open a detail, etc...
// Store state in the query string (?a=1):
//  1. The deep link will be sent to the server. Useful to see stats.
//  2. No need to setup server-side resources if the URL was modified.
//  3. Let's the back button work.
// So an implementation would call "history.pushState(obj, '', '?encoded-state-here'')".
// For now lean to just keep navigation state in URL and full state just in memory.
//
// The client side will have an API that returns a string to use as a link.
// link.close(widget)
// link.replace(widget, config-name, initial-params)
// link.child(widget, config-name)
var baseLoader = `"use strict";
system.init.set("nav","close",function(instance) {
});
system.init.set("nav","replace",function(instance, configName, config) {
});
system.init.set("nav","child",function(instance, configName, config) {
});
system.init.set("nav","create",function(cn, done) {
	let c = system.init.get(cn.Category, cn.Name);
	if(!c) {
		done("missing cn for " + cn.Category + "." + cn.Name, null);
		return;
	}
	if(typeof c.Type !== "string") {
		done("missing Type field in " + cn.Category + "." + cn.Name, null);
		return;
	}
	let ctype = system.init.get("code", c.Type);
	if(!ctype) {
		done("missing type " + c.Type, null);
		return;
	}
	let w = new ctype(c);
	done(null, w);
});

// TODO: Setup a route listener.

system.init.set("data","fetch",function(config) {
	// Create widget instance based on provided config.
	// Then do something with it.
});

var bc = system.base.config;
system.init.fetch([bc.Next], function(err) {
	if(err != null) {
		console.error("failed to fetch next:", err);
		return;
	}
	system.nav.create(bc.Next, function(err, w) {
		if(err != null) {
			console.error("failed to create next widget:", err);
			return;
		}
		system.init.set("w", "root", w);
		let root = w.ElementRoot()
		document.body.innerHTML = "";
		document.body.append(root);
	});
});
`

// TODO: Explain what this initial shim does.
var spaInitJS = []byte(`"use strict";
window.system = (function() {
	"use strict";
	var init = {};
	var sys = {init: init};
	
	init.errors = [];
	init.onError = function() {
		if(!console || !console.error) {
			return;
		}
		for(let i = 0; i < system.init.errors.length; i++) {
			let e = system.init.errors[i];
			console.error(e.Category, e.Name, e.On, e.Error, e.Input);
		}
	};
	
	function pushHTTPError(cnList, msg, done) {
		for(let i = 0; i < cnList.length; i++) {
			let cn = cnList[i];
			init.errors.push({
				Category: cn.Category,
				Name: cn.Name,
				On: "http",
				Error: msg,
				Input: "api/fetch-ui",
			});
		}
		if(typeof init.onError === "function") {
			init.onError();
		}
		if(typeof done === "function") {
			done(msg);
		}
	}
	function pushItemError(item, ex) {
		init.errors.push({
			Category: item.Category,
			Name: item.Name,
			On: item.Action,
			Error: ex,
			Input: item.Body,
		});
	}
	function processItem(item) /* ok boolean */ {
		switch(item.Action) {
		case "store":
			let o = null;
			try {
				o = JSON.parse(item.Body);
			} catch (ex) {
				pushItemError(item, ex);
				return false;
			}
			init.set(item.Category, item.Name, o);
			break;
		case "execute":
			let f = new Function(item.Body + "\n//# sourceURL=/system/" + item.Category + "/" + item.Name + ".js");
			try {
				f();
			} catch(ex) {
				pushItemError(item, ex);
				return false;
			}
			break;
		default:
			pushItemError(item, "unknown action: " + item.Action);
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
		
		if(hasError && typeof init.onError === "function") {
			init.onError();
		}
		
		if(typeof done === "function") {
			done(err);
		}
	};
	init.fetch = function(cnList, done) {
		let request = new XMLHttpRequest();
		request.responseType = "json";
		request.onerror = function(ev) {
			pushHTTPError(cnList, "unknown error, application may be down", done);
		}
		request.onload = function(ev) {
			let ok = (ev.target.status === 200);
			let resp = ev.target.response;
			if(!ok) {
				pushHTTPError(cnList, resp, done);
				return;
			}
			
			let need = [];
			for(let i = 0; i < resp.length; i++) {
				let item = resp[i];
				if(item.Require) {
					for(let ri = 0; ri < item.Require.length; ri++) {
						let r = item.Require[ri];
						if(!system.init.has(r.Category, r.Name)) {
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
		for(let i = 0; i < cnList.length; i++) {
			let cn = cnList[i];
			if(i != 0) {
				query += "&"
			}
			query += "category=" + encodeURIComponent(cn.Category) + "&name=" + encodeURIComponent(cn.Name)
		}
		
		request.open("POST", "api/fetch-ui?" + query, true);
		request.send();
	};

	init.set = function(categoryName, name, value) {
		if(!sys[categoryName]) {
			sys[categoryName] = {};
		}
		let category = sys[categoryName];
		category[name] = value;
	};
	init.has = function(categoryName, name) {
		if(!sys[categoryName]) {
			return false;
		}
		return !!sys[categoryName][name];
	};
	init.get = function(categoryName, name) {
		if(!sys[categoryName]) {
			return null;
		}
		var o = sys[categoryName][name];
		if(!o) {
			return null
		}
		return o;
	};
	return sys;
})();

system.init.fetch([{Category:"base", Name:"setup"}], function(err) {
	if(err != null) {
		console.error("failed to load application", err);
		return;
	}
});
`)
