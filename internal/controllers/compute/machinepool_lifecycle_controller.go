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
	probeTime      time.Time
	readyCondition *computev1alpha1.MachinePoolCondition
	leaseRenewTime time.Time
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
	machinePoolHealth := r.getMachinePoolHealth(machinePool.Name)
	prevReadyCondition, prevLeaseRenewTime := getPreviousHealthValues(machinePoolHealth)

	currentReadyCondition := FindMachinePoolCondition(machinePool.Status.Conditions, computev1alpha1.MachinePoolReady)
	currentLeaseRenewTime, err := r.getCurrentLeaseRenewTime(ctx, machinePool.Name)
	if err != nil {
		return ctrl.Result{}, err
	}

	changed := !ptr.Equal(prevLeaseRenewTime, currentLeaseRenewTime) || !equality.Semantic.DeepEqual(prevReadyCondition, currentReadyCondition)

	switch {
	case machinePoolHealth == nil && !changed:
		log.V(1).Info("First observation of machine pool, no prior health data or changes")
		r.setMachinePoolHealth(machinePool.Name, &MachinePoolHealth{
			probeTime: time.Now(),
		})
	case machinePoolHealth == nil && changed:
		log.V(1).Info("First observation of machine pool with existing lease/condition data")
		r.setMachinePoolHealth(machinePool.Name, &MachinePoolHealth{
			probeTime:      time.Now(),
			readyCondition: currentReadyCondition,
			leaseRenewTime: ptr.Deref(currentLeaseRenewTime, time.Time{}),
		})
	case changed:
		log.V(1).Info("Lease or ready condition changed, updating probe time")
		r.setMachinePoolHealth(machinePool.Name, &MachinePoolHealth{
			probeTime:      time.Now(),
			readyCondition: currentReadyCondition,
			leaseRenewTime: ptr.Deref(currentLeaseRenewTime, time.Time{}),
		})
	case !changed && currentReadyCondition != nil && currentReadyCondition.Status == corev1.ConditionUnknown:
		log.V(1).Info("No change but ready condition already unknown, refreshing probe time")
		r.setMachinePoolHealth(machinePool.Name, &MachinePoolHealth{
			probeTime:      time.Now(),
			readyCondition: currentReadyCondition,
			leaseRenewTime: ptr.Deref(currentLeaseRenewTime, time.Time{}),
		})

	case !changed:
		if time.Since(machinePoolHealth.probeTime) > r.GracePeriod {
			log.Info("Grace period exceeded without health update, marking machine pool status unknown", "gracePeriod", r.GracePeriod, "lastProbe", machinePoolHealth.probeTime)
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
				//TODO in this case we don't update the probe time, we probably should probably update it here
				// as well or store the probetime at the beginning or the end (and adjust where we use it for comparisons)
				return ctrl.Result{}, fmt.Errorf("error patching: %w", err)
			}

			r.setMachinePoolHealth(machinePool.Name, &MachinePoolHealth{
				probeTime:      time.Now(),
				readyCondition: &newReadyCondition,
				leaseRenewTime: ptr.Deref(currentLeaseRenewTime, time.Time{}),
			})
		} else {
			log.V(1).Info("No change, still within grace period", "gracePeriod", r.GracePeriod, "elapsed", time.Since(machinePoolHealth.probeTime))
			r.setMachinePoolHealth(machinePool.Name, &MachinePoolHealth{
				probeTime:      time.Now(),
				readyCondition: currentReadyCondition,
				leaseRenewTime: ptr.Deref(currentLeaseRenewTime, time.Time{}),
			})
		}
	}

	currentHealth := r.getMachinePoolHealth(machinePool.Name)
	// requeue when this pool's grace period runs out, but never sooner than 50ms from now.
	return ctrl.Result{RequeueAfter: max(50*time.Millisecond, time.Until(currentHealth.probeTime.Add(r.GracePeriod)))}, nil
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
