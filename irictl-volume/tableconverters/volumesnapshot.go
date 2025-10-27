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
	volumeSnapshotHeaders = []api.Header{
		{Name: "ID"},
		{Name: "VolumeID"},
		{Name: "State"},
		{Name: "Age"},
	}
)

var (
	VolumeSnapshot = tableconverter.Funcs[*iri.VolumeSnapshot]{
		Headers: tableconverter.Headers(volumeSnapshotHeaders),
		Rows: tableconverter.SingleRowFrom(func(volumeSnapshot *iri.VolumeSnapshot) (api.Row, error) {
			return api.Row{
				volumeSnapshot.Metadata.Id,
				volumeSnapshot.Spec.VolumeId,
				volumeSnapshot.Status.State.String(),
				duration.HumanDuration(time.Since(time.Unix(0, volumeSnapshot.Metadata.CreatedAt))),
			}, nil
		}),
	}
	VolumeSnapshotSlice = tableconverter.SliceFuncs[*iri.VolumeSnapshot](VolumeSnapshot)
)

func init() {
	RegistryBuilder.Register(
		tableconverter.ToTagAndTypedAny[*iri.VolumeSnapshot](VolumeSnapshot),
		tableconverter.ToTagAndTypedAny[[]*iri.VolumeSnapshot](VolumeSnapshotSlice),
	)
}
