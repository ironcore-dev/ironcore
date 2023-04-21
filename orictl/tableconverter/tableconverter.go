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

	"github.com/onmetal/onmetal-api/orictl/api"
	"github.com/onmetal/onmetal-api/utils/generic"
)

type TableConverter[E any] interface {
	ConvertToTable(v E) (*api.Table, error)
}

type TaggedAnyTableConverter interface {
	TableConverter[any]
	Tag() reflect.Type
}

type Func[E any] func(e E) (*api.Table, error)

func (f Func[E]) ConvertToTable(v E) (*api.Table, error) {
	return f(v)
}

type Funcs[E any] struct {
	Headers func() ([]api.Header, error)
	Rows    func(e E) ([]api.Row, error)
}

func Headers(headers []api.Header) func() ([]api.Header, error) {
	return func() ([]api.Header, error) {
		return headers, nil
	}
}

func SingleRowFrom[E any](f func(e E) (api.Row, error)) func(e E) ([]api.Row, error) {
	return func(e E) ([]api.Row, error) {
		row, err := f(e)
		if err != nil {
			return nil, err
		}
		return []api.Row{row}, nil
	}
}

func (f Funcs[E]) ConvertToTable(v E) (*api.Table, error) {
	headers, err := f.Headers()
	if err != nil {
		return nil, err
	}

	rows, err := f.Rows(v)
	if err != nil {
		return nil, err
	}

	return &api.Table{
		Headers: headers,
		Rows:    rows,
	}, nil
}

type SliceFuncs[E any] Funcs[E]

func (f SliceFuncs[E]) ConvertToTable(es []E) (*api.Table, error) {
	headers, err := f.Headers()
	if err != nil {
		return nil, err
	}

	var rows []api.Row
	for _, e := range es {
		moreRows, err := f.Rows(e)
		if err != nil {
			return nil, err
		}

		rows = append(rows, moreRows...)
	}

	return &api.Table{
		Headers: headers,
		Rows:    rows,
	}, nil
}

func Table[E any](table *api.Table) TableConverter[E] {
	return Func[E](func(E) (*api.Table, error) {
		return table, nil
	})
}

func MergeFuncs[E any](funcs ...Funcs[E]) Funcs[E] {
	if len(funcs) == 1 {
		return funcs[0]
	}

	return Funcs[E]{
		Headers: func() ([]api.Header, error) {
			var headers []api.Header
			for _, f := range funcs {
				h, err := f.Headers()
				if err != nil {
					return nil, err
				}
				headers = append(headers, h...)
			}
			return headers, nil
		},
		Rows: func(e E) ([]api.Row, error) {
			var rows []api.Row
			for _, f := range funcs {
				r, err := f.Rows(e)
				if err != nil {
					return nil, err
				}

				rows = permuteRows(rows, r)
			}
			return rows, nil
		},
	}
}

func permuteRows(r1, r2 []api.Row) []api.Row {
	if len(r1) == 0 {
		return r2
	}
	if len(r2) == 0 {
		return r1
	}

	perm := make([]api.Row, 0, len(r1)*len(r2))
	for i := 0; i < len(r1); i++ {
		for j := 0; j < len(r2); j++ {
			perm = append(perm, append(append(make(api.Row, 0, len(r1[i])+len(r2[j])), r1[i]...), r2[j]...))
		}
	}
	return perm
}

func Merge[E any](convs ...TableConverter[E]) TableConverter[E] {
	if len(convs) == 0 {
		return Table[E](nil)
	}
	if len(convs) == 1 {
		return convs[0]
	}

	return Func[E](func(e E) (*api.Table, error) {
		var (
			headers []api.Header
			rows    []api.Row
		)

		for _, conv := range convs {
			tab, err := conv.ConvertToTable(e)
			if err != nil {
				return nil, err
			}

			if tab == nil {
				continue
			}

			headers = append(headers, tab.Headers...)
			rows = permuteRows(rows, tab.Rows)
		}

		return &api.Table{
			Headers: headers,
			Rows:    rows,
		}, nil
	})
}

type TypedAnyConverter[E any, TC TableConverter[E]] struct {
	Converter TC
}

func (c TypedAnyConverter[E, TC]) ConvertToTable(v any) (*api.Table, error) {
	e, err := generic.Cast[E](v)
	if err != nil {
		return nil, err
	}
	return c.Converter.ConvertToTable(e)
}

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
		return fmt.Errorf("converter for type %s already registered", tag)
	}

	r.convertersByTag[tag] = conv
	return nil
}

func (r *Registry) RegisterTagged(conv TaggedAnyTableConverter) error {
	return r.Register(conv.Tag(), conv)
}

type RegistryBuilder []func(*Registry) error

func (b *RegistryBuilder) RegisterFunc(funcs ...func(*Registry) error) {
	*b = append(*b, funcs...)
}

type TagAndAnyConverter struct {
	Tag       reflect.Type
	Converter TableConverter[any]
}

func ToTagAndTypedAny[E any](conv TableConverter[E]) TagAndAnyConverter {
	return TagAndAnyConverter{
		Tag:       generic.ReflectType[E](),
		Converter: TypedAnyConverter[E, TableConverter[E]]{conv},
	}
}

func (b *RegistryBuilder) Register(taggedConvs ...TagAndAnyConverter) {
	b.RegisterFunc(func(registry *Registry) error {
		for _, taggedConv := range taggedConvs {
			if err := registry.Register(taggedConv.Tag, taggedConv.Converter); err != nil {
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
