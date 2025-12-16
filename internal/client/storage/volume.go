// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"

	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	VolumeSpecVolumeClassRefNameField    = storagev1alpha1.VolumeVolumeClassRefNameField
	VolumeSpecVolumePoolRefNameField     = storagev1alpha1.VolumeVolumePoolRefNameField
	VolumeSpecVolumeSnapshotRefNameField = storagev1alpha1.VolumeVolumeSnapshotRefNameField
)

func SetupVolumeSpecVolumeClassRefNameFieldIndexer(ctx context.Context, indexer client.FieldIndexer) error {
	return indexer.IndexField(ctx, &storagev1alpha1.Volume{}, VolumeSpecVolumeClassRefNameField, func(obj client.Object) []string {
		volume := obj.(*storagev1alpha1.Volume)
		volumeClassRef := volume.Spec.VolumeClassRef
		if volumeClassRef == nil {
			return []string{""}
		}
		return []string{volumeClassRef.Name}
	})
}

func SetupVolumeSpecVolumePoolRefNameFieldIndexer(ctx context.Context, indexer client.FieldIndexer) error {
	return indexer.IndexField(ctx, &storagev1alpha1.Volume{}, VolumeSpecVolumePoolRefNameField, func(obj client.Object) []string {
		volume := obj.(*storagev1alpha1.Volume)
		volumePoolRef := volume.Spec.VolumePoolRef
		if volumePoolRef == nil {
			return []string{""}
		}
		return []string{volumePoolRef.Name}
	})
}

func SetupVolumeSpecVolumeSnapshotRefNameFieldIndexer(ctx context.Context, indexer client.FieldIndexer) error {
	return indexer.IndexField(ctx, &storagev1alpha1.Volume{}, VolumeSpecVolumeSnapshotRefNameField, func(obj client.Object) []string {
		volume := obj.(*storagev1alpha1.Volume)
		if volume.Spec.DataSource.VolumeSnapshotRef != nil {
			return []string{volume.Spec.DataSource.VolumeSnapshotRef.Name}
		}
		return nil
	})
}
