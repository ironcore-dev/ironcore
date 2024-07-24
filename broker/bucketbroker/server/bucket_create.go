// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	bucketbrokerv1alpha1 "github.com/ironcore-dev/ironcore/broker/bucketbroker/api/v1alpha1"
	"github.com/ironcore-dev/ironcore/broker/bucketbroker/apiutils"
	iri "github.com/ironcore-dev/ironcore/iri/apis/bucket/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AggregateIronCoreBucket struct {
	Bucket       *storagev1alpha1.Bucket
	AccessSecret *corev1.Secret
}

func (s *Server) getIronCoreBucketConfig(_ context.Context, bucket *iri.Bucket) (*AggregateIronCoreBucket, error) {
	var bucketPoolRef *corev1.LocalObjectReference
	if s.bucketPoolName != "" {
		bucketPoolRef = &corev1.LocalObjectReference{
			Name: s.bucketPoolName,
		}
	}
	ironcoreBucket := &storagev1alpha1.Bucket{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: s.namespace,
			Name:      s.generateID(),
		},
		Spec: storagev1alpha1.BucketSpec{
			BucketClassRef:     &corev1.LocalObjectReference{Name: bucket.Spec.Class},
			BucketPoolRef:      bucketPoolRef,
			BucketPoolSelector: s.bucketPoolSelector,
		},
	}
	if err := apiutils.SetObjectMetadata(ironcoreBucket, bucket.Metadata); err != nil {
		return nil, err
	}
	apiutils.SetBucketManagerLabel(ironcoreBucket, bucketbrokerv1alpha1.BucketBrokerManager)

	return &AggregateIronCoreBucket{
		Bucket: ironcoreBucket,
	}, nil
}

func (s *Server) createIronCoreBucket(ctx context.Context, log logr.Logger, bucket *AggregateIronCoreBucket) (retErr error) {
	c, cleanup := s.setupCleaner(ctx, log, &retErr)
	defer cleanup()

	log.V(1).Info("Creating ironcore bucket")
	if err := s.client.Create(ctx, bucket.Bucket); err != nil {
		return fmt.Errorf("error creating ironcore bucket: %w", err)
	}
	c.Add(func(ctx context.Context) error {
		if err := s.client.Delete(ctx, bucket.Bucket); client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("error deleting ironcore bucket: %w", err)
		}
		return nil
	})

	log.V(1).Info("Patching ironcore bucket as created")
	if err := apiutils.PatchCreated(ctx, s.client, bucket.Bucket); err != nil {
		return fmt.Errorf("error patching ironcore bucket as created: %w", err)
	}

	// Reset cleaner since everything from now on operates on a consistent bucket
	c.Reset()

	accessSecret, err := s.getIronCoreBucketAccessSecretIfRequired(bucket.Bucket, s.clientGetSecretFunc(ctx))
	if err != nil {
		return err
	}

	bucket.AccessSecret = accessSecret
	return nil
}

func (s *Server) CreateBucket(ctx context.Context, req *iri.CreateBucketRequest) (res *iri.CreateBucketResponse, retErr error) {
	log := s.loggerFrom(ctx)

	log.V(1).Info("Getting bucket configuration")
	cfg, err := s.getIronCoreBucketConfig(ctx, req.Bucket)
	if err != nil {
		return nil, fmt.Errorf("error getting ironcore bucket config: %w", err)
	}

	if err := s.createIronCoreBucket(ctx, log, cfg); err != nil {
		return nil, fmt.Errorf("error creating ironcore bucket: %w", err)
	}

	v, err := s.convertAggregateIronCoreBucket(cfg)
	if err != nil {
		return nil, err
	}

	return &iri.CreateBucketResponse{
		Bucket: v,
	}, nil
}
