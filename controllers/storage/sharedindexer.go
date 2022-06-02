// Copyright 2022 OnMetal authors
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
	"github.com/onmetal/controller-utils/clientutils"
	storagev1alpha1 "github.com/onmetal/onmetal-api/apis/storage/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	VolumeSpecClaimRefNameField = ".spec.claimRef.name"
)

func NewSharedIndexer(mgr manager.Manager) *clientutils.SharedFieldIndexer {
	sharedIndexer := clientutils.NewSharedFieldIndexer(mgr.GetFieldIndexer(), mgr.GetScheme())

	sharedIndexer.MustRegister(&storagev1alpha1.Volume{}, VolumeSpecClaimRefNameField, func(object client.Object) []string {
		volume := object.(*storagev1alpha1.Volume)
		claimRef := volume.Spec.ClaimRef
		if claimRef == nil {
			return []string{""}
		}
		return []string{claimRef.Name}
	})

	return sharedIndexer
}
