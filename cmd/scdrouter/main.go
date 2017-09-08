// Copyright 2017 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// scdrouter accepts incomming connections and routes the requests to the
// correct service. It also unifies the services into a single application.
package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

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
	const redirectQueryKey = "redirect-to"
	s.rlk.RLock()
	appToken, found := s.router.App[r.Host]
	s.rlk.RUnlock()

	if !found {
		http.Error(w, "host record not found", http.StatusNotFound)
		return
	}
	app := appToken.App

	token := ""
	if c, err := r.Cookie(appToken.TokenKey); err == nil {
		token = c.Value
	}

	if app.Auth == nil {
		http.Error(w, "auth not configured", http.StatusInternalServerError)
		return
	}
	authResp, err := app.Auth.RequestAuth(r.Context(), &api.RequestAuthReq{
		Token:         token,
		Configuration: app.AuthConfig,
	})
	if err != nil {
		http.Error(w, "auth: "+err.Error(), http.StatusInternalServerError)
		return
	}
	authResp.TokenKey = appToken.TokenKey

	lb, found := app.LoginBundle[authResp.LoginState]
	if !found {
		http.Error(w, "unconfigured login state: "+authResp.LoginState.String(), http.StatusInternalServerError)
		return
	}
	ctx := r.Context()

	switch lb.ConsumeRedirect {
	case false:
		// If the redirect query value should not be consumed (like on a login application),
		// then set the redirect query key before redirecting.
		if strings.HasPrefix(r.URL.Path, lb.Prefix) == false {
			redirectQuery := url.Values{}
			if strings.Count(r.URL.Path, "/") >= 2 {
				redirectQuery.Set(redirectQueryKey, r.URL.RequestURI())
			}
			nextURL := &url.URL{Path: lb.Prefix, RawQuery: redirectQuery.Encode()}
			http.Redirect(w, r, nextURL.String(), http.StatusTemporaryRedirect)
			return
		}
	case true:
		rq := r.URL.Query()
		nextURL := rq.Get(redirectQueryKey)
		if len(nextURL) > 0 {
			// If there is not a prefix match, then remove it from
			// the query string and redirect to the same URL but without the
			// redirect.
			if !strings.HasPrefix(nextURL, lb.Prefix) {
				rq.Del(redirectQueryKey)
				r.URL.RawQuery = rq.Encode()
				nextURL = r.URL.String()
			}

			http.Redirect(w, r, nextURL, http.StatusTemporaryRedirect)
			return
		}

		// There is no redirect, but the prefix does not match the URL path.
		// Redirect to the correct prefix.
		if strings.HasPrefix(r.URL.Path, lb.Prefix) == false {
			http.Redirect(w, r, lb.Prefix, http.StatusTemporaryRedirect)
			return
		}
	}

	r.URL.Path = strings.TrimPrefix(r.URL.Path, lb.Prefix[:len(lb.Prefix)-1])

	cr, found := lb.URLRouter[r.URL.Path]
	if !found {
		http.Error(w, fmt.Sprintf("path not found %q", r.URL.Path), http.StatusNotFound)
		return
	}

	const readLimit = 1024 * 1024 * 100 // 100 MB. In the future make this property part of the AppHandler interface or RouteHandler.
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, readLimit))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var tls *api.TLSState
	if r.TLS != nil {
		tls = &api.TLSState{
			Version:           uint32(r.TLS.Version),
			HandshakeComplete: r.TLS.HandshakeComplete,
			DidResume:         r.TLS.DidResume,
			CipherSuite:       uint32(r.TLS.CipherSuite),
			ServerName:        r.TLS.ServerName,
		}
	}

	appReq := &api.HTTPRequest{
		Method: r.Method,
		URL: &api.URL{
			Host:  r.URL.Host,
			Path:  cr.PR.Name, // r.URL.Path,
			Query: api.NewKeyValueList(r.URL.Query()),
		},
		ProtoMajor:  int32(r.ProtoMajor),
		ProtoMinor:  int32(r.ProtoMinor),
		Body:        body,
		Header:      api.NewKeyValueList(r.Header),
		ContentType: r.Header.Get("Content-Type"),
		Host:        r.Host,
		RemoteAddr:  r.RemoteAddr,
		TLS:         tls,
		Auth:        authResp,
		Config:      cr.CURL,
	}

	appResp, err := cr.PR.Handler.ServeHTTP(ctx, appReq)
	if err != nil {
		/*if status, ok := err.(HTTPError); ok {
			msg := status.Msg
			if len(msg) == 0 {
				msg = status.Err.Error()
			}
			http.Error(w, msg, status.Status)
			return
		}*/
		lb, found := app.LoginBundle[api.LoginState_Error]
		if !found {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		cr, found := lb.URLRouter["/"]
		if !found {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		appReq.ContentType = "error"
		appReq.Body = []byte(err.Error())
		appResp, err = cr.PR.Handler.ServeHTTP(ctx, appReq)
		if err != nil {
			http.Error(w, "unable to render error page: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if len(appResp.ContentType) > 0 {
		w.Header().Set("Content-Type", appResp.ContentType)
	}
	if len(appResp.Encoding) > 0 {
		w.Header().Set("Content-Encoding", appResp.Encoding)
	}

	if appResp.Header != nil {
		for key, values := range appResp.Header.Values {
			for _, v := range values.Value {
				w.Header().Add(key, v)
			}
		}
	}
	w.WriteHeader(http.StatusOK)
	w.Write(appResp.Body)

}

var _ api.RouterConfigurationServer = &RouterServer{}

type serviceDef struct {
	serviceAddress string
	conn           *grpc.ClientConn
	sb             *api.ServiceBundle
}

type RouterServer struct {
	ctx context.Context

	slk      sync.Mutex
	services map[string]serviceDef

	rlk    sync.RWMutex
	router *RouterRun

	updateRouter chan *RouterRun
}

func NewRouterServer(ctx context.Context) *RouterServer {
	s := &RouterServer{
		ctx:          ctx,
		services:     make(map[string]serviceDef, 30),
		updateRouter: make(chan *RouterRun, 6),
	}
	go s.runUpdateRouter(ctx)

	return s
}
func (s *RouterServer) runUpdateRouter(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case rr := <-s.updateRouter:
			rr.resolveNames()
			if len(rr.Errors) > 0 {
				log.Printf("router: configuration errors:\n\t%s\n", strings.Join(rr.Errors, "\n\t"))
				continue
			}
			// Version routes. Assign each new RouterRun a UUID.
			// attach version to all requests.
			// Updating version has the following steps:
			//  * Add new version to remotes.
			//  * Update local router around exclusive lock.
			//  * Remove old version from remotes.
			// If there is an issue adding version to the remotes, the preivous
			// version will still work.
			err := rr.updateServices(ctx, api.ServiceConfigAction_Add)
			if err != nil {
				log.Printf("router: failed to add new router %v", err)
				continue
			}

			s.rlk.Lock()
			old := s.router
			s.router = rr
			s.rlk.Unlock()

			if old != nil {
				err = old.updateServices(ctx, api.ServiceConfigAction_Remove)
				if err != nil {
					log.Printf("router: failed to remove prior router %v", err)
				}
			}
			log.Println("router: configuration Updated")
		}
	}
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
	s.slk.Lock()
	defer s.slk.Unlock()

	delete(s.services, serviceName)
	s.updateCompleteSLocked()
}

func (s *RouterServer) updateService(serviceAddress string, conn *grpc.ClientConn, sb *api.ServiceBundle) {
	fmt.Printf("update %q\n", sb.Name)
	s.slk.Lock()
	defer s.slk.Unlock()

	s.services[sb.Name] = serviceDef{
		serviceAddress: serviceAddress,
		conn:           conn,
		sb:             sb,
	}
	s.updateCompleteSLocked()
}

var versionPrefix = ""

func init() {
	b := make([]byte, 6)
	rand.Read(b)
	versionPrefix = base64.RawURLEncoding.EncodeToString(b)
}

func NewRouterRun() *RouterRun {
	rr := &RouterRun{
		Potential:  make(map[string]*PR),
		Configured: make(map[string]*CR),
		Bundle:     make(map[string]*Bundle),
		App:        make(map[string]AppToken),
	}

	// Create a unique version of this router run.
	b := make([]byte, 6)
	rand.Read(b)
	suffix := base64.RawURLEncoding.EncodeToString(b)
	rr.Version = fmt.Sprintf("%s%d%s", versionPrefix, time.Now().Unix(), suffix)

	return rr
}

func (rr *RouterRun) updateServices(ctx context.Context, action api.ServiceConfigAction) error {
	if action == api.ServiceConfigAction_Remove {
		fmt.Printf("REMOVE %s\n", rr.Version)

		svcs := map[*serviceDef]bool{}

		for _, appToken := range rr.App {
			app := appToken.App
			for _, lbundle := range app.LoginBundle {
				for _, include := range lbundle.Bundle.Include {
					switch include.PR.Consume {
					case api.Consume_ConsumeNone:
					default:
						svcs[include.PR.Service] = true
					}
				}
			}
		}
		for s := range svcs {
			client := api.NewRoutesClient(s.conn)
			_, err := client.UpdateServiceConfig(ctx, &api.ServiceConfig{
				Action:  action,
				Version: rr.Version,
			})
			if err != nil {
				// Don't error out, we want to try to remove from each service,
				// even if one fails.
				log.Println("router: failed to remove service config %v", err)
			}
		}
		return nil
	}
	fmt.Printf("ADD %s\n", rr.Version)

	svcs := map[*serviceDef][]api.PotentialResource_ResourceType{}
	consume := map[api.PotentialResource_ResourceType]*serviceDef{}

	type UniqueResource struct {
		P map[string]*PR
		R map[string]*CR
	}
	servicePerConsumer := map[*serviceDef]map[*serviceDef]*UniqueResource{}

	// 1. Lookup where to send service from resource type.
	// 2. Add CR/PR to destination bucket.

	for _, appToken := range rr.App {
		app := appToken.App

		for _, lbundle := range app.LoginBundle {
			for _, include := range lbundle.Bundle.Include {
				switch include.PR.Consume {
				case api.Consume_ConsumeNone:
				default:
					consume[include.PR.Type] = include.Service
					svcs[include.PR.Service] = append(svcs[include.PR.Service], include.PR.Type)
				}
				fmt.Printf("\t%s <- %s\n", include.Name, include.PRName)
			}
		}
	}
	for _, appToken := range rr.App {
		app := appToken.App

		for _, lbundle := range app.LoginBundle {
			for _, include := range lbundle.Bundle.Include {
				sendTo, ok := consume[include.PR.Type]
				if !ok {
					continue
				}

				assocService, ok := servicePerConsumer[sendTo]
				if !ok {
					assocService = map[*serviceDef]*UniqueResource{}
					servicePerConsumer[sendTo] = assocService
				}
				ur, ok := assocService[include.Service]
				if !ok {
					ur = &UniqueResource{
						P: map[string]*PR{},
						R: map[string]*CR{},
					}
					assocService[include.Service] = ur
				}

				ur.P[include.PR.Name] = include.PR
				ur.R[include.Name] = include
			}
		}
	}

	for c, _ := range svcs {
		sc := &api.ServiceConfig{
			Version: rr.Version,
			Action:  action,
		}

		for epService, ur := range servicePerConsumer[c] {
			ep := &api.ServiceConfigEndpoint{
				Endpoint: epService.serviceAddress,
			}
			sc.List = append(sc.List, ep)

			for _, pr := range ur.P {
				ep.Potential = append(ep.Potential, &api.PotentialResource{
					Name:    pr.Name,
					Type:    pr.Type,
					Consume: pr.Consume,
				})
			}
			for _, cr := range ur.R {
				acr := &api.ConfiguredResource{
					Name: cr.Name,
					PotentialResourceName: cr.PRName,
				}
				switch {
				default:
					return fmt.Errorf("unknown configuration type")
				case cr.CAuth != nil:
					acr.Configuration = &api.ConfiguredResource_Auth{cr.CAuth}
				case cr.CQuery != nil:
					acr.Configuration = &api.ConfiguredResource_Query{cr.CQuery}
				case cr.CSPA != nil:
					acr.Configuration = &api.ConfiguredResource_SPACode{cr.CSPA}
				case cr.CURL != nil:
					acr.Configuration = &api.ConfiguredResource_URL{cr.CURL}
				}
				ep.Configured = append(ep.Configured, acr)
			}
		}

		client := api.NewRoutesClient(c.conn)
		_, err := client.UpdateServiceConfig(ctx, sc)
		if err != nil {
			return err
		}
	}
	return nil
}

type AppToken struct {
	TokenKey string
	App      *App
}

type RouterRun struct {
	Version    string
	Potential  map[string]*PR
	Configured map[string]*CR
	Bundle     map[string]*Bundle
	App        map[string]AppToken

	Errors []string
}
type PR struct {
	Name          string
	Type          api.PotentialResource_ResourceType
	ServiceBundle *api.ServiceBundle
	Service       *serviceDef
	Consume       api.Consume

	Handler api.HTTPClient
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
	rr.Errors = rr.Errors[:0] // Clear errors first.

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
	for _, at := range rr.App {
		a := at.App
		if len(a.AuthName) == 0 {
			rr.AddError("app on %q missing authentication", a.Host)
		} else {
			if cr, found := rr.Configured[a.AuthName]; found && cr.PR != nil && cr.PR.Service != nil {
				a.Auth = api.NewAuthClient(cr.PR.Service.conn)
				a.AuthConfig = cr.CAuth
				fmt.Printf("send traffic from %q to %q\n", a.Host, cr.PR.Service.sb.Name)
			} else {
				rr.AddError("app on %q unable to resolve authenticator %q", a.Host, a.AuthName)
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

func (s *RouterServer) updateCompleteSLocked() {
	// TODO(kardianos): A better implementation should check for conflicts and
	// deny both. It may also check for permissions or some other allowed
	// resource verification.
	rr := NewRouterRun()
	for _, s := range s.services {
		locals := s
		handler := api.NewHTTPClient(s.conn)
		for _, p := range s.sb.Potential {
			name := path.Join(s.sb.Name, p.Name)
			rr.Potential[name] = &PR{
				Name:          p.Name,
				Type:          p.Type,
				ServiceBundle: s.sb,
				Service:       &locals,
				Handler:       handler,
				Consume:       p.Consume,
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
				Service:       &locals,
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
			for _, h := range a.Host {
				rr.App[h] = AppToken{
					App: app,

					// Unique cookie key per host. Cookies are shared per hostname
					// and ignore port differences.
					TokenKey: tokenKeyName(h),
				}
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

	s.updateRouter <- rr
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
