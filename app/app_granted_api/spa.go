package app_granted_api

import (
	"context"
	"encoding/json"

	"github.com/solidcoredata/scdhttp/scdhandler"
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

func (h *spaHandler) Request(ctx context.Context, r *scdhandler.Request) (*scdhandler.Response, error) {
	resp := &scdhandler.Response{}
	switch r.URL.Path {
	case "/api/init.js":
		resp.ContentType = "	application/javascript"
		resp.Body = spaInitJS
	case "/api/fetch-ui":
		names := r.URL.Query()["name"]
		for _, name := range names {
			_ = name
		}
		// Return an array of items:
		type ReturnItem struct {
			Action   string // store | execute
			Category string // Widget, Field, code, ...
			Name     string // Text, Numeric, SearchListDetail
			Body     string // JSON, Javascript
		}
		ret := []ReturnItem{
			{Action: "execute", Category: "widget", Name: "menu", Body: widgetMenu},
			{Action: "store", Category: "base", Name: "config", Body: `{"Next": "system-menu"}`},
			{Action: "execute", Category: "base", Name: "loader", Body: baseLoader},
		}
		_ = ret
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
return function(config) {
	// Create widget instance based on provided config.
	// Then do something with it.
}
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

`

// TODO: Explain what this initial shim does.
var spaInitJS = []byte(`"use strict";
window.system = (function() {
	"use strict";
	var init = {};
	var sys = {init: init};
	
	init.bootstrap = function() {
		init.fetch("code", "base", function(ok) {
			if(!ok) {
				alert("failed to load application");
				return;
			}
		});
	};
	init.fetch = function(category, name, done) {
		var req = new XMLHttpRequest();
		req.responseType = "json";
		req.onerror = function(ev) {
			if(typeof done === "function") {
				done(false, {error: "unknown error, application may be down"});
			}
		}
		req.onload = function(ev) {
			var ok = (ev.target.status === 200);
			var resp = ev.target.response;
			if(ok) {
				for(let i = 0; i < resp.length; i++) {
					let item = resp[i];
					
					switch(item.Action) {
					case "store":
						init.set(item.Category, item.Name, item.Body);
						break;
					case "execute":
						var f = new Function(item.Body + "\n//# sourceURL=/system/" + item.Category + "/" + item.Name + ".js");
						init.set(item.Category, item.Name, f());
						break;
					default:
						throw "unknown action: " + item.Action;
					}
				}
			}
			
			if(typeof done === "function") {
				done(ok);
			}
		}
		req.open("POST", "api/fetch-ui?name=" + encodeURIComponent(category + "." + name), true);
		req.send();
	};
	init.set = function(systypeName, name, value) {
		if(!sys[systypeName]) {
			sys[systypeName] = {};
		}
		let systype = sys[systypeName];
		systype[name] = value;
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
