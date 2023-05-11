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

	corev1alpha1 "github.com/onmetal/onmetal-api/api/core/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/volume/v1alpha1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) setOnmetalVolumeResources(ctx context.Context, onmetalVolume *storagev1alpha1.Volume, resources corev1alpha1.ResourceList) error {
	baseOnmetalVolume := onmetalVolume.DeepCopy()
	onmetalVolume.Spec.Resources = resources

	if err := s.client.Patch(ctx, onmetalVolume, client.MergeFrom(baseOnmetalVolume)); err != nil {
		return fmt.Errorf("error setting resources: %w", err)
	}

	return nil
}

func (s *Server) ExpandVolume(ctx context.Context, req *ori.ExpandVolumeRequest) (*ori.ExpandVolumeResponse, error) {
	volumeID := req.VolumeId
	log := s.loggerFrom(ctx, "VolumeID", volumeID)

	onmetalVolume, err := s.getAggregateOnmetalVolume(ctx, req.VolumeId)
	if err != nil {
		return nil, err
	}

	log.V(1).Info("Expanding volume")
	if err := s.setOnmetalVolumeResources(ctx, onmetalVolume.Volume, corev1alpha1.ResourceList{
		corev1alpha1.ResourceStorage: *resource.NewQuantity(int64(req.Resources.StorageBytes), resource.DecimalSI),
	}); err != nil {
		return nil, fmt.Errorf("failed to expand volume: %w", err)
	}

	return &ori.ExpandVolumeResponse{}, nil
}
