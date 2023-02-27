// Copyright 2023 OnMetal authors
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
	"fmt"
	"net/http"
	"sync"

	utilrest "github.com/onmetal/onmetal-api/utils/rest"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/rest"
)

const FakeConfigRotatorName = "fake-config-rotator"

type FakeConfigRotatorListenerHandle struct {
	utilrest.ConfigRotatorListener
}

type FakeConfigRotator struct {
	mu sync.RWMutex

	started         bool
	clientConfig    *rest.Config
	transportConfig *rest.Config
	listeners       sets.Set[*FakeConfigRotatorListenerHandle]
}

func NewFakeConfigRotator() *FakeConfigRotator {
	return &FakeConfigRotator{
		listeners: sets.New[*FakeConfigRotatorListenerHandle](),
	}
}

func (f *FakeConfigRotator) Start(ctx context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.started {
		return fmt.Errorf("rotator already started")
	}
	f.started = true
	return nil
}

func (f *FakeConfigRotator) Started() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.started
}

func (f *FakeConfigRotator) Name() string {
	return FakeConfigRotatorName
}

func (f *FakeConfigRotator) Check(req *http.Request) error {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.clientConfig == nil {
		return fmt.Errorf("client config is unset")
	}

	if !utilrest.IsConfigValid(f.clientConfig) {
		return fmt.Errorf("client config is not valid")
	}

	return nil
}

func (f *FakeConfigRotator) Init(ctx context.Context, force bool) error {
	return nil
}

func (f *FakeConfigRotator) ClientConfig() *rest.Config {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.clientConfig
}

func (f *FakeConfigRotator) SetClientConfig(cfg *rest.Config) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.clientConfig = cfg
}

func (f *FakeConfigRotator) TransportConfig() *rest.Config {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.transportConfig
}

func (f *FakeConfigRotator) SetTransportConfig(cfg *rest.Config) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.transportConfig = cfg
}

func (f *FakeConfigRotator) AddListener(listener utilrest.ConfigRotatorListener) utilrest.ConfigRotatorListenerRegistration {
	f.mu.Lock()
	defer f.mu.Unlock()

	handle := &FakeConfigRotatorListenerHandle{listener}
	f.listeners.Insert(handle)
	return nil
}

func (f *FakeConfigRotator) RemoveListener(reg utilrest.ConfigRotatorListenerRegistration) {
	f.mu.Lock()
	defer f.mu.Unlock()

	handle, ok := reg.(*FakeConfigRotatorListenerHandle)
	if !ok {
		return
	}
	f.listeners.Delete(handle)
}

func (f *FakeConfigRotator) Listeners() []utilrest.ConfigRotatorListener {
	f.mu.RLock()
	defer f.mu.RUnlock()
	res := make([]utilrest.ConfigRotatorListener, 0, len(f.listeners))
	for handle := range f.listeners {
		res = append(res, handle.ConfigRotatorListener)
	}
	return res
}

func (f *FakeConfigRotator) EnqueueAll() {
	listeners := f.Listeners()
	for _, listener := range listeners {
		listener.Enqueue()
	}
}
