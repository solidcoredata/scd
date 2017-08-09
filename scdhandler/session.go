// Copyright 2017 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scdhandler

import (
	"context"
	"encoding/base64"
	"hash"
	"strings"
	"time"

	"github.com/minio/blake2b-simd"
)

// RequestAuth holds the authentication state for a request.
type RequestAuth struct {
	LoginState    LoginState
	Roles         []LoginRole
	ElevatedUntil time.Time

	Identity   string // Unique name each application can use to link to the internal user.
	Email      string
	GivenName  string // The given (first) name of the user.
	FamilyName string // The family (last) name of the user.

	// Key to the session cookie token.
	TokenKey string
}

type requestAuthKey struct{}

// AuthNewContext returns a child context with the RequestAuth as a value.
func AuthNewContext(ctx context.Context, rs *RequestAuth) context.Context {
	return context.WithValue(ctx, requestAuthKey{}, rs)
}

// AuthFromContext returns the RequestAuth found in the context values if found.
func AuthFromContext(ctx context.Context) (rs *RequestAuth, found bool) {
	rs, found = ctx.Value(requestAuthKey{}).(*RequestAuth)
	return rs, found
}

// Authenticator provides the authentication to the request.
type Authenticator interface {
	Init(context.Context) error
	RequestAuth(ctx context.Context, token string) (*RequestAuth, error)
}

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

type SessionManager interface {
	Authenticator

	Login(ctx context.Context, identity, password string) (tokenValue string, err error)
	Logout(ctx context.Context, sessionToken string) error
	LogoutIdentity(ctx context.Context, identity string) error
}

type SessionPasswordChanger interface {
	// NewPassword chooses a new password for the given identity.
	// The identity must be notified of this change.
	NewPassword(ctx context.Context, identity string) error

	// ResetPassword changes the identity's password. If the login state is
	// Change Password, thenn currentPassword is ignored. Otherwise the currentPassword is checked
	// against the current stored password and if valid newPassword is set.
	ResetPassword(ctx context.Context, sessionToken, currentPassword, newPassword string) error
}

type SessionElevator interface {
	SessionManager

	Elevate(ctx context.Context, sessionToken, password string, until time.Time) error
	UnElevate(ctx context.Context, sessionToken string) error
}

type SessionSigner interface {
	SessionManager

	SignRequest(ctx context.Context, sessionToken string) ([]byte, error)
	SignResponse(ctx context.Context, sessionToken string, response []byte) error
}
