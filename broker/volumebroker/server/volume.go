// Copyright 2022 IronCore authors
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

	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	"github.com/ironcore-dev/ironcore/broker/volumebroker/apiutils"
	ori "github.com/ironcore-dev/ironcore/ori/apis/volume/v1alpha1"
)

func (s *Server) convertAggregateIronCoreVolume(volume *AggregateIronCoreVolume) (*ori.Volume, error) {
	metadata, err := apiutils.GetObjectMetadata(volume.Volume)
	if err != nil {
		return nil, err
	}

	resources, err := s.convertIronCoreVolumeResources(volume.Volume.Spec.Resources)
	if err != nil {
		return nil, err
	}

	state, err := s.convertIronCoreVolumeState(volume.Volume.Status.State)
	if err != nil {
		return nil, err
	}

	access, err := s.convertIronCoreVolumeAccess(volume)
	if err != nil {
		return nil, err
	}

	return &ori.Volume{
		Metadata: metadata,
		Spec: &ori.VolumeSpec{
			Image:      volume.Volume.Spec.Image,
			Class:      volume.Volume.Spec.VolumeClassRef.Name,
			Resources:  resources,
			Encryption: s.convertIronCoreVolumeEncryption(volume),
		},
		Status: &ori.VolumeStatus{
			State:  state,
			Access: access,
		},
	}, nil
}

var ironcoreVolumeStateToORIState = map[storagev1alpha1.VolumeState]ori.VolumeState{
	storagev1alpha1.VolumeStatePending:   ori.VolumeState_VOLUME_PENDING,
	storagev1alpha1.VolumeStateAvailable: ori.VolumeState_VOLUME_AVAILABLE,
	storagev1alpha1.VolumeStateError:     ori.VolumeState_VOLUME_ERROR,
}

func (s *Server) convertIronCoreVolumeState(state storagev1alpha1.VolumeState) (ori.VolumeState, error) {
	if state, ok := ironcoreVolumeStateToORIState[state]; ok {
		return state, nil
	}
	return 0, fmt.Errorf("unknown ironcore volume state %q", state)
}

func (s *Server) convertIronCoreVolumeResources(resources corev1alpha1.ResourceList) (*ori.VolumeResources, error) {
	storage := resources.Storage()
	if storage.IsZero() {
		return nil, fmt.Errorf("volume does not specify storage resource")
	}

	return &ori.VolumeResources{
		StorageBytes: storage.Value(),
	}, nil
}

func (s *Server) convertIronCoreVolumeEncryption(volume *AggregateIronCoreVolume) *ori.EncryptionSpec {
	if volume.EncryptionSecret == nil {
		return nil
	}

	return &ori.EncryptionSpec{
		SecretData: volume.EncryptionSecret.Data,
	}
}

func (s *Server) convertIronCoreVolumeAccess(volume *AggregateIronCoreVolume) (*ori.VolumeAccess, error) {
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
			return nil, fmt.Errorf("access secret specified but not contained in aggregate ironcore volume")
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
