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

package server

import (
	"context"
	"fmt"

	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	bucketbrokerv1alpha1 "github.com/onmetal/onmetal-api/broker/bucketbroker/api/v1alpha1"
	"github.com/onmetal/onmetal-api/broker/common"
	ori "github.com/onmetal/onmetal-api/ori/apis/bucket/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) listManagedAndCreated(ctx context.Context, list client.ObjectList) error {
	return s.client.List(ctx, list,
		client.InNamespace(s.namespace),
		client.MatchingLabels{
			bucketbrokerv1alpha1.ManagerLabel: bucketbrokerv1alpha1.BucketBrokerManager,
			bucketbrokerv1alpha1.CreatedLabel: "true",
		},
	)
}

func (s *Server) listAggregateOnmetalBuckets(ctx context.Context) ([]AggregateOnmetalBucket, error) {
	onmetalBucketList := &storagev1alpha1.BucketList{}
	if err := s.listManagedAndCreated(ctx, onmetalBucketList); err != nil {
		return nil, fmt.Errorf("error listing onmetal buckets: %w", err)
	}

	secretList := &corev1.SecretList{}
	if err := s.client.List(ctx, secretList,
		client.InNamespace(s.namespace),
	); err != nil {
		return nil, fmt.Errorf("error listing secrets: %w", err)
	}

	secretByNameGetter, err := common.NewObjectGetter[string, *corev1.Secret](
		corev1.Resource("secrets"),
		common.ByObjectName[*corev1.Secret](),
		common.ObjectSlice[string](secretList.Items),
	)
	if err != nil {
		return nil, fmt.Errorf("error constructing secret getter: %w", err)
	}

	var res []AggregateOnmetalBucket
	for i := range onmetalBucketList.Items {
		onmetalBucket := &onmetalBucketList.Items[i]
		aggregateOnmetalBucket, err := s.aggregateOnmetalBucket(onmetalBucket, secretByNameGetter.Get)
		if err != nil {
			return nil, fmt.Errorf("error aggregating onmetal bucket %s: %w", onmetalBucket.Name, err)
		}

		res = append(res, *aggregateOnmetalBucket)
	}

	return res, nil
}

func (s *Server) clientGetSecretFunc(ctx context.Context) func(string) (*corev1.Secret, error) {
	return func(name string) (*corev1.Secret, error) {
		secret := &corev1.Secret{}
		if err := s.client.Get(ctx, client.ObjectKey{Namespace: s.namespace, Name: name}, secret); err != nil {
			return nil, err
		}
		return secret, nil
	}
}

func (s *Server) getOnmetalBucketAccessSecretIfRequired(
	onmetalBucket *storagev1alpha1.Bucket,
	getSecret func(string) (*corev1.Secret, error),
) (*corev1.Secret, error) {
	if onmetalBucket.Status.State != storagev1alpha1.BucketStateAvailable {
		return nil, nil
	}

	access := onmetalBucket.Status.Access
	if access == nil {
		return nil, nil
	}

	secretRef := access.SecretRef
	if secretRef == nil {
		return nil, nil
	}

	secretName := secretRef.Name
	return getSecret(secretName)
}

func (s *Server) aggregateOnmetalBucket(
	onmetalBucket *storagev1alpha1.Bucket,
	getSecret func(string) (*corev1.Secret, error),
) (*AggregateOnmetalBucket, error) {
	accessSecret, err := s.getOnmetalBucketAccessSecretIfRequired(onmetalBucket, getSecret)
	if err != nil {
		return nil, fmt.Errorf("error getting onmetal bucket access secret: %w", err)
	}

	return &AggregateOnmetalBucket{
		Bucket:       onmetalBucket,
		AccessSecret: accessSecret,
	}, nil
}

func (s *Server) getAggregateOnmetalBucket(ctx context.Context, id string) (*AggregateOnmetalBucket, error) {
	onmetalBucket := &storagev1alpha1.Bucket{}
	if err := s.getManagedAndCreated(ctx, id, onmetalBucket); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("error getting onmetal bucket %s: %w", id, err)
		}
		return nil, status.Errorf(codes.NotFound, "bucket %s not found", id)
	}

	return s.aggregateOnmetalBucket(onmetalBucket, s.clientGetSecretFunc(ctx))
}

func (s *Server) listBuckets(ctx context.Context) ([]*ori.Bucket, error) {
	onmetalBuckets, err := s.listAggregateOnmetalBuckets(ctx)
	if err != nil {
		return nil, fmt.Errorf("error listing buckets: %w", err)
	}

	var res []*ori.Bucket
	for _, onmetalBucket := range onmetalBuckets {
		bucket, err := s.convertAggregateOnmetalBucket(&onmetalBucket)
		if err != nil {
			return nil, err
		}

		res = append(res, bucket)
	}
	return res, nil
}

func (s *Server) filterBuckets(buckets []*ori.Bucket, filter *ori.BucketFilter) []*ori.Bucket {
	if filter == nil {
		return buckets
	}

	var (
		res []*ori.Bucket
		sel = labels.SelectorFromSet(filter.LabelSelector)
	)
	for _, oriBucket := range buckets {
		if !sel.Matches(labels.Set(oriBucket.Metadata.Labels)) {
			continue
		}

		res = append(res, oriBucket)
	}
	return res
}

func (s *Server) getBucket(ctx context.Context, id string) (*ori.Bucket, error) {
	onmetalBucket, err := s.getAggregateOnmetalBucket(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.convertAggregateOnmetalBucket(onmetalBucket)
}

func (s *Server) ListBuckets(ctx context.Context, req *ori.ListBucketsRequest) (*ori.ListBucketsResponse, error) {
	if filter := req.Filter; filter != nil && filter.Id != "" {
		bucket, err := s.getBucket(ctx, filter.Id)
		if err != nil {
			if status.Code(err) != codes.NotFound {
				return nil, err
			}
			return &ori.ListBucketsResponse{
				Buckets: []*ori.Bucket{},
			}, nil
		}

		return &ori.ListBucketsResponse{
			Buckets: []*ori.Bucket{bucket},
		}, nil
	}

	buckets, err := s.listBuckets(ctx)
	if err != nil {
		return nil, err
	}

	buckets = s.filterBuckets(buckets, req.Filter)

	return &ori.ListBucketsResponse{
		Buckets: buckets,
	}, nil
}
