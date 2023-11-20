// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package tableconverters

import (
	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
	"github.com/ironcore-dev/ironcore/irictl/api"
	"github.com/ironcore-dev/ironcore/irictl/tableconverter"
	"k8s.io/apimachinery/pkg/api/resource"
)

var (
	volumeClassHeaders = []api.Header{
		{Name: "Name"},
		{Name: "TPS"},
		{Name: "IOPS"},
		{Name: "Quantity"},
	}
)

var (
	VolumeClassStatus = tableconverter.Funcs[*iri.VolumeClassStatus]{
		Headers: tableconverter.Headers(volumeClassHeaders),
		Rows: tableconverter.SingleRowFrom(func(status *iri.VolumeClassStatus) (api.Row, error) {
			return api.Row{
				status.VolumeClass.Name,
				resource.NewQuantity(status.VolumeClass.Capabilities.Tps, resource.BinarySI).String(),
				resource.NewQuantity(status.VolumeClass.Capabilities.Iops, resource.DecimalSI).String(),
				resource.NewQuantity(status.Quantity, resource.BinarySI).String(),
			}, nil
		}),
	}
	VolumeClassStatusSlice = tableconverter.SliceFuncs[*iri.VolumeClassStatus](VolumeClassStatus)
)

func init() {
	RegistryBuilder.Register(
		tableconverter.ToTagAndTypedAny[*iri.VolumeClassStatus](VolumeClassStatus),
		tableconverter.ToTagAndTypedAny[[]*iri.VolumeClassStatus](VolumeClassStatusSlice),
	)
}
