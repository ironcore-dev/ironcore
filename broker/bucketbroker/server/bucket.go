// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"fmt"

	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	"github.com/ironcore-dev/ironcore/broker/bucketbroker/apiutils"
	iri "github.com/ironcore-dev/ironcore/iri/apis/bucket/v1alpha1"
)

func (s *Server) convertAggregateIronCoreBucket(bucket *AggregateIronCoreBucket) (*iri.Bucket, error) {
	metadata, err := apiutils.GetObjectMetadata(bucket.Bucket)
	if err != nil {
		return nil, err
	}

	state, err := s.convertIronCoreBucketState(bucket.Bucket.Status.State)
	if err != nil {
		return nil, err
	}

	access, err := s.convertIronCoreBucketAccess(bucket)
	if err != nil {
		return nil, err
	}

	return &iri.Bucket{
		Metadata: metadata,
		Spec: &iri.BucketSpec{
			Class: bucket.Bucket.Spec.BucketClassRef.Name,
		},
		Status: &iri.BucketStatus{
			State:  state,
			Access: access,
		},
	}, nil
}

var ironcoreBucketStateToIRIState = map[storagev1alpha1.BucketState]iri.BucketState{
	storagev1alpha1.BucketStatePending:   iri.BucketState_BUCKET_PENDING,
	storagev1alpha1.BucketStateAvailable: iri.BucketState_BUCKET_AVAILABLE,
	storagev1alpha1.BucketStateError:     iri.BucketState_BUCKET_ERROR,
}

func (s *Server) convertIronCoreBucketState(state storagev1alpha1.BucketState) (iri.BucketState, error) {
	if state, ok := ironcoreBucketStateToIRIState[state]; ok {
		return state, nil
	}
	return 0, fmt.Errorf("unknown ironcore bucket state %q", state)
}

func (s *Server) convertIronCoreBucketAccess(bucket *AggregateIronCoreBucket) (*iri.BucketAccess, error) {
	if bucket.Bucket.Status.State != storagev1alpha1.BucketStateAvailable {
		return nil, nil
	}

	access := bucket.Bucket.Status.Access
	if access == nil {
		return nil, nil
	}

	var secretData map[string][]byte
	if secretRef := access.SecretRef; secretRef != nil {
		if bucket.AccessSecret == nil {
			return nil, fmt.Errorf("access secret specified but not contained in aggregate ironcore bucket")
		}
		secretData = bucket.AccessSecret.Data
	}

	return &iri.BucketAccess{
		Endpoint:   access.Endpoint,
		SecretData: secretData,
	}, nil
}
