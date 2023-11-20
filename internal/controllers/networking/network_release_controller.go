// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package networking

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	"golang.org/x/exp/slices"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/lru"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type NetworkReleaseReconciler struct {
	client.Client
	APIReader client.Reader

	AbsenceCache *lru.Cache
}

//+kubebuilder:rbac:groups=networking.ironcore.dev,resources=networks,verbs=get;list;watch;update;patch

func (r *NetworkReleaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	network := &networkingv1alpha1.Network{}
	if err := r.Get(ctx, req.NamespacedName, network); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileExists(ctx, log, network)
}

func (r *NetworkReleaseReconciler) reconcileExists(ctx context.Context, log logr.Logger, network *networkingv1alpha1.Network) (ctrl.Result, error) {
	if !network.DeletionTimestamp.IsZero() {
		log.V(1).Info("Network is already deleting, nothing to do")
		return ctrl.Result{}, nil
	}

	return r.reconcile(ctx, log, network)
}

func (r *NetworkReleaseReconciler) filterExistingNetworkPeeringClaimRefs(
	ctx context.Context,
	network *networkingv1alpha1.Network,
) ([]networkingv1alpha1.NetworkPeeringClaimRef, error) {
	var filtered []networkingv1alpha1.NetworkPeeringClaimRef
	for _, peeringClaimRef := range network.Spec.PeeringClaimRefs {
		ok, err := r.networkPeeringClaimExists(ctx, network, peeringClaimRef)
		if err != nil {
			return nil, err
		}

		if ok {
			filtered = append(filtered, peeringClaimRef)
		}
	}
	return filtered, nil
}

func (r *NetworkReleaseReconciler) networkPeeringClaimExists(
	ctx context.Context,
	network *networkingv1alpha1.Network,
	peeringClaimRef networkingv1alpha1.NetworkPeeringClaimRef,
) (bool, error) {
	if _, ok := r.AbsenceCache.Get(peeringClaimRef.UID); ok {
		return false, nil
	}

	claimer := &metav1.PartialObjectMetadata{
		TypeMeta: metav1.TypeMeta{
			APIVersion: networkingv1alpha1.SchemeGroupVersion.String(),
			Kind:       "Network",
		},
	}
	claimerKey := client.ObjectKey{Namespace: network.Namespace, Name: peeringClaimRef.Name}
	if err := r.APIReader.Get(ctx, claimerKey, claimer); err != nil {
		if !apierrors.IsNotFound(err) {
			return false, fmt.Errorf("error getting claiming network %s: %w", peeringClaimRef.Name, err)
		}

		r.AbsenceCache.Add(peeringClaimRef.UID, nil)
		return false, nil
	}
	return true, nil
}

func (r *NetworkReleaseReconciler) releaseNetwork(
	ctx context.Context,
	network *networkingv1alpha1.Network,
	filteredPeeringClaimRefs []networkingv1alpha1.NetworkPeeringClaimRef,
) error {
	baseNetwork := network.DeepCopy()
	network.Spec.PeeringClaimRefs = filteredPeeringClaimRefs
	if err := r.Patch(ctx, network, client.StrategicMergeFrom(baseNetwork, client.MergeFromWithOptimisticLock{})); err != nil {
		return fmt.Errorf("error patching network: %w", err)
	}
	return nil
}

func (r *NetworkReleaseReconciler) reconcile(ctx context.Context, log logr.Logger, network *networkingv1alpha1.Network) (ctrl.Result, error) {
	log.V(1).Info("Reconcile")

	if len(network.Spec.PeeringClaimRefs) == 0 {
		log.V(1).Info("Network is not claimed, nothing to do")
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Filtering for existing peering claim references")
	filteredPeeringClaimRefs, err := r.filterExistingNetworkPeeringClaimRefs(ctx, network)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error filtering for existing network peering claim refs: %w", err)
	}
	if slices.Equal(network.Spec.PeeringClaimRefs, filteredPeeringClaimRefs) {
		log.V(1).Info("All network peering claim refs are still up-to-date")
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Some network peering claims do not exist, releasing network")
	if err := r.releaseNetwork(ctx, network, filteredPeeringClaimRefs); err != nil {
		if !apierrors.IsConflict(err) {
			return ctrl.Result{}, fmt.Errorf("error releasing network: %w", err)
		}
		log.V(1).Info("Network was updated, requeueing")
		return ctrl.Result{Requeue: true}, nil
	}

	log.V(1).Info("Reconciled")
	return ctrl.Result{}, nil
}

func (r *NetworkReleaseReconciler) networkClaimedPredicate() predicate.Predicate {
	return predicate.NewPredicateFuncs(func(obj client.Object) bool {
		network := obj.(*networkingv1alpha1.Network)
		return len(network.Spec.PeeringClaimRefs) > 0
	})
}

func (r *NetworkReleaseReconciler) enqueueByNetwork() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []ctrl.Request {
		network := obj.(*networkingv1alpha1.Network)
		log := ctrl.LoggerFrom(ctx)

		networkList := &networkingv1alpha1.NetworkList{}
		if err := r.List(ctx, networkList); err != nil {
			log.Error(err, "Error listing networks")
			return nil
		}

		var reqs []ctrl.Request
		for _, targetNetwork := range networkList.Items {
			var found bool
			for _, claimRef := range targetNetwork.Spec.PeeringClaimRefs {
				if claimRef.UID == network.UID {
					found = true
					break
				}
			}
			if !found {
				continue
			}

			reqs = append(reqs, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(&targetNetwork)})
		}
		return reqs
	})
}

func (r *NetworkReleaseReconciler) networkDeletingPredicate() predicate.Predicate {
	return predicate.NewPredicateFuncs(func(obj client.Object) bool {
		network := obj.(*networkingv1alpha1.Network)
		return !network.DeletionTimestamp.IsZero()
	})
}

func (r *NetworkReleaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("networkrelease").
		For(
			&networkingv1alpha1.Network{},
			builder.WithPredicates(r.networkClaimedPredicate()),
		).
		Watches(
			&networkingv1alpha1.Network{},
			r.enqueueByNetwork(),
			builder.WithPredicates(r.networkDeletingPredicate()),
		).
		Complete(r)
}
