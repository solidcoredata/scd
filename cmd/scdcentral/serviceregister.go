// Copyright 2018 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type ServiceRegister struct {
	reg Registry
}

func NewServiceRegister(reg Registry) *ServiceRegister {
	return &ServiceRegister{reg: reg}
}

func (sr *ServiceRegister) Register(s ...Configurable) error {
	return nil
}

type Configurable interface {
	Config()
}
