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
	"fmt"

	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	machinebrokerv1alpha1 "github.com/onmetal/onmetal-api/machinebroker/api/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type volumeFilter struct {
	machineID string
	name      string
}

func (s *Server) listOnmetalVolumes(ctx context.Context, filter volumeFilter) ([]OnmetalVolume, error) {
	opts := []client.ListOption{
		client.InNamespace(s.namespace),
	}
	if filter.name != "" || filter.machineID != "" {
		labels := map[string]string{}
		if filter.machineID != "" {
			labels[machinebrokerv1alpha1.MachineIDLabel] = filter.machineID
		}
		if filter.name != "" {
			labels[machinebrokerv1alpha1.VolumeNameLabel] = filter.name
		}

		opts = append(opts, client.MatchingLabels(labels))
	}

	onmetalVolumeList := &storagev1alpha1.VolumeList{}
	if err := s.client.List(ctx, onmetalVolumeList, opts...); err != nil {
		return nil, fmt.Errorf("error listing onmetal storage volumes: %w", err)
	}

	var res []OnmetalVolume
	for _, volume := range onmetalVolumeList.Items {
		volume := volume
		res = append(res, OnmetalVolume{Volume: &volume})
	}

	var onmetalMachines []computev1alpha1.Machine
	if filter.machineID != "" {
		onmetalMachine, err := s.getOnmetalMachine(ctx, filter.machineID)
		if err != nil {
			if status.Code(err) != codes.NotFound {
				return nil, fmt.Errorf("error getting machine %s: %w", filter.machineID, err)
			}
		} else {
			onmetalMachines = []computev1alpha1.Machine{*onmetalMachine}
		}
	} else {
		var err error
		onmetalMachines, err = s.listOnmetalMachines(ctx)
		if err != nil {
			return nil, fmt.Errorf("error listing onmetal machines: %w", err)
		}
	}

	for _, onmetalMachine := range onmetalMachines {
		onmetalMachine := onmetalMachine
		for _, volume := range onmetalMachine.Spec.Volumes {
			if volume.EmptyDisk != nil {
				volume := volume
				res = append(res, OnmetalVolume{
					EmptyDisk: &OnmetalEmptyDiskVolume{
						Machine: &onmetalMachine,
						Volume:  &volume,
					},
				})
			}
		}
	}

	return res, nil
}

func (s *Server) listVolumes(ctx context.Context, filter volumeFilter) ([]*ori.Volume, error) {
	onmetalVolumes, err := s.listOnmetalVolumes(ctx, filter)
	if err != nil {
		return nil, err
	}

	res := make([]*ori.Volume, len(onmetalVolumes))
	for i, onmetalVolume := range onmetalVolumes {
		volume, err := s.convertOnmetalVolume(&onmetalVolume)
		if err != nil {
			return nil, err
		}

		res[i] = volume
	}
	return res, nil
}

func (s *Server) ListVolumes(ctx context.Context, req *ori.ListVolumesRequest) (*ori.ListVolumesResponse, error) {
	if filter := req.Filter; filter != nil && filter.MachineId != "" && filter.Name != "" {
		volume, err := s.getVolume(ctx, filter.MachineId, filter.Name)
		if err != nil {
			if status.Code(err) != codes.NotFound {
				return nil, err
			}
			return &ori.ListVolumesResponse{
				Volumes: []*ori.Volume{},
			}, nil
		}
		return &ori.ListVolumesResponse{
			Volumes: []*ori.Volume{volume},
		}, nil
	}

	volumes, err := s.listVolumes(ctx, volumeFilter{
		machineID: req.GetFilter().GetMachineId(),
		name:      req.GetFilter().GetName(),
	})
	if err != nil {
		return nil, err
	}
	return &ori.ListVolumesResponse{
		Volumes: volumes,
	}, nil
}
