// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"

	iri "github.com/ironcore-dev/ironcore/iri/apis/bucket/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func (s *Server) DeleteBucket(ctx context.Context, req *iri.DeleteBucketRequest) (*iri.DeleteBucketResponse, error) {
	bucketID := req.BucketId
	log := s.loggerFrom(ctx, "BucketID", bucketID)

	ironcoreBucket, err := s.getAggregateIronCoreBucket(ctx, req.BucketId)
	if err != nil {
		return nil, err
	}

	log.V(1).Info("Deleting bucket")
	if err := s.client.Delete(ctx, ironcoreBucket.Bucket); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("error deleting ironcore bucket: %w", err)
		}
		return nil, status.Errorf(codes.NotFound, "bucket %s not found", bucketID)
	}

	return &iri.DeleteBucketResponse{}, nil
}
