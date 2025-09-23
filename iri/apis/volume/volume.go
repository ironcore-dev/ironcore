// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package volume

import (
	"context"

	api "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
)

type RuntimeService interface {
	Version(context.Context, *api.VersionRequest) (*api.VersionResponse, error)
	ListEvents(context.Context, *api.ListEventsRequest) (*api.ListEventsResponse, error)
	ListVolumes(context.Context, *api.ListVolumesRequest) (*api.ListVolumesResponse, error)
	CreateVolume(context.Context, *api.CreateVolumeRequest) (*api.CreateVolumeResponse, error)
	ExpandVolume(context.Context, *api.ExpandVolumeRequest) (*api.ExpandVolumeResponse, error)
	DeleteVolume(context.Context, *api.DeleteVolumeRequest) (*api.DeleteVolumeResponse, error)
	CreateVolumeSnapshot(context.Context, *api.CreateVolumeSnapshotRequest) (*api.CreateVolumeSnapshotResponse, error)
	DeleteVolumeSnapshot(context.Context, *api.DeleteVolumeSnapshotRequest) (*api.DeleteVolumeSnapshotResponse, error)
	ListVolumeSnapshots(context.Context, *api.ListVolumeSnapshotsRequest) (*api.ListVolumeSnapshotsResponse, error)

	Status(context.Context, *api.StatusRequest) (*api.StatusResponse, error)
}
