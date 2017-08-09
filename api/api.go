package api

import "context"

//go:generate protoc --go_out=plugins=grpc:. auth.proto handler.proto

type requestAuthKey struct{}

// AuthNewContext returns a child context with the RequestAuth as a value.
func AuthNewContext(ctx context.Context, rs *RequestAuthResp) context.Context {
	return context.WithValue(ctx, requestAuthKey{}, rs)
}

// AuthFromContext returns the RequestAuth found in the context values if found.
func AuthFromContext(ctx context.Context) (rs *RequestAuthResp, found bool) {
	rs, found = ctx.Value(requestAuthKey{}).(*RequestAuthResp)
	return rs, found
}
