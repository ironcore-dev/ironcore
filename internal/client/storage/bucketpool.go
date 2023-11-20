// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"

	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	BucketPoolAvailableBucketClassesField = "bucketpool-available-bucket-classes"
)

func SetupBucketPoolAvailableBucketClassesFieldIndexer(ctx context.Context, indexer client.FieldIndexer) error {
	return indexer.IndexField(ctx, &storagev1alpha1.BucketPool{}, BucketPoolAvailableBucketClassesField, func(object client.Object) []string {
		bucketPool := object.(*storagev1alpha1.BucketPool)

		names := make([]string, 0, len(bucketPool.Status.AvailableBucketClasses))
		for _, availableBucketClass := range bucketPool.Status.AvailableBucketClasses {
			names = append(names, availableBucketClass.Name)
		}

		if len(names) == 0 {
			return []string{""}
		}
		return names
	})
}
