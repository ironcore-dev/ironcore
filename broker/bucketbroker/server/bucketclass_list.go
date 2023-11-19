// Copyright 2022 IronCore authors
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

	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/bucket/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) getTargetIronCoreBucketPools(ctx context.Context) ([]storagev1alpha1.BucketPool, error) {
	if s.bucketPoolName != "" {
		ironcoreBucketPool := &storagev1alpha1.BucketPool{}
		ironcoreBucketPoolKey := client.ObjectKey{Name: s.bucketPoolName}
		if err := s.client.Get(ctx, ironcoreBucketPoolKey, ironcoreBucketPool); err != nil {
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

func (s *Server) gatherAvailableBucketClassNames(ironcoreBucketPools []storagev1alpha1.BucketPool) sets.Set[string] {
	res := sets.New[string]()
	for _, ironcoreBucketPool := range ironcoreBucketPools {
		for _, availableBucketClass := range ironcoreBucketPool.Status.AvailableBucketClasses {
			res.Insert(availableBucketClass.Name)
		}
	}
	return res
}

func (s *Server) filterIronCoreBucketClasses(
	availableBucketClassNames sets.Set[string],
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

func (s *Server) convertIronCoreBucketClass(bucketClass *storagev1alpha1.BucketClass) (*iri.BucketClass, error) {
	tps := bucketClass.Capabilities.TPS()
	iops := bucketClass.Capabilities.IOPS()

	return &iri.BucketClass{
		Name: bucketClass.Name,
		Capabilities: &iri.BucketClassCapabilities{
			Tps:  tps.Value(),
			Iops: iops.Value(),
		},
	}, nil
}

func (s *Server) ListBucketClasses(ctx context.Context, req *iri.ListBucketClassesRequest) (*iri.ListBucketClassesResponse, error) {
	log := s.loggerFrom(ctx)

	log.V(1).Info("Getting target ironcore bucket pools")
	ironcoreBucketPools, err := s.getTargetIronCoreBucketPools(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting target ironcore bucket pools: %w", err)
	}

	log.V(1).Info("Gathering available bucket class names")
	availableIronCoreBucketClassNames := s.gatherAvailableBucketClassNames(ironcoreBucketPools)

	if len(availableIronCoreBucketClassNames) == 0 {
		log.V(1).Info("No available bucket classes")
		return &iri.ListBucketClassesResponse{BucketClasses: []*iri.BucketClass{}}, nil
	}

	log.V(1).Info("Listing ironcore bucket classes")
	ironcoreBucketClassList := &storagev1alpha1.BucketClassList{}
	if err := s.client.List(ctx, ironcoreBucketClassList); err != nil {
		return nil, fmt.Errorf("error listing ironcore bucket classes: %w", err)
	}

	availableIronCoreBucketClasses := s.filterIronCoreBucketClasses(availableIronCoreBucketClassNames, ironcoreBucketClassList.Items)
	bucketClasses := make([]*iri.BucketClass, 0, len(availableIronCoreBucketClasses))
	for _, ironcoreBucketClass := range availableIronCoreBucketClasses {
		bucketClass, err := s.convertIronCoreBucketClass(&ironcoreBucketClass)
		if err != nil {
			return nil, fmt.Errorf("error converting ironcore bucket class %s: %w", ironcoreBucketClass.Name, err)
		}

		bucketClasses = append(bucketClasses, bucketClass)
	}

	log.V(1).Info("Returning bucket classes")
	return &iri.ListBucketClassesResponse{
		BucketClasses: bucketClasses,
	}, nil
}
