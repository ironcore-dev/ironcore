// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"

	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	volumebrokerv1alpha1 "github.com/ironcore-dev/ironcore/broker/volumebroker/api/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) listManagedAndCreatedVolumeSnapshots(ctx context.Context, list client.ObjectList) error {
	return s.client.List(ctx, list,
		client.InNamespace(s.namespace),
		client.MatchingLabels{
			volumebrokerv1alpha1.ManagerLabel: volumebrokerv1alpha1.VolumeBrokerManager,
			volumebrokerv1alpha1.CreatedLabel: "true",
		},
	)
}

func (s *Server) listIronCoreVolumeSnapshots(ctx context.Context) ([]*storagev1alpha1.VolumeSnapshot, error) {
	ironcoreVolumeSnapshotList := &storagev1alpha1.VolumeSnapshotList{}
	if err := s.listManagedAndCreatedVolumeSnapshots(ctx, ironcoreVolumeSnapshotList); err != nil {
		return nil, fmt.Errorf("error listing ironcore volume snapshots: %w", err)
	}

	var res []*storagev1alpha1.VolumeSnapshot
	for i := range ironcoreVolumeSnapshotList.Items {
		res = append(res, &ironcoreVolumeSnapshotList.Items[i])
	}

	return res, nil
}

func (s *Server) filterVolumeSnapshots(volumeSnapshots []*iri.VolumeSnapshot, filter *iri.VolumeSnapshotFilter) []*iri.VolumeSnapshot {
	if filter == nil {
		return volumeSnapshots
	}

	var (
		res []*iri.VolumeSnapshot
		sel = labels.SelectorFromSet(filter.LabelSelector)
	)
	for _, iriVolumeSnapshot := range volumeSnapshots {
		if !sel.Matches(labels.Set(iriVolumeSnapshot.Metadata.Labels)) {
			continue
		}

		res = append(res, iriVolumeSnapshot)
	}

	return res
}

func (s *Server) getIronCoreVolumeSnapshot(ctx context.Context, id string) (*storagev1alpha1.VolumeSnapshot, error) {
	ironcoreVolumeSnapshot := &storagev1alpha1.VolumeSnapshot{}
	if err := s.getManagedAndCreated(ctx, id, ironcoreVolumeSnapshot); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("error getting ironcore volume snapshot %s: %w", id, err)
		}
		return nil, status.Errorf(codes.NotFound, "volume snapshot %s not found", id)
	}

	return ironcoreVolumeSnapshot, nil
}

func (s *Server) listVolumeSnapshots(ctx context.Context) ([]*iri.VolumeSnapshot, error) {
	ironcoreVolumeSnapshots, err := s.listIronCoreVolumeSnapshots(ctx)
	if err != nil {
		return nil, fmt.Errorf("error listing volume snapshots: %w", err)
	}

	var res []*iri.VolumeSnapshot
	for _, ironcoreVolumeSnapshot := range ironcoreVolumeSnapshots {
		volumeSnapshot, err := s.convertIronCoreVolumeSnapshot(ctx, ironcoreVolumeSnapshot)
		if err != nil {
			return nil, err
		}

		res = append(res, volumeSnapshot)
	}
	return res, nil
}

func (s *Server) getVolumeSnapshot(ctx context.Context, id string) (*iri.VolumeSnapshot, error) {
	ironcoreVolumeSnapshot, err := s.getIronCoreVolumeSnapshot(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.convertIronCoreVolumeSnapshot(ctx, ironcoreVolumeSnapshot)
}

func (s *Server) ListVolumeSnapshots(ctx context.Context, req *iri.ListVolumeSnapshotsRequest) (*iri.ListVolumeSnapshotsResponse, error) {
	if filter := req.Filter; filter != nil && filter.Id != "" {
		volumeSnapshot, err := s.getVolumeSnapshot(ctx, filter.Id)
		if err != nil {
			if status.Code(err) != codes.NotFound {
				return nil, err
			}
			return &iri.ListVolumeSnapshotsResponse{
				VolumeSnapshots: []*iri.VolumeSnapshot{},
			}, nil
		}

		return &iri.ListVolumeSnapshotsResponse{
			VolumeSnapshots: []*iri.VolumeSnapshot{volumeSnapshot},
		}, nil
	}

	volumeSnapshots, err := s.listVolumeSnapshots(ctx)
	if err != nil {
		return nil, err
	}

	volumeSnapshots = s.filterVolumeSnapshots(volumeSnapshots, req.Filter)

	return &iri.ListVolumeSnapshotsResponse{
		VolumeSnapshots: volumeSnapshots,
	}, nil
}
