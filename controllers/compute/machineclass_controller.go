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
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	util "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

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
	mClass := &computev1alpha1.MachineClass{}
	if err := r.Get(ctx, req.NamespacedName, mClass); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := r.addFinalizerIfNone(ctx, mClass); err != nil {
		return ctrl.Result{}, fmt.Errorf("adding the finalizer if none: %w", err)
	}

	if !mClass.DeletionTimestamp.IsZero() {
		return r.reconcileDeletion(ctx, mClass)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MachineClassReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&computev1alpha1.MachineClass{}).
		Complete(r)
}

func (r *MachineClassReconciler) addFinalizerIfNone(ctx context.Context, mClass *computev1alpha1.MachineClass) error {
	if !util.ContainsFinalizer(mClass, computev1alpha1.MachineClassFinalizer) {
		old := mClass.DeepCopy()
		util.AddFinalizer(mClass, computev1alpha1.MachineClassFinalizer)
		if err := r.Patch(ctx, mClass, client.MergeFrom(old)); err != nil {
			return fmt.Errorf("adding the finalizer: %w", err)
		}
	}
	return nil
}

func (r *MachineClassReconciler) reconcileDeletion(ctx context.Context, mClass *computev1alpha1.MachineClass) (ctrl.Result, error) {
	// List the machines currently using the MachineClass
	mList := &computev1alpha1.MachineList{}
	if err := r.List(ctx, mList, client.InNamespace(mClass.Namespace), client.MatchingFields{machineClassNameField: mClass.Name}); err != nil {
		return ctrl.Result{}, fmt.Errorf("listing the machines using the MachineClass: %w", err)
	}

	// Check if there's still any machine using the MachineClass
	if len(mList.Items) != 0 {
		return ctrl.Result{}, errMachineClassDeletionForbidden
	}

	// Remove the finalizer in the machineclass and persist the new state
	old := mClass.DeepCopy()
	util.RemoveFinalizer(mClass, computev1alpha1.MachineClassFinalizer)
	if err := r.Patch(ctx, mClass, client.MergeFrom(old)); err != nil {
		return ctrl.Result{}, fmt.Errorf("removing the finalizer: %w", err)
	}
	return ctrl.Result{}, nil
}

var errMachineClassDeletionForbidden = errors.New("forbidden to delete the machineclass used by a machine")
