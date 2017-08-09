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

	"github.com/solidcoredata/scd/app/compose"
	"github.com/solidcoredata/scd/app/granted"
	"github.com/solidcoredata/scd/app/none"
	"github.com/solidcoredata/scd/app/spa"
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

	h := &scdhandler.RouteHandler{
		Router: scdhandler.HostRouter{
			"localhost:9786": scdhandler.LoginStateRouter{
				Authenticator: authSession,
				State: map[scdhandler.LoginState]scdhandler.AppHandler{
					scdhandler.LoginNone: compose.NewHandler(authSession, "/app/", true, granted.NewSessionHandler(authSession), spa.NewHandler(), granted.NewUIHandler()),
				},
			},
			"localhost:9787": scdhandler.LoginStateRouter{
				Authenticator: authSession,
				State: map[scdhandler.LoginState]scdhandler.AppHandler{
					scdhandler.LoginNone:    compose.NewHandler(authSession, "/login/", false, none.NewSessionHandler(authSession), none.NewUIHandler()),
					scdhandler.LoginGranted: compose.NewHandler(authSession, "/app/", true, granted.NewSessionHandler(authSession), spa.NewHandler(), granted.NewUIHandler()),
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
			err := http.ListenAndServe(serveon, h)
			if err != nil {
				fmt.Fprintf(os.Stderr, "main: unable to listen and serve on %q: %v\n", serveon, err)
			}
		}(serveon)
	}
	fmt.Println("ready")
	select {}
}
