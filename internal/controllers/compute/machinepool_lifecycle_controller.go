// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package compute

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/go-logr/logr"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	"github.com/ironcore-dev/ironcore/utils/equality"
	coordinationv1 "k8s.io/api/coordination/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type MachinePoolLifecycleReconciler struct {
	client.Client
	GracePeriod time.Duration

	healthDataMu sync.RWMutex
	healthData   map[string]*MachinePoolHealth
}

type MachinePoolHealth struct {
	lastObservedTime       time.Time
	lastChangeDetectedTime time.Time
	readyCondition         *computev1alpha1.MachinePoolCondition
	leaseRenewTime         time.Time
}

func (r *MachinePoolLifecycleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	machinePool := &computev1alpha1.MachinePool{}
	if err := r.Get(ctx, req.NamespacedName, machinePool); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileExists(ctx, log, machinePool)
}

func (r *MachinePoolLifecycleReconciler) getMachinePoolHealth(machinePoolName string) *MachinePoolHealth {
	r.healthDataMu.RLock()
	defer r.healthDataMu.RUnlock()

	return r.healthData[machinePoolName]
}

func (r *MachinePoolLifecycleReconciler) setMachinePoolHealth(machinePoolName string, machinePoolHealh *MachinePoolHealth) {
	r.healthDataMu.Lock()
	defer r.healthDataMu.Unlock()

	if r.healthData == nil {
		r.healthData = make(map[string]*MachinePoolHealth)
	}

	r.healthData[machinePoolName] = machinePoolHealh
}

func FindMachinePoolCondition(conditions []computev1alpha1.MachinePoolCondition, typ computev1alpha1.MachinePoolConditionType) *computev1alpha1.MachinePoolCondition {
	idx := slices.IndexFunc(conditions, func(cond computev1alpha1.MachinePoolCondition) bool {
		return cond.Type == typ
	})

	if idx < 0 {
		return nil
	}

	return &conditions[idx]
}

func SetMachinePoolCondition(conditions []computev1alpha1.MachinePoolCondition, cond computev1alpha1.MachinePoolCondition) []computev1alpha1.MachinePoolCondition {
	idx := slices.IndexFunc(conditions, func(c computev1alpha1.MachinePoolCondition) bool {
		return c.Type == cond.Type
	})

	cond.LastUpdateTime = metav1.Now()

	if idx < 0 || conditions[idx].Status != cond.Status {
		cond.LastTransitionTime = metav1.Now()
	}

	if idx < 0 {
		return append(conditions, cond)
	}
	conditions[idx] = cond
	return conditions
}

func getPreviousHealthValues(machinePoolHealth *MachinePoolHealth) (*computev1alpha1.MachinePoolCondition, *time.Time) {
	var (
		prevReadyCondition *computev1alpha1.MachinePoolCondition
		prevLeaseRenewTime *time.Time
	)
	if machinePoolHealth != nil {
		prevReadyCondition = machinePoolHealth.readyCondition
		if !machinePoolHealth.leaseRenewTime.IsZero() {
			prevLeaseRenewTime = &machinePoolHealth.leaseRenewTime
		}
	}
	return prevReadyCondition, prevLeaseRenewTime
}

func (r *MachinePoolLifecycleReconciler) getCurrentLeaseRenewTime(ctx context.Context, machinePoolName string) (*time.Time, error) {
	lease := &coordinationv1.Lease{}
	if err := r.Get(ctx, client.ObjectKey{
		Namespace: computev1alpha1.NamespaceMachinePoolLease,
		Name:      machinePoolName,
	}, lease); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("getting machine pool lease: %w", err)
		}
		return nil, nil
	}
	if lease.Spec.RenewTime != nil {
		return ptr.To(lease.Spec.RenewTime.Time), nil
	}
	return nil, nil
}

func (r *MachinePoolLifecycleReconciler) reconcileExists(ctx context.Context, log logr.Logger, machinePool *computev1alpha1.MachinePool) (ctrl.Result, error) {
	now := time.Now()
	prev := r.getMachinePoolHealth(machinePool.Name)
	prevReadyCondition, prevLeaseRenewTime := getPreviousHealthValues(prev)

	currentReadyCondition := FindMachinePoolCondition(machinePool.Status.Conditions, computev1alpha1.MachinePoolReady)
	currentLeaseRenewTime, err := r.getCurrentLeaseRenewTime(ctx, machinePool.Name)
	if err != nil {
		return ctrl.Result{}, err
	}

	changed := !ptr.Equal(prevLeaseRenewTime, currentLeaseRenewTime) || !equality.Semantic.DeepEqual(prevReadyCondition, currentReadyCondition)

	next := &MachinePoolHealth{
		lastObservedTime: now,
		readyCondition:   currentReadyCondition,
		leaseRenewTime:   ptr.Deref(currentLeaseRenewTime, time.Time{}),
	}

	switch {
	case prev == nil:
		log.V(1).Info("First observation of machine pool")
		next.lastChangeDetectedTime = now
	case changed:
		log.V(1).Info("Lease or ready condition changed")
		next.lastChangeDetectedTime = now
	default:
		next.lastChangeDetectedTime = prev.lastChangeDetectedTime

		if time.Since(prev.lastChangeDetectedTime) > r.GracePeriod {
			if currentReadyCondition != nil && currentReadyCondition.Status == corev1.ConditionUnknown {
				log.V(1).Info("Grace period exceeded, ready condition already unknown — no patch needed",
					"gracePeriod", r.GracePeriod, "lastChangeDetected", prev.lastChangeDetectedTime)
			} else {
				log.Info("Grace period exceeded without health update, marking machine pool status unknown",
					"gracePeriod", r.GracePeriod, "lastChangeDetected", prev.lastChangeDetectedTime)
				patch := client.StrategicMergeFrom(machinePool.DeepCopy())
				newReadyCondition := computev1alpha1.MachinePoolCondition{
					Type:               computev1alpha1.MachinePoolReady,
					Status:             corev1.ConditionUnknown,
					Reason:             "MachinePoolStatusUnknown",
					Message:            "machinepoollet stopped posting machine pool status.",
					ObservedGeneration: machinePool.Generation,
				}
				machinePool.Status.Conditions = SetMachinePoolCondition(machinePool.Status.Conditions, newReadyCondition)

				if err := r.Status().Patch(ctx, machinePool, patch); err != nil {
					// On patch failure, leave health state untouched so the next reconcile retries.
					return ctrl.Result{}, fmt.Errorf("error patching: %w", err)
				}
				next.readyCondition = &newReadyCondition
			}
		} else {
			log.V(1).Info("No change, still within grace period",
				"gracePeriod", r.GracePeriod, "elapsed", time.Since(prev.lastChangeDetectedTime))
		}
	}

	r.setMachinePoolHealth(machinePool.Name, next)

	// requeue when this pool's grace period runs out, but never sooner than 50ms from now.
	return ctrl.Result{RequeueAfter: max(50*time.Millisecond, time.Until(next.lastChangeDetectedTime.Add(r.GracePeriod)))}, nil
}

func (r *MachinePoolLifecycleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("MachinePoolLifecycle").
		For(&computev1alpha1.MachinePool{}).
		Watches(
			&coordinationv1.Lease{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []ctrl.Request {
				return []ctrl.Request{{NamespacedName: client.ObjectKey{Name: obj.GetName()}}}
			}),
			builder.WithPredicates(predicate.NewPredicateFuncs(func(obj client.Object) bool {
				return obj.GetNamespace() == computev1alpha1.NamespaceMachinePoolLease
			})),
		).
		Complete(r)
}
