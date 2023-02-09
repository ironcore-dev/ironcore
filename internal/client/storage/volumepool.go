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

package storage

import (
	"context"

	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
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
