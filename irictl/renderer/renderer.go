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

package renderer

import (
	"errors"
	"fmt"
	"io"
)

var ErrRendererNotFound = errors.New("renderer not found")

type Registry struct {
	rendererByName map[string]Renderer
}

func NewRegistry() *Registry {
	return &Registry{
		rendererByName: map[string]Renderer{},
	}
}

func (r *Registry) Register(name string, renderer Renderer) error {
	if _, ok := r.rendererByName[name]; ok {
		return fmt.Errorf("renderer %q already registered", name)
	}
	r.rendererByName[name] = renderer
	return nil
}

func (r *Registry) Get(name string) (Renderer, error) {
	renderer, ok := r.rendererByName[name]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrRendererNotFound, name)
	}
	return renderer, nil
}

type RegistryBuilder []func(*Registry) error

func (r *RegistryBuilder) AddToRegistry(registry *Registry) error {
	for _, f := range *r {
		if err := f(registry); err != nil {
			return err
		}
	}
	return nil
}

func (r *RegistryBuilder) Add(funcs ...func(*Registry) error) {
	for _, f := range funcs {
		*r = append(*r, f)
	}
}

func (r *RegistryBuilder) Register(namesAndRenderers ...any) {
	n := len(namesAndRenderers)
	if (n % 2) != 0 {
		panic(fmt.Errorf("uneven length %d of names and renderers supplied", n))
	}

	var (
		names     []string
		renderers []Renderer
	)

	for i := 0; i < n; i += 2 {
		nameV := namesAndRenderers[i]
		name, ok := nameV.(string)
		if !ok {
			panic(fmt.Errorf("name of pair %d is not a string but %T", i, nameV))
		}

		rendererV := namesAndRenderers[i+1]
		renderer, ok := rendererV.(Renderer)
		if !ok {
			panic(fmt.Errorf("renderer of pair %d is not a Renderer but %T", i, rendererV))
		}

		names = append(names, name)
		renderers = append(renderers, renderer)
	}

	r.Add(func(registry *Registry) error {
		for i := 0; i < n/2; i++ {
			name := names[i]
			renderer := renderers[i]
			if err := registry.Register(name, renderer); err != nil {
				return err
			}
		}
		return nil
	})
}

type Renderer interface {
	Render(v any, w io.Writer) error
}
