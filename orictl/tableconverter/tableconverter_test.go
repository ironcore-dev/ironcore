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

package tableconverter_test

import (
	"github.com/onmetal/onmetal-api/orictl/api"
	. "github.com/onmetal/onmetal-api/orictl/tableconverter"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Tableconverter", func() {
	Describe("Zip", func() {
		It("should zip the two table converters", func() {
			h1 := []api.Header{{Name: "foo"}}
			f1 := Funcs[int]{
				Headers: Headers(h1),
				Rows: func(n int) ([]api.Row, error) {
					return []api.Row{{n}, {n + 1}}, nil
				},
			}

			h2 := []api.Header{{Name: "bar"}}
			f2 := Funcs[int]{
				Headers: Headers(h2),
				Rows: func(n int) ([]api.Row, error) {
					return []api.Row{{n}, {n * n}}, nil
				},
			}

			f := Zip[int](f1, f2)
			table, err := f.ConvertToTable(2)
			Expect(err).NotTo(HaveOccurred())
			Expect(table).To(Equal(&api.Table{
				Headers: []api.Header{{Name: "foo"}, {Name: "bar"}},
				Rows: []api.Row{
					{2, 2},
					{3, 4},
				},
			}))
		})
	})
})
