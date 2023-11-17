// Copyright 2022 IronCore authors
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

	"github.com/ironcore-dev/ironcore/utils/generic"
	"github.com/ironcore-dev/ironcore/utils/klog"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/gcustom"
	"github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

// BeControlledBy matches any object that is controlled by the given owner.
func BeControlledBy(owner client.Object) types.GomegaMatcher {
	return gcustom.MakeMatcher(func(obj client.Object) (bool, error) {
		return metav1.IsControlledBy(obj, owner), nil
	}).WithTemplate("Expected:\n{{.FormattedActual}}\n{{.To}} be controlled by {{.Data 1}}", klog.KObjUID(owner))
}

// SetupNamespace sets up a namespace before each test and tears the namespace down after each test.
func SetupNamespace(c *client.Client) *corev1.Namespace {
	return SetupObjectStruct[*corev1.Namespace](c, func(ns *corev1.Namespace) {
		*ns = corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-ns",
			},
		}
	})
}

func SetupObject(c *client.Client, obj client.Object, f func()) {
	ginkgo.BeforeEach(func(ctx ginkgo.SpecContext) {
		f()
		gomega.Expect((*c).Create(ctx, obj)).To(gomega.Succeed())
		ginkgo.DeferCleanup(DeleteIfExists(c, obj))
	})
}

func SetupObjectStruct[O interface {
	client.Object
	*OStruct
}, OStruct any](c *client.Client, f func(obj O)) O {
	obj := O(generic.ZeroPointer[OStruct]())
	SetupObject(c, obj, func() {
		f(obj)
	})
	return obj
}

// DeleteIfExists returns a function to clean up an object if it exists.
func DeleteIfExists(c *client.Client, obj client.Object) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		return client.IgnoreNotFound((*c).Delete(ctx, obj))
	}
}
