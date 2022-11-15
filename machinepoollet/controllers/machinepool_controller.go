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
	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	"github.com/onmetal/onmetal-api/machinepoollet/mcm"
	ori "github.com/onmetal/onmetal-api/ori/apis/compute/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type MachinePoolReconciler struct {
	client.Client
	MachinePoolName    string
	MachineRuntime     ori.MachineRuntimeClient
	MachineClassMapper mcm.MachineClassMapper
}

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
		if !errors.Is(err, mcm.ErrNoMatchingMachineClass) || !errors.Is(err, mcm.ErrAmbiguousMatchingMachineClass) {
			return false, fmt.Errorf("error getting machine class for %s: %w", machineClass.Name, err)
		}
		return false, nil
	}
	return true, nil
}

func (r *MachinePoolReconciler) reconcile(ctx context.Context, log logr.Logger, machinePool *computev1alpha1.MachinePool) (ctrl.Result, error) {
	log.V(1).Info("Reconcile")

	log.V(1).Info("Listing machine classes")
	machineClassList := &computev1alpha1.MachineClassList{}
	if err := r.List(ctx, machineClassList); err != nil {
		return ctrl.Result{}, fmt.Errorf("error listing machine classes: %w", err)
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
	machinePool.Status.AvailableMachineClasses = supported
	if err := r.Status().Patch(ctx, machinePool, client.MergeFrom(base)); err != nil {
		return ctrl.Result{}, fmt.Errorf("error patchign machine pool status: %w", err)
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
			&source.Kind{Type: &computev1alpha1.MachineClass{}},
			handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
				return []ctrl.Request{{NamespacedName: client.ObjectKey{Name: r.MachinePoolName}}}
			}),
		).
		Complete(r)
}
