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

	"github.com/onmetal/controller-utils/set"
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/bucket/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) getTargetOnmetalBucketPools(ctx context.Context) ([]storagev1alpha1.BucketPool, error) {
	if s.bucketPoolName != "" {
		onmetalBucketPool := &storagev1alpha1.BucketPool{}
		onmetalBucketPoolKey := client.ObjectKey{Name: s.bucketPoolName}
		if err := s.client.Get(ctx, onmetalBucketPoolKey, onmetalBucketPool); err != nil {
			if !apierrors.IsNotFound(err) {
				return nil, fmt.Errorf("error getting bucket pool %s: %w", s.bucketPoolName, err)
			}
			return nil, nil
		}
	}

	bucketPoolList := &storagev1alpha1.BucketPoolList{}
	if err := s.client.List(ctx, bucketPoolList,
		client.MatchingLabels(s.bucketPoolSelector),
	); err != nil {
		return nil, fmt.Errorf("error listing bucket pools: %w", err)
	}
	return bucketPoolList.Items, nil
}

func (s *Server) gatherAvailableBucketClassNames(onmetalBucketPools []storagev1alpha1.BucketPool) set.Set[string] {
	res := set.New[string]()
	for _, onmetalBucketPool := range onmetalBucketPools {
		for _, availableBucketClass := range onmetalBucketPool.Status.AvailableBucketClasses {
			res.Insert(availableBucketClass.Name)
		}
	}
	return res
}

func (s *Server) filterOnmetalBucketClasses(
	availableBucketClassNames set.Set[string],
	bucketClasses []storagev1alpha1.BucketClass,
) []storagev1alpha1.BucketClass {
	var filtered []storagev1alpha1.BucketClass
	for _, bucketClass := range bucketClasses {
		if !availableBucketClassNames.Has(bucketClass.Name) {
			continue
		}

		filtered = append(filtered, bucketClass)
	}
	return filtered
}

func (s *Server) convertOnmetalBucketClass(bucketClass *storagev1alpha1.BucketClass) (*ori.BucketClass, error) {
	tps := bucketClass.Capabilities.TPS()
	iops := bucketClass.Capabilities.IOPS()

	return &ori.BucketClass{
		Name: bucketClass.Name,
		Capabilities: &ori.BucketClassCapabilities{
			Tps:  tps.Value(),
			Iops: iops.Value(),
		},
	}, nil
}

func (s *Server) ListBucketClasses(ctx context.Context, req *ori.ListBucketClassesRequest) (*ori.ListBucketClassesResponse, error) {
	log := s.loggerFrom(ctx)

	log.V(1).Info("Getting target onmetal bucket pools")
	onmetalBucketPools, err := s.getTargetOnmetalBucketPools(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting target onmetal bucket pools: %w", err)
	}

	log.V(1).Info("Gathering available bucket class names")
	availableOnmetalBucketClassNames := s.gatherAvailableBucketClassNames(onmetalBucketPools)

	if len(availableOnmetalBucketClassNames) == 0 {
		log.V(1).Info("No available bucket classes")
		return &ori.ListBucketClassesResponse{BucketClasses: []*ori.BucketClass{}}, nil
	}

	log.V(1).Info("Listing onmetal bucket classes")
	onmetalBucketClassList := &storagev1alpha1.BucketClassList{}
	if err := s.client.List(ctx, onmetalBucketClassList); err != nil {
		return nil, fmt.Errorf("error listing onmetal bucket classes: %w", err)
	}

	availableOnmetalBucketClasses := s.filterOnmetalBucketClasses(availableOnmetalBucketClassNames, onmetalBucketClassList.Items)
	bucketClasses := make([]*ori.BucketClass, 0, len(availableOnmetalBucketClasses))
	for _, onmetalBucketClass := range availableOnmetalBucketClasses {
		bucketClass, err := s.convertOnmetalBucketClass(&onmetalBucketClass)
		if err != nil {
			return nil, fmt.Errorf("error converting onmetal bucket class %s: %w", onmetalBucketClass.Name, err)
		}

		bucketClasses = append(bucketClasses, bucketClass)
	}

	log.V(1).Info("Returning bucket classes")
	return &ori.ListBucketClassesResponse{
		BucketClasses: bucketClasses,
	}, nil
}
