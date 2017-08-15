// Copyright 2017 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package service contains helpers for creating services.
package service

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"sync"

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

// ServiceConfiguration
type ServiceConfigration interface {
	ServiceBundle() chan *api.ServiceBundle
	RequestHanderServer() (api.RequestHanderServer, bool)
	AuthServer() (api.AuthServer, bool)
}

func Setup(ctx context.Context, sc ServiceConfigration) {
	var bindAddress, routerAddress string
	flag.StringVar(&bindAddress, "bind", "localhost:0", "address and port to bind to")
	flag.StringVar(&routerAddress, "router", "", "optionally notify specified router")
	flag.Parse()

	if len(bindAddress) == 0 {
		onErr(printDefaults, `missing "bind" argument`)
	}
	server := grpc.NewServer()
	api.RegisterRoutesServer(server, newRoutes(ctx, sc))

	if handler, is := sc.RequestHanderServer(); is {
		api.RegisterRequestHanderServer(server, handler)
	}
	if handler, is := sc.AuthServer(); is {
		api.RegisterAuthServer(server, handler)
	}

	l, err := net.Listen("tcp", bindAddress)
	if err != nil {
		onErrf(printMessage, `unable to listen on %q: %v`, bindAddress, err)
	}
	defer l.Close()
	if len(routerAddress) > 0 {
		go registerOnRouter(ctx, routerAddress, l.Addr().String())
	}

	err = server.Serve(l)
	if err != nil {
		onErrf(printMessage, `failed to serve on %q: %v`, bindAddress, err)
	}
}

func registerOnRouter(ctx context.Context, routerAddress, serviceAddress string) {
	// TODO(kardianos): This needs to send the service address every time the router is started.
	conn, err := grpc.DialContext(ctx, routerAddress, grpc.WithInsecure())
	if err != nil {
		onErrf(printMessage, "unable to connect to router: %v", err)
	}
	defer conn.Close()

	client := api.NewRouterConfigurationClient(conn)
	_, err = client.Notify(ctx, &api.NotifyReq{ServiceAddress: serviceAddress})
	if err != nil {
		onErrf(printMessage, "unable to update router: %v", err)
	}
}

type routesService struct {
	doneLock sync.RWMutex
	done     bool

	update *sync.Map
	latest chan chan *api.ServiceBundle
}

func newRoutes(ctx context.Context, sc ServiceConfigration) *routesService {
	r := &routesService{
		update: &sync.Map{},
		latest: make(chan chan *api.ServiceBundle, 5),
	}
	go func() {
		var latest *api.ServiceBundle

		for {
			select {
			case sb, ok := <-sc.ServiceBundle():
				if !ok {
					r.doneLock.Lock()
					r.done = true
					defer r.doneLock.Unlock()

					del := make([]chan *api.ServiceBundle, 0, 5)
					r.update.Range(func(key interface{}, value interface{}) bool {
						u := key.(chan *api.ServiceBundle)
						del = append(del, u)
						return true
					})
					for _, key := range del {
						r.update.Delete(key)
					}
					return
				}
				latest = sb
				r.update.Range(func(key interface{}, value interface{}) bool {
					u := key.(chan *api.ServiceBundle)
					u <- sb
					return true
				})
			case lchan := <-r.latest:
				if latest != nil {
					lchan <- latest
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return r
}

// Update connected services information, such as other service locations and SPA code and configs.
func (r *routesService) UpdateServiceBundle(arg0 *google_protobuf1.Empty, server api.Routes_UpdateServiceBundleServer) error {
	r.doneLock.RLock()
	if r.done {
		r.doneLock.RUnlock()
		return grpc.ErrServerStopped
	}

	update := make(chan *api.ServiceBundle, 5)
	r.update.Store(update, true)
	defer r.update.Delete(update)
	r.doneLock.RUnlock()

	r.latest <- update
	for {
		select {
		case bundle, ok := <-update:
			if !ok {
				return nil
			}
			err := server.Send(bundle)
			if err != nil {
				// TODO(kardianos): determine what do do on error.
				continue
			}
		case <-server.Context().Done():
			return nil
		}
	}
	return nil
}

func (r *routesService) UpdateServiceConfig(ctx context.Context, config *api.ServiceConfig) (*google_protobuf1.Empty, error) {
	return nil, nil
}
