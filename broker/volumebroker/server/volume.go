// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"fmt"

	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	"github.com/ironcore-dev/ironcore/broker/volumebroker/apiutils"
	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
)

func (s *Server) convertAggregateIronCoreVolume(volume *AggregateIronCoreVolume) (*iri.Volume, error) {
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

	return &iri.Volume{
		Metadata: metadata,
		Spec: &iri.VolumeSpec{
			Image:      volume.Volume.Spec.Image,
			Class:      volume.Volume.Spec.VolumeClassRef.Name,
			Resources:  resources,
			Encryption: s.convertIronCoreVolumeEncryption(volume),
		},
		Status: &iri.VolumeStatus{
			State:  state,
			Access: access,
		},
	}, nil
}

var ironcoreVolumeStateToIRIState = map[storagev1alpha1.VolumeState]iri.VolumeState{
	storagev1alpha1.VolumeStatePending:   iri.VolumeState_VOLUME_PENDING,
	storagev1alpha1.VolumeStateAvailable: iri.VolumeState_VOLUME_AVAILABLE,
	storagev1alpha1.VolumeStateError:     iri.VolumeState_VOLUME_ERROR,
}

func (s *Server) convertIronCoreVolumeState(state storagev1alpha1.VolumeState) (iri.VolumeState, error) {
	if state, ok := ironcoreVolumeStateToIRIState[state]; ok {
		return state, nil
	}
	return 0, fmt.Errorf("unknown ironcore volume state %q", state)
}

func (s *Server) convertIronCoreVolumeResources(resources corev1alpha1.ResourceList) (*iri.VolumeResources, error) {
	storage := resources.Storage()
	if storage.IsZero() {
		return nil, fmt.Errorf("volume does not specify storage resource")
	}

	return &iri.VolumeResources{
		StorageBytes: storage.Value(),
	}, nil
}

func (s *Server) convertIronCoreVolumeEncryption(volume *AggregateIronCoreVolume) *iri.EncryptionSpec {
	if volume.EncryptionSecret == nil {
		return nil
	}

	return &iri.EncryptionSpec{
		SecretData: volume.EncryptionSecret.Data,
	}
}

func (s *Server) convertIronCoreVolumeAccess(volume *AggregateIronCoreVolume) (*iri.VolumeAccess, error) {
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

	return &iri.VolumeAccess{
		Driver:     access.Driver,
		Handle:     access.Handle,
		Attributes: access.VolumeAttributes,
		SecretData: secretData,
	}, nil
}
