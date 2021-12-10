/*
 * Copyright (c) 2021 by the OnMetal authors.
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

package compute

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
)

// MachineClassReconciler reconciles a MachineClass object
type MachineClassReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=compute.onmetal.de,resources=machineclasses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=compute.onmetal.de,resources=machineclasses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=compute.onmetal.de,resources=machineclasses/finalizers,verbs=update

// Reconcile moves the current state of the cluster closer to the desired state
func (r *MachineClassReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	machineClass := &computev1alpha1.MachineClass{}
	if err := r.Get(ctx, req.NamespacedName, machineClass); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileExists(ctx, log, machineClass)
}

// SetupWithManager sets up the controller with the Manager.
func (r *MachineClassReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Index the field of machineclass name for listing machines in machineclass controller
	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&computev1alpha1.Machine{},
		machineClassNameField,
		func(object client.Object) []string {
			m := object.(*computev1alpha1.Machine)
			if m.Spec.MachineClass.Name == "" {
				return nil
			}
			return []string{m.Spec.MachineClass.Name}
		},
	); err != nil {
		return fmt.Errorf("indexing the field %s: %w", machineClassNameField, err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&computev1alpha1.MachineClass{}).
		Watches(
			&source.Kind{Type: &computev1alpha1.Machine{}},
			handler.Funcs{
				DeleteFunc: func(e event.DeleteEvent, q workqueue.RateLimitingInterface) {
					m := e.Object.(*computev1alpha1.Machine)
					q.Add(ctrl.Request{NamespacedName: types.NamespacedName{Name: m.Spec.MachineClass.Name}})
				},
			},
		).
		Complete(r)
}

func (r *MachineClassReconciler) delete(ctx context.Context, log logr.Logger, machineClass *computev1alpha1.MachineClass) (ctrl.Result, error) {
	if !controllerutil.ContainsFinalizer(machineClass, computev1alpha1.MachineClassFinalizer) {
		return ctrl.Result{}, nil
	}

	// List the machines currently using the MachineClass
	machineList := &computev1alpha1.MachineList{}
	if err := r.List(ctx, machineList, client.InNamespace(machineClass.Namespace), client.MatchingFields{machineClassNameField: machineClass.Name}); err != nil {
		return ctrl.Result{}, fmt.Errorf("listing the machines using the MachineClass: %w", err)
	}

	// Check if there's still any machine using the MachineClass
	if len(machineList.Items) != 0 {
		// List the names of the machines still using the machineclass
		machineNames := make([]string, 0, len(machineList.Items))
		for _, machine := range machineList.Items {
			machineNames = append(machineNames, machine.Name)
		}
		log.Info("Forbidden to delete the machineclass which is still used by machines", "machine names", machineNames)

		return ctrl.Result{}, nil
	}

	// Remove the finalizer in the machineclass and persist the new state
	old := machineClass.DeepCopy()
	controllerutil.RemoveFinalizer(machineClass, computev1alpha1.MachineClassFinalizer)
	if err := r.Patch(ctx, machineClass, client.MergeFrom(old)); err != nil {
		return ctrl.Result{}, fmt.Errorf("removing the finalizer: %w", err)
	}
	log.V(1).Info("Successfully removed finalizer")

	return ctrl.Result{}, nil
}

func (r *MachineClassReconciler) reconcile(ctx context.Context, log logr.Logger, machineClass *computev1alpha1.MachineClass) (ctrl.Result, error) {
	if !controllerutil.ContainsFinalizer(machineClass, computev1alpha1.MachineClassFinalizer) {
		old := machineClass.DeepCopy()
		controllerutil.AddFinalizer(machineClass, computev1alpha1.MachineClassFinalizer)
		if err := r.Patch(ctx, machineClass, client.MergeFrom(old)); err != nil {
			return ctrl.Result{}, fmt.Errorf("adding the finalizer: %w", err)
		}
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *MachineClassReconciler) reconcileExists(ctx context.Context, log logr.Logger, machineClass *computev1alpha1.MachineClass) (ctrl.Result, error) {
	if !machineClass.DeletionTimestamp.IsZero() {
		return r.delete(ctx, log, machineClass)
	}
	return r.reconcile(ctx, log, machineClass)
}
