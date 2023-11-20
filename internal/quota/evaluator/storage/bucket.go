// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

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
	bucketResource          = storagev1alpha1.Resource("buckets")
	bucketCountResourceName = corev1alpha1.ObjectCountQuotaResourceNameFor(bucketResource)

	BucketResourceNames = sets.New(
		bucketCountResourceName,
		corev1alpha1.ResourceRequestsStorage,
	)
)

type bucketEvaluator struct {
	capabilities generic.CapabilitiesReader
}

func NewBucketEvaluator(capabilities generic.CapabilitiesReader) quota.Evaluator {
	return &bucketEvaluator{
		capabilities: capabilities,
	}
}

func (m *bucketEvaluator) Type() client.Object {
	return &storagev1alpha1.Bucket{}
}

func (m *bucketEvaluator) MatchesResourceName(name corev1alpha1.ResourceName) bool {
	return BucketResourceNames.Has(name)
}

func (m *bucketEvaluator) MatchesResourceScopeSelectorRequirement(item client.Object, req corev1alpha1.ResourceScopeSelectorRequirement) (bool, error) {
	bucket := item.(*storagev1alpha1.Bucket)

	switch req.ScopeName {
	case corev1alpha1.ResourceScopeBucketClass:
		return bucketMatchesBucketClassScope(bucket, req.Operator, req.Values), nil
	default:
		return false, nil
	}
}

func bucketMatchesBucketClassScope(bucket *storagev1alpha1.Bucket, op corev1alpha1.ResourceScopeSelectorOperator, values []string) bool {
	bucketClassRef := bucket.Spec.BucketClassRef

	switch op {
	case corev1alpha1.ResourceScopeSelectorOperatorExists:
		return bucketClassRef != nil
	case corev1alpha1.ResourceScopeSelectorOperatorDoesNotExist:
		return bucketClassRef == nil
	case corev1alpha1.ResourceScopeSelectorOperatorIn:
		return slices.Contains(values, bucketClassRef.Name)
	case corev1alpha1.ResourceScopeSelectorOperatorNotIn:
		if bucketClassRef == nil {
			return false
		}
		return !slices.Contains(values, bucketClassRef.Name)
	default:
		return false
	}
}

func toExternalBucketOrError(obj client.Object) (*storagev1alpha1.Bucket, error) {
	switch t := obj.(type) {
	case *storagev1alpha1.Bucket:
		return t, nil
	case *storage.Bucket:
		bucket := &storagev1alpha1.Bucket{}
		if err := internalstoragev1alpha1.Convert_storage_Bucket_To_v1alpha1_Bucket(t, bucket, nil); err != nil {
			return nil, err
		}
		return bucket, nil
	default:
		return nil, fmt.Errorf("expect *storage.Bucket or *storagev1alpha1.Bucket but got %v", t)
	}
}

func (m *bucketEvaluator) Usage(ctx context.Context, item client.Object) (corev1alpha1.ResourceList, error) {
	_, err := toExternalBucketOrError(item)
	if err != nil {
		return nil, err
	}

	return corev1alpha1.ResourceList{
		// TODO: return more detailed usage
		bucketCountResourceName: resource.MustParse("1"),
	}, nil
}
