// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"

	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func (s *Server) DeleteVolumeSnapshot(ctx context.Context, req *iri.DeleteVolumeSnapshotRequest) (*iri.DeleteVolumeSnapshotResponse, error) {
	volumeSnapshotID := req.VolumeSnapshotId
	log := s.loggerFrom(ctx, "VolumeSnapshotID", volumeSnapshotID)

	ironcoreVolumeSnapshot, err := s.getIronCoreVolumeSnapshot(ctx, req.VolumeSnapshotId)
	if err != nil {
		return nil, err
	}

	log.V(1).Info("Deleting volume snapshot")
	if err := s.client.Delete(ctx, ironcoreVolumeSnapshot); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("error deleting ironcore volume snapshot: %w", err)
		}
		return nil, status.Errorf(codes.NotFound, "volume snapshot %s not found", volumeSnapshotID)
	}

	return &iri.DeleteVolumeSnapshotResponse{}, nil
}
