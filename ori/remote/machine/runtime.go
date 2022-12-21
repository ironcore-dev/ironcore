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

package machine

import (
	"context"
	"fmt"

	"github.com/onmetal/onmetal-api/ori/apis/machine"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type remoteRuntime struct {
	client ori.MachineRuntimeClient
}

func NewRemoteRuntime(endpoint string) (machine.RuntimeService, error) {
	conn, err := grpc.Dial(endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("error dialing: %w", err)
	}

	return &remoteRuntime{
		client: ori.NewMachineRuntimeClient(conn),
	}, nil
}

func (r *remoteRuntime) ListMachines(ctx context.Context, req *ori.ListMachinesRequest) (*ori.ListMachinesResponse, error) {
	return r.client.ListMachines(ctx, req)
}

func (r *remoteRuntime) CreateMachine(ctx context.Context, req *ori.CreateMachineRequest) (*ori.CreateMachineResponse, error) {
	return r.client.CreateMachine(ctx, req)
}

func (r *remoteRuntime) DeleteMachine(ctx context.Context, req *ori.DeleteMachineRequest) (*ori.DeleteMachineResponse, error) {
	return r.client.DeleteMachine(ctx, req)
}

func (r *remoteRuntime) UpdateMachineAnnotations(ctx context.Context, req *ori.UpdateMachineAnnotationsRequest) (*ori.UpdateMachineAnnotationsResponse, error) {
	return r.client.UpdateMachineAnnotations(ctx, req)
}

func (r *remoteRuntime) CreateVolumeAttachment(ctx context.Context, req *ori.CreateVolumeAttachmentRequest) (*ori.CreateVolumeAttachmentResponse, error) {
	return r.client.CreateVolumeAttachment(ctx, req)
}

func (r *remoteRuntime) DeleteVolumeAttachment(ctx context.Context, req *ori.DeleteVolumeAttachmentRequest) (*ori.DeleteVolumeAttachmentResponse, error) {
	return r.client.DeleteVolumeAttachment(ctx, req)
}

func (r *remoteRuntime) CreateNetworkInterfaceAttachment(ctx context.Context, req *ori.CreateNetworkInterfaceAttachmentRequest) (*ori.CreateNetworkInterfaceAttachmentResponse, error) {
	return r.client.CreateNetworkInterfaceAttachment(ctx, req)
}

func (r *remoteRuntime) DeleteNetworkInterfaceAttachment(ctx context.Context, req *ori.DeleteNetworkInterfaceAttachmentRequest) (*ori.DeleteNetworkInterfaceAttachmentResponse, error) {
	return r.client.DeleteNetworkInterfaceAttachment(ctx, req)
}

func (r *remoteRuntime) ListVolumes(ctx context.Context, req *ori.ListVolumesRequest) (*ori.ListVolumesResponse, error) {
	return r.client.ListVolumes(ctx, req)
}

func (r *remoteRuntime) CreateVolume(ctx context.Context, req *ori.CreateVolumeRequest) (*ori.CreateVolumeResponse, error) {
	return r.client.CreateVolume(ctx, req)
}

func (r *remoteRuntime) DeleteVolume(ctx context.Context, req *ori.DeleteVolumeRequest) (*ori.DeleteVolumeResponse, error) {
	return r.client.DeleteVolume(ctx, req)
}

func (r *remoteRuntime) ListNetworkInterfaces(ctx context.Context, req *ori.ListNetworkInterfacesRequest) (*ori.ListNetworkInterfacesResponse, error) {
	return r.client.ListNetworkInterfaces(ctx, req)
}

func (r *remoteRuntime) CreateNetworkInterface(ctx context.Context, req *ori.CreateNetworkInterfaceRequest) (*ori.CreateNetworkInterfaceResponse, error) {
	return r.client.CreateNetworkInterface(ctx, req)
}

func (r *remoteRuntime) DeleteNetworkInterface(ctx context.Context, req *ori.DeleteNetworkInterfaceRequest) (*ori.DeleteNetworkInterfaceResponse, error) {
	return r.client.DeleteNetworkInterface(ctx, req)
}

func (r *remoteRuntime) UpdateNetworkInterfaceIPs(ctx context.Context, req *ori.UpdateNetworkInterfaceIPsRequest) (*ori.UpdateNetworkInterfaceIPsResponse, error) {
	return r.client.UpdateNetworkInterfaceIPs(ctx, req)
}

func (r *remoteRuntime) CreateNetworkInterfaceVirtualIP(ctx context.Context, req *ori.CreateNetworkInterfaceVirtualIPRequest) (*ori.CreateNetworkInterfaceVirtualIPResponse, error) {
	return r.client.CreateNetworkInterfaceVirtualIP(ctx, req)
}

func (r *remoteRuntime) UpdateNetworkInterfaceVirtualIP(ctx context.Context, req *ori.UpdateNetworkInterfaceVirtualIPRequest) (*ori.UpdateNetworkInterfaceVirtualIPResponse, error) {
	return r.client.UpdateNetworkInterfaceVirtualIP(ctx, req)
}

func (r *remoteRuntime) DeleteNetworkInterfaceVirtualIP(ctx context.Context, req *ori.DeleteNetworkInterfaceVirtualIPRequest) (*ori.DeleteNetworkInterfaceVirtualIPResponse, error) {
	return r.client.DeleteNetworkInterfaceVirtualIP(ctx, req)
}

func (r *remoteRuntime) CreateNetworkInterfacePrefix(ctx context.Context, req *ori.CreateNetworkInterfacePrefixRequest) (*ori.CreateNetworkInterfacePrefixResponse, error) {
	return r.client.CreateNetworkInterfacePrefix(ctx, req)
}

func (r *remoteRuntime) DeleteNetworkInterfacePrefix(ctx context.Context, req *ori.DeleteNetworkInterfacePrefixRequest) (*ori.DeleteNetworkInterfacePrefixResponse, error) {
	return r.client.DeleteNetworkInterfacePrefix(ctx, req)
}

func (r *remoteRuntime) CreateNetworkInterfaceLoadBalancerTarget(ctx context.Context, req *ori.CreateNetworkInterfaceLoadBalancerTargetRequest) (*ori.CreateNetworkInterfaceLoadBalancerTargetResponse, error) {
	return r.client.CreateNetworkInterfaceLoadBalancerTarget(ctx, req)
}

func (r *remoteRuntime) DeleteNetworkInterfaceLoadBalancerTarget(ctx context.Context, req *ori.DeleteNetworkInterfaceLoadBalancerTargetRequest) (*ori.DeleteNetworkInterfaceLoadBalancerTargetResponse, error) {
	return r.client.DeleteNetworkInterfaceLoadBalancerTarget(ctx, req)
}

func (r *remoteRuntime) ListMachineClasses(ctx context.Context, req *ori.ListMachineClassesRequest) (*ori.ListMachineClassesResponse, error) {
	return r.client.ListMachineClasses(ctx, req)
}
