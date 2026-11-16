// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"

	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	VolumeSnapshotVolumeRefNameField = storagev1alpha1.VolumeSnapshotVolumeRefNameField
)

func SetupVolumeSnapshotVolumeRefNameFieldIndexer(ctx context.Context, indexer client.FieldIndexer) error {
	return indexer.IndexField(ctx, &storagev1alpha1.VolumeSnapshot{}, VolumeSnapshotVolumeRefNameField, func(obj client.Object) []string {
		volumeSnapshot := obj.(*storagev1alpha1.VolumeSnapshot)
		volumeRef := volumeSnapshot.Spec.VolumeRef
		if volumeRef == nil {
			return []string{""}
		}
		return []string{volumeRef.Name}
	})
}
