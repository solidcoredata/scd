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

	"github.com/solidcoredata/scd/api"
	"github.com/solidcoredata/scd/handler"

	"google.golang.org/grpc"
)

func main() {
	h, err := getRouteHandler()
	if err != nil {
		fmt.Fprintf(os.Stderr, "main: unable to create router: %v\n", err)
		return
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

func getRouteHandler() (*handler.RouteHandler, error) {
	ctx := context.TODO()
	const (
		authServer    = "localhost:7001"
		exampleServer = "localhost:7002"
	)
	authConn, err := grpc.DialContext(ctx, authServer, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	authClient := api.NewAuthClient(authConn)

	exampleConn, err := grpc.DialContext(ctx, exampleServer, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	exampleClient := api.NewAppHandlerClient(exampleConn)

	h := &handler.RouteHandler{
		Router: handler.HostRouter{
			"localhost:9786": handler.LoginStateRouter{
				Authenticator: authClient,
				State: map[handler.LoginState]api.AppHandlerClient{
					//handler.LoginNone: compose.NewHandler(authSession, "/app/", true, granted.NewSessionHandler(authSession), spa.NewHandler(), granted.NewUIHandler()),
					handler.LoginNone: exampleClient,
				},
			},
			/*"localhost:9787": handler.LoginStateRouter{
				Authenticator: authSession,
				State: map[handler.LoginState]handler.AppHandler{
					handler.LoginNone:    compose.NewHandler(authSession, "/login/", false, none.NewSessionHandler(authSession), none.NewUIHandler()),
					handler.LoginGranted: compose.NewHandler(authSession, "/app/", true, granted.NewSessionHandler(authSession), spa.NewHandler(), granted.NewUIHandler()),
				},
			},*/
		},
		Loggerf: func(f string, a ...interface{}) {
			log.Printf(f, a...)
		},
	}

	err = h.Init(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("unable to init route handler: %v", err)
	}
	return h, nil
}
