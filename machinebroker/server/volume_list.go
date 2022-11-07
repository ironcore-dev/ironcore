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
	"context"
	"errors"
	"fmt"

	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/apis/storage/v1alpha1"
	machinebrokerv1alpha1 "github.com/onmetal/onmetal-api/machinebroker/api/v1alpha1"
	"github.com/onmetal/onmetal-api/machinebroker/apiutils"
	ori "github.com/onmetal/onmetal-api/ori/apis/runtime/v1alpha1"
	"github.com/onmetal/onmetal-api/utils/slices"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type onmetalStorageVolumeFilter struct {
	machineID string
}

func (s *Server) listOnmetalStorageVolumes(ctx context.Context, filter *onmetalStorageVolumeFilter) ([]storagev1alpha1.Volume, error) {
	labels := map[string]string{}
	if filter != nil {
		if filter.machineID != "" {
			labels[machinebrokerv1alpha1.MachineIDLabel] = filter.machineID
		}
	}

	onmetalStorageVolumeList := &storagev1alpha1.VolumeList{}
	if err := s.client.List(ctx, onmetalStorageVolumeList,
		client.InNamespace(s.namespace),
		client.MatchingLabels(labels),
	); err != nil {
		return nil, fmt.Errorf("error listing onmetal storage volumes: %w", err)
	}

	return onmetalStorageVolumeList.Items, nil
}

func (s *Server) getMachineVolumes(ctx context.Context, machineID string) ([]*ori.Volume, error) {
	onmetalMachine, err := s.getOnmetalMachine(ctx, machineID)
	if err != nil {
		return nil, err
	}

	onmetalStorageVolumes, err := s.listOnmetalStorageVolumes(ctx, &onmetalStorageVolumeFilter{machineID: machineID})
	if err != nil {
		return nil, err
	}

	onmetalStorageVolumeByName := slices.ToMap(onmetalStorageVolumes, func(onmetalStorageVolume storagev1alpha1.Volume) string {
		return onmetalStorageVolume.Name
	})

	return s.getOnmetalMachineVolumes(ctx, onmetalMachine, onmetalStorageVolumeByName)
}

func (s *Server) getOnmetalMachineVolumes(
	_ context.Context,
	onmetalMachine *computev1alpha1.Machine,
	onmetalStorageVolumeByName map[string]storagev1alpha1.Volume,
) ([]*ori.Volume, error) {
	machineMetadata, err := apiutils.GetMetadataAnnotation(onmetalMachine)
	if err != nil {
		return nil, err
	}

	var volumes []*ori.Volume
	for _, onmetalVolume := range onmetalMachine.Spec.Volumes {
		var onmetalStorageVolume *storagev1alpha1.Volume
		if onmetalStorageVolumeName := computev1alpha1.MachineVolumeName(onmetalMachine.Name, onmetalVolume); onmetalStorageVolumeName != "" {
			onmetalStorageVolume = &storagev1alpha1.Volume{}
			var ok bool
			*onmetalStorageVolume, ok = onmetalStorageVolumeByName[onmetalStorageVolumeName]
			if !ok {
				return nil, fmt.Errorf("onmetal storage volume %s not found", onmetalStorageVolumeName)
			}
		}

		volume, err := OnmetalVolumeToVolume(onmetalMachine.Name, machineMetadata, &onmetalVolume, onmetalStorageVolume)
		if err != nil {
			return nil, err
		}

		volumes = append(volumes, volume)
	}
	return volumes, nil
}

func (s *Server) listMachineVolumes(ctx context.Context) ([]*ori.Volume, error) {
	onmetalMachines, err := s.listOnmetalMachines(ctx)
	if err != nil {
		return nil, err
	}

	onmetalStorageVolumes, err := s.listOnmetalStorageVolumes(ctx, nil)
	if err != nil {
		return nil, err
	}

	onmetalStorageVolumeByName := slices.ToMap(onmetalStorageVolumes, func(onmetalStorageVolume storagev1alpha1.Volume) string {
		return onmetalStorageVolume.Name
	})

	var volumes []*ori.Volume
	for _, onmetalMachine := range onmetalMachines {
		machineVolumes, err := s.getOnmetalMachineVolumes(ctx, &onmetalMachine, onmetalStorageVolumeByName)
		if err != nil {
			return nil, err
		}

		volumes = append(volumes, machineVolumes...)
	}

	return volumes, nil
}

func (s *Server) ListVolumes(ctx context.Context, req *ori.ListVolumesRequest) (*ori.ListVolumesResponse, error) {
	if filter := req.Filter; filter != nil && filter.MachineId != "" {
		volumes, err := s.getMachineVolumes(ctx, filter.MachineId)
		if err != nil {
			if !errors.Is(err, ErrMachineNotFound) {
				return nil, err
			}
			return &ori.ListVolumesResponse{
				Volumes: []*ori.Volume{},
			}, nil
		}
		return &ori.ListVolumesResponse{
			Volumes: volumes,
		}, nil
	}

	volumes, err := s.listMachineVolumes(ctx)
	if err != nil {
		return nil, err
	}
	return &ori.ListVolumesResponse{
		Volumes: volumes,
	}, nil
}
