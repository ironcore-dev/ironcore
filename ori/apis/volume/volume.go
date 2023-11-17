// Copyright 2023 IronCore authors
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

	api "github.com/ironcore-dev/ironcore/ori/apis/volume/v1alpha1"
)

type RuntimeService interface {
	ListVolumes(context.Context, *api.ListVolumesRequest) (*api.ListVolumesResponse, error)
	CreateVolume(context.Context, *api.CreateVolumeRequest) (*api.CreateVolumeResponse, error)
	ExpandVolume(ctx context.Context, request *api.ExpandVolumeRequest) (*api.ExpandVolumeResponse, error)
	DeleteVolume(context.Context, *api.DeleteVolumeRequest) (*api.DeleteVolumeResponse, error)

	Status(context.Context, *api.StatusRequest) (*api.StatusResponse, error)
}
