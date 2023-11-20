// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/go-logr/logr"
	"github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	storageclient "github.com/ironcore-dev/ironcore/internal/client/storage"
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
)

type BucketScheduler struct {
	record.EventRecorder
	client.Client
}

//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch
//+kubebuilder:rbac:groups=storage.ironcore.dev,resources=buckets,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=storage.ironcore.dev,resources=buckets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=storage.ironcore.dev,resources=bucketpools,verbs=get;list;watch

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
		client.MatchingFields{storageclient.BucketPoolAvailableBucketClassesField: bucket.Spec.BucketClassRef.Name},
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

func filterBucket(bucket *storagev1alpha1.Bucket) bool {
	return bucket.DeletionTimestamp.IsZero() &&
		bucket.Spec.BucketPoolRef == nil &&
		bucket.Spec.BucketClassRef != nil
}

func (s *BucketScheduler) enqueueRequestsByBucketPool() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, object client.Object) []ctrl.Request {
		pool := object.(*storagev1alpha1.BucketPool)
		log := ctrl.LoggerFrom(ctx)
		if !pool.DeletionTimestamp.IsZero() {
			return nil
		}

		list := &storagev1alpha1.BucketList{}
		if err := s.List(ctx, list, client.MatchingFields{storageclient.BucketSpecBucketPoolRefNameField: ""}); err != nil {
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
	})
}

func (s *BucketScheduler) SetupWithManager(mgr manager.Manager) error {
	// Only schedule buckets that are not deleting, have no bucket pool and no bucket class set

	return ctrl.NewControllerManagedBy(mgr).
		Named("bucket-scheduler").
		For(&storagev1alpha1.Bucket{},
			builder.WithPredicates(predicate.NewPredicateFuncs(func(object client.Object) bool {
				bucket := object.(*storagev1alpha1.Bucket)
				return filterBucket(bucket)
			})),
		).
		// Enqueue unscheduled buckets if a bucket pool w/ required bucket classes becomes available.
		Watches(
			&storagev1alpha1.BucketPool{},
			s.enqueueRequestsByBucketPool(),
		).
		Complete(s)
}
