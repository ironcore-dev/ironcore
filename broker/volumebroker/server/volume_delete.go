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

	ori "github.com/onmetal/onmetal-api/ori/apis/volume/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *Server) DeleteVolume(ctx context.Context, req *ori.DeleteVolumeRequest) (*ori.DeleteVolumeResponse, error) {
	volumeID := req.VolumeId
	log := s.loggerFrom(ctx, "VolumeID", volumeID)

	onmetalVolume, err := s.getAggregateOnmetalVolume(ctx, req.VolumeId)
	if err != nil {
		return nil, err
	}

	if encryption := onmetalVolume.Volume.Spec.Encryption; encryption != nil {
		log.V(1).Info("Deleting encryption secret")
		if err := s.client.Delete(ctx, &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
			Name:      encryption.SecretRef.Name,
			Namespace: s.namespace,
		}}); err != nil {
			if !apierrors.IsNotFound(err) {
				return nil, fmt.Errorf("error deleting onmetal encryption secret: %w", err)
			}
			return nil, status.Errorf(codes.NotFound, "secret %s not found", encryption.SecretRef.Name)
		}
	}

	log.V(1).Info("Deleting volume")
	if err := s.client.Delete(ctx, onmetalVolume.Volume); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("error deleting onmetal volume: %w", err)
		}
		return nil, status.Errorf(codes.NotFound, "volume %s not found", volumeID)
	}

	return &ori.DeleteVolumeResponse{}, nil
}
