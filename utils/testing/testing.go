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

package testing

import (
	"context"
	"math/rand"
	"strings"
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

// LowerCaseAlphabetCharset is a charset consisting of lower-case alphabet letters.
const LowerCaseAlphabetCharset = "abcdefghijklmnopqrstuvwxyz"

// RandomStringOptions are options for RandomString.
type RandomStringOptions struct {
	// Charset overrides the default RandomString charset if non-empty.
	Charset string
}

// ApplyToRandomString implements RandomStringOption.
func (o *RandomStringOptions) ApplyToRandomString(o2 *RandomStringOptions) {
	if o.Charset != "" {
		o2.Charset = o.Charset
	}
}

// ApplyOptions applies the slice of RandomStringOption to the RandomStringOptions.
func (o *RandomStringOptions) ApplyOptions(opts []RandomStringOption) {
	for _, opt := range opts {
		opt.ApplyToRandomString(o)
	}
}

// RandomStringOption is an option to RandomString.
type RandomStringOption interface {
	// ApplyToRandomString modifies the given RandomStringOptions with the option settings.
	ApplyToRandomString(o *RandomStringOptions)
}

// Charset specifies an explicit charset to use.
type Charset string

// ApplyToRandomString implements RandomStringOption.
func (s Charset) ApplyToRandomString(o *RandomStringOptions) {
	o.Charset = string(s)
}

// RandomString generates a random string of length n with the given options.
// If n is negative, RandomString panics.
func RandomString(n int, opts ...RandomStringOption) string {
	if n < 0 {
		panic("RandomString: negative length")
	}

	o := RandomStringOptions{}
	o.ApplyOptions(opts)

	charset := o.Charset
	if charset == "" {
		charset = LowerCaseAlphabetCharset
	}

	var sb strings.Builder
	for i := 0; i < n; i++ {
		sb.WriteRune(rune(charset[rand.Intn(len(charset))]))
	}
	return sb.String()
}
