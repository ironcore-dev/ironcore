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
	"fmt"

	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	"github.com/onmetal/onmetal-api/broker/bucketbroker/apiutils"
	ori "github.com/onmetal/onmetal-api/ori/apis/bucket/v1alpha1"
)

func (s *Server) convertAggregateOnmetalBucket(bucket *AggregateOnmetalBucket) (*ori.Bucket, error) {
	metadata, err := apiutils.GetObjectMetadata(bucket.Bucket)
	if err != nil {
		return nil, err
	}

	state, err := s.convertOnmetalBucketState(bucket.Bucket.Status.State)
	if err != nil {
		return nil, err
	}

	access, err := s.convertOnmetalBucketAccess(bucket)
	if err != nil {
		return nil, err
	}

	return &ori.Bucket{
		Metadata: metadata,
		Spec: &ori.BucketSpec{
			Class: bucket.Bucket.Spec.BucketClassRef.Name,
		},
		Status: &ori.BucketStatus{
			State:  state,
			Access: access,
		},
	}, nil
}

var onmetalBucketStateToORIState = map[storagev1alpha1.BucketState]ori.BucketState{
	storagev1alpha1.BucketStatePending:   ori.BucketState_BUCKET_PENDING,
	storagev1alpha1.BucketStateAvailable: ori.BucketState_BUCKET_AVAILABLE,
	storagev1alpha1.BucketStateError:     ori.BucketState_BUCKET_ERROR,
}

func (s *Server) convertOnmetalBucketState(state storagev1alpha1.BucketState) (ori.BucketState, error) {
	if state, ok := onmetalBucketStateToORIState[state]; ok {
		return state, nil
	}
	return 0, fmt.Errorf("unknown onmetal bucket state %q", state)
}

func (s *Server) convertOnmetalBucketAccess(bucket *AggregateOnmetalBucket) (*ori.BucketAccess, error) {
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
			return nil, fmt.Errorf("access secret specified but not contained in aggregate onmetal bucket")
		}
		secretData = bucket.AccessSecret.Data
	}

	return &ori.BucketAccess{
		Endpoint:   access.Endpoint,
		SecretData: secretData,
	}, nil
}
