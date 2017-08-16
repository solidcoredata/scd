// Copyright 2017 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// scdrouter accepts incomming connections and routes the requests to the
// correct service. It also unifies the services into a single application.
package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/solidcoredata/scd/api"

	google_protobuf1 "github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
)

const (
	printMessage  = 1
	printDefaults = 2
)

func onErr(t byte, msg string) {
	if len(msg) > 0 {
		fmt.Fprint(os.Stderr, msg)
		fmt.Fprintln(os.Stderr)
	}
	switch t {
	case printMessage:
		os.Exit(1)
	case printDefaults:
		flag.PrintDefaults()
		os.Exit(2)
	}
}
func onErrf(t byte, f string, v ...interface{}) {
	onErr(t, fmt.Sprintf(f, v...))
}

func main() {
	const bindRPC = ":9301"

	server := grpc.NewServer()
	s := &RouterServer{}
	api.RegisterRouterConfigurationServer(server, s)

	l, err := net.Listen("tcp", bindRPC)
	if err != nil {
		onErrf(printMessage, `unable to listen on %q: %v`, bindRPC, err)
	}
	defer l.Close()

	err = server.Serve(l)
	if err != nil {
		onErrf(printMessage, `failed to serve on %q: %v`, bindRPC, err)
	}
}

var _ api.RouterConfigurationServer = &RouterServer{}

type RouterServer struct{}

func (s *RouterServer) Notify(ctx context.Context, n *api.NotifyReq) (*google_protobuf1.Empty, error) {
	// For testing attempt to hit the service right back to ensure the service
	// address is good.
	ok := false
	conn, err := grpc.DialContext(ctx, n.ServiceAddress, grpc.WithInsecure())
	if err == nil {
		defer conn.Close()

		client := api.NewRoutesClient(conn)
		_, err = client.UpdateServiceConfig(ctx, &api.ServiceConfig{})
		if err == nil {
			ok = true
		}
	}
	fmt.Printf("service=%q ok=%t\n", n.ServiceAddress, ok)
	return &google_protobuf1.Empty{}, nil
}
func (s *RouterServer) Update(ctx context.Context, u *api.UpdateReq) (*api.UpdateResp, error) {
	fmt.Printf("Update: action=%q bind=%q bundle=%q host=%q", u.Action, u.Bind, u.Bundle, u.Host)
	return &api.UpdateResp{}, nil
}
