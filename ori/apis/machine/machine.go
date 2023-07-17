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

	api "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
)

type RuntimeService interface {
	Version(context.Context, *api.VersionRequest) (*api.VersionResponse, error)

	ListMachines(context.Context, *api.ListMachinesRequest) (*api.ListMachinesResponse, error)
	CreateMachine(context.Context, *api.CreateMachineRequest) (*api.CreateMachineResponse, error)
	DeleteMachine(context.Context, *api.DeleteMachineRequest) (*api.DeleteMachineResponse, error)

	UpdateMachineAnnotations(context.Context, *api.UpdateMachineAnnotationsRequest) (*api.UpdateMachineAnnotationsResponse, error)
	UpdateMachinePower(context.Context, *api.UpdateMachinePowerRequest) (*api.UpdateMachinePowerResponse, error)

	CreateVolumeAttachment(context.Context, *api.CreateVolumeAttachmentRequest) (*api.CreateVolumeAttachmentResponse, error)
	DeleteVolumeAttachment(context.Context, *api.DeleteVolumeAttachmentRequest) (*api.DeleteVolumeAttachmentResponse, error)

	CreateNetworkInterfaceAttachment(context.Context, *api.CreateNetworkInterfaceAttachmentRequest) (*api.CreateNetworkInterfaceAttachmentResponse, error)
	DeleteNetworkInterfaceAttachment(context.Context, *api.DeleteNetworkInterfaceAttachmentRequest) (*api.DeleteNetworkInterfaceAttachmentResponse, error)

	ListVolumes(context.Context, *api.ListVolumesRequest) (*api.ListVolumesResponse, error)
	CreateVolume(context.Context, *api.CreateVolumeRequest) (*api.CreateVolumeResponse, error)
	DeleteVolume(context.Context, *api.DeleteVolumeRequest) (*api.DeleteVolumeResponse, error)

	ListNetworkInterfaces(context.Context, *api.ListNetworkInterfacesRequest) (*api.ListNetworkInterfacesResponse, error)
	CreateNetworkInterface(context.Context, *api.CreateNetworkInterfaceRequest) (*api.CreateNetworkInterfaceResponse, error)
	DeleteNetworkInterface(context.Context, *api.DeleteNetworkInterfaceRequest) (*api.DeleteNetworkInterfaceResponse, error)

	UpdateNetworkInterfaceIPs(context.Context, *api.UpdateNetworkInterfaceIPsRequest) (*api.UpdateNetworkInterfaceIPsResponse, error)

	UpdateNetwork(context.Context, *api.UpdateNetworkRequest) (*api.UpdateNetworkResponse, error)

	CreateNetworkInterfaceVirtualIP(context.Context, *api.CreateNetworkInterfaceVirtualIPRequest) (*api.CreateNetworkInterfaceVirtualIPResponse, error)
	UpdateNetworkInterfaceVirtualIP(context.Context, *api.UpdateNetworkInterfaceVirtualIPRequest) (*api.UpdateNetworkInterfaceVirtualIPResponse, error)
	DeleteNetworkInterfaceVirtualIP(context.Context, *api.DeleteNetworkInterfaceVirtualIPRequest) (*api.DeleteNetworkInterfaceVirtualIPResponse, error)

	CreateNetworkInterfacePrefix(context.Context, *api.CreateNetworkInterfacePrefixRequest) (*api.CreateNetworkInterfacePrefixResponse, error)
	DeleteNetworkInterfacePrefix(context.Context, *api.DeleteNetworkInterfacePrefixRequest) (*api.DeleteNetworkInterfacePrefixResponse, error)

	CreateNetworkInterfaceLoadBalancerTarget(context.Context, *api.CreateNetworkInterfaceLoadBalancerTargetRequest) (*api.CreateNetworkInterfaceLoadBalancerTargetResponse, error)
	DeleteNetworkInterfaceLoadBalancerTarget(context.Context, *api.DeleteNetworkInterfaceLoadBalancerTargetRequest) (*api.DeleteNetworkInterfaceLoadBalancerTargetResponse, error)

	CreateNetworkInterfaceNAT(context.Context, *api.CreateNetworkInterfaceNATRequest) (*api.CreateNetworkInterfaceNATResponse, error)
	DeleteNetworkInterfaceNAT(context.Context, *api.DeleteNetworkInterfaceNATRequest) (*api.DeleteNetworkInterfaceNATResponse, error)

	ListMachineClasses(context.Context, *api.ListMachineClassesRequest) (*api.ListMachineClassesResponse, error)

	Exec(context.Context, *api.ExecRequest) (*api.ExecResponse, error)
}
