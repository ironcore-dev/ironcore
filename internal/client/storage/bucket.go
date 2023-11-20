// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"

	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	BucketSpecBucketClassRefNameField = storagev1alpha1.BucketBucketClassRefNameField
	BucketSpecBucketPoolRefNameField  = storagev1alpha1.BucketBucketPoolRefNameField
)

func SetupBucketSpecBucketClassRefNameFieldIndexer(ctx context.Context, indexer client.FieldIndexer) error {
	return indexer.IndexField(ctx, &storagev1alpha1.Bucket{}, BucketSpecBucketClassRefNameField, func(obj client.Object) []string {
		bucket := obj.(*storagev1alpha1.Bucket)
		bucketClassRef := bucket.Spec.BucketClassRef
		if bucketClassRef == nil {
			return []string{""}
		}
		return []string{bucketClassRef.Name}
	})
}

func SetupBucketSpecBucketPoolRefNameFieldIndexer(ctx context.Context, indexer client.FieldIndexer) error {
	return indexer.IndexField(ctx, &storagev1alpha1.Bucket{}, BucketSpecBucketPoolRefNameField, func(obj client.Object) []string {
		bucket := obj.(*storagev1alpha1.Bucket)
		bucketPoolRef := bucket.Spec.BucketPoolRef
		if bucketPoolRef == nil {
			return []string{""}
		}
		return []string{bucketPoolRef.Name}
	})
}
