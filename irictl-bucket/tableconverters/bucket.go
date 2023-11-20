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
	"time"

	iri "github.com/ironcore-dev/ironcore/iri/apis/bucket/v1alpha1"
	"github.com/ironcore-dev/ironcore/irictl/api"
	"github.com/ironcore-dev/ironcore/irictl/tableconverter"
	"k8s.io/apimachinery/pkg/util/duration"
)

var (
	bucketHeaders = []api.Header{
		{Name: "ID"},
		{Name: "Class"},
		{Name: "State"},
		{Name: "Age"},
	}
)

var (
	Bucket = tableconverter.Funcs[*iri.Bucket]{
		Headers: tableconverter.Headers(bucketHeaders),
		Rows: tableconverter.SingleRowFrom(func(bucket *iri.Bucket) (api.Row, error) {
			return api.Row{
				bucket.Metadata.Id,
				bucket.Spec.Class,
				bucket.Status.State.String(),
				duration.HumanDuration(time.Since(time.Unix(0, bucket.Metadata.CreatedAt))),
			}, nil
		}),
	}
	BucketSlice = tableconverter.SliceFuncs[*iri.Bucket](Bucket)
)

func init() {
	RegistryBuilder.Register(
		tableconverter.ToTagAndTypedAny[*iri.Bucket](Bucket),
		tableconverter.ToTagAndTypedAny[[]*iri.Bucket](BucketSlice),
	)
}
