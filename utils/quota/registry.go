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

package quota

import (
	"sync"

	"golang.org/x/exp/maps"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

type simpleRegistry struct {
	mu sync.RWMutex

	scheme  *runtime.Scheme
	entries map[schema.GroupKind]Evaluator
}

func NewRegistry(scheme *runtime.Scheme) Registry {
	return &simpleRegistry{
		scheme:  scheme,
		entries: make(map[schema.GroupKind]Evaluator),
	}
}

func AddAllToRegistry(registry Registry, evaluators []Evaluator) error {
	for _, evaluator := range evaluators {
		if err := registry.Add(evaluator); err != nil {
			return err
		}
	}
	return nil
}

func (r *simpleRegistry) Add(e Evaluator) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gvk, err := apiutil.GVKForObject(e.Type(), r.scheme)
	if err != nil {
		return err
	}

	r.entries[gvk.GroupKind()] = e
	return nil
}

func (r *simpleRegistry) Remove(obj client.Object) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gvk, err := apiutil.GVKForObject(obj, r.scheme)
	if err != nil {
		// Don't return an error here - this can't be registered
		// so we can safely return nil here.
		return nil
	}

	delete(r.entries, gvk.GroupKind())
	return nil
}

func (r *simpleRegistry) Get(obj client.Object) (Evaluator, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	gvk, err := apiutil.GVKForObject(obj, r.scheme)
	if err != nil {
		// Don't return an error here - this can't be registered
		// so we can safely return nil here.
		return nil, nil
	}

	return r.entries[gvk.GroupKind()], nil
}

func (r *simpleRegistry) List() []Evaluator {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return maps.Values(r.entries)
}
