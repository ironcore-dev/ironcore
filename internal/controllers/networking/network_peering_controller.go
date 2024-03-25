// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package networking

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type NetworkPeeringReconciler struct {
	client.Client
}

//+kubebuilder:rbac:groups=networking.ironcore.dev,resources=networks,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=networking.ironcore.dev,resources=networks/status,verbs=get;update;patch

func (r *NetworkPeeringReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	network := &networkingv1alpha1.Network{}
	if err := r.Get(ctx, req.NamespacedName, network); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileExists(ctx, log, network)
}

func (r *NetworkPeeringReconciler) reconcileExists(ctx context.Context, log logr.Logger, network *networkingv1alpha1.Network) (ctrl.Result, error) {
	if !network.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	return r.reconcile(ctx, log, network)
}

func (r *NetworkPeeringReconciler) reconcile(ctx context.Context, log logr.Logger, network *networkingv1alpha1.Network) (ctrl.Result, error) {
	log.V(1).Info("Reconcile")

	var peeringClaimRefs []networkingv1alpha1.NetworkPeeringClaimRef
	var peeringNames []string

	for _, peering := range network.Spec.Peerings {
		peeringClaimRef, err := r.reconcilePeering(ctx, log, network, peering)

		if err != nil {
			return ctrl.Result{}, fmt.Errorf("[network peering %s] %w", peering.Name, err)
		}

		if peeringClaimRef != (networkingv1alpha1.NetworkPeeringClaimRef{}) {
			peeringClaimRefs = append(peeringClaimRefs, peeringClaimRef)
		}

		if peering.Name != "" {
			peeringNames = append(peeringNames, peering.Name)
		}
	}

	if len(peeringClaimRefs) > 0 {
		log.V(1).Info("Peering claim refs require network spec update")
		if err := r.updateSpec(ctx, log, network, peeringClaimRefs); err != nil {
			return ctrl.Result{}, fmt.Errorf("error updating network spec: %w", err)
		}
	}

	if len(peeringNames) > 0 {
		log.V(1).Info("Network peering status require network status update")
		if err := r.updateStatus(ctx, log, network, peeringNames); err != nil {
			return ctrl.Result{}, fmt.Errorf("error updating network status: %w", err)
		}
	}

	log.V(1).Info("Reconciled")
	return ctrl.Result{}, nil
}

func (r *NetworkPeeringReconciler) updateStatus(
	ctx context.Context,
	log logr.Logger,
	network *networkingv1alpha1.Network,
	peeringNames []string,
) error {
	base := network.DeepCopy()
	newStatusPeerings := make([]networkingv1alpha1.NetworkPeeringStatus, 0, len(peeringNames))
	for _, name := range peeringNames {
		newStatusPeerings = append(newStatusPeerings, networkingv1alpha1.NetworkPeeringStatus{
			Name:  name,
			State: networkingv1alpha1.NetworkPeeringStateInitial,
		})
	}
	network.Status.Peerings = newStatusPeerings

	log.V(1).Info("Updating network status peerings", "", network.Status.Peerings)
	if err := r.Status().Patch(ctx, network, client.StrategicMergeFrom(base)); err != nil {
		return fmt.Errorf("error updating network status peerings: %w", err)
	}
	log.V(1).Info("Updated network status peerings")

	return nil
}

func (r *NetworkPeeringReconciler) updateSpec(
	ctx context.Context,
	log logr.Logger,
	network *networkingv1alpha1.Network,
	peeringClaimRefs []networkingv1alpha1.NetworkPeeringClaimRef,
) error {
	base := network.DeepCopy()
	network.Spec.PeeringClaimRefs = peeringClaimRefs
	log.V(1).Info("Updating network spec incoming peerings")
	if err := r.Patch(ctx, network, client.StrategicMergeFrom(base)); err != nil {
		return fmt.Errorf("error updating network spec incoming peerings: %w", err)
	}

	log.V(1).Info("Updated network spec incoming peerings")
	return nil
}

func (r *NetworkPeeringReconciler) reconcilePeering(
	ctx context.Context,
	log logr.Logger,
	network *networkingv1alpha1.Network,
	peering networkingv1alpha1.NetworkPeering,
) (networkingv1alpha1.NetworkPeeringClaimRef, error) {
	networkKey := client.ObjectKeyFromObject(network)

	targetNetwork := &networkingv1alpha1.Network{}
	targetNetworkRef := peering.NetworkRef
	targetNetworkNamespace := targetNetworkRef.Namespace
	if targetNetworkNamespace == "" {
		targetNetworkNamespace = network.Namespace
	}
	targetNetworkKey := client.ObjectKey{Namespace: targetNetworkNamespace, Name: targetNetworkRef.Name}
	log = log.WithValues("TargetNetworkKey", targetNetworkKey)

	log.V(1).Info("Getting target network")
	if err := r.Get(ctx, targetNetworkKey, targetNetwork); err != nil {
		if !apierrors.IsNotFound(err) {
			return networkingv1alpha1.NetworkPeeringClaimRef{}, fmt.Errorf("error getting target network %s: %w", targetNetworkKey, err)
		}

		log.V(1).Info("Target network not found")
		return networkingv1alpha1.NetworkPeeringClaimRef{}, nil
	}

	for _, targetPeering := range targetNetwork.Spec.Peerings {
		targetPeeringNetworkRef := targetPeering.NetworkRef
		targetPeeringNetworkNamespace := targetPeeringNetworkRef.Namespace
		if targetPeeringNetworkNamespace == "" {
			targetPeeringNetworkNamespace = targetNetwork.Namespace
		}
		targetPeeringNetworkKey := client.ObjectKey{Namespace: targetPeeringNetworkNamespace, Name: targetPeeringNetworkRef.Name}

		if targetPeeringNetworkKey != networkKey {
			continue
		}

		if targetNetwork.Status.State != networkingv1alpha1.NetworkStateAvailable {
			log.V(1).Info("Target network is not available yet")
			return networkingv1alpha1.NetworkPeeringClaimRef{}, nil
		}

		log.V(1).Info("Target network peering matches")
		peeringClaimRef := networkingv1alpha1.NetworkPeeringClaimRef{
			Namespace: targetNetwork.Namespace,
			Name:      targetNetwork.Name,
			UID:       targetNetwork.UID,
		}

		return peeringClaimRef, nil
	}

	log.V(1).Info("No matching target peering found")
	return networkingv1alpha1.NetworkPeeringClaimRef{}, nil
}

func (r *NetworkPeeringReconciler) enqueuePeeringReferencedNetworks() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []ctrl.Request {
		network := obj.(*networkingv1alpha1.Network)
		reqs := sets.New[ctrl.Request]()
		for _, peering := range network.Spec.Peerings {
			ref := peering.NetworkRef
			refNamespace := ref.Namespace
			if refNamespace == "" {
				refNamespace = network.Namespace
			}

			refKey := client.ObjectKey{Namespace: refNamespace, Name: ref.Name}
			reqs.Insert(ctrl.Request{NamespacedName: refKey})
		}
		return reqs.UnsortedList()
	})
}

func (r *NetworkPeeringReconciler) networkStateAvailablePredicate() predicate.Predicate {
	return predicate.NewPredicateFuncs(func(obj client.Object) bool {
		network := obj.(*networkingv1alpha1.Network)
		return network.Status.State == networkingv1alpha1.NetworkStateAvailable
	})
}

func (r *NetworkPeeringReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("networkpeering").
		For(
			&networkingv1alpha1.Network{},
			builder.WithPredicates(r.networkStateAvailablePredicate()),
		).
		Watches(
			&networkingv1alpha1.Network{},
			r.enqueuePeeringReferencedNetworks(),
			builder.WithPredicates(r.networkStateAvailablePredicate()),
		).
		Complete(r)
}
