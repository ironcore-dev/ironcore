// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"errors"
	"fmt"

	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	storageclient "github.com/ironcore-dev/ironcore/internal/client/storage"
	"github.com/ironcore-dev/ironcore/utils/quota"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/go-logr/logr"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	"github.com/ironcore-dev/ironcore/iri/apis/volume"
	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
	"github.com/ironcore-dev/ironcore/poollet/volumepoollet/vcm"
	ironcoreclient "github.com/ironcore-dev/ironcore/utils/client"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type VolumePoolReconciler struct {
	client.Client
	VolumePoolName    string
	VolumeRuntime     volume.RuntimeService
	VolumeClassMapper vcm.VolumeClassMapper
}

//+kubebuilder:rbac:groups=storage.ironcore.dev,resources=volumepools,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=storage.ironcore.dev,resources=volumepools/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=storage.ironcore.dev,resources=volumeclasses,verbs=get;list;watch

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

func (r *VolumePoolReconciler) supportsVolumeClass(ctx context.Context, volumeClass *storagev1alpha1.VolumeClass) (*iri.VolumeClass, *resource.Quantity, error) {
	iriCapabilities := getIRIVolumeClassCapabilities(volumeClass)

	class, quantity, err := r.VolumeClassMapper.GetVolumeClassFor(ctx, volumeClass.Name, iriCapabilities)
	if err != nil {
		if !errors.Is(err, vcm.ErrNoMatchingVolumeClass) && !errors.Is(err, vcm.ErrAmbiguousMatchingVolumeClass) {
			return nil, nil, fmt.Errorf("error getting volume class for %s: %w", volumeClass.Name, err)
		}
		return nil, nil, nil
	}
	return class, quantity, nil
}

func (r *VolumePoolReconciler) calculateCapacity(
	ctx context.Context,
	log logr.Logger,
	volumes []storagev1alpha1.Volume,
	volumeClassList []storagev1alpha1.VolumeClass,
) (capacity, allocatable corev1alpha1.ResourceList, supported []corev1.LocalObjectReference, err error) {
	log.V(1).Info("Determining supported volume classes, capacity and allocatable")

	capacity = corev1alpha1.ResourceList{}
	for _, volumeClass := range volumeClassList {
		class, quantity, err := r.supportsVolumeClass(ctx, &volumeClass)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("error checking whether volume class %s is supported: %w", volumeClass.Name, err)
		}
		if class == nil {
			continue
		}

		supported = append(supported, corev1.LocalObjectReference{Name: volumeClass.Name})
		capacity[corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeVolumeClass, volumeClass.Name)] = *quantity
	}

	usedResources := corev1alpha1.ResourceList{}
	for _, volume := range volumes {
		className := volume.Spec.VolumeClassRef.Name
		res, ok := usedResources[corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeVolumeClass, className)]
		if !ok {
			usedResources[corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeVolumeClass, className)] = *volume.Spec.Resources.Storage()
			continue
		}

		res.Add(*volume.Spec.Resources.Storage())
	}

	return capacity, quota.SubtractWithNonNegativeResult(capacity, usedResources), supported, nil
}

func (r *VolumePoolReconciler) updateStatus(ctx context.Context, log logr.Logger, volumePool *storagev1alpha1.VolumePool, volumes []storagev1alpha1.Volume, volumeClassList []storagev1alpha1.VolumeClass) error {
	capacity, allocatable, supported, err := r.calculateCapacity(ctx, log, volumes, volumeClassList)
	if err != nil {
		return fmt.Errorf("error calculating pool resources:%w", err)
	}

	base := volumePool.DeepCopy()
	volumePool.Status.State = storagev1alpha1.VolumePoolStateAvailable
	volumePool.Status.AvailableVolumeClasses = supported
	volumePool.Status.Capacity = capacity
	volumePool.Status.Allocatable = allocatable

	if err := r.Status().Patch(ctx, volumePool, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error patching volume pool status: %w", err)
	}

	return nil
}

func (r *VolumePoolReconciler) reconcile(ctx context.Context, log logr.Logger, volumePool *storagev1alpha1.VolumePool) (ctrl.Result, error) {
	log.V(1).Info("Reconcile")

	log.V(1).Info("Ensuring no reconcile annotation")
	modified, err := ironcoreclient.PatchEnsureNoReconcileAnnotation(ctx, r.Client, volumePool)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error ensuring no reconcile annotation: %w", err)
	}
	if modified {
		log.V(1).Info("Removed reconcile annotation, requeueing")
		return ctrl.Result{RequeueAfter: 1}, nil
	}

	log.V(1).Info("Listing volume classes")
	volumeClassList := &storagev1alpha1.VolumeClassList{}
	if err := r.List(ctx, volumeClassList); err != nil {
		return ctrl.Result{}, fmt.Errorf("error listing volume classes: %w", err)
	}

	log.V(1).Info("Listing volumes in pool")
	volumeList := &storagev1alpha1.VolumeList{}
	if err := r.List(ctx, volumeList, client.MatchingFields{
		storageclient.VolumeSpecVolumePoolRefNameField: r.VolumePoolName,
	}); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to list volumes in pool: %w", err)
	}

	log.V(1).Info("Updating volume pool status")
	if err := r.updateStatus(ctx, log, volumePool, volumeList.Items, volumeClassList.Items); err != nil {
		return ctrl.Result{}, fmt.Errorf("error updating status: %w", err)
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
			&storagev1alpha1.VolumeClass{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []ctrl.Request {
				return []ctrl.Request{{NamespacedName: client.ObjectKey{Name: r.VolumePoolName}}}
			}),
		).
		Complete(r)
}
