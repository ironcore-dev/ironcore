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
	"time"

	"github.com/go-logr/logr"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/onmetal/controller-utils/conditionutils"

	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
)

var (
	pendingStateRequeueAfter = 30 * time.Second
)

// MachinePoolReconciler reconciles a MachinePool object
type MachinePoolReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	ReadyDuration time.Duration
}

//+kubebuilder:rbac:groups=compute.onmetal.de,resources=machinepools,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=compute.onmetal.de,resources=machinepools/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=compute.onmetal.de,resources=machinepools/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MachinePool object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *MachinePoolReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	pool := &computev1alpha1.MachinePool{}
	if err := r.Get(ctx, req.NamespacedName, pool); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileExists(ctx, log, pool)
}

func (r *MachinePoolReconciler) reconcileExists(ctx context.Context, log logr.Logger, pool *computev1alpha1.MachinePool) (ctrl.Result, error) {
	cond := &computev1alpha1.MachinePoolCondition{}
	ok, err := conditionutils.FindSlice(pool.Status.Conditions, string(computev1alpha1.MachinePoolConditionTypeReady), cond)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed while searching 'Ready' condition: %w", err)
	}

	outdatedPool := pool.DeepCopy()
	requeueAfter := r.ReadyDuration
	if ok {
		if cond.LastUpdateTime.Add(r.ReadyDuration).After(time.Now()) {
			pool.Status.State = computev1alpha1.MachinePoolStateReady
		} else {
			pool.Status.State = computev1alpha1.MachinePoolStatePending
			requeueAfter = pendingStateRequeueAfter
		}
	} else {
		pool.Status.State = computev1alpha1.MachinePoolStatePending
		requeueAfter = pendingStateRequeueAfter
	}

	if err := r.Status().Patch(ctx, pool, client.MergeFrom(outdatedPool)); err != nil {
		return ctrl.Result{}, fmt.Errorf("could not update status: %w", err)
	}

	return ctrl.Result{RequeueAfter: requeueAfter}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MachinePoolReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&computev1alpha1.MachinePool{}).
		Complete(r)
}
