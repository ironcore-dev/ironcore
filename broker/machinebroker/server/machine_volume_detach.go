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

	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) DetachVolume(ctx context.Context, req *ori.DetachVolumeRequest) (*ori.DetachVolumeResponse, error) {
	machineID := req.MachineId
	volumeName := req.Name
	log := s.loggerFrom(ctx, "MachineID", machineID, "VolumeName", volumeName)

	log.V(1).Info("Getting onmetal machine")
	onmetalMachine, err := s.getOnmetalMachine(ctx, machineID)
	if err != nil {
		return nil, err
	}

	idx := onmetalMachineVolumeIndex(onmetalMachine, volumeName)
	if idx < 0 {
		return nil, grpcstatus.Errorf(codes.NotFound, "machine %s volume %s not found", machineID, volumeName)
	}

	onmetalMachineVolume := onmetalMachine.Spec.Volumes[idx]

	log.V(1).Info("Patching onmetal machine volumes")
	baseOnmetalMachine := onmetalMachine.DeepCopy()
	onmetalMachine.Spec.Volumes = slices.Delete(onmetalMachine.Spec.Volumes, idx, idx+1)
	if err := s.cluster.Client().Patch(ctx, onmetalMachine, client.StrategicMergeFrom(baseOnmetalMachine)); err != nil {
		return nil, fmt.Errorf("error patching onmetal machine volumes: %w", err)
	}

	switch {
	case onmetalMachineVolume.VolumeRef != nil:
		onmetalVolumeName := onmetalMachineVolume.VolumeRef.Name
		log = log.WithValues("OnmetalVolumeName", onmetalVolumeName)
		onmetalVolume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: s.cluster.Namespace(),
				Name:      onmetalVolumeName,
			},
		}
		log.V(1).Info("Deleting onmetal volume")
		if err := s.cluster.Client().Delete(ctx, onmetalVolume); client.IgnoreNotFound(err) != nil {
			return nil, fmt.Errorf("error deleting onmetal volume %s: %w", onmetalVolumeName, err)
		}
	case onmetalMachineVolume.EmptyDisk != nil:
		log.V(1).Info("No need to cleanujp empty disk")
	default:
		return nil, fmt.Errorf("unrecognized onmetal machine volume %#v", onmetalMachineVolume)
	}

	return &ori.DetachVolumeResponse{}, nil
}
