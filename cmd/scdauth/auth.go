// Copyright 2017 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	crand "crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"time"

	"github.com/solidcoredata/scd/api"
)

// BUG(kardianos): normalize identity and passwords.
// BUG(kardianos): hash and salt passwords.
// BUG(kardianos): compute or set next login state.

// MemoryUser holds information in an easy way to setup.
//
// For testing only.
type MemoryUser struct {
	ID         int64
	Identity   string
	Email      string
	GivenName  string
	FamilyName string

	Password           string
	PasswordSalt       []byte
	PasswordMethod     int64
	ForceResetPassword bool

	Roles []int64
}

type MemoryDevices struct {
	Identity     string
	Name         string
	Registration []byte
	Counter      int64
}

type MemorySession struct {
	ID         int64
	Identity   string
	Token      string
	LoginState api.LoginState

	ElevatedUntil time.Time
	TimeExpire    time.Time
	TimeLastHit   time.Time
}

// AuthenticateMemory provides an in-memory authenticator for request authentication.
type AuthenticateMemory struct {
	lk sync.RWMutex

	// UserSetup maps Identity to the user information.
	UserSetup map[string]*MemoryUser

	// Tokens maps a token string to a session.
	Tokens map[string]*MemorySession
}

func NewAuthenticateMemory(users ...*MemoryUser) *AuthenticateMemory {
	am := &AuthenticateMemory{
		Tokens:    make(map[string]*MemorySession, 50),
		UserSetup: make(map[string]*MemoryUser, len(users)),
	}
	for _, u := range users {
		am.UserSetup[u.Identity] = u
	}
	return am
}

// RequestAuth implements the scdhandler.Authenticator.
func (am *AuthenticateMemory) RequestAuth(ctx context.Context, token string) (*api.RequestAuthResp, error) {
	ra := &api.RequestAuthResp{}
	am.lk.RLock()
	defer am.lk.RUnlock()

	t, ok := am.Tokens[token]
	if !ok {
		ra.LoginState = api.LoginState_None
		return ra, nil
	}
	u, ok := am.UserSetup[t.Identity]
	if !ok {
		return ra, nil
	}
	ra.Identity = t.Identity
	ra.Email = u.Email
	ra.LoginState = t.LoginState
	// ra.ElevatedUntil.Seconds = t.ElevatedUntil.Unix()
	ra.GivenName = u.GivenName
	ra.FamilyName = u.FamilyName
	ra.Roles = u.Roles
	return ra, nil
}

func (am *AuthenticateMemory) random(length int) (tokenValue []byte, err error) {
	tokenValue = make([]byte, length)
	n, err := crand.Read(tokenValue)
	if err != nil {
		return nil, err
	}
	if n != length {
		return nil, errors.New("short random read")
	}
	return tokenValue, nil
}

func (am *AuthenticateMemory) NewPassword(ctx context.Context, identity string) error {
	return nil
}

func (am *AuthenticateMemory) ResetPassword(ctx context.Context, sessionToken, oldpassword, newpassword string) error {
	return nil
}

var errLoginFailed = errors.New("auth: failed to login")

func (am *AuthenticateMemory) Login(ctx context.Context, identity, password string) (tokenValue string, err error) {
	am.lk.Lock()
	defer am.lk.Unlock()

	u, exists := am.UserSetup[identity]
	if !exists {
		return "", errLoginFailed
	}
	// TODO(kardianos): yeah, this is bad. But we just need the mechanics to function, not work.
	if u.Password != password {
		return "", errLoginFailed
	}

	var t *MemorySession

	for i := 0; i < 5; i++ {
		tokenBytes, err := am.random(160)
		if err != nil {
			return "", err
		}
		tokenValue = base64.RawURLEncoding.EncodeToString(tokenBytes)

		t, exists = am.Tokens[tokenValue]
		if exists {
			continue
		}
		t = &MemorySession{
			Identity:    identity,
			Token:       tokenValue,
			LoginState:  api.LoginState_Granted, // TODO(kardianos): compute as required.
			TimeExpire:  time.Now().Add(28 * 24 * time.Hour),
			TimeLastHit: time.Now(),
		}
		am.Tokens[tokenValue] = t

		return tokenValue, nil
	}
	return "", errLoginFailed
}

func (am *AuthenticateMemory) Logout(ctx context.Context, sessionToken string) error {
	am.lk.Lock()
	defer am.lk.Unlock()

	delete(am.Tokens, sessionToken)
	return nil
}
func (am *AuthenticateMemory) LogoutIdentity(ctx context.Context, identity string) error {
	am.lk.Lock()
	defer am.lk.Unlock()

	tokens := make([]string, 0, 3)
	for tokenValue, t := range am.Tokens {
		if t.Identity == identity {
			tokens = append(tokens, tokenValue)
		}
	}
	for _, tokenValue := range tokens {
		delete(am.Tokens, tokenValue)
	}

	return nil
}
