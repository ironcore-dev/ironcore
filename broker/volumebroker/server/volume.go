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
	"fmt"

	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	"github.com/onmetal/onmetal-api/broker/volumebroker/apiutils"
	ori "github.com/onmetal/onmetal-api/ori/apis/volume/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

func (s *Server) convertAggregateOnmetalVolume(volume *AggregateOnmetalVolume) (*ori.Volume, error) {
	metadata, err := apiutils.GetObjectMetadata(volume.Volume)
	if err != nil {
		return nil, err
	}

	resources, err := s.convertOnmetalVolumeResources(volume.Volume.Spec.Resources)
	if err != nil {
		return nil, err
	}

	state, err := s.convertOnmetalVolumeState(volume.Volume.Status.State)
	if err != nil {
		return nil, err
	}

	access, err := s.convertOnmetalVolumeAccess(volume)
	if err != nil {
		return nil, err
	}

	return &ori.Volume{
		Metadata: metadata,
		Spec: &ori.VolumeSpec{
			Image:     volume.Volume.Spec.Image,
			Class:     volume.Volume.Spec.VolumeClassRef.Name,
			Resources: resources,
		},
		Status: &ori.VolumeStatus{
			State:  state,
			Access: access,
		},
	}, nil
}

var onmetalVolumeStateToORIState = map[storagev1alpha1.VolumeState]ori.VolumeState{
	storagev1alpha1.VolumeStatePending:   ori.VolumeState_VOLUME_PENDING,
	storagev1alpha1.VolumeStateAvailable: ori.VolumeState_VOLUME_AVAILABLE,
	storagev1alpha1.VolumeStateError:     ori.VolumeState_VOLUME_ERROR,
}

func (s *Server) convertOnmetalVolumeState(state storagev1alpha1.VolumeState) (ori.VolumeState, error) {
	if state, ok := onmetalVolumeStateToORIState[state]; ok {
		return state, nil
	}
	return 0, fmt.Errorf("unknown onmetal volume state %q", state)
}

func (s *Server) convertOnmetalVolumeResources(resources corev1.ResourceList) (*ori.VolumeResources, error) {
	storage := resources.Storage()
	if storage.IsZero() {
		return nil, fmt.Errorf("volume does not specify storage resource")
	}

	return &ori.VolumeResources{
		StorageBytes: storage.AsDec().UnscaledBig().Uint64(),
	}, nil
}

func (s *Server) convertOnmetalVolumeAccess(volume *AggregateOnmetalVolume) (*ori.VolumeAccess, error) {
	if volume.Volume.Status.State != storagev1alpha1.VolumeStateAvailable {
		return nil, nil
	}

	access := volume.Volume.Status.Access
	if access == nil {
		return nil, nil
	}

	var secretData map[string][]byte
	if secretRef := access.SecretRef; secretRef != nil {
		if volume.AccessSecret == nil {
			return nil, fmt.Errorf("access secret specified but not contained in aggregate onmetal volume")
		}
		secretData = volume.AccessSecret.Data
	}

	return &ori.VolumeAccess{
		Driver:     access.Driver,
		Handle:     access.Handle,
		Attributes: access.VolumeAttributes,
		SecretData: secretData,
	}, nil
}
