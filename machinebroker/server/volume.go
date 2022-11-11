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

package server

import (
	"fmt"

	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/apis/storage/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/runtime/v1alpha1"
)

func (s *Server) convertOnmetalVolume(
	machineID string,
	machineMetadata *ori.MachineMetadata,
	onmetalVolume *computev1alpha1.Volume,
	onmetalStorageVolume *storagev1alpha1.Volume,
) (*ori.Volume, error) {
	var access *ori.VolumeAccess
	if onmetalVolume.VolumeRef != nil || onmetalVolume.Ephemeral != nil {
		onmetalVolumeAccess := onmetalStorageVolume.Status.Access
		if onmetalVolumeAccess == nil {
			return nil, fmt.Errorf("onmetal volume %s/%s does not specify access", onmetalStorageVolume.Namespace, onmetalStorageVolume.Name)
		}

		access = &ori.VolumeAccess{
			Driver: onmetalVolumeAccess.Driver,
			Handle: onmetalVolumeAccess.Handle,
		}
	}

	var emptyDisk *ori.EmptyDisk
	if onmetalEmptyDisk := onmetalVolume.EmptyDisk; onmetalEmptyDisk != nil {
		var sizeLimitBytes uint64
		if sizeLimit := onmetalEmptyDisk.SizeLimit; sizeLimit != nil {
			sizeLimitBytes = uint64(sizeLimit.Value())
		}

		emptyDisk = &ori.EmptyDisk{
			SizeLimitBytes: sizeLimitBytes,
		}
	}

	return &ori.Volume{
		MachineId:       machineID,
		MachineMetadata: machineMetadata,
		Name:            onmetalVolume.Name,
		Device:          onmetalVolume.Device,
		Access:          access,
		EmptyDisk:       emptyDisk,
	}, nil
}
