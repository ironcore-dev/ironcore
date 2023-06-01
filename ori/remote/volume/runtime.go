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

package volume

import (
	"context"
	"fmt"

	"github.com/onmetal/onmetal-api/ori/apis/volume"
	ori "github.com/onmetal/onmetal-api/ori/apis/volume/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type remoteRuntime struct {
	client ori.VolumeRuntimeClient
}

func NewRemoteRuntime(endpoint string) (volume.RuntimeService, error) {
	conn, err := grpc.Dial(endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("error dialing: %w", err)
	}

	return &remoteRuntime{
		client: ori.NewVolumeRuntimeClient(conn),
	}, nil
}

func (r *remoteRuntime) ListVolumes(ctx context.Context, request *ori.ListVolumesRequest) (*ori.ListVolumesResponse, error) {
	return r.client.ListVolumes(ctx, request)
}

func (r *remoteRuntime) CreateVolume(ctx context.Context, request *ori.CreateVolumeRequest) (*ori.CreateVolumeResponse, error) {
	return r.client.CreateVolume(ctx, request)
}

func (r *remoteRuntime) ExpandVolume(ctx context.Context, request *ori.ExpandVolumeRequest) (*ori.ExpandVolumeResponse, error) {
	return r.client.ExpandVolume(ctx, request)
}

func (r *remoteRuntime) DeleteVolume(ctx context.Context, request *ori.DeleteVolumeRequest) (*ori.DeleteVolumeResponse, error) {
	return r.client.DeleteVolume(ctx, request)
}

func (r *remoteRuntime) ListVolumeClasses(ctx context.Context, request *ori.ListVolumeClassesRequest) (*ori.ListVolumeClassesResponse, error) {
	return r.client.ListVolumeClasses(ctx, request)
}

func (r *remoteRuntime) ListMetricDescriptors(ctx context.Context, request *ori.ListMetricDescriptorsRequest) (*ori.ListMetricDescriptorsResponse, error) {
	return r.client.ListMetricDescriptors(ctx, request)
}

func (r *remoteRuntime) ListMetrics(ctx context.Context, request *ori.ListMetricsRequest) (*ori.ListMetricsResponse, error) {
	return r.client.ListMetrics(ctx, request)
}
