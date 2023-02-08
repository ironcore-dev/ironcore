// Copyright 2021 OnMetal authors
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
	"math/rand"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/onmetal/onmetal-api/api/common/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
)

const (
	bucketPoolStatusAvailableBucketClassesNameField = ".status.availableBucketClasses[*].name"
	bucketSpecBucketPoolNameField                   = ".spec.bucketPool.name"
)

type BucketScheduler struct {
	record.EventRecorder
	client.Client
}

//+kubebuilder:rbac:groups=storage.api.onmetal.de,resources=buckets,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=storage.api.onmetal.de,resources=buckets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=storage.api.onmetal.de,resources=bucketpools,verbs=get;list;watch

// Reconcile reconciles the desired with the actual state.
func (s *BucketScheduler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	bucket := &storagev1alpha1.Bucket{}
	if err := s.Get(ctx, req.NamespacedName, bucket); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !bucket.DeletionTimestamp.IsZero() {
		log.Info("Bucket is already deleting")
		return ctrl.Result{}, nil
	}
	if bucket.Spec.BucketPoolRef != nil {
		log.Info("Bucket is already assigned", "BucketPoolRef", bucket.Spec.BucketPoolRef)
		return ctrl.Result{}, nil
	}
	return s.schedule(ctx, log, bucket)
}

func (s *BucketScheduler) schedule(ctx context.Context, log logr.Logger, bucket *storagev1alpha1.Bucket) (ctrl.Result, error) {
	log.Info("Scheduling bucket")
	list := &storagev1alpha1.BucketPoolList{}
	if err := s.List(ctx, list,
		client.MatchingFields{bucketPoolStatusAvailableBucketClassesNameField: bucket.Spec.BucketClassRef.Name},
		client.MatchingLabels(bucket.Spec.BucketPoolSelector),
	); err != nil {
		return ctrl.Result{}, fmt.Errorf("error listing bucket pools: %w", err)
	}

	var available []storagev1alpha1.BucketPool
	for _, bucketPool := range list.Items {
		if bucketPool.DeletionTimestamp.IsZero() {
			available = append(available, bucketPool)
		}
	}
	if len(available) == 0 {
		log.Info("No bucket pool available for bucket class", "BucketClass", bucket.Spec.BucketClassRef.Name)
		s.Eventf(bucket, corev1.EventTypeNormal, "CannotSchedule", "No BucketPoolRef found for BucketClass %s", bucket.Spec.BucketClassRef.Name)
		return ctrl.Result{}, nil
	}

	// Filter bucket pools by checking if the bucket tolerates all the taints of a bucket pool
	var filtered []storagev1alpha1.BucketPool
	for _, pool := range available {
		if v1alpha1.TolerateTaints(bucket.Spec.Tolerations, pool.Spec.Taints) {
			filtered = append(filtered, pool)
		}
	}
	if len(filtered) == 0 {
		log.Info("No bucket pool tolerated by the bucket", "Tolerations", bucket.Spec.Tolerations)
		s.Eventf(bucket, corev1.EventTypeNormal, "CannotSchedule", "No BucketPoolRef tolerated by %s", &bucket.Spec.Tolerations)
		return ctrl.Result{}, nil
	}
	available = filtered

	// Get a random pool to distribute evenly.
	// TODO: Instead of random distribution, try to come up w/ metrics that include usage of each pool to
	// avoid unfortunate random distribution of items.
	pool := available[rand.Intn(len(available))]
	log = log.WithValues("BucketPoolRef", pool.Name)
	base := bucket.DeepCopy()
	bucket.Spec.BucketPoolRef = &corev1.LocalObjectReference{Name: pool.Name}
	log.Info("Patching bucket")
	if err := s.Patch(ctx, bucket, client.MergeFrom(base)); err != nil {
		return ctrl.Result{}, fmt.Errorf("error scheduling bucket on pool: %w", err)
	}

	log.Info("Successfully assigned bucket")
	return ctrl.Result{}, nil
}

func (s *BucketScheduler) SetupWithManager(mgr manager.Manager) error {
	ctx := context.Background()
	log := ctrl.Log.WithName("bucket-scheduler").WithName("setup")
	ctx = ctrl.LoggerInto(ctx, log)

	if err := mgr.GetFieldIndexer().IndexField(ctx, &storagev1alpha1.BucketPool{}, bucketPoolStatusAvailableBucketClassesNameField, func(obj client.Object) []string {
		pool := obj.(*storagev1alpha1.BucketPool)
		names := make([]string, 0, len(pool.Status.AvailableBucketClasses))
		for _, availableBucketClass := range pool.Status.AvailableBucketClasses {
			names = append(names, availableBucketClass.Name)
		}
		return names
	}); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(ctx, &storagev1alpha1.Bucket{}, bucketSpecBucketPoolNameField, func(obj client.Object) []string {
		bucket := obj.(*storagev1alpha1.Bucket)
		bucketPoolRef := bucket.Spec.BucketPoolRef
		if bucketPoolRef == nil {
			return []string{""}
		}
		return []string{bucketPoolRef.Name}
	}); err != nil {
		return err
	}

	// Only schedule buckets that are not deleting, have no bucket pool and no bucket class set
	filterBucket := func(bucket *storagev1alpha1.Bucket) bool {
		return bucket.DeletionTimestamp.IsZero() &&
			bucket.Spec.BucketPoolRef == nil &&
			bucket.Spec.BucketClassRef != nil
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named("bucket-scheduler").
		For(&storagev1alpha1.Bucket{},
			builder.WithPredicates(predicate.NewPredicateFuncs(func(object client.Object) bool {
				bucket := object.(*storagev1alpha1.Bucket)
				return filterBucket(bucket)
			})),
		).
		// Enqueue unscheduled buckets if a bucket pool w/ required bucket classes becomes available.
		Watches(&source.Kind{Type: &storagev1alpha1.BucketPool{}},
			handler.EnqueueRequestsFromMapFunc(func(object client.Object) []ctrl.Request {
				pool := object.(*storagev1alpha1.BucketPool)
				if !pool.DeletionTimestamp.IsZero() {
					return nil
				}

				list := &storagev1alpha1.BucketList{}
				if err := s.List(ctx, list, client.MatchingFields{bucketSpecBucketPoolNameField: ""}); err != nil {
					log.Error(err, "error listing unscheduled buckets")
					return nil
				}

				availableClassNames := sets.NewString()
				for _, availableBucketClass := range pool.Status.AvailableBucketClasses {
					availableClassNames.Insert(availableBucketClass.Name)
				}

				var requests []ctrl.Request
				for _, bucket := range list.Items {
					if !filterBucket(&bucket) {
						continue
					}

					if !availableClassNames.Has(bucket.Spec.BucketClassRef.Name) {
						continue
					}

					if !labels.SelectorFromSet(bucket.Spec.BucketPoolSelector).Matches(labels.Set(pool.Labels)) {
						continue
					}

					requests = append(requests, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(&bucket)})
				}
				return requests
			}),
		).
		Complete(s)
}
