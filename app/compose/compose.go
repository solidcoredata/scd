// Copyright 2017 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package compose

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/solidcoredata/scdhttp/scdhandler"
)

// TODO(kardianos): define an interface that allows returning static assets and can handle dynamic or static page requests.
/*
	Provide a list of static assets.
	Optionally map the static assets to a custom path.
	A custom path may be "/" to serve a static root HTTP page.

	Provide a list of dynamic handlers and their mount points.
	A custom path for a dynamic handler may be "/" to serve a dynamic root HTTP page.

	Use same interface for logged in state as well.

	This handlers methods are uniformally coposed.
	Handler may provide endpoints or provide endpoints.
	Both should be declared up front.

	Each component should declare what it uses. (yes)
*/

func NewHandler(session scdhandler.SessionManager, prefix string, consumeRedirect bool, cc ...scdhandler.AppComponentHandler) scdhandler.AppHandler {
	return &handler{
		ses:             session,
		prefix:          prefix,
		consumeRedirect: consumeRedirect,
		components:      cc,
	}
}

type handler struct {
	ses             scdhandler.SessionManager
	prefix          string
	consumeRedirect bool
	components      []scdhandler.AppComponentHandler
	routes          map[string]scdhandler.AppComponentHandler
}

var _ scdhandler.AppHandler = &handler{}

func (h *handler) URLPartition() (prefix string, consumeRedirect bool) {
	return h.prefix, h.consumeRedirect
}

func (h *handler) Init(ctx context.Context) error {
	for _, c := range h.components {
		err := c.Init(ctx)
		if err != nil {
			return err
		}
	}
	rc := RouteConflict{}
	h.routes = make(map[string]scdhandler.AppComponentHandler, 30)
	for _, c := range h.components {
		mounts, err := c.ProvideMounts(ctx)
		if err != nil {
			return err
		}
		for _, m := range mounts {
			_, has := h.routes[m.At]
			if has {
				rc.Add(m.At, c)
				continue
			}
			h.routes[m.At] = c
		}
	}
	if len(rc) != 0 {
		return rc
	}
	if _, found := h.routes["/"]; !found {
		return fmt.Errorf(`missing route for root "/" path`)
	}
	return nil
}

// Request should be routed by the r.URL.Path field.
func (h *handler) Request(ctx context.Context, r *scdhandler.Request) (*scdhandler.Response, error) {
	ch, found := h.routes[r.URL.Path]
	if !found {
		return nil, scdhandler.HTTPError{Status: http.StatusNotFound}
	}
	return ch.Request(ctx, r)
}

type conflict struct {
	At   string
	With []scdhandler.AppComponentHandler
}

func (c *conflict) Error() string {
	buf := bytes.Buffer{}
	buf.WriteString("path conflict at ")
	buf.WriteString(c.At)
	buf.WriteString(" between: ")
	for i, ch := range c.With {
		if i != 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(fmt.Sprintf("%T", ch))
	}
	return buf.String()
}

type RouteConflict map[string]*conflict

func (rc RouteConflict) Error() string {
	buf := bytes.Buffer{}
	first := true
	for _, c := range rc {
		if !first {
			buf.WriteString("\n")
		}
		first = false
		buf.WriteString(c.Error())
	}
	return buf.String()
}
func (rc RouteConflict) Add(at string, ch scdhandler.AppComponentHandler) {
	c, has := rc[at]
	if !has {
		c = &conflict{
			At: at,
		}
		rc[at] = c
	}
	c.With = append(c.With, ch)
}
