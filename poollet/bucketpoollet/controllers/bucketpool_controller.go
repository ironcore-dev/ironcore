// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	iriBucket "github.com/ironcore-dev/ironcore/iri/apis/bucket"
	"github.com/ironcore-dev/ironcore/poollet/bucketpoollet/bcm"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type BucketPoolReconciler struct {
	client.Client
	BucketPoolName    string
	BucketRuntime     iriBucket.RuntimeService
	BucketClassMapper bcm.BucketClassMapper
}

//+kubebuilder:rbac:groups=storage.ironcore.dev,resources=bucketpools,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=storage.ironcore.dev,resources=bucketpools/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=storage.ironcore.dev,resources=bucketclasses,verbs=get;list;watch

func (r *BucketPoolReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	bucketPool := &storagev1alpha1.BucketPool{}
	if err := r.Get(ctx, req.NamespacedName, bucketPool); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileExists(ctx, log, bucketPool)
}

func (r *BucketPoolReconciler) reconcileExists(ctx context.Context, log logr.Logger, bucketPool *storagev1alpha1.BucketPool) (ctrl.Result, error) {
	if !bucketPool.DeletionTimestamp.IsZero() {
		return r.delete(ctx, log, bucketPool)
	}
	return r.reconcile(ctx, log, bucketPool)
}

func (r *BucketPoolReconciler) delete(ctx context.Context, log logr.Logger, bucketPool *storagev1alpha1.BucketPool) (ctrl.Result, error) {
	log.V(1).Info("Delete")
	log.V(1).Info("Deleted")
	return ctrl.Result{}, nil
}

func (r *BucketPoolReconciler) supportsBucketClass(ctx context.Context, log logr.Logger, bucketClass *storagev1alpha1.BucketClass) (bool, error) {
	iriCapabilities, err := getIRIBucketClassCapabilities(bucketClass)
	if err != nil {
		return false, fmt.Errorf("error getting iri mahchine class capabilities: %w", err)
	}

	_, err = r.BucketClassMapper.GetBucketClassFor(ctx, bucketClass.Name, iriCapabilities)
	if err != nil {
		if !errors.Is(err, bcm.ErrNoMatchingBucketClass) && !errors.Is(err, bcm.ErrAmbiguousMatchingBucketClass) {
			return false, fmt.Errorf("error getting bucket class for %s: %w", bucketClass.Name, err)
		}
		return false, nil
	}
	return true, nil
}

func (r *BucketPoolReconciler) reconcile(ctx context.Context, log logr.Logger, bucketPool *storagev1alpha1.BucketPool) (ctrl.Result, error) {
	log.V(1).Info("Reconcile")

	log.V(1).Info("Listing bucket classes")
	bucketClassList := &storagev1alpha1.BucketClassList{}
	if err := r.List(ctx, bucketClassList); err != nil {
		return ctrl.Result{}, fmt.Errorf("error listing bucket classes: %w", err)
	}

	log.V(1).Info("Determining supported bucket classes")
	var supported []corev1.LocalObjectReference
	for _, bucketClass := range bucketClassList.Items {
		ok, err := r.supportsBucketClass(ctx, log, &bucketClass)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("error checking whether bucket class %s is supported: %w", bucketClass.Name, err)
		}
		if !ok {
			continue
		}

		supported = append(supported, corev1.LocalObjectReference{Name: bucketClass.Name})
	}

	log.V(1).Info("Updating bucket pool status")
	base := bucketPool.DeepCopy()
	bucketPool.Status.AvailableBucketClasses = supported
	if err := r.Status().Patch(ctx, bucketPool, client.MergeFrom(base)); err != nil {
		return ctrl.Result{}, fmt.Errorf("error patchign bucket pool status: %w", err)
	}

	log.V(1).Info("Reconciled")
	return ctrl.Result{}, nil
}

func (r *BucketPoolReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(
			&storagev1alpha1.BucketPool{},
			builder.WithPredicates(
				predicate.NewPredicateFuncs(func(obj client.Object) bool {
					return obj.GetName() == r.BucketPoolName
				}),
			),
		).
		Watches(
			&storagev1alpha1.BucketClass{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []ctrl.Request {
				return []ctrl.Request{{NamespacedName: client.ObjectKey{Name: r.BucketPoolName}}}
			}),
		).
		Complete(r)
}
