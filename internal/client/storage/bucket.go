// Copyright 2023 IronCore authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
