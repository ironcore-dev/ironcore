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
	"github.com/onmetal/onmetal-api/ori/apis/machine"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	machinepoolletclient "github.com/onmetal/onmetal-api/poollet/machinepoollet/client"
	"github.com/onmetal/onmetal-api/poollet/machinepoollet/mcm"
	onmetalapiclient "github.com/onmetal/onmetal-api/utils/client"
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

func (r *MachinePoolReconciler) supportsMachineClass(ctx context.Context, log logr.Logger, machineClass *computev1alpha1.MachineClass) (bool, error) {
	oriCapabilities, err := getORIMachineClassCapabilities(machineClass)
	if err != nil {
		return false, fmt.Errorf("error getting ori mahchine class capabilities: %w", err)
	}

	_, err = r.MachineClassMapper.GetMachineClassFor(ctx, machineClass.Name, oriCapabilities)
	if err != nil {
		if !errors.Is(err, mcm.ErrNoMatchingMachineClass) && !errors.Is(err, mcm.ErrAmbiguousMatchingMachineClass) {
			return false, fmt.Errorf("error getting machine class for %s: %w", machineClass.Name, err)
		}
		return false, nil
	}
	return true, nil
}

func (r *MachinePoolReconciler) updateResources(log logr.Logger, status *computev1alpha1.MachinePoolStatus, poolInfo *ori.PoolInfoResponse, machines []computev1alpha1.Machine, machineClasses []computev1alpha1.MachineClass) {
	log.V(2).Info("Updating capacity resources")
	status.Capacity = map[corev1alpha1.ResourceName]resource.Quantity{
		corev1alpha1.ResourceCPU:          *resource.NewQuantity(poolInfo.StaticCpu, resource.DecimalSI),
		corev1alpha1.SharedResourceCPU:    *resource.NewQuantity(poolInfo.SharedCpu, resource.DecimalSI),
		corev1alpha1.ResourceMemory:       *resource.NewQuantity(int64(poolInfo.StaticMemory), resource.BinarySI),
		corev1alpha1.SharedResourceMemory: *resource.NewQuantity(int64(poolInfo.SharedMemory), resource.BinarySI),
	}

	if len(machines) == 0 {
		status.Allocatable = status.Capacity
		return
	}

	log.V(2).Info("Updating allocatable resources")
	machineClassCount := map[string]int{}
	for _, onmetalMachine := range machines {
		machineClassCount[onmetalMachine.Spec.MachineClassRef.Name]++
	}

	var (
		sharedCPU, staticCPU       int64
		sharedMemory, staticMemory uint64
	)
	for _, machineClass := range machineClasses {
		count, found := machineClassCount[machineClass.Name]
		if !found {
			continue
		}

		switch machineClass.Mode {
		case computev1alpha1.ModeDistinct:
			staticCPU += machineClass.Capabilities.CPU().AsDec().UnscaledBig().Int64() * int64(count)
			staticMemory += machineClass.Capabilities.Memory().AsDec().UnscaledBig().Uint64() * uint64(count)
		case computev1alpha1.ModeShared:
			sharedCPU += machineClass.Capabilities.CPU().AsDec().UnscaledBig().Int64() * int64(count)
			sharedMemory += machineClass.Capabilities.Memory().AsDec().UnscaledBig().Uint64() * uint64(count)
		default:
			log.V(1).Info("Unknown machine class mode", "mode", machineClass.Mode)
		}
	}

	status.Allocatable = map[corev1alpha1.ResourceName]resource.Quantity{
		corev1alpha1.ResourceCPU:    *resource.NewQuantity(Max(0, poolInfo.StaticCpu-staticCPU), resource.DecimalSI),
		corev1alpha1.ResourceMemory: *resource.NewQuantity(Max(0, int64(poolInfo.StaticMemory-staticMemory)), resource.BinarySI),

		corev1alpha1.SharedResourceCPU:    *resource.NewQuantity(Max(0, poolInfo.SharedCpu-sharedCPU), resource.DecimalSI),
		corev1alpha1.SharedResourceMemory: *resource.NewQuantity(Max(0, int64(poolInfo.SharedMemory-sharedMemory)), resource.BinarySI),
	}

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
	machinesInPool := &computev1alpha1.MachineList{}
	if err := r.List(ctx, machinesInPool, client.MatchingFields{
		machinepoolletclient.MachineMachinePoolRefNameField: r.MachinePoolName,
	}); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to list machines in pool: %w", err)
	}

	log.V(1).Info("Fetching pool info")
	poolInfo, err := r.MachineRuntime.PoolInfo(ctx, &ori.PoolInfoRequest{})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to fetch pool info: %w", err)
	}

	log.V(1).Info("Determining supported machine classes")
	var supported []corev1.LocalObjectReference
	for _, machineClass := range machineClassList.Items {
		ok, err := r.supportsMachineClass(ctx, log, &machineClass)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("error checking whether machine class %s is supported: %w", machineClass.Name, err)
		}
		if !ok {
			continue
		}

		supported = append(supported, corev1.LocalObjectReference{Name: machineClass.Name})
	}

	log.V(1).Info("Updating machine pool status")
	base := machinePool.DeepCopy()
	machinePool.Status.State = computev1alpha1.MachinePoolStateReady
	machinePool.Status.AvailableMachineClasses = supported
	machinePool.Status.Addresses = r.Addresses
	machinePool.Status.DaemonEndpoints.MachinepoolletEndpoint.Port = r.Port
	r.updateResources(log, &machinePool.Status, poolInfo, machinesInPool.Items, machineClassList.Items)

	if err := r.Status().Patch(ctx, machinePool, client.MergeFrom(base)); err != nil {
		return ctrl.Result{}, fmt.Errorf("error patching machine pool status: %w", err)
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
