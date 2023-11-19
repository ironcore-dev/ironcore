// Copyright 2022 IronCore authors
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

package machine

import (
	"context"
	"fmt"

	"github.com/ironcore-dev/ironcore/iri/apis/machine"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type remoteRuntime struct {
	client iri.MachineRuntimeClient
}

func NewRemoteRuntime(endpoint string) (machine.RuntimeService, error) {
	conn, err := grpc.Dial(endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("error dialing: %w", err)
	}

	return &remoteRuntime{
		client: iri.NewMachineRuntimeClient(conn),
	}, nil
}

func (r *remoteRuntime) Version(ctx context.Context, req *iri.VersionRequest) (*iri.VersionResponse, error) {
	return r.client.Version(ctx, req)
}

func (r *remoteRuntime) ListMachines(ctx context.Context, req *iri.ListMachinesRequest) (*iri.ListMachinesResponse, error) {
	return r.client.ListMachines(ctx, req)
}

func (r *remoteRuntime) CreateMachine(ctx context.Context, req *iri.CreateMachineRequest) (*iri.CreateMachineResponse, error) {
	return r.client.CreateMachine(ctx, req)
}

func (r *remoteRuntime) DeleteMachine(ctx context.Context, req *iri.DeleteMachineRequest) (*iri.DeleteMachineResponse, error) {
	return r.client.DeleteMachine(ctx, req)
}

func (r *remoteRuntime) UpdateMachineAnnotations(ctx context.Context, req *iri.UpdateMachineAnnotationsRequest) (*iri.UpdateMachineAnnotationsResponse, error) {
	return r.client.UpdateMachineAnnotations(ctx, req)
}

func (r *remoteRuntime) UpdateMachinePower(ctx context.Context, req *iri.UpdateMachinePowerRequest) (*iri.UpdateMachinePowerResponse, error) {
	return r.client.UpdateMachinePower(ctx, req)
}

func (r *remoteRuntime) AttachVolume(ctx context.Context, req *iri.AttachVolumeRequest) (*iri.AttachVolumeResponse, error) {
	return r.client.AttachVolume(ctx, req)
}

func (r *remoteRuntime) DetachVolume(ctx context.Context, req *iri.DetachVolumeRequest) (*iri.DetachVolumeResponse, error) {
	return r.client.DetachVolume(ctx, req)
}

func (r *remoteRuntime) AttachNetworkInterface(ctx context.Context, req *iri.AttachNetworkInterfaceRequest) (*iri.AttachNetworkInterfaceResponse, error) {
	return r.client.AttachNetworkInterface(ctx, req)
}

func (r *remoteRuntime) DetachNetworkInterface(ctx context.Context, req *iri.DetachNetworkInterfaceRequest) (*iri.DetachNetworkInterfaceResponse, error) {
	return r.client.DetachNetworkInterface(ctx, req)
}

func (r *remoteRuntime) Status(ctx context.Context, req *iri.StatusRequest) (*iri.StatusResponse, error) {
	return r.client.Status(ctx, req)
}

func (r *remoteRuntime) Exec(ctx context.Context, req *iri.ExecRequest) (*iri.ExecResponse, error) {
	return r.client.Exec(ctx, req)
}
