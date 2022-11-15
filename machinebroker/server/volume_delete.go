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

	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/apis/storage/v1alpha1"
	machinebrokerv1alpha1 "github.com/onmetal/onmetal-api/machinebroker/api/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/compute/v1alpha1"
	"golang.org/x/exp/slices"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) DeleteVolume(ctx context.Context, req *ori.DeleteVolumeRequest) (*ori.DeleteVolumeResponse, error) {
	machineID := req.MachineId
	volumeName := req.VolumeName
	log := s.loggerFrom(ctx, "MachineID", machineID, "VolumeName", volumeName)

	log.V(1).Info("Getting machine")
	onmetalMachine, err := s.getOnmetalMachine(ctx, machineID)
	if err != nil {
		return nil, err
	}

	idx := slices.IndexFunc(onmetalMachine.Spec.Volumes,
		func(volume computev1alpha1.Volume) bool {
			return volume.Name == volumeName
		},
	)
	if idx < 0 {
		log.V(1).Info("Volume not present in machine")
		return &ori.DeleteVolumeResponse{}, nil
	}

	log.V(1).Info("Deleting volume from machine")
	baseOnmetalMachine := onmetalMachine.DeepCopy()
	onmetalMachine.Spec.Volumes = slices.Delete(onmetalMachine.Spec.Volumes, idx, idx+1)
	if err := s.client.Patch(ctx, onmetalMachine, client.MergeFrom(baseOnmetalMachine)); err != nil {
		return nil, fmt.Errorf("error deleting volume from machine: %w", err)
	}

	log.V(1).Info("Deleting volume")
	if err := s.client.DeleteAllOf(ctx, &storagev1alpha1.Volume{},
		client.InNamespace(s.namespace),
		client.MatchingLabels{
			machinebrokerv1alpha1.MachineIDLabel:  machineID,
			machinebrokerv1alpha1.VolumeNameLabel: volumeName,
		},
	); err != nil {
		return nil, fmt.Errorf("error deleting volume: %w", err)
	}

	log.V(1).Info("Deleting secret")
	if err := s.client.DeleteAllOf(ctx, &corev1.Secret{},
		client.InNamespace(s.namespace),
		client.MatchingLabels{
			machinebrokerv1alpha1.MachineIDLabel:  machineID,
			machinebrokerv1alpha1.VolumeNameLabel: volumeName,
		},
	); err != nil {
		return nil, fmt.Errorf("error deleting secret: %w", err)
	}

	return &ori.DeleteVolumeResponse{}, nil
}
