// Copyright 2017 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/solidcoredata/scd/app/app_granted_api"
	"github.com/solidcoredata/scd/app/app_granted_ui"
	"github.com/solidcoredata/scd/app/app_none_api"
	"github.com/solidcoredata/scd/app/app_none_ui"
	"github.com/solidcoredata/scd/app/compose"
	"github.com/solidcoredata/scd/auth/auth_memory"
	"github.com/solidcoredata/scd/scdhandler"
)

func main() {
	authSession := &auth_memory.AuthenticateMemory{
		UserSetup: map[string]*auth_memory.MemoryUser{
			"user1": &auth_memory.MemoryUser{
				Identity:   "user1",
				GivenName:  "Myfirst",
				FamilyName: "Mylast",
				Password:   "password1",
			},
		},
	}
	// The authenticator needs to be per URL, NOT per RouteHandler or app.
	// On second thought, the authentication system should be tied to a single cookie name.
	// The reason I wanted per URL cookie names is to prevent programs running on different ports
	// from overwriting each other's cookies. One way to prevent this and to work around this
	// at least for now is to use a consistent hash scheme of the Hostname.
	//
	// The RouteHandler should not have Authenticator, it should be on the LoginStateRouter.
	// Additionally the LoginStateRouter should be able to pass along the authenticator to
	// all applications.
	//
	// I'd like to be able to support multiple QA environments on the same Host.
	// I can probably do this by configuring a special LoginGranted handler that
	// maps paths to nested-applications.
	//
	// In addition to the State handler, I'd also like to define various service
	// descriptions that can be implemented by a service. Then we define a
	// list of services that implement service descriptions.
	// We then move the Authenticator into a service description
	// and move the session manager into a service description. The
	// AuthenticateMemory service is used to satisfy both service descriptions.
	// The API component handlers for login and logout both require the service
	// description for session management.
	//
	// Service include: Authenticator, Session Manager, Database Querier,
	// Report Engine, Scheduler, Job Runner.
	//
	// The service registry could itself be a component, able to be swapped out.
	// The service registry could be a name mapper.
	//
	// Another component that should be first class is the notifier.
	//
	// Need a (central?) configuration manager for all components.
	// Need to allow a well known configuration be the basis
	// for other varients with extra properties.
	// Need to register a type of configuration, then delcare configuration
	// instances.
	// Configuration of the current running instance should be separate from
	// the new running instance. Need to be able to transition away from
	// one system to another gracefully with GOAWAY and similar messages.
	//
	// It would be good to have a standardized way to handle file resources:
	// templates, images, HTML, favicon, and any other client or server side
	// resources (anothing not code).
	//
	// Widget Registry. Component A loads widget instances 1,2,3 into the registry.
	// Component B loads widget instances 4,5,6 into the registry.
	// Data sources need to be able to be routed to the correct database.
	// A Widget needs to have a handle to both the database and the table.
	// When a system that contains widgets are loaded, we also record the
	// databases used by that system. So we actually have "service + database + table"
	// to record for data routing.
	// 1. Create Thing1 that has a database handle + schema.
	// 2. Create a registry for Thing1.
	// 3. Let Thing1 pipe a ui widget to the client.
	// 4. Let /api/data send data requests to Thing1's database.
	h := &scdhandler.RouteHandler{
		Router: scdhandler.HostRouter{
			"localhost:9786": scdhandler.LoginStateRouter{
				Authenticator: authSession,
				State: map[scdhandler.LoginState]scdhandler.AppHandler{
					scdhandler.LoginNone:    compose.NewHandler(authSession, "/login/", false, app_none_api.NewHandler(authSession), app_none_ui.NewHandler()),
					scdhandler.LoginGranted: compose.NewHandler(authSession, "/app/", true, app_granted_api.NewSessionHandler(authSession), app_granted_api.NewSPAHandler(), app_granted_ui.NewHandler()),
				},
			},
			"localhost:9787": scdhandler.LoginStateRouter{
				Authenticator: authSession,
				State: map[scdhandler.LoginState]scdhandler.AppHandler{
					scdhandler.LoginNone:    compose.NewHandler(authSession, "/login/", false, app_none_api.NewHandler(authSession), app_none_ui.NewHandler()),
					scdhandler.LoginGranted: compose.NewHandler(authSession, "/app/", true, app_granted_api.NewSessionHandler(authSession), app_granted_api.NewSPAHandler(), app_granted_ui.NewHandler()),
				},
			},
		},
		Loggerf: func(f string, a ...interface{}) {
			log.Printf(f, a...)
		},
	}

	err := h.Init(context.TODO())
	if err != nil {
		log.Fatalf("unable to init route handler: %v", err)
	}

	for serveon := range h.Router {
		go func(serveon string) {
			err := http.ListenAndServe(serveon, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				h.ServeHTTP(w, r)
			}))
			if err != nil {
				fmt.Fprintf(os.Stderr, "main: unable to listen and serve on %q: %v\n", serveon, err)
			}
		}(serveon)
	}
	fmt.Println("ready")
	select {}
}
