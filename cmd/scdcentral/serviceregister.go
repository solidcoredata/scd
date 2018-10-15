// Copyright 2018 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"log"
	"time"
)

type Configurable interface {
	Config()
}

type ServiceRegister struct {
	reg Registry
	ctx context.Context

	lease    string
	interval time.Duration
}

func NewServiceRegister(ctx context.Context, reg Registry) (*ServiceRegister, error) {
	interval := time.Second * 9
	lease, err := reg.NewLease(ctx, interval)
	if err != nil {
		return nil, err
	}
	return &ServiceRegister{
		reg: reg,
		ctx: ctx,

		lease:    lease,
		interval: interval,
	}, nil
}

func (sr *ServiceRegister) async() {
	timer := time.NewTicker(sr.interval / 3)
	for {
		select {
		case <-timer.C:
			err := sr.reg.UpdateLease(sr.ctx, sr.lease)
			if err != nil {
				sr.asyncError(err)
			}
		case <-sr.ctx.Done():
			timer.Stop()
			err := sr.reg.DeleteLease(sr.ctx, sr.lease)
			if err != nil {
				sr.asyncError(err)
			}
			return
		}
	}
}

func (sr *ServiceRegister) asyncError(err error) {
	log.Println(err)
}

func (sr *ServiceRegister) Register(ctx context.Context, s ...Configurable) error {
	if len(s) == 0 {
		return nil
	}
	tx, err := sr.reg.Begin(ctx)
	if err != nil {
		return err
	}
	for _, c := range s {
		c.Config()
	}
	return tx.Commit(ctx)
}
