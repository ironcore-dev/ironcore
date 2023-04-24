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
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"text/template"

	"github.com/onmetal/onmetal-api/orictl/api"
	"github.com/onmetal/onmetal-api/utils/generic"
	"golang.org/x/exp/slices"
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

func StaticTable[E any](table *api.Table) TableConverter[E] {
	return Func[E](func(E) (*api.Table, error) {
		return table, nil
	})
}

func ZipTables(tables ...*api.Table) (*api.Table, error) {
	if len(tables) == 0 {
		return &api.Table{}, nil
	}
	if len(tables) == 1 {
		return tables[1], nil
	}

	res := &api.Table{
		Headers: slices.Clone(tables[0].Headers),
		Rows:    slices.Clone(tables[0].Rows),
	}

	for _, table := range tables[1:] {
		if len(res.Rows) != len(table.Rows) {
			return nil, fmt.Errorf("row length mismatch: expected %d but got %d", res.Rows, table.Rows)
		}

		res.Headers = append(res.Headers, table.Headers...)
		for i := 0; i < len(res.Rows); i++ {
			res.Rows[i] = append(res.Rows[i], table.Rows[i]...)
		}
	}
	return res, nil
}

func Zip[E any](convs ...TableConverter[E]) TableConverter[E] {
	if len(convs) == 0 {
		return StaticTable[E](&api.Table{})
	}
	if len(convs) == 1 {
		return convs[0]
	}
	return Func[E](func(e E) (*api.Table, error) {
		tab, err := convs[0].ConvertToTable(e)
		if err != nil {
			return nil, err
		}

		for _, conv := range convs[1:] {
			newTab, err := conv.ConvertToTable(e)
			if err != nil {
				return nil, err
			}

			tab, err = ZipTables(tab, newTab)
			if err != nil {
				return nil, err
			}
		}

		return tab, nil
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

type TemplateTableColumn struct {
	Name     string
	Template *template.Template
}

func TemplateTableBuilder[E any](columns ...TemplateTableColumn) Funcs[E] {
	headers := make([]api.Header, len(columns))
	for i, col := range columns {
		headers[i] = api.Header{Name: col.Name}
	}

	return Funcs[E]{
		Headers: Headers(headers),
		Rows: SingleRowFrom(func(e E) (api.Row, error) {
			data, err := json.Marshal(e)
			if err != nil {
				return nil, err
			}

			var jsonV any
			if err := json.Unmarshal(data, &jsonV); err != nil {
				return nil, err
			}

			var sb strings.Builder
			row := make(api.Row, len(columns))
			for i, col := range columns {
				if err := col.Template.Execute(&sb, jsonV); err != nil {
					return nil, fmt.Errorf("[column %s] error executing template: %w", col.Name, err)
				}

				row[i] = sb.String()
				sb.Reset()
			}

			return row, nil
		}),
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
