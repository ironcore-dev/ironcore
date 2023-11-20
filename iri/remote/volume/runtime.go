// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package volume

import (
	"context"
	"fmt"

	"github.com/ironcore-dev/ironcore/iri/apis/volume"
	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type remoteRuntime struct {
	client iri.VolumeRuntimeClient
}

func NewRemoteRuntime(endpoint string) (volume.RuntimeService, error) {
	conn, err := grpc.Dial(endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("error dialing: %w", err)
	}

	return &remoteRuntime{
		client: iri.NewVolumeRuntimeClient(conn),
	}, nil
}

func (r *remoteRuntime) ListVolumes(ctx context.Context, request *iri.ListVolumesRequest) (*iri.ListVolumesResponse, error) {
	return r.client.ListVolumes(ctx, request)
}

func (r *remoteRuntime) CreateVolume(ctx context.Context, request *iri.CreateVolumeRequest) (*iri.CreateVolumeResponse, error) {
	return r.client.CreateVolume(ctx, request)
}

func (r *remoteRuntime) ExpandVolume(ctx context.Context, request *iri.ExpandVolumeRequest) (*iri.ExpandVolumeResponse, error) {
	return r.client.ExpandVolume(ctx, request)
}

func (r *remoteRuntime) DeleteVolume(ctx context.Context, request *iri.DeleteVolumeRequest) (*iri.DeleteVolumeResponse, error) {
	return r.client.DeleteVolume(ctx, request)
}

func (r *remoteRuntime) Status(ctx context.Context, request *iri.StatusRequest) (*iri.StatusResponse, error) {
	return r.client.Status(ctx, request)
}
