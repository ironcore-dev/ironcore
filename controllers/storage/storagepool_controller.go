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
	"time"

	corev1 "k8s.io/api/core/v1"

	"github.com/go-logr/logr"
	"github.com/onmetal/controller-utils/conditionutils"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	storagev1alpha1 "github.com/onmetal/onmetal-api/apis/storage/v1alpha1"
)

// StoragePoolReconciler reconciles a StoragePool object
type StoragePoolReconciler struct {
	client.Client
	Scheme                 *runtime.Scheme
	StoragePoolGracePeriod time.Duration
}

//+kubebuilder:rbac:groups=storage.onmetal.de,resources=storagepools,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=storage.onmetal.de,resources=storagepools/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=storage.onmetal.de,resources=storagepools/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *StoragePoolReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	pool := &storagev1alpha1.StoragePool{}
	if err := r.Get(ctx, req.NamespacedName, pool); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileExists(ctx, log, pool)
}

func (r *StoragePoolReconciler) reconcileExists(ctx context.Context, log logr.Logger, pool *storagev1alpha1.StoragePool) (ctrl.Result, error) {
	cond := &storagev1alpha1.StoragePoolCondition{}
	ok := conditionutils.MustFindSlice(pool.Status.Conditions, string(storagev1alpha1.StoragePoolConditionTypeReady), cond)
	if !ok {
		log.Info("Didn't found ready condition for StoragePool")
	}

	outdatedPool := pool.DeepCopy()
	switch cond.Status {
	case corev1.ConditionTrue:
		if cond.LastUpdateTime.Add(r.StoragePoolGracePeriod).After(time.Now()) {
			pool.Status.State = storagev1alpha1.StoragePoolStateAvailable
		} else {
			pool.Status.State = storagev1alpha1.StoragePoolStatePending
		}
	default:
		pool.Status.State = storagev1alpha1.StoragePoolStatePending
	}

	if err := r.Status().Patch(ctx, pool, client.MergeFrom(outdatedPool)); err != nil {
		return ctrl.Result{}, fmt.Errorf("could not update status: %w", err)
	}

	return ctrl.Result{RequeueAfter: r.StoragePoolGracePeriod}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *StoragePoolReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&storagev1alpha1.StoragePool{}).
		Complete(r)
}
