// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"

	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) DetachVolume(ctx context.Context, req *iri.DetachVolumeRequest) (*iri.DetachVolumeResponse, error) {
	machineID := req.MachineId
	volumeName := req.Name
	log := s.loggerFrom(ctx, "MachineID", machineID, "VolumeName", volumeName)

	log.V(1).Info("Getting ironcore machine")
	ironcoreMachine, err := s.getIronCoreMachine(ctx, machineID)
	if err != nil {
		return nil, err
	}

	idx := ironcoreMachineVolumeIndex(ironcoreMachine, volumeName)
	if idx < 0 {
		return nil, grpcstatus.Errorf(codes.NotFound, "machine %s volume %s not found", machineID, volumeName)
	}

	ironcoreMachineVolume := ironcoreMachine.Spec.Volumes[idx]

	log.V(1).Info("Patching ironcore machine volumes")
	baseIronCoreMachine := ironcoreMachine.DeepCopy()
	ironcoreMachine.Spec.Volumes = slices.Delete(ironcoreMachine.Spec.Volumes, idx, idx+1)
	if err := s.cluster.Client().Patch(ctx, ironcoreMachine, client.StrategicMergeFrom(baseIronCoreMachine)); err != nil {
		return nil, fmt.Errorf("error patching ironcore machine volumes: %w", err)
	}

	switch {
	case ironcoreMachineVolume.VolumeRef != nil:
		ironcoreVolumeName := ironcoreMachineVolume.VolumeRef.Name
		log = log.WithValues("IronCoreVolumeName", ironcoreVolumeName)
		ironcoreVolume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: s.cluster.Namespace(),
				Name:      ironcoreVolumeName,
			},
		}
		log.V(1).Info("Deleting ironcore volume")
		if err := s.cluster.Client().Delete(ctx, ironcoreVolume); client.IgnoreNotFound(err) != nil {
			return nil, fmt.Errorf("error deleting ironcore volume %s: %w", ironcoreVolumeName, err)
		}
	case ironcoreMachineVolume.EmptyDisk != nil:
		log.V(1).Info("No need to clean up empty disk")
	default:
		return nil, fmt.Errorf("unrecognized ironcore machine volume %#v", ironcoreMachineVolume)
	}

	return &iri.DetachVolumeResponse{}, nil
}
