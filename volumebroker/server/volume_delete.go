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
	ori "github.com/onmetal/onmetal-api/ori/apis/volume/v1alpha1"
	volumebrokerv1alpha1 "github.com/onmetal/onmetal-api/volumebroker/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) DeleteVolume(ctx context.Context, req *ori.DeleteVolumeRequest) (*ori.DeleteVolumeResponse, error) {
	volumeID := req.VolumeId
	log := s.loggerFrom(ctx, "VolumeID", volumeID)

	var errs []error

	log.V(1).Info("Deleting volume")
	if err := s.client.DeleteAllOf(ctx, &storagev1alpha1.Volume{},
		client.InNamespace(s.namespace),
		client.MatchingLabels{volumebrokerv1alpha1.VolumeIDLabel: volumeID},
	); err != nil {
		errs = append(errs, fmt.Errorf("error deleting volume: %w", err))
	}

	if len(errs) > 0 {
		return &ori.DeleteVolumeResponse{}, fmt.Errorf("error(s) deleting volume: %v", errs)
	}
	return &ori.DeleteVolumeResponse{}, nil
}
