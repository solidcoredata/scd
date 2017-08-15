// Copyright 2017 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package handler provides various HTTP handlers for a server setup.
//
// The design is that each "application" can be plugged into the system.
// Eventually this will run in a k8s style components that are configured.
// At the moment use code to create servers.
//
// The most important design will be the end application under login_granted
// that will proof out various aspects of the system, including just-in-time
// component handling.
package handler

import (
	"context"
	"encoding/base64"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/solidcoredata/scd/api"

	"github.com/minio/blake2b-simd"
)

// LoginState is a single value for each possible state a login can be in
// when checked.
type LoginState int64

// LoginRole is given meaning by the application and may be composed with
// each other for additional meaning.
// Multiple LoginRoles may be assigned to a request.
type LoginRole int64

//go:generate stringer -type=LoginState

// Known login states.
const (
	LoginError          LoginState = 0 // Error in system configuration, load, or disconnected backend component.
	LoginNone           LoginState = 1 // No credentials.
	LoginGranted        LoginState = 2 // User has valid credentials.
	LoginU2F            LoginState = 3 // Require user to identify with second factor.
	LoginChangePassword LoginState = 4 // Require user to change password.
)

// TODO how to return this correctly from an API?
type HTTPError struct {
	Status int
	Msg    string
	Err    error
}

func (e HTTPError) Error() string {
	msg := e.Msg
	if len(msg) == 0 {
		msg = e.Err.Error()
	}
	return fmt.Sprintf("%d: %s", e.Status, msg)
}

// LoginStateRouter links login states with a routable handler.
type LoginStateRouter struct {
	State map[LoginState]api.AppHandlerClient

	// Authenticator is sent the token as found in the request cookie.
	// It is called after detaching the cookie value and before any routing has happened.
	Authenticator api.AuthClient
}

// URLRouter links an incomming URL Host with a LoginStateRouter.
type HostRouter map[string]LoginStateRouter

// RouteHandler routes requests after coming in from the edge of the system.
// It ensures requests are authenticated to any sensitive areas. It gives downstream
// components isolation.
type RouteHandler struct {
	// Router maps the request to a URL and login state.
	Router HostRouter

	// Loggerf allows feedback to the system.
	// TODO(kardianos): probably not the final interface.
	Loggerf func(f string, v ...interface{})
}

func (h *RouteHandler) Init(ctx context.Context) error {
	uniqueAuth := make(map[api.AuthClient]bool, len(h.Router))
	uniqueApp := make(map[api.AppHandlerClient]bool, len(h.Router)*4)

	// Check each application handler to ensure the path prefix all
	// start and end with a "/".
	//
	// Add each authenticator and app to a unique map before calling Init on each.
	for host, logins := range h.Router {
		for state, login := range logins.State {
			partition, err := login.URLPartition(ctx, nil)
			if err != nil {
				return err
			}
			if len(partition.Prefix) == 0 || partition.Prefix[0] != '/' || partition.Prefix[len(partition.Prefix)-1] != '/' {
				return fmt.Errorf(`URL Parition must begin and end with a slash "/" for %q for state %s.`, host, state)
			}
			uniqueApp[login] = true
		}
		uniqueAuth[logins.Authenticator] = true
	}

	return nil
}

func (h *RouteHandler) logf(f string, v ...interface{}) {
	if h.Loggerf == nil {
		return
	}
	h.Loggerf(f, v...)
}

const redirectQueryKey = "redirect-to"

var tokenKeyHMAC = []byte(`solidcoredata`)
var tokenKeyHasher hash.Hash

func init() {
	h, err := blake2b.New(&blake2b.Config{
		Size: 4,
		Key:  tokenKeyHMAC,
	})
	if err != nil {
		panic("unable to create token key hasher")
	}
	tokenKeyHasher = h
}
func TokenKeyName(host string) string {
	return strings.TrimRight(base64.RawURLEncoding.EncodeToString(tokenKeyHasher.Sum([]byte(host))), "=")
}

func (h *RouteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get values of credential tokens from request.
	// Get URL.
	// Send credential tokens to authentcation server.
	// Attach result to request context.
	// Send request to correct application server.

	urlrouter, ok := h.Router[r.Host]
	if !ok {
		h.logf("host 404: host=%q\n", r.URL.String())
		http.Error(w, "host 404", http.StatusNotFound)
		return
	}

	tokenKey := TokenKeyName(r.Host)

	rs := &api.RequestAuthResp{}
	if c, err := r.Cookie(tokenKey); err == nil {
		rs, err = urlrouter.Authenticator.RequestAuth(r.Context(), &api.RequestAuthReq{Token: c.Value})
		if err != nil {
			h.logf("scdhandler: unable to check auth %v", err)
			http.Error(w, "failed to check auth", http.StatusInternalServerError)
			return
		}
	}
	rs.TokenKey = tokenKey

	next, ok := urlrouter.State[LoginState(rs.LoginState)]
	if !ok {
		http.Error(w, "login state 404", http.StatusNotFound)
		return
	}

	// Handle redirects to correct URL partition and redirects
	// to protected applications.
	partition, err := next.URLPartition(r.Context(), nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	prefix := partition.Prefix
	switch partition.ConsumeRedirect {
	case false:
		// If the redirect query value should not be consumed (like on a login application),
		// then set the redirect query key before redirecting.
		if strings.HasPrefix(r.URL.Path, prefix) == false {
			redirectQuery := url.Values{}
			if strings.Count(r.URL.Path, "/") >= 2 {
				redirectQuery.Set(redirectQueryKey, r.URL.RequestURI())
			}
			nextURL := &url.URL{Path: prefix, RawQuery: redirectQuery.Encode()}
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
			if !strings.HasPrefix(nextURL, prefix) {
				rq.Del(redirectQueryKey)
				r.URL.RawQuery = rq.Encode()
				nextURL = r.URL.String()
			}

			http.Redirect(w, r, nextURL, http.StatusTemporaryRedirect)
			return
		}

		// There is no redirect, but the prefix does not match the URL path.
		// Redirect to the correct prefix.
		if strings.HasPrefix(r.URL.Path, prefix) == false {
			http.Redirect(w, r, prefix, http.StatusTemporaryRedirect)
			return
		}
	}

	ctx := api.AuthNewContext(r.Context(), rs)
	r.URL.Path = strings.TrimPrefix(r.URL.Path, prefix[:len(prefix)-1])

	const readLimit = 1024 * 1024 * 100 // 100 MB. In the future make this property part of the AppHandler interface or RouteHandler.
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, readLimit))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	appReq := &api.RequestReq{
		Method: r.Method,
		// URL:         r.URL,
		ProtoMajor: int32(r.ProtoMajor),
		ProtoMinor: int32(r.ProtoMinor),
		Body:       body,
		// Header:      r.Header,
		ContentType: r.Header.Get("Content-Type"),
		Host:        r.Host,
		RemoteAddr:  r.RemoteAddr,
		// TLS:         r.TLS,
	}

	appResp, err := next.Request(ctx, appReq)
	if err != nil {
		if status, ok := err.(HTTPError); ok {
			msg := status.Msg
			if len(msg) == 0 {
				msg = status.Err.Error()
			}
			http.Error(w, msg, status.Status)
			return
		}
		next, ok := urlrouter.State[LoginState(rs.LoginState)]
		if !ok {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		appReq.ContentType = "error"
		appReq.Body = []byte(err.Error())
		appResp, err = next.Request(ctx, appReq)
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
	// TODO fixme
	/*
		for key, values := range appResp.Header {
			for _, v := range values {
				w.Header().Add(key, v)
			}
		}
	*/
	w.WriteHeader(http.StatusOK)
	w.Write(appResp.Body)
}
