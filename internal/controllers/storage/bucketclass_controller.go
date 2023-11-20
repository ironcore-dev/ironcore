/*
 * Copyright (c) 2021 by the IronCore authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package storage

import (
	"context"
	"fmt"
	"sort"

	"github.com/go-logr/logr"
	"github.com/ironcore-dev/controller-utils/clientutils"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	storageclient "github.com/ironcore-dev/ironcore/internal/client/storage"
	"github.com/ironcore-dev/ironcore/utils/slices"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
)

// BucketClassReconciler reconciles a BucketClass object
type BucketClassReconciler struct {
	client.Client
	APIReader client.Reader
}

//+kubebuilder:rbac:groups=storage.ironcore.dev,resources=bucketclasses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=storage.ironcore.dev,resources=bucketclasses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=storage.ironcore.dev,resources=bucketclasses/finalizers,verbs=update
//+kubebuilder:rbac:groups=storage.ironcore.dev,resources=buckets,verbs=get;list;watch

// Reconcile moves the current state of the cluster closer to the desired state.
func (r *BucketClassReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	bucketClass := &storagev1alpha1.BucketClass{}
	if err := r.Get(ctx, req.NamespacedName, bucketClass); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileExists(ctx, log, bucketClass)
}

func (r *BucketClassReconciler) listReferencingBucketsWithReader(
	ctx context.Context,
	rd client.Reader,
	bucketClass *storagev1alpha1.BucketClass,
) ([]storagev1alpha1.Bucket, error) {
	bucketList := &storagev1alpha1.BucketList{}
	if err := rd.List(ctx, bucketList,
		client.InNamespace(bucketClass.Namespace),
		client.MatchingFields{storageclient.BucketSpecBucketClassRefNameField: bucketClass.Name},
	); err != nil {
		return nil, fmt.Errorf("error listing the buckets using the bucket class: %w", err)
	}
	return bucketList.Items, nil
}

func (r *BucketClassReconciler) collectBucketNames(buckets []storagev1alpha1.Bucket) []string {
	bucketNames := slices.MapRef(buckets, func(bucket *storagev1alpha1.Bucket) string {
		return bucket.Name
	})
	sort.Strings(bucketNames)
	return bucketNames
}

func (r *BucketClassReconciler) delete(ctx context.Context, log logr.Logger, bucketClass *storagev1alpha1.BucketClass) (ctrl.Result, error) {
	if !controllerutil.ContainsFinalizer(bucketClass, storagev1alpha1.BucketClassFinalizer) {
		return ctrl.Result{}, nil
	}

	buckets, err := r.listReferencingBucketsWithReader(ctx, r.Client, bucketClass)
	if err != nil {
		return ctrl.Result{}, err
	}
	if len(buckets) > 0 {
		log.V(1).Info("Bucket class is still in use", "ReferencingBucketNames", r.collectBucketNames(buckets))
		return ctrl.Result{Requeue: true}, nil
	}

	buckets, err = r.listReferencingBucketsWithReader(ctx, r.APIReader, bucketClass)
	if err != nil {
		return ctrl.Result{}, err
	}
	if len(buckets) > 0 {
		log.V(1).Info("Bucket class is still in use", "ReferencingBucketNames", r.collectBucketNames(buckets))
		return ctrl.Result{Requeue: true}, nil
	}

	log.V(1).Info("Bucket Class is not used anymore, removing finalizer")
	if err := clientutils.PatchRemoveFinalizer(ctx, r.Client, bucketClass, storagev1alpha1.BucketClassFinalizer); err != nil {
		return ctrl.Result{}, err
	}

	log.V(1).Info("Successfully removed finalizer")
	return ctrl.Result{}, nil
}

func (r *BucketClassReconciler) reconcile(ctx context.Context, log logr.Logger, bucketClass *storagev1alpha1.BucketClass) (ctrl.Result, error) {
	log.V(1).Info("Ensuring finalizer")
	if modified, err := clientutils.PatchEnsureFinalizer(ctx, r.Client, bucketClass, storagev1alpha1.BucketClassFinalizer); err != nil || modified {
		return ctrl.Result{}, err
	}

	log.V(1).Info("Finalizer is present")
	return ctrl.Result{}, nil
}

func (r *BucketClassReconciler) reconcileExists(ctx context.Context, log logr.Logger, bucketClass *storagev1alpha1.BucketClass) (ctrl.Result, error) {
	if !bucketClass.DeletionTimestamp.IsZero() {
		return r.delete(ctx, log, bucketClass)
	}
	return r.reconcile(ctx, log, bucketClass)
}

// SetupWithManager sets up the controller with the Manager.
func (r *BucketClassReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&storagev1alpha1.BucketClass{}).
		Watches(
			&storagev1alpha1.Bucket{},
			handler.Funcs{
				DeleteFunc: func(ctx context.Context, event event.DeleteEvent, queue workqueue.RateLimitingInterface) {
					bucket := event.Object.(*storagev1alpha1.Bucket)
					bucketClassRef := bucket.Spec.BucketClassRef
					if bucketClassRef == nil {
						return
					}

					queue.Add(ctrl.Request{NamespacedName: types.NamespacedName{Name: bucketClassRef.Name}})
				},
			},
		).
		Complete(r)
}
