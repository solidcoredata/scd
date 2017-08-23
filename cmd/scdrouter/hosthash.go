// Copyright 2017 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/base64"
	"hash"
	"strings"

	"github.com/minio/blake2b-simd"
)

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
func tokenKeyName(host string) string {
	return strings.TrimRight(base64.RawURLEncoding.EncodeToString(tokenKeyHasher.Sum([]byte(host))), "=")
}
