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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
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

var errStorageClassDeletionForbidden = errors.New("forbidden to delete the storageclass used by a volume")

// StorageClassReconciler reconciles a StorageClass object
type StorageClassReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Events record.EventRecorder
}

//+kubebuilder:rbac:groups=storage.onmetal.de,resources=storageclasses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=storage.onmetal.de,resources=storageclasses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=storage.onmetal.de,resources=storageclasses/finalizers,verbs=update

// Reconcile moves the current state of the cluster closer to the desired state.
func (r *StorageClassReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	sc := &storagev1alpha1.StorageClass{}
	if err := r.Get(ctx, req.NamespacedName, sc); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !controllerutil.ContainsFinalizer(sc, storagev1alpha1.StorageClassFinalizer) {
		old := sc.DeepCopy()
		controllerutil.AddFinalizer(sc, storagev1alpha1.StorageClassFinalizer)
		if err := r.Patch(ctx, sc, client.MergeFrom(old)); err != nil {
			return ctrl.Result{}, fmt.Errorf("adding the finalizer: %w", err)
		}

		// Requeue since the storageclass can be simultaneously updated by multiple parties
		return ctrl.Result{Requeue: true}, nil
	}

	// return ctrl.Result{}, nil
	return r.reconcileExists(ctx, log, sc)
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

func (r *StorageClassReconciler) delete(ctx context.Context, log logr.Logger, sc *storagev1alpha1.StorageClass) (ctrl.Result, error) {
	// List the volumes currently using the storageclass
	vList := &storagev1alpha1.VolumeList{}
	if err := r.List(ctx, vList, client.InNamespace(sc.Namespace), client.MatchingFields{storageClassNameField: sc.Name}); err != nil {
		return ctrl.Result{}, fmt.Errorf("listing the volumes using the storageclass: %w", err)
	}

	// Check if there's still any volume using the storageclass
	if vv := vList.Items; len(vv) != 0 {
		// List the volume names still using the storageclass in the error message
		volumeNames := ""
		for i := range vv {
			volumeNames += vv[i].Name + ", "
		}
		err := errors.New(fmt.Sprintf("the following volumes still using the volumeclass: %s", volumeNames))

		log.Error(err, "Forbidden to delete the volumeclass which is still used by volumes")
		r.Events.Eventf(sc, corev1.EventTypeWarning, "ForbiddenToDelete", err.Error())
		return ctrl.Result{}, nil
	}

	// Remove the finalizer in the storageclass and persist the new state
	old := sc.DeepCopy()
	controllerutil.RemoveFinalizer(sc, storagev1alpha1.StorageClassFinalizer)
	if err := r.Patch(ctx, sc, client.MergeFrom(old)); err != nil {
		return ctrl.Result{}, fmt.Errorf("removing the finalizer: %w", err)
	}
	return ctrl.Result{}, nil
}

func (r *StorageClassReconciler) reconcile(ctx context.Context, log logr.Logger, sc *storagev1alpha1.StorageClass) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (r *StorageClassReconciler) reconcileExists(ctx context.Context, log logr.Logger, sc *storagev1alpha1.StorageClass) (ctrl.Result, error) {
	if !sc.DeletionTimestamp.IsZero() {
		return r.delete(ctx, log, sc)
	}
	return r.reconcile(ctx, log, sc)
}
