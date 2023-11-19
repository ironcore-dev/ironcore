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

package tableconverters

import (
	iri "github.com/ironcore-dev/ironcore/iri/apis/bucket/v1alpha1"
	"github.com/ironcore-dev/ironcore/irictl/api"
	"github.com/ironcore-dev/ironcore/irictl/tableconverter"
)

var (
	bucketClassHeaders = []api.Header{
		{Name: "Name"},
		{Name: "TPS"},
		{Name: "IOPS"},
	}
)

var (
	BucketClass = tableconverter.Funcs[*iri.BucketClass]{
		Headers: tableconverter.Headers(bucketClassHeaders),
		Rows: tableconverter.SingleRowFrom(func(class *iri.BucketClass) (api.Row, error) {
			return api.Row{
				class.Name,
				class.Capabilities.Tps,
				class.Capabilities.Iops,
			}, nil
		}),
	}
	BucketClassSlice = tableconverter.SliceFuncs[*iri.BucketClass](BucketClass)
)

func init() {
	RegistryBuilder.Register(
		tableconverter.ToTagAndTypedAny[*iri.BucketClass](BucketClass),
		tableconverter.ToTagAndTypedAny[[]*iri.BucketClass](BucketClassSlice),
	)
}
