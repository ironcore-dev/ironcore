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

package storage

import (
	"context"
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	util "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	storagev1alpha1 "github.com/onmetal/onmetal-api/apis/storage/v1alpha1"
)

// StorageClassReconciler reconciles a StorageClass object
type StorageClassReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=storage.onmetal.de,resources=storageclasses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=storage.onmetal.de,resources=storageclasses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=storage.onmetal.de,resources=storageclasses/finalizers,verbs=update

// Reconcile moves the current state of the cluster closer to the desired state.
func (r *StorageClassReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	sc := &storagev1alpha1.StorageClass{}
	if err := r.Get(ctx, req.NamespacedName, sc); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := r.addFinalizerIfNone(ctx, sc); err != nil {
		return ctrl.Result{}, fmt.Errorf("adding the finalizer if none: %w", err)
	}

	if !sc.DeletionTimestamp.IsZero() {
		return r.reconcileDeletion(ctx, sc)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *StorageClassReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&storagev1alpha1.StorageClass{}).
		Complete(r)
}

func (r *StorageClassReconciler) addFinalizerIfNone(ctx context.Context, sc *storagev1alpha1.StorageClass) error {
	if !util.ContainsFinalizer(sc, storagev1alpha1.StorageClassFinalizer) {
		old := sc.DeepCopy()
		util.AddFinalizer(sc, storagev1alpha1.StorageClassFinalizer)
		if err := r.Patch(ctx, sc, client.MergeFrom(old)); err != nil {
			return fmt.Errorf("adding the finalizer: %w", err)
		}
	}
	return nil
}

func (r *StorageClassReconciler) reconcileDeletion(ctx context.Context, sc *storagev1alpha1.StorageClass) (ctrl.Result, error) {
	// List the volumes currently using the storageclass
	vList := &storagev1alpha1.VolumeList{}
	if err := r.List(ctx, vList, client.InNamespace(sc.Namespace), client.MatchingFields{storageClassNameField: sc.Name}); err != nil {
		return ctrl.Result{}, fmt.Errorf("listing the volumes using the storageclass: %w", err)
	}

	// Check if there's still any volume using the storageclass
	if len(vList.Items) != 0 {
		return ctrl.Result{}, errStorageClassDeletionForbidden
	}

	// Remove the finalizer in the storageclass and persist the new state
	old := sc.DeepCopy()
	util.RemoveFinalizer(sc, storagev1alpha1.StorageClassFinalizer)
	if err := r.Patch(ctx, sc, client.MergeFrom(old)); err != nil {
		return ctrl.Result{}, fmt.Errorf("removing the finalizer: %w", err)
	}
	return ctrl.Result{}, nil
}

var errStorageClassDeletionForbidden = errors.New("forbidden to delete the storageclass used by a volume")
