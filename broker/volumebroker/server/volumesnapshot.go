// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"fmt"

	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	"github.com/ironcore-dev/ironcore/broker/volumebroker/apiutils"
	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
)

func (s *Server) convertIronCoreVolumeSnapshot(ironcoreVolumeSnapshot *storagev1alpha1.VolumeSnapshot) (*iri.VolumeSnapshot, error) {
	metadata, err := apiutils.GetObjectMetadata(ironcoreVolumeSnapshot)
	if err != nil {
		return nil, fmt.Errorf("error getting object metadata: %w", err)
	}

	state, err := s.convertIronCoreVolumeSnapshotState(ironcoreVolumeSnapshot.Status.State)
	if err != nil {
		return nil, fmt.Errorf("error converting volume snapshot state: %w", err)
	}

	iriVolumeSnapshot := &iri.VolumeSnapshot{
		Metadata: metadata,
		Status: &iri.VolumeSnapshotStatus{
			State: state,
		},
	}

	if ironcoreVolumeSnapshot.Status.Size != nil {
		iriVolumeSnapshot.Status.Size = ironcoreVolumeSnapshot.Status.Size.Value()
	}

	return iriVolumeSnapshot, nil
}

var ironcoreVolumeSnapshotStateToIRIState = map[storagev1alpha1.VolumeSnapshotState]iri.VolumeSnapshotState{
	storagev1alpha1.VolumeSnapshotStatePending: iri.VolumeSnapshotState_VOLUME_SNAPSHOT_PENDING,
	storagev1alpha1.VolumeSnapshotStateReady:   iri.VolumeSnapshotState_VOLUME_SNAPSHOT_READY,
	storagev1alpha1.VolumeSnapshotStateFailed:  iri.VolumeSnapshotState_VOLUME_SNAPSHOT_FAILED,
}

func (s *Server) convertIronCoreVolumeSnapshotState(state storagev1alpha1.VolumeSnapshotState) (iri.VolumeSnapshotState, error) {
	if state, ok := ironcoreVolumeSnapshotStateToIRIState[state]; ok {
		return state, nil
	}
	return 0, fmt.Errorf("unknown ironcore volume snapshot state %q", state)
}
