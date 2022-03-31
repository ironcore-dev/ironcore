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

	"github.com/go-logr/logr"
	"github.com/onmetal/controller-utils/clientutils"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	storagev1alpha1 "github.com/onmetal/onmetal-api/apis/storage/v1alpha1"
)

// StorageClassReconciler reconciles a StorageClass object
type StorageClassReconciler struct {
	client.Client
	APIReader client.Reader
	Scheme    *runtime.Scheme
}

//+kubebuilder:rbac:groups=storage.onmetal.de,resources=storageclasses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=storage.onmetal.de,resources=storageclasses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=storage.onmetal.de,resources=storageclasses/finalizers,verbs=update
//+kubebuilder:rbac:groups=storage.onmetal.de,resources=volumes,verbs=get;list;watch

// Reconcile moves the current state of the cluster closer to the desired state.
func (r *StorageClassReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	storageClass := &storagev1alpha1.StorageClass{}
	if err := r.Get(ctx, req.NamespacedName, storageClass); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileExists(ctx, log, storageClass)
}

func (r *StorageClassReconciler) listReferencingVolumes(ctx context.Context, storageClass *storagev1alpha1.StorageClass) ([]storagev1alpha1.Volume, error) {
	volumeList := &storagev1alpha1.VolumeList{}
	if err := r.APIReader.List(ctx, volumeList, client.InNamespace(storageClass.Namespace)); err != nil {
		return nil, fmt.Errorf("error listing the volumes using the storage class: %w", err)
	}

	var volumes []storagev1alpha1.Volume
	for _, volume := range volumeList.Items {
		if volume.Spec.StorageClassRef.Name == storageClass.Name {
			volumes = append(volumes, volume)
		}
	}
	return volumes, nil
}

func (r *StorageClassReconciler) delete(ctx context.Context, log logr.Logger, storageClass *storagev1alpha1.StorageClass) (ctrl.Result, error) {
	if !controllerutil.ContainsFinalizer(storageClass, storagev1alpha1.StorageClassFinalizer) {
		return ctrl.Result{}, nil
	}

	volumes, err := r.listReferencingVolumes(ctx, storageClass)
	if err != nil {
		return ctrl.Result{}, err
	}

	if len(volumes) != 0 {
		volumeNames := make([]string, 0, len(volumes))
		for _, volume := range volumes {
			volumeNames = append(volumeNames, volume.Name)
		}

		log.Info("Storage class is still in use", "ReferencingVolumeNames", volumeNames)
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Storage class is not used anymore, removing finalizer")
	if err := clientutils.PatchRemoveFinalizer(ctx, r.Client, storageClass, storagev1alpha1.StorageClassFinalizer); err != nil {
		return ctrl.Result{}, err
	}

	log.V(1).Info("Successfully removed finalizer")
	return ctrl.Result{}, nil
}

func (r *StorageClassReconciler) reconcile(ctx context.Context, log logr.Logger, storageClass *storagev1alpha1.StorageClass) (ctrl.Result, error) {
	log.V(1).Info("Ensuring finalizer")
	if modified, err := clientutils.PatchEnsureFinalizer(ctx, r.Client, storageClass, storagev1alpha1.StorageClassFinalizer); err != nil || modified {
		return ctrl.Result{}, err
	}

	log.V(1).Info("Finalizer is present")
	return ctrl.Result{}, nil
}

func (r *StorageClassReconciler) reconcileExists(ctx context.Context, log logr.Logger, storageClass *storagev1alpha1.StorageClass) (ctrl.Result, error) {
	if !storageClass.DeletionTimestamp.IsZero() {
		return r.delete(ctx, log, storageClass)
	}
	return r.reconcile(ctx, log, storageClass)
}

// SetupWithManager sets up the controller with the Manager.
func (r *StorageClassReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&storagev1alpha1.StorageClass{}).
		Watches(
			&source.Kind{Type: &storagev1alpha1.Volume{}},
			handler.Funcs{
				DeleteFunc: func(event event.DeleteEvent, queue workqueue.RateLimitingInterface) {
					volume := event.Object.(*storagev1alpha1.Volume)
					queue.Add(ctrl.Request{NamespacedName: types.NamespacedName{Name: volume.Spec.StorageClassRef.Name}})
				},
			},
		).
		Complete(r)
}
