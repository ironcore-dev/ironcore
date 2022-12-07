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

package controllers

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/volume/v1alpha1"
	"github.com/onmetal/onmetal-api/poollet/volumepoollet/vcm"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type VolumePoolReconciler struct {
	client.Client
	VolumePoolName    string
	VolumeRuntime     ori.VolumeRuntimeClient
	VolumeClassMapper vcm.VolumeClassMapper
}

//+kubebuilder:rbac:groups=storage.api.onmetal.de,resources=volumepools,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=storage.api.onmetal.de,resources=volumepools/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=storage.api.onmetal.de,resources=volumeclasses,verbs=get;list;watch

func (r *VolumePoolReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	volumePool := &storagev1alpha1.VolumePool{}
	if err := r.Get(ctx, req.NamespacedName, volumePool); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileExists(ctx, log, volumePool)
}

func (r *VolumePoolReconciler) reconcileExists(ctx context.Context, log logr.Logger, volumePool *storagev1alpha1.VolumePool) (ctrl.Result, error) {
	if !volumePool.DeletionTimestamp.IsZero() {
		return r.delete(ctx, log, volumePool)
	}
	return r.reconcile(ctx, log, volumePool)
}

func (r *VolumePoolReconciler) delete(ctx context.Context, log logr.Logger, volumePool *storagev1alpha1.VolumePool) (ctrl.Result, error) {
	log.V(1).Info("Delete")
	log.V(1).Info("Deleted")
	return ctrl.Result{}, nil
}

func (r *VolumePoolReconciler) supportsVolumeClass(ctx context.Context, log logr.Logger, volumeClass *storagev1alpha1.VolumeClass) (bool, error) {
	oriCapabilities, err := getORIVolumeClassCapabilities(volumeClass)
	if err != nil {
		return false, fmt.Errorf("error getting ori mahchine class capabilities: %w", err)
	}

	_, err = r.VolumeClassMapper.GetVolumeClassFor(ctx, volumeClass.Name, oriCapabilities)
	if err != nil {
		if !errors.Is(err, vcm.ErrNoMatchingVolumeClass) && !errors.Is(err, vcm.ErrAmbiguousMatchingVolumeClass) {
			return false, fmt.Errorf("error getting volume class for %s: %w", volumeClass.Name, err)
		}
		return false, nil
	}
	return true, nil
}

func (r *VolumePoolReconciler) reconcile(ctx context.Context, log logr.Logger, volumePool *storagev1alpha1.VolumePool) (ctrl.Result, error) {
	log.V(1).Info("Reconcile")

	log.V(1).Info("Listing volume classes")
	volumeClassList := &storagev1alpha1.VolumeClassList{}
	if err := r.List(ctx, volumeClassList); err != nil {
		return ctrl.Result{}, fmt.Errorf("error listing volume classes: %w", err)
	}

	log.V(1).Info("Determining supported volume classes")
	var supported []corev1.LocalObjectReference
	for _, volumeClass := range volumeClassList.Items {
		ok, err := r.supportsVolumeClass(ctx, log, &volumeClass)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("error checking whether volume class %s is supported: %w", volumeClass.Name, err)
		}
		if !ok {
			continue
		}

		supported = append(supported, corev1.LocalObjectReference{Name: volumeClass.Name})
	}

	log.V(1).Info("Updating volume pool status")
	base := volumePool.DeepCopy()
	volumePool.Status.AvailableVolumeClasses = supported
	if err := r.Status().Patch(ctx, volumePool, client.MergeFrom(base)); err != nil {
		return ctrl.Result{}, fmt.Errorf("error patchign volume pool status: %w", err)
	}

	log.V(1).Info("Reconciled")
	return ctrl.Result{}, nil
}

func (r *VolumePoolReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(
			&storagev1alpha1.VolumePool{},
			builder.WithPredicates(
				predicate.NewPredicateFuncs(func(obj client.Object) bool {
					return obj.GetName() == r.VolumePoolName
				}),
			),
		).
		Watches(
			&source.Kind{Type: &storagev1alpha1.VolumeClass{}},
			handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
				return []ctrl.Request{{NamespacedName: client.ObjectKey{Name: r.VolumePoolName}}}
			}),
		).
		Complete(r)
}
