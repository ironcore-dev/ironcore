// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"

	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) setIronCoreVolumeResources(ctx context.Context, ironcoreVolume *storagev1alpha1.Volume, resources corev1alpha1.ResourceList) error {
	baseIronCoreVolume := ironcoreVolume.DeepCopy()
	ironcoreVolume.Spec.Resources = resources

	if err := s.client.Patch(ctx, ironcoreVolume, client.MergeFrom(baseIronCoreVolume)); err != nil {
		return fmt.Errorf("error setting resources: %w", err)
	}

	return nil
}

func (s *Server) ExpandVolume(ctx context.Context, req *iri.ExpandVolumeRequest) (*iri.ExpandVolumeResponse, error) {
	volumeID := req.VolumeId
	log := s.loggerFrom(ctx, "VolumeID", volumeID)

	ironcoreVolume, err := s.getAggregateIronCoreVolume(ctx, req.VolumeId)
	if err != nil {
		return nil, err
	}

	log.V(1).Info("Expanding volume")
	if err := s.setIronCoreVolumeResources(ctx, ironcoreVolume.Volume, corev1alpha1.ResourceList{
		corev1alpha1.ResourceStorage: *resource.NewQuantity(int64(req.Resources.StorageBytes), resource.DecimalSI),
	}); err != nil {
		return nil, fmt.Errorf("failed to expand volume: %w", err)
	}

	return &iri.ExpandVolumeResponse{}, nil
}
