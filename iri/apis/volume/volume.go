// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package volume

import (
	"context"

	api "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
)

type RuntimeService interface {
	ListEvents(context.Context, *api.ListEventsRequest) (*api.ListEventsResponse, error)
	ListVolumes(context.Context, *api.ListVolumesRequest) (*api.ListVolumesResponse, error)
	CreateVolume(context.Context, *api.CreateVolumeRequest) (*api.CreateVolumeResponse, error)
	ExpandVolume(ctx context.Context, request *api.ExpandVolumeRequest) (*api.ExpandVolumeResponse, error)
	DeleteVolume(context.Context, *api.DeleteVolumeRequest) (*api.DeleteVolumeResponse, error)

	Status(context.Context, *api.StatusRequest) (*api.StatusResponse, error)
}
