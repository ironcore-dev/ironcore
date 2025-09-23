// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	brokerutils "github.com/ironcore-dev/ironcore/broker/common/utils"
	volumebrokerv1alpha1 "github.com/ironcore-dev/ironcore/broker/volumebroker/api/v1alpha1"
	"github.com/ironcore-dev/ironcore/broker/volumebroker/apiutils"
	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
	volumepoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/volumepoollet/api/v1alpha1"
	utilsmaps "github.com/ironcore-dev/ironcore/utils/maps"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) getIronCoreVolumeSnapshotConfig(ctx context.Context, volumeSnapshot *iri.VolumeSnapshot) (*storagev1alpha1.VolumeSnapshot, error) {
	volumeID := volumeSnapshot.Spec.VolumeId
	if volumeID == "" {
		return nil, status.Errorf(codes.InvalidArgument, "volume ID is required")
	}

	volume := &storagev1alpha1.Volume{}
	if err := s.getManagedAndCreated(ctx, volumeID, volume); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("error getting volume %s: %w", volumeID, err)
		}
		return nil, status.Errorf(codes.NotFound, "volume with ID %s not found", volumeID)
	}

	labels := brokerutils.PrepareDownwardAPILabels(
		volumeSnapshot.Metadata.Labels,
		s.brokerDownwardAPILabels,
		volumepoolletv1alpha1.VolumeSnapshotDownwardAPIPrefix,
	)

	ironcoreVolumeSnapshot := &storagev1alpha1.VolumeSnapshot{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: s.namespace,
			Name:      s.idGen.Generate(),
			Labels: utilsmaps.AppendMap(labels, map[string]string{
				volumebrokerv1alpha1.ManagerLabel: volumebrokerv1alpha1.VolumeBrokerManager,
			}),
			Annotations: volumeSnapshot.Metadata.Annotations,
		},
		Spec: storagev1alpha1.VolumeSnapshotSpec{
			VolumeRef: &corev1.LocalObjectReference{
				Name: volume.Name,
			},
		},
	}

	if err := apiutils.SetObjectMetadata(ironcoreVolumeSnapshot, volumeSnapshot.Metadata); err != nil {
		return nil, err
	}

	return ironcoreVolumeSnapshot, nil
}

func (s *Server) createIronCoreVolumeSnapshot(ctx context.Context, log logr.Logger, volumeSnapshot *storagev1alpha1.VolumeSnapshot) (retErr error) {
	c, cleanup := s.setupCleaner(ctx, log, &retErr)
	defer cleanup()

	log.V(1).Info("Creating ironcore volume snapshot")
	if err := s.client.Create(ctx, volumeSnapshot); err != nil {
		return fmt.Errorf("error creating ironcore volume snapshot: %w", err)
	}
	c.Add(func(ctx context.Context) error {
		if err := s.client.Delete(ctx, volumeSnapshot); client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("error deleting ironcore volume snapshot: %w", err)
		}
		return nil
	})

	log.V(1).Info("Patching ironcore volume snapshot as created")
	if err := apiutils.PatchCreated(ctx, s.client, volumeSnapshot); err != nil {
		return fmt.Errorf("error patching ironcore volume snapshot as created: %w", err)
	}

	// Reset cleaner since everything from now on operates on a consistent volume snapshot
	c.Reset()

	return nil
}

func (s *Server) CreateVolumeSnapshot(ctx context.Context, req *iri.CreateVolumeSnapshotRequest) (res *iri.CreateVolumeSnapshotResponse, retErr error) {
	log := s.loggerFrom(ctx)

	log.V(1).Info("Getting volume snapshot configuration")
	ironcoreVolumeSnapshot, err := s.getIronCoreVolumeSnapshotConfig(ctx, req.VolumeSnapshot)
	if err != nil {
		return nil, fmt.Errorf("error getting ironcore volume snapshot config: %w", err)
	}

	if err := s.createIronCoreVolumeSnapshot(ctx, log, ironcoreVolumeSnapshot); err != nil {
		return nil, fmt.Errorf("error creating ironcore volume snapshot: %w", err)
	}

	v, err := s.convertIronCoreVolumeSnapshot(ctx, ironcoreVolumeSnapshot)
	if err != nil {
		return nil, err
	}

	return &iri.CreateVolumeSnapshotResponse{
		VolumeSnapshot: v,
	}, nil
}
