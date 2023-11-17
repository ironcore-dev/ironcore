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
	"fmt"

	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	"github.com/ironcore-dev/ironcore/internal/apis/storage"
	internalstoragev1alpha1 "github.com/ironcore-dev/ironcore/internal/apis/storage/v1alpha1"
	"github.com/ironcore-dev/ironcore/internal/quota/evaluator/generic"
	"github.com/ironcore-dev/ironcore/utils/quota"
	"golang.org/x/exp/slices"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	volumeResource          = storagev1alpha1.Resource("volumes")
	volumeCountResourceName = corev1alpha1.ObjectCountQuotaResourceNameFor(volumeResource)

	VolumeResourceNames = sets.New(
		volumeCountResourceName,
		corev1alpha1.ResourceRequestsStorage,
	)
)

type volumeEvaluator struct {
	capabilities generic.CapabilitiesReader
}

func NewVolumeEvaluator(capabilities generic.CapabilitiesReader) quota.Evaluator {
	return &volumeEvaluator{
		capabilities: capabilities,
	}
}

func (m *volumeEvaluator) Type() client.Object {
	return &storagev1alpha1.Volume{}
}

func (m *volumeEvaluator) MatchesResourceName(name corev1alpha1.ResourceName) bool {
	return VolumeResourceNames.Has(name)
}

func (m *volumeEvaluator) MatchesResourceScopeSelectorRequirement(item client.Object, req corev1alpha1.ResourceScopeSelectorRequirement) (bool, error) {
	volume := item.(*storagev1alpha1.Volume)

	switch req.ScopeName {
	case corev1alpha1.ResourceScopeVolumeClass:
		return volumeMatchesVolumeClassScope(volume, req.Operator, req.Values), nil
	default:
		return false, nil
	}
}

func volumeMatchesVolumeClassScope(volume *storagev1alpha1.Volume, op corev1alpha1.ResourceScopeSelectorOperator, values []string) bool {
	volumeClassRef := volume.Spec.VolumeClassRef

	switch op {
	case corev1alpha1.ResourceScopeSelectorOperatorExists:
		return volumeClassRef != nil
	case corev1alpha1.ResourceScopeSelectorOperatorDoesNotExist:
		return volumeClassRef == nil
	case corev1alpha1.ResourceScopeSelectorOperatorIn:
		return slices.Contains(values, volumeClassRef.Name)
	case corev1alpha1.ResourceScopeSelectorOperatorNotIn:
		if volumeClassRef == nil {
			return false
		}
		return !slices.Contains(values, volumeClassRef.Name)
	default:
		return false
	}
}

func toExternalVolumeOrError(obj client.Object) (*storagev1alpha1.Volume, error) {
	switch t := obj.(type) {
	case *storagev1alpha1.Volume:
		return t, nil
	case *storage.Volume:
		volume := &storagev1alpha1.Volume{}
		if err := internalstoragev1alpha1.Convert_storage_Volume_To_v1alpha1_Volume(t, volume, nil); err != nil {
			return nil, err
		}
		return volume, nil
	default:
		return nil, fmt.Errorf("expect *storage.Volume or *storagev1alpha1.Volume but got %v", t)
	}
}

func (m *volumeEvaluator) Usage(ctx context.Context, item client.Object) (corev1alpha1.ResourceList, error) {
	volume, err := toExternalVolumeOrError(item)
	if err != nil {
		return nil, err
	}

	return corev1alpha1.ResourceList{
		volumeCountResourceName:              resource.MustParse("1"),
		corev1alpha1.ResourceRequestsStorage: *volume.Spec.Resources.Storage(),
	}, nil
}
