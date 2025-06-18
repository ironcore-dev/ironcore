// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"

	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) UpdateVolume(ctx context.Context, req *iri.UpdateVolumeRequest) (*iri.UpdateVolumeResponse, error) {
	machineID := req.MachineId
	volume := req.Volume
	log := s.loggerFrom(ctx, "MachineID", machineID, "VolumeName", volume.Name)

	log.V(1).Info("Getting ironcore machine")
	ironcoreMachine, err := s.getIronCoreMachine(ctx, machineID)
	if err != nil {
		return nil, err
	}

	volumeIndex := ironcoreMachineVolumeIndex(ironcoreMachine, volume.Name)
	if volumeIndex < 0 {
		return nil, status.Errorf(codes.NotFound, "machine %s volume %s not found", machineID, volume.Name)
	}

	base := ironcoreMachine.DeepCopy()
	ironcoreMachine.Spec.Volumes[volumeIndex].Device = ptr.To(volume.Device)

	if volume.Connection != nil {
		volumeRef := ironcoreMachine.Spec.Volumes[volumeIndex].VolumeRef
		if volumeRef == nil {
			return nil, status.Errorf(codes.InvalidArgument, "machine %s volume %s is not a referenced volume", machineID, volume.Name)
		}

		log.V(1).Info("Getting ironcore volume")
		ironcoreVolume := &storagev1alpha1.Volume{}
		if err := s.cluster.Client().Get(ctx, client.ObjectKey{Name: volumeRef.Name, Namespace: ironcoreMachine.Namespace}, ironcoreVolume); err != nil {
			return nil, fmt.Errorf("error getting ironcore volume: %w", err)
		}

		baseVolume := ironcoreVolume.DeepCopy()
		if ironcoreVolume.Status.Resources == nil {
			ironcoreVolume.Status.Resources = corev1alpha1.ResourceList{}
		}
		ironcoreVolume.Status.Resources[corev1alpha1.ResourceStorage] = *resource.NewQuantity(volume.Connection.EffectiveStorageBytes, resource.BinarySI)

		if ironcoreVolume.Status.Access == nil {
			ironcoreVolume.Status.Access = &storagev1alpha1.VolumeAccess{}
		}
		ironcoreVolume.Status.Access.Driver = volume.Connection.Driver
		ironcoreVolume.Status.Access.Handle = volume.Connection.Handle
		ironcoreVolume.Status.Access.VolumeAttributes = volume.Connection.Attributes

		log.V(1).Info("Patching ironcore volume status")
		if err := s.cluster.Client().Status().Patch(ctx, ironcoreVolume, client.MergeFrom(baseVolume)); err != nil {
			return nil, fmt.Errorf("error updating volume status: %w", err)
		}
	}

	log.V(1).Info("Patching ironcore machine volume")
	if err := s.cluster.Client().Patch(ctx, ironcoreMachine, client.MergeFrom(base)); err != nil {
		return nil, fmt.Errorf("error updating machine volume: %w", err)
	}

	return &iri.UpdateVolumeResponse{}, nil
}
