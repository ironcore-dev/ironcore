// Copyright 2023 IronCore authors
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
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/ironcore-dev/ironcore/utils/certificate"
	"k8s.io/apimachinery/pkg/util/sets"
)

const FakeRotatorName = "fake-rotator"

type FakeRotatorListenerHandle struct {
	certificate.RotatorListener
}

type FakeRotator struct {
	mu sync.RWMutex

	started   bool
	cert      *tls.Certificate
	listeners sets.Set[*FakeRotatorListenerHandle]
}

func NewFakeRotator() *FakeRotator {
	return &FakeRotator{
		listeners: sets.New[*FakeRotatorListenerHandle](),
	}
}

func (f *FakeRotator) Start(ctx context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.started {
		return fmt.Errorf("rotator already started")
	}
	f.started = true
	return nil
}

func (f *FakeRotator) Started() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.started
}

func (f *FakeRotator) Init(ctx context.Context, force bool) error {
	return nil
}

func (f *FakeRotator) Certificate() *tls.Certificate {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.cert
}

func (f *FakeRotator) SetCertificate(cert *tls.Certificate) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.cert = cert
}

func (f *FakeRotator) AddListener(listener certificate.RotatorListener) certificate.RotatorListenerRegistration {
	f.mu.Lock()
	defer f.mu.Unlock()

	handle := &FakeRotatorListenerHandle{listener}
	f.listeners.Insert(handle)
	return nil
}

func (f *FakeRotator) RemoveListener(reg certificate.RotatorListenerRegistration) {
	f.mu.Lock()
	defer f.mu.Unlock()

	handle, ok := reg.(*FakeRotatorListenerHandle)
	if !ok {
		return
	}
	f.listeners.Delete(handle)
}

func (f *FakeRotator) Listeners() []certificate.RotatorListener {
	f.mu.RLock()
	defer f.mu.RUnlock()
	res := make([]certificate.RotatorListener, 0, len(f.listeners))
	for handle := range f.listeners {
		res = append(res, handle.RotatorListener)
	}
	return res
}

func (f *FakeRotator) EnqueueAll() {
	listeners := f.Listeners()
	for _, listener := range listeners {
		listener.Enqueue()
	}
}

func (f *FakeRotator) Name() string {
	return FakeRotatorName
}

func (f *FakeRotator) Check(_ *http.Request) error {
	f.mu.RLock()
	defer f.mu.RUnlock()
	cert := f.cert
	if cert == nil {
		return fmt.Errorf("certificate has not yet been issued")
	}

	leaf, err := certificate.TLSCertificateLeaf(cert)
	if err != nil {
		return fmt.Errorf("error getting certificate leaf: %w", err)
	}

	if time.Now().After(leaf.NotAfter) {
		return fmt.Errorf("certificate is expired")
	}
	return nil
}
