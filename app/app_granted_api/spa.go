// Copyright 2017 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app_granted_api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/solidcoredata/scd/scdhandler"
)

// There are a few types of items for a SPA:
// * Widget Type (Search List Detail, Split Layout, ...).
// * Widget Instance (Account Search List Detail, Master View, ...).
// * Database Instance (on a given query service).
// * Proceedure into custom code.
//
// How to go from Widget Type to Widget Instance? How to compose
// multiple Widget Instances? How do I setup the database schema
// and schema version transition logic?
//
// 1. Setting up all the configurations, schemas, and schema version trasition logic
//    needs to be done offline in a compile step.
//    a. Existing system descriptions need to be able to be pulled while offline
//       and stored locally to be compiled against. Pulling the schema to another
//       system also pulls historical versions so it can be incrementally compiled
//       against.
//    b. Systems that rely on another system need to be able to exert back-pressure
//       on a migration process so the two systems can be migrated:
//       A -> A', B -> B', A' -> A'', B' -> B''.
// 2. The underlying method to setup an SPA is:
//    a. /api/init.js returns a loader script.
//    b. The loader script looks for a special widget instance named  "root".
//    c. Each widget instance should contain the names of the required widgets,
//       and the configuration for the widget instance.
//    d. The loader then loads any each required widget, if not already loaded,
//       and then calls "CurrentWidgetType.Create(config)" which returns an
//       interface, including "RootElement() HTMLElement". This returned
//       element is attached as the parent wants. The base loader calls instance.Open()
//       attaches the root element to the document.body.
//    e. A varient of the above is the root loader could just call instance.Open()
//       and assume the root loader is actually headless. This would allow
//       making the first loader smaller and the second stage provide
//       various other artifacts, such as a data cache and call layer. Unsure.

// SPAItem needs to return a list of widget types, widget instances, and databases
// available.
type SPAItem interface {
	WidgetType() []WidgetType
	WidgetInstance() []WidgetInstance
	DatabaseQueryer() []DatabaseQueryer
}

type WidgetType struct{}
type WidgetInstance struct{}
type DatabaseQueryer struct{}

type SPAItemRegister interface {
	scdhandler.AppComponentHandler

	RegisterSPAItem(item SPAItem) error
}

// TODO: Also implement the widget/database registry here.
type spaHandler struct {
	// Database list
	// Widget type list linked to resources for file.
	// Widget configuration list.
}

var _ scdhandler.AppComponentHandler = &spaHandler{}

func NewSPAHandler() SPAItemRegister {
	return &spaHandler{}
}

func (h *spaHandler) Init(ctx context.Context) error {
	return nil
}

func (h *spaHandler) RegisterSPAItem(item SPAItem) error {
	return nil
}

func (h *spaHandler) ProvideMounts(ctx context.Context) ([]scdhandler.MountProvide, error) {
	return []scdhandler.MountProvide{
		{At: "/api/init.js"},
		{At: "/api/fetch-ui"},
		{At: "/api/fetch-data"},
	}, nil
}

// Return an array of items:
type ReturnItem struct {
	Action   string // store | execute
	Category string // Widget, Field, code, ...
	Name     string // Text, Numeric, SearchListDetail
	Require  []CN
	Body     string // JSON, Javascript
}

type CN struct{ Category, Name string }

var requestMap = map[CN][]*ReturnItem{
	CN{"code", "base"}: []*ReturnItem{
		// {Action: "store", Category: "base", Name: "loader", Body: `{"Next":{"Category":"config","Name":"example1.solidcoredata.org/system-menu"}}`},
		{Action: "execute", Category: "base", Name: "loader", Body: baseLoader, Require: []CN{{"config", "example1.solidcoredata.org/system-menu"}}},
	},
	CN{"config", "example1.solidcoredata.org/system-menu"}: []*ReturnItem{
		{Action: "store", Category: "base", Name: "config", Require: []CN{{"code", "solidcoredata.org/system-menu"}}, Body: `[{"Name": "File", "Location":"file"},{"Name": "Edit", "Location":"edit"}]`},
	},
	CN{"code", "solidcoredata.org/system-menu"}: []*ReturnItem{
		{Action: "execute", Category: "code", Name: "solidcoredata.org/system-menu", Body: widgetMenu},
	},
}

func (h *spaHandler) Request(ctx context.Context, r *scdhandler.Request) (*scdhandler.Response, error) {
	resp := &scdhandler.Response{}
	switch r.URL.Path {
	case "/api/init.js":
		resp.ContentType = "	application/javascript"
		resp.Body = spaInitJS
	case "/api/fetch-ui":
		cats := r.URL.Query()["category"]
		names := r.URL.Query()["name"]
		if len(cats) != len(names) {
			return nil, errors.New("api/fetch-ui: category and name have un-equal lengths")
		}
		ret := make([]*ReturnItem, 0, len(cats)+2)
		for i := range cats {
			c, n := cats[i], names[i]
			riList, found := requestMap[CN{c, n}]
			if !found {
				return nil, fmt.Errorf("api/fetch-ui: category=%q name=%q not found", c, n)
			}
			ret = append(ret, riList...)
		}
		var err error
		resp.ContentType = "application/json"
		resp.Body, err = json.Marshal(ret)
		return resp, err
	case "/api/fetch-data":
	}
	return resp, nil
}

// Method of navigation:
//
// Tree of data
// Main motivation for any navigation: seamless developer refresh of UI component.
// JS API will take simple arguments and return a URL string that can be provided
// in a link, or set directly to location.href.

var widgetMenu = `"use strict";
system.init.set("code","widget-menu",function(config) {
	// Create widget instance based on provided config.
	// Then do something with it.
});
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
var bc = system.base.config;

system.init.set("nav","close",function(instance) {
});
system.init.set("nav","replace",function(instance, configName, config) {
});
system.init.set("nav","child",function(instance, configName, config) {
});

system.init.set("nav","create",function(category, name, done) {
	var hasDone = function(ok) {
		var requested = system.init.get(category, name);
		if(!ok || !requested) {
			done(false);
			return;
		}
		
	};
	if(system.init.has(category,name) === false) {
		system.init.fetch([{Category:category, Name:name}], hasDone);
	}
});

system.init.set("data","fetch",function(config) {
	// Create widget instance based on provided config.
	// Then do something with it.
});
system.nav.create("base", "config", function(ok) {
	
});
`

// TODO: Explain what this initial shim does.
var spaInitJS = []byte(`"use strict";
window.system = (function() {
	"use strict";
	var init = {};
	var sys = {init: init};

	init.bootstrap = function() {
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
		init.fetch([{Category:"code", Name:"base"}], function(ok) {
			if(!ok) {
				alert("failed to load application");
				return;
			}
		});
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
			done(false);
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
	function finishItem(ok, resp, done) {
		let hasError = false;
		if(!ok) {
			hasError = true;
		}
		if(ok && resp) {
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
			done(ok);
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
				finishItem(true, resp, done);
				return;
			}
			init.fetch(need, function(ok) {
				finishItem(ok, resp, done);
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
system.init.bootstrap();
`)

/*
	// 1. Pull down JS.
	// 2. Execute the JS.
	// 3. The JS will register stuff in "System.StuffType.Name = function / constructor / object".

	// 1. Pull down JSON.
	// 2. Pull down any dependent JS (see above) that doesn't exist locally.
	// 3. Put the config in "App.ConfigType.Name = config".

	<Load Page>
	Request: "config.menu-system" -> (Config) + Require "code.menu"
	Request: "code.menu" -> (new Function())(), ensure "code.menu" now exists.
	Run: instance.menu-system = code.menu(config.menu-system) -> Attach to DOM
	<Click Menu Item>
	Request: "config.page1 -> (Config) + Require "code.search-list-detail"
	Request: "code.search-list-detail" -> (new Function())(), ensure "code.search-list-detail" now exists.
	Run: instance.page1 = code.search-list-detail(config.page1) -> Attach to DOM


	// system.data.table.GetFromSystem(...)
	// system.init.bootstrap()

	// Fetch "root" from "/api/fetch-ui".
	// Return the JSON Object.
	var result1 = {
		Name: "root",
		RequiredProgramParts: [
			"data-access",
			"menu-system",
		],
		Config: {
			Type: "menu-system",
			Style: "pull-down-menus",
			Menus: {
				"File": [
					{Display: "Open", Action: "my-open"},
					{Display: "Close", Action: "my-close"},
				],
			},
		},
	};
	var result2 = {
		Name: "root",
		RequiredProgramParts: [
			"base",
			"ui-base",
			"data-access",
			"ui-loader",
		],
		Config: {
			Type: "ui-loader",
			NextConfig: "root-container",
		},
	};
*/
