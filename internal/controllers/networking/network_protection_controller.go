// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package networking

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/ironcore-dev/controller-utils/clientutils"
	"github.com/ironcore-dev/controller-utils/metautils"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	"github.com/ironcore-dev/ironcore/internal/client/networking"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
)

const (
	networkFinalizer = "networking.ironcore.dev/network"
)

type NetworkProtectionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=networking.ironcore.dev,resources=networks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.ironcore.dev,resources=networks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=networking.ironcore.dev,resources=networks/finalizers,verbs=update
//+kubebuilder:rbac:groups=networking.ironcore.dev,resources=networkinterfaces,verbs=get;list;watch
//+kubebuilder:rbac:groups=networking.ironcore.dev,resources=loadbalancers,verbs=get;list;watch
//+kubebuilder:rbac:groups=networking.ironcore.dev,resources=natgateways,verbs=get;list;watch

func (r *NetworkProtectionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	network := &networkingv1alpha1.Network{}
	if err := r.Get(ctx, req.NamespacedName, network); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileExists(ctx, log, network)
}

func (r *NetworkProtectionReconciler) reconcileExists(ctx context.Context, log logr.Logger, network *networkingv1alpha1.Network) (ctrl.Result, error) {
	if !network.DeletionTimestamp.IsZero() {
		return r.delete(ctx, log, network)
	}
	return r.reconcile(ctx, log, network)
}

func (r *NetworkProtectionReconciler) delete(ctx context.Context, log logr.Logger, network *networkingv1alpha1.Network) (ctrl.Result, error) {
	log.Info("Deleting Network")

	if ok, err := r.isNetworkInUse(ctx, log, network); err != nil || ok {
		return ctrl.Result{Requeue: ok}, err
	}

	log.V(1).Info("Removing finalizer from Network as the Network is not in use")
	if _, err := clientutils.PatchEnsureNoFinalizer(ctx, r.Client, network, networkFinalizer); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to remove finalizer from network: %w", err)
	}

	log.Info("Successfully deleted Network")
	return ctrl.Result{}, nil
}

func (r *NetworkProtectionReconciler) reconcile(ctx context.Context, log logr.Logger, network *networkingv1alpha1.Network) (ctrl.Result, error) {
	log.Info("Reconcile Network")

	log.V(1).Info("Ensuring finalizer on Network")
	if _, err := clientutils.PatchEnsureFinalizer(ctx, r.Client, network, networkFinalizer); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to patch finalizer from network: %w", err)
	}

	log.Info("Successfully reconciled Network")
	return ctrl.Result{}, nil
}

func (r *NetworkProtectionReconciler) isNetworkInUseByType(
	ctx context.Context,
	log logr.Logger,
	network *networkingv1alpha1.Network,
	obj client.Object,
	networkField string,
) (bool, error) {
	gvk, err := apiutil.GVKForObject(network, r.Scheme)
	if err != nil {
		return false, fmt.Errorf("error getting gvk for object: %w", err)
	}

	_, list, err := metautils.NewListForObject(r.Scheme, obj)
	if err != nil {
		return false, fmt.Errorf("error creating list for object: %w", err)
	}

	if err := r.List(ctx, list,
		client.InNamespace(network.Namespace),
		client.MatchingFields{networkField: network.Name},
	); err != nil {
		return false, fmt.Errorf("failed to list : %w", err)
	}

	var names []string
	if err := metautils.EachListItem(list, func(obj client.Object) error {
		if obj.GetDeletionTimestamp().IsZero() {
			names = append(names, obj.GetName())
		}
		return nil
	}); err != nil {
		return false, fmt.Errorf("error iterating list: %w", err)
	}

	if len(names) > 0 {
		log.V(1).Info("Network is in use", "GVK", gvk, "Names", names)
		return true, nil
	}
	return false, nil
}

func (r *NetworkProtectionReconciler) isNetworkInUse(ctx context.Context, log logr.Logger, network *networkingv1alpha1.Network) (bool, error) {
	log.V(1).Info("Checking if the network is in use")

	typesAndFields := []struct {
		Type  client.Object
		Field string
	}{
		{
			Type:  &networkingv1alpha1.NetworkInterface{},
			Field: networking.NetworkInterfaceSpecNetworkRefNameField,
		},
		{
			Type:  &networkingv1alpha1.LoadBalancer{},
			Field: networking.LoadBalancerNetworkNameField,
		},
		{
			Type:  &networkingv1alpha1.NATGateway{},
			Field: networking.NATGatewayNetworkNameField,
		},
		{
			Type:  &networkingv1alpha1.Network{},
			Field: networking.NetworkSpecPeeringClaimRefNamesField,
		},
	}

	for _, typeAndField := range typesAndFields {
		ok, err := r.isNetworkInUseByType(ctx, log, network, typeAndField.Type, typeAndField.Field)
		if err != nil {
			return false, fmt.Errorf("error checking if network is in use by %T: %w", typeAndField.Type, err)
		}
		if ok {
			return ok, err
		}
	}

	return false, nil
}

func (r *NetworkProtectionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("networkprotection").
		For(&networkingv1alpha1.Network{}).
		Watches(
			&networkingv1alpha1.NetworkInterface{},
			r.enqueueByNetworkInterface(),
		).
		Watches(
			&networkingv1alpha1.LoadBalancer{},
			r.enqueueByLoadBalancer(),
		).
		Watches(
			&networkingv1alpha1.NATGateway{},
			r.enqueueByNATGateway(),
		).
		Watches(
			&networkingv1alpha1.Network{},
			r.enqueueByPeeredNetwork(),
		).
		Complete(r)
}

func (r *NetworkProtectionReconciler) enqueueByNetworkInterface() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []ctrl.Request {
		nic := obj.(*networkingv1alpha1.NetworkInterface)

		var res []ctrl.Request
		networkKey := types.NamespacedName{
			Namespace: nic.Namespace,
			Name:      nic.Spec.NetworkRef.Name,
		}
		res = append(res, ctrl.Request{NamespacedName: networkKey})

		return res
	})
}

func (r *NetworkProtectionReconciler) enqueueByLoadBalancer() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []ctrl.Request {
		loadBalancer := obj.(*networkingv1alpha1.LoadBalancer)

		var res []ctrl.Request
		networkKey := types.NamespacedName{
			Namespace: loadBalancer.Namespace,
			Name:      loadBalancer.Spec.NetworkRef.Name,
		}
		res = append(res, ctrl.Request{NamespacedName: networkKey})

		return res
	})
}

func (r *NetworkProtectionReconciler) enqueueByNATGateway() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []ctrl.Request {
		natGateway := obj.(*networkingv1alpha1.NATGateway)

		var res []ctrl.Request
		networkKey := types.NamespacedName{
			Namespace: natGateway.Namespace,
			Name:      natGateway.Spec.NetworkRef.Name,
		}
		res = append(res, ctrl.Request{NamespacedName: networkKey})

		return res
	})
}

func (r *NetworkProtectionReconciler) enqueueByPeeredNetwork() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []ctrl.Request {
		network := obj.(*networkingv1alpha1.Network)
		log := ctrl.LoggerFrom(ctx)

		networkList := &networkingv1alpha1.NetworkList{}
		if err := r.List(ctx, networkList); err != nil {
			log.Error(err, "Error listing networks")
			return nil
		}

		var res []ctrl.Request
		for _, targetNetwork := range networkList.Items {
			var found bool
			for _, claimRef := range targetNetwork.Spec.PeeringClaimRefs {
				if claimRef.UID == network.UID {
					networkKey := types.NamespacedName{
						Namespace: claimRef.Namespace,
						Name:      claimRef.Name,
					}
					res = append(res, ctrl.Request{NamespacedName: networkKey})
					found = true
					break
				}
			}
			if found {
				break
			}
		}

		return res
	})
}
