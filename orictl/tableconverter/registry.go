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

package tableconverter

import (
	"fmt"
	"reflect"

	"github.com/ironcore-dev/ironcore/orictl/api"
)

type Registry struct {
	convertersByTag map[reflect.Type]TableConverter[any]
}

func NewRegistry() *Registry {
	return &Registry{
		convertersByTag: map[reflect.Type]TableConverter[any]{},
	}
}

func (r *Registry) ConvertToTable(e any) (*api.Table, error) {
	tag := reflect.TypeOf(e)
	converter, ok := r.convertersByTag[tag]
	if !ok {
		return nil, fmt.Errorf("no converter found for type %T", e)
	}
	return converter.ConvertToTable(e)
}

func (r *Registry) Register(tag reflect.Type, conv TableConverter[any]) error {
	if _, ok := r.convertersByTag[tag]; ok {
		return fmt.Errorf("converter for type %s %w", tag, ErrAlreadyExists)
	}

	r.convertersByTag[tag] = conv
	return nil
}

func (r *Registry) Lookup(tag reflect.Type) (TableConverter[any], error) {
	converter, ok := r.convertersByTag[tag]
	if !ok {
		return nil, fmt.Errorf("converter for tag %s %w", tag, ErrNotFound)
	}

	return converter, nil
}

func (r *Registry) Delete(tag reflect.Type) error {
	if _, ok := r.convertersByTag[tag]; !ok {
		return fmt.Errorf("converter for tag %s %w", tag, ErrNotFound)
	}
	delete(r.convertersByTag, tag)
	return nil
}
