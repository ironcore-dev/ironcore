// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package tableconverter

import (
	"sort"

	irimeta "github.com/ironcore-dev/ironcore/iri/apis/meta/v1alpha1"
	"github.com/ironcore-dev/ironcore/irictl/api"
)

func WellKnownLabels[E irimeta.Object](labels map[string]string) Funcs[E] {
	headers := make([]api.Header, 0, len(labels))
	for name := range labels {
		headers = append(headers, api.Header{Name: name})
	}
	sort.Slice(headers, func(i, j int) bool {
		return headers[i].Name < headers[j].Name
	})

	return Funcs[E]{
		Headers: Headers(headers),
		Rows: SingleRowFrom(func(e E) (api.Row, error) {
			row := make(api.Row, len(headers))

			objLabels := e.GetMetadata().GetLabels()
			for i := range headers {
				row[i] = objLabels[labels[headers[i].Name]]
			}

			return row, nil
		}),
	}
}
