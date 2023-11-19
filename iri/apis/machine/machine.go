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

	api "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
)

type RuntimeService interface {
	Version(context.Context, *api.VersionRequest) (*api.VersionResponse, error)
	ListMachines(context.Context, *api.ListMachinesRequest) (*api.ListMachinesResponse, error)
	CreateMachine(context.Context, *api.CreateMachineRequest) (*api.CreateMachineResponse, error)
	DeleteMachine(context.Context, *api.DeleteMachineRequest) (*api.DeleteMachineResponse, error)
	UpdateMachineAnnotations(context.Context, *api.UpdateMachineAnnotationsRequest) (*api.UpdateMachineAnnotationsResponse, error)
	UpdateMachinePower(context.Context, *api.UpdateMachinePowerRequest) (*api.UpdateMachinePowerResponse, error)
	AttachVolume(context.Context, *api.AttachVolumeRequest) (*api.AttachVolumeResponse, error)
	DetachVolume(context.Context, *api.DetachVolumeRequest) (*api.DetachVolumeResponse, error)
	AttachNetworkInterface(context.Context, *api.AttachNetworkInterfaceRequest) (*api.AttachNetworkInterfaceResponse, error)
	DetachNetworkInterface(context.Context, *api.DetachNetworkInterfaceRequest) (*api.DetachNetworkInterfaceResponse, error)
	Status(context.Context, *api.StatusRequest) (*api.StatusResponse, error)
	Exec(context.Context, *api.ExecRequest) (*api.ExecResponse, error)
}
