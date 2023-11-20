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
