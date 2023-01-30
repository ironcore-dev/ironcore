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

package bucket

import (
	"time"

	ori "github.com/onmetal/onmetal-api/ori/apis/bucket/v1alpha1"
	"github.com/onmetal/onmetal-api/orictl/table"
	"github.com/onmetal/onmetal-api/orictl/table/tableconverter"
	"k8s.io/apimachinery/pkg/util/duration"
)

var (
	bucketHeaders = []table.Header{
		{Name: "ID"},
		{Name: "Class"},
		{Name: "State"},
		{Name: "Age"},
	}
)

var Bucket, BucketSlice = tableconverter.ForType[*ori.Bucket]( //nolint:revive
	func() ([]table.Header, error) {
		return bucketHeaders, nil
	},
	func(bucket *ori.Bucket) ([]table.Row, error) {
		return []table.Row{
			{
				bucket.Metadata.Id,
				bucket.Spec.Class,
				bucket.Status.State.String(),
				duration.HumanDuration(time.Since(time.Unix(0, bucket.Metadata.CreatedAt))),
			},
		}, nil
	},
)

func init() {
	RegistryBuilder.Register(
		tableconverter.ToTaggedAny(Bucket),
		tableconverter.ToTaggedAny(BucketSlice),
	)
}
