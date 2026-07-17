// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package compute

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	computeclient "github.com/ironcore-dev/ironcore/internal/client/compute"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/events"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type MachineEvictionReconciler struct {
	events.EventRecorder
	client.Client
}

// +kubebuilder:rbac:groups=compute.ironcore.dev,resources=machines,verbs=get;list;watch;delete
// +kubebuilder:rbac:groups=compute.ironcore.dev,resources=machinepools,verbs=get;list;watch

// Reconcile reconciles the desired with the actual state.
func (r *MachineEvictionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	machine := &computev1alpha1.Machine{}
	if err := r.Get(ctx, req.NamespacedName, machine); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileExists(ctx, log, machine)
}

func (r *MachineEvictionReconciler) reconcileExists(ctx context.Context, log logr.Logger, machine *computev1alpha1.Machine) (ctrl.Result, error) {
	if !machine.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}
	if machine.Spec.MachinePoolRef == nil {
		return ctrl.Result{}, nil
	}

	machinePool := &computev1alpha1.MachinePool{}
	if err := r.Get(ctx, client.ObjectKey{Name: machine.Spec.MachinePoolRef.Name}, machinePool); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if shouldEvict(machinePool.Spec.Taints, machine.Spec.Tolerations) {
		log.V(2).Info("Evicting machine", "machine", machine.Name)
		if err := r.Delete(ctx, machine); err != nil {
			return ctrl.Result{}, fmt.Errorf("error evicting machine: %w", err)
		}
		r.Eventf(machine, nil, corev1.EventTypeNormal, "Evicted", "Eviction",
			"Evicted from MachinePool %s: does not tolerate NoExecute taint", machinePool.Name)
	}

	return ctrl.Result{}, nil
}

func shouldEvict(taints []commonv1alpha1.Taint, tolerations []commonv1alpha1.Toleration) bool {
	for _, taint := range taints {
		if taint.Effect != commonv1alpha1.TaintEffectNoExecute {
			continue
		}

		if commonv1alpha1.ToleratesTaint(tolerations, &taint) {
			continue
		}

		return true
	}

	return false
}

func (r *MachineEvictionReconciler) enqueueMachinesInPool() handler.MapFunc {
	return func(ctx context.Context, obj client.Object) []reconcile.Request {
		log := ctrl.LoggerFrom(ctx)

		machineList := &computev1alpha1.MachineList{}
		if err := r.List(ctx, machineList,
			client.MatchingFields{computeclient.MachineSpecMachinePoolRefNameField: obj.GetName()},
		); err != nil {
			log.Error(err, "Error listing machines bound to pool", "MachinePool", obj.GetName())
			return nil
		}

		reqs := make([]reconcile.Request, 0, len(machineList.Items))
		for i := range machineList.Items {
			reqs = append(reqs, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(&machineList.Items[i])})
		}
		return reqs
	}
}

func (r *MachineEvictionReconciler) isMachineAssigned() predicate.Predicate {
	return predicate.NewPredicateFuncs(func(obj client.Object) bool {
		machine, ok := obj.(*computev1alpha1.Machine)
		if !ok {
			return false
		}

		return machine.Spec.MachinePoolRef != nil
	})
}

func (r *MachineEvictionReconciler) SetupWithManager(mgr manager.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("machine-eviction").
		For(
			&computev1alpha1.Machine{},
			builder.WithPredicates(r.isMachineAssigned()),
		).
		Watches(
			&computev1alpha1.MachinePool{},
			handler.EnqueueRequestsFromMapFunc(r.enqueueMachinesInPool()),
		).
		Complete(r)
}
