// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

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
