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
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/go-logr/logr"

	storagev1alpha1 "github.com/onmetal/onmetal-api/apis/storage/v1alpha1"
)

const storageClassNameField = ".spec.storageclass.name"

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
	log := ctrl.LoggerFrom(ctx)
	storageclass := &storagev1alpha1.StorageClass{}
	if err := r.Get(ctx, req.NamespacedName, storageclass); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileExists(ctx, log, storageclass)
}

// SetupWithManager sets up the controller with the Manager.
func (r *StorageClassReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Index the field of storageclass name for listing volumes in storageclass controller
	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&storagev1alpha1.Volume{},
		storageClassNameField,
		func(object client.Object) []string {
			m := object.(*storagev1alpha1.Volume)
			if m.Spec.StorageClass.Name == "" {
				return nil
			}
			return []string{m.Spec.StorageClass.Name}
		},
	); err != nil {
		return fmt.Errorf("indexing the field %s: %w", storageClassNameField, err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&storagev1alpha1.StorageClass{}).
		Watches(
			&source.Kind{Type: &storagev1alpha1.Volume{}},
			handler.Funcs{
				DeleteFunc: func(e event.DeleteEvent, q workqueue.RateLimitingInterface) {
					v := e.Object.(*storagev1alpha1.Volume)
					q.Add(ctrl.Request{NamespacedName: types.NamespacedName{Name: v.Spec.StorageClass.Name}})
				},
			},
		).
		Complete(r)
}

func (r *StorageClassReconciler) delete(ctx context.Context, log logr.Logger, storageclass *storagev1alpha1.StorageClass) (ctrl.Result, error) {
	// List the volumes currently using the storageclass
	volumeList := &storagev1alpha1.VolumeList{}
	if err := r.List(ctx, volumeList, client.InNamespace(storageclass.Namespace), client.MatchingFields{storageClassNameField: storageclass.Name}); err != nil {
		return ctrl.Result{}, fmt.Errorf("listing the volumes using the storageclass: %w", err)
	}

	// Check if there's still any volume using the storageclass
	if len(volumeList.Items) != 0 {
		// List the names of the volumes still using the storageclass
		volumeNames := make([]string, 0, len(volumeList.Items))
		for _, volume := range volumeList.Items {
			volumeNames = append(volumeNames, volume.Name)
		}
		log.Info("Forbidden to delete the volumeclass which is still used by volumes", "volume names", volumeNames)

		return ctrl.Result{}, nil
	}

	// Remove the finalizer in the storageclass and persist the new state
	old := storageclass.DeepCopy()
	controllerutil.RemoveFinalizer(storageclass, storagev1alpha1.StorageClassFinalizer)
	if err := r.Patch(ctx, storageclass, client.MergeFrom(old)); err != nil {
		return ctrl.Result{}, fmt.Errorf("removing the finalizer: %w", err)
	}
	log.V(1).Info("Successfully removed finalizer")

	return ctrl.Result{}, nil
}

func (r *StorageClassReconciler) reconcile(ctx context.Context, log logr.Logger, storageclass *storagev1alpha1.StorageClass) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (r *StorageClassReconciler) reconcileExists(ctx context.Context, log logr.Logger, storageclass *storagev1alpha1.StorageClass) (ctrl.Result, error) {
	if !storageclass.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(storageclass, storagev1alpha1.StorageClassFinalizer) {
			return ctrl.Result{}, nil
		}
		return r.delete(ctx, log, storageclass)
	}

	if !controllerutil.ContainsFinalizer(storageclass, storagev1alpha1.StorageClassFinalizer) {
		old := storageclass.DeepCopy()
		controllerutil.AddFinalizer(storageclass, storagev1alpha1.StorageClassFinalizer)
		if err := r.Patch(ctx, storageclass, client.MergeFrom(old)); err != nil {
			return ctrl.Result{}, fmt.Errorf("adding the finalizer: %w", err)
		}
		return ctrl.Result{}, nil
	}

	return r.reconcile(ctx, log, storageclass)
}
