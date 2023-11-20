// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"

	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	VolumePoolAvailableVolumeClassesField = "volumepool-available-volume-classes"
)

func SetupVolumePoolAvailableVolumeClassesFieldIndexer(ctx context.Context, indexer client.FieldIndexer) error {
	return indexer.IndexField(ctx, &storagev1alpha1.VolumePool{}, VolumePoolAvailableVolumeClassesField, func(object client.Object) []string {
		volumePool := object.(*storagev1alpha1.VolumePool)

		names := make([]string, 0, len(volumePool.Status.AvailableVolumeClasses))
		for _, availableVolumeClass := range volumePool.Status.AvailableVolumeClasses {
			names = append(names, availableVolumeClass.Name)
		}

		if len(names) == 0 {
			return []string{""}
		}
		return names
	})
}
