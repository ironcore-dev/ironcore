// Copyright 2022 OnMetal authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package testutils

import (
	"context"
	"sync"
	"time"

	"github.com/onsi/ginkgo/v2"
)

type DelegatingContext struct {
	lock sync.RWMutex
	ctx  context.Context
}

func (d *DelegatingContext) Deadline() (deadline time.Time, ok bool) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.ctx.Deadline()
}

func (d *DelegatingContext) Done() <-chan struct{} {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.ctx.Done()
}

func (d *DelegatingContext) Err() error {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.ctx.Err()
}

func (d *DelegatingContext) Value(key interface{}) interface{} {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.ctx.Value(key)
}

func (d *DelegatingContext) Fulfill(ctx context.Context) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.ctx = ctx
}

func NewDelegatingContext(ctx context.Context) *DelegatingContext {
	return &DelegatingContext{ctx: ctx}
}

func SetupContext() context.Context {
	initCtx, initCancel := context.WithCancel(context.Background())
	delegCtx := NewDelegatingContext(initCtx)

	ginkgo.BeforeEach(func() {
		ctx, cancel := context.WithCancel(context.Background())
		ginkgo.DeferCleanup(cancel)

		delegCtx.Fulfill(ctx)
		if initCancel != nil {
			initCancel()
			initCancel = nil
		}
	})

	return delegCtx
}
