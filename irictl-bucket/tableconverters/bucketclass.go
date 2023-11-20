// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

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
