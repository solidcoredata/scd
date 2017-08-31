// Copyright 2017 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package service contains helpers for creating services.
package service

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/solidcoredata/scd/api"

	google_protobuf1 "github.com/golang/protobuf/ptypes/empty"
	"github.com/google/trillian/client/backoff"
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
	HTTPServer() (api.HTTPServer, bool)
	AuthServer() (api.AuthServer, bool)
	SPAServer() (api.SPAServer, bool)
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

	if handler, is := sc.HTTPServer(); is {
		api.RegisterHTTPServer(server, handler)
	}
	if handler, is := sc.AuthServer(); is {
		api.RegisterAuthServer(server, handler)
	}
	if handler, is := sc.SPAServer(); is {
		api.RegisterSPAServer(server, handler)
	}

	l, err := net.Listen("tcp", bindAddress)
	if err != nil {
		onErrf(printMessage, `unable to listen on %q: %v`, bindAddress, err)
	}
	defer l.Close()
	if len(routerAddress) > 0 {
		serviceAddress, err := resolveServiceAddress(l.Addr(), routerAddress)
		if err != nil {
			// Error with an exit because this service won't register
			// to the router as expected. This is not expected to error
			// without some type of fatal configuration error.
			onErrf(printMessage, `unable to register with router %v`, err)
		}
		go registerOnRouter(ctx, routerAddress, serviceAddress)
	}

	err = server.Serve(l)
	if err != nil {
		onErrf(printMessage, `failed to serve on %q: %v`, bindAddress, err)
	}
}

// resolveServiceAddress attempts to return the routable IP:port address
// of the local system if the bind address is unspecified (bind to all ports).
//
// On any issue it falls back to the bind address.
func resolveServiceAddress(laddr net.Addr, routerAddress string) (string, error) {
	addr, is := laddr.(*net.TCPAddr)
	if !is {
		return laddr.String(), nil
	}
	if !addr.IP.IsUnspecified() {
		return addr.String(), nil
	}

	list, err := net.InterfaceAddrs()
	if err != nil {
		return "", fmt.Errorf("service: unable to iterate interfaces %v", err)
	}
	// Attempt to find IP address of router address in the case of multiple
	// local host interfaces.
	routerHost := strings.Split(routerAddress, ":")[0]
	routerIP := net.ParseIP(routerHost)
	if routerIP == nil {
		routerNet, _ := net.ResolveIPAddr("ip", routerHost)
		if routerNet != nil {
			routerIP = routerNet.IP
		}
	}
	// Still return a loopback interface if no other router IP can be found.
	// May want to split this out in the future, so local IP addresses
	// are detected and handled in another way.
	loopback := ""
	for _, a := range list {
		a, is := a.(*net.IPNet)
		if !is {
			continue
		}
		if routerIP != nil && !a.Contains(routerIP) {
			continue
		}
		if a.IP.IsLoopback() {
			loopback = fmt.Sprintf("%s:%d", a.IP, addr.Port)
			continue
		}
		return fmt.Sprintf("%s:%d", a.IP, addr.Port), nil
	}
	if len(loopback) > 0 {
		return loopback, nil
	}
	return "", fmt.Errorf("service: unable to find external service address for %q", laddr.String())
}

// registerOnRouter informs the router at routerAddress this service at serviceAddress.
func registerOnRouter(ctx context.Context, routerAddress, serviceAddress string) {
	conn, err := grpc.DialContext(ctx, routerAddress, grpc.WithInsecure(), grpc.WithBackoffConfig(grpc.BackoffConfig{MaxDelay: time.Second * 5}))
	if err != nil {
		onErrf(printMessage, "unable to connect to router: %v", err)
	}
	defer conn.Close()
	client := api.NewRouterConfigurationClient(conn)

	notReady := errors.New("remote not ready")
	bo := &backoff.Backoff{
		Min:    time.Millisecond * 400,
		Max:    time.Second * 5,
		Jitter: true,
		Factor: 1.2,
	}
	bo.Retry(ctx, func() error {
		_, err := client.Notify(ctx, &api.NotifyReq{ServiceAddress: serviceAddress})
		if err != nil {
			return err
		}
		bo.Reset()
		if conn.WaitForStateChange(ctx, grpc.Ready) {
			return notReady
		}
		return nil // Context was canceled, return.
	})
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
	return &google_protobuf1.Empty{}, nil
}
