// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package tableconverters

import (
	"time"

	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
	"github.com/ironcore-dev/ironcore/irictl/api"
	"github.com/ironcore-dev/ironcore/irictl/tableconverter"
	"k8s.io/apimachinery/pkg/util/duration"
)

var (
	volumeHeaders = []api.Header{
		{Name: "ID"},
		{Name: "Class"},
		{Name: "Image"},
		{Name: "State"},
		{Name: "Age"},
	}
)

var (
	Volume = tableconverter.Funcs[*iri.Volume]{
		Headers: tableconverter.Headers(volumeHeaders),
		Rows: tableconverter.SingleRowFrom(func(volume *iri.Volume) (api.Row, error) {
			return api.Row{
				volume.Metadata.Id,
				volume.Spec.Class,
				volume.Spec.Image,
				volume.Status.State.String(),
				duration.HumanDuration(time.Since(time.Unix(0, volume.Metadata.CreatedAt))),
			}, nil
		}),
	}
	VolumeSlice = tableconverter.SliceFuncs[*iri.Volume](Volume)
)

func init() {
	RegistryBuilder.Register(
		tableconverter.ToTagAndTypedAny[*iri.Volume](Volume),
		tableconverter.ToTagAndTypedAny[[]*iri.Volume](VolumeSlice),
	)
}
