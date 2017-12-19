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
	"path/filepath"
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

// Configuration
type Configration interface {
	HTTPServer() (api.HTTPServer, bool)
	AuthServer() (api.AuthServer, bool)
	BundleUpdate(*api.ServiceBundle)
}

type RemoteService struct {
	Conn    *grpc.ClientConn
	Address string
}

type Service struct {
	r *routesService
}

func (s *Service) ResConn(version string) (map[string]ConnRes, bool) {
	s.r.setupLock.RLock()
	cr, found := s.r.setupVersion[version]
	s.r.setupLock.RUnlock()
	if !found {
		return nil, false
	}
	return cr.lookup, found
}

func (s *Service) SPA(name string) (*ResourceFile, bool) {
	s.r.spaLock.RLock()
	rf, found := s.r.spa[name]
	s.r.spaLock.RUnlock()
	return rf, found
}

func New() *Service {
	s := &Service{}
	return s
}

func (s *Service) Setup(ctx context.Context, sc Configration) {
	var bindAddress, routerAddress string
	flag.StringVar(&bindAddress, "bind", "localhost:0", "address and port to bind to")
	flag.StringVar(&routerAddress, "router", "", "optionally notify specified router")
	flag.Parse()

	server := grpc.NewServer()
	r, err := newRoutes(ctx, sc)
	if err != nil {
		onErrf(printMessage, "unable to create routes: %v", err)
	}
	s.r = r
	api.RegisterRoutesServer(server, s.r)

	if len(bindAddress) == 0 {
		onErr(printDefaults, `missing "bind" argument`)
	}

	if handler, is := sc.HTTPServer(); is {
		api.RegisterHTTPServer(server, handler)
	}
	if handler, is := sc.AuthServer(); is {
		api.RegisterAuthServer(server, handler)
	}
	api.RegisterSPAServer(server, s.r)

	l, err := net.Listen("tcp", bindAddress)
	if err != nil {
		onErrf(printMessage, `unable to listen on %q: %v`, bindAddress, err)
	}
	defer l.Close()
	if len(routerAddress) > 0 {
		serviceAddress, err := resolveServiceAddress(l.Addr(), routerAddress)
		fmt.Printf("address=%s\n", serviceAddress)
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

type ConnRes struct {
	Conn     *grpc.ClientConn
	Resource *api.Resource
}

type setup struct {
	lookup map[string]ConnRes

	// RPC connections specific to this version.
	conns map[string]*grpc.ClientConn
}

type routesService struct {
	sc Configration

	spaLock sync.RWMutex
	spa     map[string]*ResourceFile

	bundle chan *api.ServiceBundle
	config chan *api.ServiceConfig

	setupLock    sync.RWMutex
	setupVersion map[string]*setup
	conns        map[string]*grpc.ClientConn // All rpc connections created the server.
}

func newRoutes(ctx context.Context, sc Configration) (*routesService, error) {
	r := &routesService{
		sc: sc,

		bundle:       make(chan *api.ServiceBundle, 5),
		config:       make(chan *api.ServiceConfig, 5),
		setupVersion: make(map[string]*setup, 7),
		conns:        make(map[string]*grpc.ClientConn, 7),
	}

	execPath, err := os.Executable()
	if err != nil {
		return nil, err
	}
	_, execName := filepath.Split(execPath)
	execName = strings.TrimSuffix(execName, filepath.Ext(execName))
	scr, err := NewSCReader(filepath.Join("res", execName+".jsonnet"))
	if err != nil {
		return nil, err
	}
	scr.changes <- nil

	go r.run(ctx, scr)
	return r, nil
}

func (r *routesService) run(ctx context.Context, scr *SCReader) {
	for {
		select {
		case <-scr.changes:
			sb, spa, err := scr.open()
			if err != nil {
				fmt.Printf("failed to read file config: %v\n", err)
			} else {
				fmt.Printf("updated config for %q\n", sb.Name)
				r.bundle <- sb
				r.spaLock.Lock()
				r.spa = spa
				r.spaLock.Unlock()

				r.sc.BundleUpdate(sb)
			}
		case sc := <-r.config:
			switch sc.Action {
			case api.ServiceConfigAction_Remove:
				// Lookup all endpoint connections.
				// If any are unused then close them.
				r.setupLock.Lock()
				delete(r.setupVersion, sc.Version)
				closeConns := make([]string, 3)
				for ep, conn := range r.conns {
					found := false
					for _, ver := range r.setupVersion {
						if _, verFound := ver.conns[ep]; verFound {
							found = true
							break
						}
					}
					if !found {
						conn.Close()
						closeConns = append(closeConns, ep)
					}
				}
				for _, ep := range closeConns {
					delete(r.conns, ep)
				}
				r.setupLock.Unlock()

			case api.ServiceConfigAction_Add:
				// Lookup all endpoint connections.
				// If any are new then create them and add them.
				setup := &setup{
					lookup: make(map[string]ConnRes, len(sc.List)*10),
					conns:  make(map[string]*grpc.ClientConn, len(sc.List)),
				}

				r.setupLock.Lock()
				for _, sce := range sc.List {
					conn, found := r.conns[sce.Endpoint]
					if !found {
						cc, err := grpc.DialContext(ctx, sce.Endpoint, grpc.WithInsecure())
						if err != nil {
							fmt.Printf("Failed to dial rpc %q: %v\n", sce.Endpoint, err)
							continue
						}
						conn = cc
						r.conns[sce.Endpoint] = conn
					}
					setup.conns[sce.Endpoint] = conn
					for _, res := range sce.Resource {
						setup.lookup[res.Name] = ConnRes{
							Conn:     conn,
							Resource: res,
						}
					}
				}
				r.setupVersion[sc.Version] = setup
				r.setupLock.Unlock()
			}
		case <-ctx.Done():
			scr.Close()
			return
		}
	}
}

// Update connected services information, such as other service locations and SPA code and configs.
func (r *routesService) UpdateServiceBundle(arg0 *google_protobuf1.Empty, server api.Routes_UpdateServiceBundleServer) error {
	// TODO(kardianos): This will only work for a single service connected currently.
	// r.bundle is used directly, rather then through a manafold.
	for {
		select {
		case bundle, ok := <-r.bundle:
			if !ok {
				return nil
			}
			err := server.Send(bundle)
			if err != nil {
				// TODO(kardianos): determine what do do on error.
				continue
			}
		case <-server.Context().Done():
			return grpc.ErrServerStopped
		}
	}
}

// TODO(kardianos): include a version string in each resource so the client
// can be updated in real time without reloading everything. This string
// could be a content hash, file mod time, or a version number.

func (r *routesService) UpdateServiceConfig(ctx context.Context, config *api.ServiceConfig) (*google_protobuf1.Empty, error) {
	fmt.Printf("%v Service Config: version=%s\n", config.Action, config.Version)
	for _, ep := range config.List {
		fmt.Printf("\tService %q at %q\n", ep.Name, ep.Endpoint)
		for _, r := range ep.Resource {
			fmt.Printf("\t\tResource %q, parent %q, type=%s\n", r.Name, r.Parent, r.Type)
			if r.Type == api.ResourceSPACode {
				fmt.Printf("\t\t\tconfig=%s\n", string(r.Configuration))
			}
			for _, in := range r.Include {
				fmt.Printf("\t\t\tinclude=%s\n", in)
			}
		}
	}
	r.config <- config
	return &google_protobuf1.Empty{}, nil
}

func (r *routesService) FetchUI(ctx context.Context, req *api.FetchUIRequest) (*api.FetchUIResponse, error) {
	resp := &api.FetchUIResponse{}
	r.spaLock.RLock()
	defer r.spaLock.RUnlock()

	for _, name := range req.List {
		body, found := r.spa[name]
		if !found {
			continue
		}
		resp.List = append(resp.List, &api.FetchUIItem{Name: name, Body: body.Content})
	}
	return resp, nil
}
