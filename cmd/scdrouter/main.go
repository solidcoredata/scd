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
	"log"
	"net"
	"net/http"
	"os"
	"path"
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

func main() {
	const (
		bindRPC  = ":9301"
		bindHTTP = ":8301"
	)
	ctx := context.TODO()

	s := NewRouterServer(ctx)

	go s.startRPC(ctx, bindRPC)
	go s.startHTTP(ctx, bindHTTP)
	select {}
}
func (s *RouterServer) startRPC(ctx context.Context, bindRPC string) {
	server := grpc.NewServer()
	api.RegisterRouterConfigurationServer(server, s)

	l, err := net.Listen("tcp", bindRPC)
	if err != nil {
		onErrf(printMessage, `unable to listen on %q: %v`, bindRPC, err)
	}
	defer l.Close()

	err = server.Serve(l)
	if err != nil {
		onErrf(printMessage, `failed to serve RPC on %q: %v`, bindRPC, err)
	}
}
func (s *RouterServer) startHTTP(ctx context.Context, bindHTTP string) {
	err := http.ListenAndServe(bindHTTP, s)
	if err != nil {
		onErrf(printMessage, `failed to listen and serve HTTP on %q: %v`, bindHTTP, err)
	}
}

func (s *RouterServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.lk.RLock()
	app, found := s.router.App[r.Host]
	s.lk.RUnlock()

	if !found {
		http.Error(w, "host record not found", http.StatusNotFound)
		return
	}

	_ = app.AuthConfig
	// TODO(kardianos): attach token and auth config to request.
	authResp, err := app.Auth.RequestAuth(r.Context(), &api.RequestAuthReq{
		Token: "X",
	})
	if err != nil {
		http.Error(w, "auth: "+err.Error(), http.StatusInternalServerError)
		return
	}

	lb, found := app.LoginBundle[authResp.LoginState]
	if !found {
		http.Error(w, "unconfigured login state: "+authResp.LoginState.String(), http.StatusInternalServerError)
		return
	}

	cr, found := lb.URLRouter[r.URL.Path]
	if !found {
		http.Error(w, "path not found", http.StatusNotFound)
		return
	}

	_ = cr
	// TODO(kardianos): In setup setup client router. Then here translate the HTTP
	// request and call API to server.
}

var _ api.RouterConfigurationServer = &RouterServer{}

type serviceDef struct {
	serviceAddress string
	conn           *grpc.ClientConn
	sb             *api.ServiceBundle
}

type RouterServer struct {
	ctx context.Context

	lk       sync.RWMutex
	services map[string]serviceDef
	router   *RouterRun
	// complete setup
}

func NewRouterServer(ctx context.Context) *RouterServer {
	s := &RouterServer{
		ctx:      ctx,
		services: make(map[string]serviceDef, 30),
	}
	return s
}

func (s *RouterServer) updateServiceAddress(serviceAddress string) {
	conn, err := grpc.DialContext(s.ctx, serviceAddress, grpc.WithInsecure())
	if err != nil {
		log.Printf("unable to dial service %q %v", serviceAddress, err)
		return
	}
	defer conn.Close()

	client := api.NewRoutesClient(conn)
	sbc, err := client.UpdateServiceBundle(s.ctx, &google_protobuf1.Empty{})

	sb, err := sbc.Recv()
	if err != nil {
		log.Printf("failed to receive from %q %v", serviceAddress, err)
		return
	}
	defer s.removeService(sb.Name)
	s.updateService(serviceAddress, conn, sb)

	go func() {
		for {
			sb, err := sbc.Recv()
			if err != nil {
				log.Printf("failed to receive in loop %q %v", serviceAddress, err)
				return
			}
			s.updateService(serviceAddress, conn, sb)
		}
	}()
	conn.WaitForStateChange(s.ctx, grpc.Ready)
}

func (s *RouterServer) removeService(serviceName string) {
	fmt.Printf("remove %q\n", serviceName)
	s.lk.Lock()
	defer s.lk.Unlock()

	delete(s.services, serviceName)

	s.updateCompleteLocked()
}

func (s *RouterServer) updateService(serviceAddress string, conn *grpc.ClientConn, sb *api.ServiceBundle) {
	fmt.Printf("update %q\n", sb.Name)
	s.lk.Lock()
	defer s.lk.Unlock()

	s.services[sb.Name] = serviceDef{
		serviceAddress: serviceAddress,
		conn:           conn,
		sb:             sb,
	}

	s.updateCompleteLocked()
}

func NewRouterRun() *RouterRun {
	rr := &RouterRun{
		Potential:  make(map[string]*PR),
		Configured: make(map[string]*CR),
		Bundle:     make(map[string]*Bundle),
		App:        make(map[string]*App),
	}
	return rr
}

type RouterRun struct {
	Potential  map[string]*PR
	Configured map[string]*CR
	Bundle     map[string]*Bundle
	App        map[string]*App

	Errors []string
}
type PR struct {
	Name          string
	Type          api.PotentialResource_ResourceType
	ServiceBundle *api.ServiceBundle
	Service       *serviceDef
}
type CR struct {
	Name          string
	PRName        string
	PR            *PR
	CAuth         *api.ConfigureAuth
	CSPA          *api.ConfigureSPACode
	CURL          *api.ConfigureURL
	CQuery        *api.ConfigureQuery
	ServiceBundle *api.ServiceBundle
	Service       *serviceDef
}
type Bundle struct {
	Name        string
	IncludeName []string
	Include     []*CR
}
type LoginBundle struct {
	LoginState      api.LoginState
	Prefix          string
	ConsumeRedirect bool
	BundleName      string
	Bundle          *Bundle

	// TODO(kardianos): this should use some type of prefix tree to handle
	// folder paths.
	URLRouter map[string]*CR
}
type App struct {
	Host        []string
	AuthName    string
	LoginBundle map[api.LoginState]*LoginBundle
	Auth        api.AuthClient
	AuthConfig  *api.ConfigureAuth
}

func (rr *RouterRun) AddError(f string, v ...interface{}) {
	rr.Errors = append(rr.Errors, fmt.Sprintf(f, v...))
}

func (rr *RouterRun) resolveNames() {
	for _, c := range rr.Configured {
		pr, found := rr.Potential[c.PRName]
		if !found {
			rr.AddError("missing potential resource %q required by %q", c.PRName, c.Name)
			continue
		}
		c.PR = pr
	}
	for _, b := range rr.Bundle {
		for _, iname := range b.IncludeName {
			cr, found := rr.Configured[iname]
			if !found {
				rr.AddError("missing configured resource %q required by %q", iname, b.Name)
				continue
			}
			b.Include = append(b.Include, cr)
		}
	}
	for _, a := range rr.App {
		if len(a.AuthName) == 0 {
			rr.AddError("app on %q missing authentication", a.Host)
		} else {
			if cr, found := rr.Configured[a.AuthName]; found && cr.Service != nil {
				a.Auth = api.NewAuthClient(cr.PR.Service.conn)
				a.AuthConfig = cr.CAuth
			} else {
				rr.AddError("app on %q unable to resolve authenticator %q", a.AuthName)
			}
		}
		for _, lb := range a.LoginBundle {
			b, found := rr.Bundle[lb.BundleName]
			if !found {
				rr.AddError("missing bundle %q for app on %q for state %v", lb.BundleName, a.Host, lb.LoginState)
				continue
			}
			lb.Bundle = b
			// TODO(kardianos): process URL resources and create a per LoginBundle URL tree.
			// TODO(kardianos): process SPA resources and create a per LoginBundle SPA lookup.
			for _, cr := range b.Include {
				switch {
				case cr.CURL != nil:
					lb.URLRouter[cr.CURL.MapTo] = cr
				case cr.CSPA != nil:
				}
			}
		}
	}
}

func (s *RouterServer) updateCompleteLocked() {
	// TODO(kardianos): A better implementation should check for conflicts and
	// deny both. It may also check for permissions or some other allowed
	// resource verification.
	rr := NewRouterRun()
	for _, s := range s.services {
		for _, p := range s.sb.Potential {
			name := path.Join(s.sb.Name, p.Name)
			rr.Potential[name] = &PR{
				Name:          name,
				Type:          p.Type,
				ServiceBundle: s.sb,
				Service:       &s,
			}
		}
		for _, cr := range s.sb.Configured {
			name := path.Join(s.sb.Name, cr.Name)
			rr.Configured[name] = &CR{
				Name:          name,
				PRName:        cr.PotentialResourceName,
				CAuth:         cr.GetAuth(),
				CSPA:          cr.GetSPACode(),
				CURL:          cr.GetURL(),
				CQuery:        cr.GetQuery(),
				ServiceBundle: s.sb,
				Service:       &s,
			}
		}
		for _, b := range s.sb.Bundle {
			name := path.Join(s.sb.Name, b.Name)
			rr.Bundle[name] = &Bundle{
				Name:        name,
				IncludeName: b.Include,
			}
		}
		for _, a := range s.sb.Application {
			app := &App{
				Host:        a.Host,
				AuthName:    a.AuthConfiguredResource,
				LoginBundle: make(map[api.LoginState]*LoginBundle, len(a.LoginBundle)),
			}
			for _, h := range app.Host {
				rr.App[h] = app
			}
			for _, lb := range a.LoginBundle {
				app.LoginBundle[lb.LoginState] = &LoginBundle{
					LoginState:      lb.LoginState,
					Prefix:          lb.Prefix,
					ConsumeRedirect: lb.ConsumeRedirect,
					BundleName:      lb.Bundle,

					URLRouter: make(map[string]*CR),
				}
			}
		}
	}

	rr.resolveNames()
	s.router = rr
}

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
	go s.updateServiceAddress(n.ServiceAddress)
	return &google_protobuf1.Empty{}, nil
}
func (s *RouterServer) Update(ctx context.Context, u *api.UpdateReq) (*api.UpdateResp, error) {
	fmt.Printf("Update: action=%q bind=%q bundle=%q host=%q", u.Action, u.Bind, u.Bundle, u.Host)
	return &api.UpdateResp{}, nil
}
