// Copyright 2023 OnMetal authors
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

package client

import (
	"context"

	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	VolumeSpecVolumeClassRefNameField = storagev1alpha1.VolumeVolumeClassRefNameField
	BucketSpecBucketClassRefNameField = storagev1alpha1.BucketBucketClassRefNameField
)

func SetupVolumeSpecVolumeClassRefNameFieldIndexer(ctx context.Context, indexer client.FieldIndexer) error {
	return indexer.IndexField(ctx, &storagev1alpha1.Volume{}, VolumeSpecVolumeClassRefNameField, func(obj client.Object) []string {
		volume := obj.(*storagev1alpha1.Volume)
		volumeClassRef := volume.Spec.VolumeClassRef
		if volumeClassRef == nil {
			return nil
		}
		return []string{volumeClassRef.Name}
	})
}

func SetupBucketSpecBucketClassRefNameFieldIndexer(ctx context.Context, indexer client.FieldIndexer) error {
	return indexer.IndexField(ctx, &storagev1alpha1.Bucket{}, BucketSpecBucketClassRefNameField, func(obj client.Object) []string {
		bucket := obj.(*storagev1alpha1.Bucket)
		bucketClassRef := bucket.Spec.BucketClassRef
		if bucketClassRef == nil {
			return nil
		}
		return []string{bucketClassRef.Name}
	})
}
