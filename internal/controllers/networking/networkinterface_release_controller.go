// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package networking

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/lru"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type NetworkInterfaceReleaseReconciler struct {
	client.Client
	APIReader client.Reader

	AbsenceCache *lru.Cache
}

//+kubebuilder:rbac:groups=networking.ironcore.dev,resources=networkinterfaces,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=compute.ironcore.dev,resources=machines,verbs=get;list;watch

func (r *NetworkInterfaceReleaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	nic := &networkingv1alpha1.NetworkInterface{}
	if err := r.Get(ctx, req.NamespacedName, nic); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileExists(ctx, log, nic)
}

func (r *NetworkInterfaceReleaseReconciler) reconcileExists(ctx context.Context, log logr.Logger, nic *networkingv1alpha1.NetworkInterface) (ctrl.Result, error) {
	if !nic.DeletionTimestamp.IsZero() {
		log.V(1).Info("Network interface is already deleting, nothing to do")
		return ctrl.Result{}, nil
	}

	return r.reconcile(ctx, log, nic)
}

func (r *NetworkInterfaceReleaseReconciler) networkInterfaceClaimExists(ctx context.Context, nic *networkingv1alpha1.NetworkInterface) (bool, error) {
	claimRef := nic.Spec.MachineRef
	if _, ok := r.AbsenceCache.Get(claimRef.UID); ok {
		return false, nil
	}

	claimer := &metav1.PartialObjectMetadata{
		TypeMeta: metav1.TypeMeta{
			APIVersion: computev1alpha1.SchemeGroupVersion.String(),
			Kind:       "Machine",
		},
	}
	claimerKey := client.ObjectKey{Namespace: nic.Namespace, Name: claimRef.Name}
	if err := r.APIReader.Get(ctx, claimerKey, claimer); err != nil {
		if !apierrors.IsNotFound(err) {
			return false, fmt.Errorf("error getting claiming machine %s: %w", claimRef.Name, err)
		}

		r.AbsenceCache.Add(claimRef.UID, nil)
		return false, nil
	}
	return true, nil
}

func (r *NetworkInterfaceReleaseReconciler) releaseNetworkInterface(ctx context.Context, nic *networkingv1alpha1.NetworkInterface) error {
	baseNic := nic.DeepCopy()
	nic.Spec.MachineRef = nil
	if err := r.Patch(ctx, nic, client.StrategicMergeFrom(baseNic, client.MergeFromWithOptimisticLock{})); err != nil {
		return fmt.Errorf("error patching network interface: %w", err)
	}
	return nil
}

func (r *NetworkInterfaceReleaseReconciler) reconcile(ctx context.Context, log logr.Logger, nic *networkingv1alpha1.NetworkInterface) (ctrl.Result, error) {
	log.V(1).Info("Reconcile")

	if nic.Spec.MachineRef == nil {
		log.V(1).Info("Network interface is not claimed, nothing to do")
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Checking whether network interface claimer exists")
	ok, err := r.networkInterfaceClaimExists(ctx, nic)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error checking whether network interface claimer exists: %w", err)
	}
	if ok {
		log.V(1).Info("Network interface claimer is still present")
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Network interface claimer does not exist, releasing network interface")
	if err := r.releaseNetworkInterface(ctx, nic); err != nil {
		if !apierrors.IsConflict(err) {
			return ctrl.Result{}, fmt.Errorf("error releasing network interface: %w", err)
		}
		log.V(1).Info("Network interface was updated, requeueing")
		return ctrl.Result{RequeueAfter: 1}, nil
	}

	log.V(1).Info("Reconciled")
	return ctrl.Result{}, nil
}

func (r *NetworkInterfaceReleaseReconciler) networkInterfaceClaimedPredicate() predicate.Predicate {
	return predicate.NewPredicateFuncs(func(obj client.Object) bool {
		nic := obj.(*networkingv1alpha1.NetworkInterface)
		return nic.Spec.MachineRef != nil
	})
}

func (r *NetworkInterfaceReleaseReconciler) enqueueByMachine() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []ctrl.Request {
		machine := obj.(*computev1alpha1.Machine)
		log := ctrl.LoggerFrom(ctx)

		nicList := &networkingv1alpha1.NetworkInterfaceList{}
		if err := r.List(ctx, nicList,
			client.InNamespace(machine.Namespace),
		); err != nil {
			log.Error(err, "Error listing network interfaces")
			return nil
		}

		var reqs []ctrl.Request
		for _, nic := range nicList.Items {
			claimRef := nic.Spec.MachineRef
			if claimRef == nil {
				continue
			}

			if claimRef.UID != machine.UID {
				continue
			}

			reqs = append(reqs, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(&nic)})
		}
		return reqs
	})
}

func (r *NetworkInterfaceReleaseReconciler) machineDeletingPredicate() predicate.Predicate {
	return predicate.NewPredicateFuncs(func(obj client.Object) bool {
		machine := obj.(*computev1alpha1.Machine)
		return !machine.DeletionTimestamp.IsZero()
	})
}

func (r *NetworkInterfaceReleaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("networkinterfacerelease").
		For(
			&networkingv1alpha1.NetworkInterface{},
			builder.WithPredicates(r.networkInterfaceClaimedPredicate()),
		).
		Watches(
			&computev1alpha1.Machine{},
			r.enqueueByMachine(),
			builder.WithPredicates(r.machineDeletingPredicate()),
		).
		Complete(r)
}
