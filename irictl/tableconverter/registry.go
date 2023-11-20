// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package tableconverter

import (
	"fmt"
	"reflect"

	"github.com/ironcore-dev/ironcore/irictl/api"
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
