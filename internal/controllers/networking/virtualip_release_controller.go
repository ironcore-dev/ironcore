// Copyright 2023 IronCore authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package networking

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
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

type VirtualIPReleaseReconciler struct {
	client.Client
	APIReader client.Reader

	AbsenceCache *lru.Cache
}

//+kubebuilder:rbac:groups=networking.ironcore.dev,resources=virtualips,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=networking.ironcore.dev,resources=networkinterfaces,verbs=get;list;watch

func (r *VirtualIPReleaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	virtualIP := &networkingv1alpha1.VirtualIP{}
	if err := r.Get(ctx, req.NamespacedName, virtualIP); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileExists(ctx, log, virtualIP)
}

func (r *VirtualIPReleaseReconciler) reconcileExists(ctx context.Context, log logr.Logger, virtualIP *networkingv1alpha1.VirtualIP) (ctrl.Result, error) {
	if !virtualIP.DeletionTimestamp.IsZero() {
		log.V(1).Info("Virtual IP is already deleting, nothing to do")
		return ctrl.Result{}, nil
	}

	return r.reconcile(ctx, log, virtualIP)
}

func (r *VirtualIPReleaseReconciler) virtualIPClaimExists(ctx context.Context, virtualIP *networkingv1alpha1.VirtualIP) (bool, error) {
	claimRef := virtualIP.Spec.TargetRef
	if _, ok := r.AbsenceCache.Get(claimRef.UID); ok {
		return false, nil
	}

	claimer := &metav1.PartialObjectMetadata{
		TypeMeta: metav1.TypeMeta{
			APIVersion: networkingv1alpha1.SchemeGroupVersion.String(),
			Kind:       "NetworkInterface",
		},
	}
	claimerKey := client.ObjectKey{Namespace: virtualIP.Namespace, Name: claimRef.Name}
	if err := r.APIReader.Get(ctx, claimerKey, claimer); err != nil {
		if !apierrors.IsNotFound(err) {
			return false, fmt.Errorf("error getting claiming virtual IP %s: %w", claimRef.Name, err)
		}

		r.AbsenceCache.Add(claimRef.UID, nil)
		return false, nil
	}
	return true, nil
}

func (r *VirtualIPReleaseReconciler) releaseVirtualIP(ctx context.Context, virtualIP *networkingv1alpha1.VirtualIP) error {
	baseNic := virtualIP.DeepCopy()
	virtualIP.Spec.TargetRef = nil
	if err := r.Patch(ctx, virtualIP, client.StrategicMergeFrom(baseNic, client.MergeFromWithOptimisticLock{})); err != nil {
		return fmt.Errorf("error patching virtual IP: %w", err)
	}
	return nil
}

func (r *VirtualIPReleaseReconciler) reconcile(ctx context.Context, log logr.Logger, virtualIP *networkingv1alpha1.VirtualIP) (ctrl.Result, error) {
	log.V(1).Info("Reconcile")

	if virtualIP.Spec.TargetRef == nil {
		log.V(1).Info("Virtual IP is not claimed, nothing to do")
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Checking whether virtual IP claimer exists")
	ok, err := r.virtualIPClaimExists(ctx, virtualIP)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error checking whether virtual IP claimer exists: %w", err)
	}
	if ok {
		log.V(1).Info("Virtual IP claimer is still present")
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Virtual IP claimer does not exist, releasing virtual IP")
	if err := r.releaseVirtualIP(ctx, virtualIP); err != nil {
		if !apierrors.IsConflict(err) {
			return ctrl.Result{}, fmt.Errorf("error releasing virtual IP: %w", err)
		}
		log.V(1).Info("Virtual IP was updated, requeueing")
		return ctrl.Result{Requeue: true}, nil
	}

	log.V(1).Info("Reconciled")
	return ctrl.Result{}, nil
}

func (r *VirtualIPReleaseReconciler) virtualIPClaimedPredicate() predicate.Predicate {
	return predicate.NewPredicateFuncs(func(obj client.Object) bool {
		virtualIP := obj.(*networkingv1alpha1.VirtualIP)
		return virtualIP.Spec.TargetRef != nil
	})
}

func (r *VirtualIPReleaseReconciler) enqueueByNetworkInterface() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []ctrl.Request {
		nic := obj.(*networkingv1alpha1.NetworkInterface)
		log := ctrl.LoggerFrom(ctx)

		virtualIPList := &networkingv1alpha1.VirtualIPList{}
		if err := r.List(ctx, virtualIPList,
			client.InNamespace(nic.Namespace),
		); err != nil {
			log.Error(err, "Error listing virtual IPs")
			return nil
		}

		var reqs []ctrl.Request
		for _, virtualIP := range virtualIPList.Items {
			claimRef := virtualIP.Spec.TargetRef
			if claimRef == nil {
				continue
			}

			if claimRef.UID != nic.UID {
				continue
			}

			reqs = append(reqs, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(&virtualIP)})
		}
		return reqs
	})
}

func (r *VirtualIPReleaseReconciler) networkInterfaceDeletingPredicate() predicate.Predicate {
	return predicate.NewPredicateFuncs(func(obj client.Object) bool {
		nic := obj.(*networkingv1alpha1.NetworkInterface)
		return !nic.DeletionTimestamp.IsZero()
	})
}

func (r *VirtualIPReleaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("virtualiprelease").
		For(
			&networkingv1alpha1.VirtualIP{},
			builder.WithPredicates(r.virtualIPClaimedPredicate()),
		).
		Watches(
			&networkingv1alpha1.NetworkInterface{},
			r.enqueueByNetworkInterface(),
			builder.WithPredicates(r.networkInterfaceDeletingPredicate()),
		).
		Complete(r)
}
