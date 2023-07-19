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
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	corev1alpha1 "github.com/onmetal/onmetal-api/api/core/v1alpha1"
	computeclient "github.com/onmetal/onmetal-api/internal/client/compute"
	"github.com/onmetal/onmetal-api/ori/apis/machine"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"github.com/onmetal/onmetal-api/poollet/machinepoollet/mcm"
	onmetalapiclient "github.com/onmetal/onmetal-api/utils/client"
	"github.com/onmetal/onmetal-api/utils/quota"
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

//+kubebuilder:rbac:groups=compute.api.onmetal.de,resources=machinepools,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=compute.api.onmetal.de,resources=machinepools/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=compute.api.onmetal.de,resources=machineclasses,verbs=get;list;watch

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

func (r *MachinePoolReconciler) supportsMachineClass(ctx context.Context, log logr.Logger, machineClass *computev1alpha1.MachineClass) (*ori.MachineClass, int64, error) {
	oriCapabilities, err := getORIMachineClassCapabilities(machineClass)
	if err != nil {
		return nil, 0, fmt.Errorf("error getting ori mahchine class capabilities: %w", err)
	}

	class, quantity, err := r.MachineClassMapper.GetMachineClassFor(ctx, machineClass.Name, oriCapabilities)
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
	machinePool *computev1alpha1.MachinePool,
	machines []computev1alpha1.Machine,
	machineClassList []computev1alpha1.MachineClass,
) (corev1alpha1.ResourceList, corev1alpha1.ResourceList, []corev1.LocalObjectReference, error) {
	log.V(1).Info("Determining supported machine classes, capacity and allocatable")

	capacity := corev1alpha1.ResourceList{}
	usedResources := corev1alpha1.ResourceList{}

	var supported []corev1.LocalObjectReference
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
	capacity, allocatable, supported, err := r.calculateCapacity(ctx, log, machinePool, machines, machineClassList)
	if err != nil {
		//ToDo
		return fmt.Errorf("failed to ... :%w", err)
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
	modified, err := onmetalapiclient.PatchEnsureNoReconcileAnnotation(ctx, r.Client, machinePool)
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
