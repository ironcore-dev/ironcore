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
	ori "github.com/onmetal/onmetal-api/ori/apis/storage/v1alpha1"
	volumebrokerv1alpha1 "github.com/onmetal/onmetal-api/volumebroker/api/v1alpha1"
	"github.com/onmetal/onmetal-api/volumebroker/apiutils"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) convertOnmetalVolume(volume *storagev1alpha1.Volume) (*ori.Volume, error) {
	id := volume.Labels[volumebrokerv1alpha1.VolumeIDLabel]

	metadata, err := apiutils.GetMetadataAnnotation(volume)
	if err != nil {
		return nil, err
	}

	labels, err := apiutils.GetLabelsAnnotation(volume)
	if err != nil {
		return nil, err
	}

	annotations, err := apiutils.GetAnnotationsAnnotation(volume)
	if err != nil {
		return nil, err
	}

	return &ori.Volume{
		Id:          id,
		Metadata:    metadata,
		Annotations: annotations,
		Labels:      labels,
	}, nil
}

var onmetalVolumeStateToORIState = map[storagev1alpha1.VolumeState]ori.VolumeState{
	storagev1alpha1.VolumeStatePending:   ori.VolumeState_VOLUME_PENDING,
	storagev1alpha1.VolumeStateAvailable: ori.VolumeState_VOLUME_AVAILABLE,
	storagev1alpha1.VolumeStateError:     ori.VolumeState_VOLUME_ERROR,
}

func (s *Server) convertOnmetalVolumeState(state storagev1alpha1.VolumeState) ori.VolumeState {
	if state, ok := onmetalVolumeStateToORIState[state]; ok {
		return state
	}
	return ori.VolumeState_VOLUME_PENDING
}

func (s *Server) convertOnmetalVolumeAccess(ctx context.Context, access *storagev1alpha1.VolumeAccess) (*ori.VolumeAccess, error) {
	if access == nil {
		return nil, nil
	}

	var secretData map[string][]byte
	if secretRef := access.SecretRef; secretRef != nil {
		onmetalSecret := &corev1.Secret{}
		onmetalSecretKey := client.ObjectKey{Namespace: s.namespace, Name: secretRef.Name}
		if err := s.client.Get(ctx, onmetalSecretKey, onmetalSecret); err != nil {
			return nil, fmt.Errorf("error getting onmetal secret %s: %w", onmetalSecretKey, err)
		}

		secretData = onmetalSecret.Data
	}

	return &ori.VolumeAccess{
		Driver:     access.Driver,
		Handle:     access.Handle,
		Attributes: access.VolumeAttributes,
		SecretData: secretData,
	}, nil
}
