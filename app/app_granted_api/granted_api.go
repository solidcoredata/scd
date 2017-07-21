package app_granted_api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/solidcoredata/scdhttp/scdhandler"
)

type handler struct {
	ses scdhandler.SessionManager
}

var _ scdhandler.AppComponentHandler = &handler{}

func NewHandler(session scdhandler.SessionManager) scdhandler.AppComponentHandler {
	return &handler{
		ses: session,
	}
}

func (h *handler) Init(ctx context.Context) error {
	return nil
}

func (h *handler) RequireMounts(ctx context.Context) ([]scdhandler.MountConsume, error) {
	return nil, nil
}
func (h *handler) OptionalMounts(ctx context.Context) ([]scdhandler.MountConsume, error) {
	return nil, nil
}
func (h *handler) ProvideMounts(ctx context.Context) ([]scdhandler.MountProvide, error) {
	return []scdhandler.MountProvide{
		{At: "/api/logout"},
	}, nil
}

func (h *handler) Request(ctx context.Context, r *scdhandler.Request) (*scdhandler.Response, error) {
	resp := &scdhandler.Response{}
	switch r.URL.Path {
	case "/api/logout":
		rs, found := scdhandler.AuthFromContext(ctx)
		if !found {
			panic("no auth context")
		}
		c, err := r.Cookie(rs.TokenKey)
		if err != nil {
			// If there is no cookie, user may already be logged out.
			return resp, nil
		}
		err = h.ses.Logout(ctx, c.Value)
		if err != nil {
			return nil, fmt.Errorf("unable to logout: %v", err)
		}
		resp.Header = make(map[string][]string, 1)
		// TODO(kardianos): set exire time, secure=true, strict origin.
		c = &http.Cookie{
			Name:   rs.TokenKey,
			Path:   "/",
			MaxAge: -1,
		}
		resp.Header.Add("Set-Cookie", c.String())
	}
	return resp, nil
}
