// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	computeclient "github.com/ironcore-dev/ironcore/internal/client/compute"
	"github.com/ironcore-dev/ironcore/iri/apis/machine"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	"github.com/ironcore-dev/ironcore/poollet/machinepoollet/mcm"
	ironcoreclient "github.com/ironcore-dev/ironcore/utils/client"
	"github.com/ironcore-dev/ironcore/utils/quota"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type MachinePoolReconciler struct {
	client.Client

	// MachinePoolName is the name of the computev1alpha1.MachinePool to report / update.
	MachinePoolName string
	// Addresses are the addresses the machinepoollet server is available on.
	Addresses []computev1alpha1.MachinePoolAddress
	// Port is the port the machinepoollet server is available on.
	Port int32

	MachineRuntime     machine.RuntimeService
	MachineClassMapper mcm.MachineClassMapper
}

//+kubebuilder:rbac:groups=compute.ironcore.dev,resources=machinepools,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=compute.ironcore.dev,resources=machinepools/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=compute.ironcore.dev,resources=machineclasses,verbs=get;list;watch

func (r *MachinePoolReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	machinePool := &computev1alpha1.MachinePool{}
	if err := r.Get(ctx, req.NamespacedName, machinePool); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileExists(ctx, log, machinePool)
}

func (r *MachinePoolReconciler) reconcileExists(ctx context.Context, log logr.Logger, machinePool *computev1alpha1.MachinePool) (ctrl.Result, error) {
	if !machinePool.DeletionTimestamp.IsZero() {
		return r.delete(ctx, log, machinePool)
	}
	return r.reconcile(ctx, log, machinePool)
}

func (r *MachinePoolReconciler) delete(ctx context.Context, log logr.Logger, machinePool *computev1alpha1.MachinePool) (ctrl.Result, error) {
	log.V(1).Info("Delete")
	log.V(1).Info("Deleted")
	return ctrl.Result{}, nil
}

func (r *MachinePoolReconciler) supportsMachineClass(ctx context.Context, log logr.Logger, machineClass *computev1alpha1.MachineClass) (*iri.MachineClass, int64, error) {
	iriCapabilities, err := getIRIMachineClassCapabilities(machineClass)
	if err != nil {
		return nil, 0, fmt.Errorf("error getting iri mahchine class capabilities: %w", err)
	}

	class, quantity, err := r.MachineClassMapper.GetMachineClassFor(ctx, machineClass.Name, iriCapabilities)
	if err != nil {
		if !errors.Is(err, mcm.ErrNoMatchingMachineClass) && !errors.Is(err, mcm.ErrAmbiguousMatchingMachineClass) {
			return nil, 0, fmt.Errorf("error getting machine class for %s: %w", machineClass.Name, err)
		}
		return nil, 0, nil
	}
	return class, quantity, nil
}

func (r *MachinePoolReconciler) calculateCapacity(
	ctx context.Context,
	log logr.Logger,
	machines []computev1alpha1.Machine,
	machineClassList []computev1alpha1.MachineClass,
) (capacity, allocatable corev1alpha1.ResourceList, supported []corev1.LocalObjectReference, err error) {
	log.V(1).Info("Determining supported machine classes, capacity and allocatable")

	capacity = corev1alpha1.ResourceList{}
	for _, machineClass := range machineClassList {
		class, quantity, err := r.supportsMachineClass(ctx, log, &machineClass)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("error checking whether machine class %s is supported: %w", machineClass.Name, err)
		}
		if class == nil {
			continue
		}

		supported = append(supported, corev1.LocalObjectReference{Name: machineClass.Name})
		capacity[corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeMachineClass, machineClass.Name)] = *resource.NewQuantity(quantity, resource.DecimalSI)
	}

	usedResources := corev1alpha1.ResourceList{}
	for _, machine := range machines {
		className := machine.Spec.MachineClassRef.Name
		res, ok := usedResources[corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeMachineClass, className)]
		if !ok {
			usedResources[corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeMachineClass, className)] = *resource.NewQuantity(1, resource.DecimalSI)
			continue
		}

		res.Add(resource.MustParse("1"))
	}

	return capacity, quota.SubtractWithNonNegativeResult(capacity, usedResources), supported, nil
}

func (r *MachinePoolReconciler) updateStatus(ctx context.Context, log logr.Logger, machinePool *computev1alpha1.MachinePool, machines []computev1alpha1.Machine, machineClassList []computev1alpha1.MachineClass) error {
	capacity, allocatable, supported, err := r.calculateCapacity(ctx, log, machines, machineClassList)
	if err != nil {
		return fmt.Errorf("error calculating pool resources:%w", err)
	}

	base := machinePool.DeepCopy()
	machinePool.Status.State = computev1alpha1.MachinePoolStateReady
	machinePool.Status.AvailableMachineClasses = supported
	machinePool.Status.Addresses = r.Addresses
	machinePool.Status.Capacity = capacity
	machinePool.Status.Allocatable = allocatable
	machinePool.Status.DaemonEndpoints.MachinepoolletEndpoint.Port = r.Port

	if err := r.Status().Patch(ctx, machinePool, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error patching machine pool status: %w", err)
	}

	return nil
}

func (r *MachinePoolReconciler) reconcile(ctx context.Context, log logr.Logger, machinePool *computev1alpha1.MachinePool) (ctrl.Result, error) {
	log.V(1).Info("Reconcile")

	log.V(1).Info("Ensuring no reconcile annotation")
	modified, err := ironcoreclient.PatchEnsureNoReconcileAnnotation(ctx, r.Client, machinePool)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error ensuring no reconcile annotation: %w", err)
	}
	if modified {
		log.V(1).Info("Removed reconcile annotation, requeueing")
		return ctrl.Result{Requeue: true}, nil
	}

	log.V(1).Info("Listing machine classes")
	machineClassList := &computev1alpha1.MachineClassList{}
	if err := r.List(ctx, machineClassList); err != nil {
		return ctrl.Result{}, fmt.Errorf("error listing machine classes: %w", err)
	}

	log.V(1).Info("Listing machines in pool")
	machineList := &computev1alpha1.MachineList{}
	if err := r.List(ctx, machineList, client.MatchingFields{
		computeclient.MachineSpecMachinePoolRefNameField: r.MachinePoolName,
	}); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to list machines in pool: %w", err)
	}

	log.V(1).Info("Updating machine pool status")
	if err := r.updateStatus(ctx, log, machinePool, machineList.Items, machineClassList.Items); err != nil {
		return ctrl.Result{}, fmt.Errorf("error updating status: %w", err)
	}

	log.V(1).Info("Reconciled")
	return ctrl.Result{}, nil
}

func (r *MachinePoolReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(
			&computev1alpha1.MachinePool{},
			builder.WithPredicates(
				predicate.NewPredicateFuncs(func(obj client.Object) bool {
					return obj.GetName() == r.MachinePoolName
				}),
			),
		).
		Watches(
			&computev1alpha1.MachineClass{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []ctrl.Request {
				return []ctrl.Request{{NamespacedName: client.ObjectKey{Name: r.MachinePoolName}}}
			}),
		).
		Watches(
			&computev1alpha1.Machine{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []ctrl.Request {
				return []ctrl.Request{{NamespacedName: client.ObjectKey{Name: r.MachinePoolName}}}
			}),
			builder.WithPredicates(
				MachineRunsInMachinePoolPredicate(r.MachinePoolName),
			),
		).
		Complete(r)
}
