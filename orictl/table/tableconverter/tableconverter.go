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

package tableconverter

import (
	"fmt"
	"reflect"

	"github.com/onmetal/onmetal-api/orictl/table"
)

type TableConverter[E any] interface {
	ConvertToTable(v E) (*table.Table, error)
}

type TaggedAnyTableConverter interface {
	TableConverter[any]
	Tag() reflect.Type
}

type Func[E any] func(e E) (*table.Table, error)

func (f Func[E]) ConvertToTable(v E) (*table.Table, error) {
	return f(v)
}

func ForType[E any](
	headersFunc func() ([]table.Header, error),
	rowsFunc func(e E) ([]table.Row, error),
) (TableConverter[E], TableConverter[[]E]) {
	return Func[E](func(e E) (*table.Table, error) {
			headers, err := headersFunc()
			if err != nil {
				return nil, err
			}

			rows, err := rowsFunc(e)
			if err != nil {
				return nil, err
			}

			return &table.Table{
				Headers: headers,
				Rows:    rows,
			}, nil
		}), Func[[]E](func(es []E) (*table.Table, error) {
			headers, err := headersFunc()
			if err != nil {
				return nil, err
			}

			var rows []table.Row
			for _, e := range es {
				moreRows, err := rowsFunc(e)
				if err != nil {
					return nil, err
				}

				rows = append(rows, moreRows...)
			}
			return &table.Table{
				Headers: headers,
				Rows:    rows,
			}, nil
		})
}

type taggedAnyTableConverter struct {
	tag     reflect.Type
	convert func(e any) (*table.Table, error)
}

func (t *taggedAnyTableConverter) Tag() reflect.Type {
	return t.tag
}

func (t *taggedAnyTableConverter) ConvertToTable(e any) (*table.Table, error) {
	actualType := reflect.TypeOf(e)
	if actualType != t.tag {
		return nil, fmt.Errorf("expected type %s but got %T", t.tag, e)
	}
	return t.convert(e)
}

func ToTaggedAny[E any](conv TableConverter[E]) TaggedAnyTableConverter {
	var zero E
	tag := reflect.TypeOf(zero)
	return &taggedAnyTableConverter{
		tag: tag,
		convert: func(e any) (*table.Table, error) {
			if asE, ok := e.(E); ok {
				return conv.ConvertToTable(asE)
			}
			return nil, fmt.Errorf("cannot convert %T to %T", e, zero)
		},
	}
}

type Registry struct {
	convertersByTag map[reflect.Type]TableConverter[any]
}

func NewRegistry() *Registry {
	return &Registry{
		convertersByTag: map[reflect.Type]TableConverter[any]{},
	}
}

func (r *Registry) ConvertToTable(e any) (*table.Table, error) {
	tag := reflect.TypeOf(e)
	converter, ok := r.convertersByTag[tag]
	if !ok {
		return nil, fmt.Errorf("no converter found for type %T", e)
	}
	return converter.ConvertToTable(e)
}

func (r *Registry) Register(tag reflect.Type, conv TableConverter[any]) error {
	if _, ok := r.convertersByTag[tag]; ok {
		return fmt.Errorf("converter for type %s already registered", tag)
	}

	r.convertersByTag[tag] = conv
	return nil
}

func RegisterToRegistry[E any](registry *Registry, conv TableConverter[E]) error {
	tagged := ToTaggedAny(conv)
	return registry.Register(tagged.Tag(), tagged)
}

type RegistryBuilder []func(*Registry) error

func (b *RegistryBuilder) RegisterFunc(funcs ...func(*Registry) error) {
	*b = append(*b, funcs...)
}

func (b *RegistryBuilder) Register(taggedConvs ...TaggedAnyTableConverter) {
	b.RegisterFunc(func(registry *Registry) error {
		for _, taggedConv := range taggedConvs {
			if err := registry.Register(taggedConv.Tag(), taggedConv); err != nil {
				return err
			}
		}
		return nil
	})
}

func (b *RegistryBuilder) AddToRegistry(r *Registry) error {
	for _, f := range *b {
		if err := f(r); err != nil {
			return err
		}
	}
	return nil
}
