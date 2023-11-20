// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package tableconverter_test

import (
	"github.com/ironcore-dev/ironcore/irictl/api"
	. "github.com/ironcore-dev/ironcore/irictl/tableconverter"
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
