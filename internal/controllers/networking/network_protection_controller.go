/*
 * Copyright (c) 2022 by the OnMetal authors.
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

package networking

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/onmetal/controller-utils/clientutils"
	"github.com/onmetal/controller-utils/metautils"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	"github.com/onmetal/onmetal-api/internal/client/networking"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	networkFinalizer = "networking.api.onmetal.de/network"
)

type NetworkProtectionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=networks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=networks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=networks/finalizers,verbs=update
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=networkinterfaces,verbs=get;list;watch
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=loadbalancers,verbs=get;list;watch
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=natgateways,verbs=get;list;watch

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
	}

	for _, typeAndField := range typesAndFields {
		ok, err := r.isNetworkInUseByType(ctx, log, network, typeAndField.Type, typeAndField.Field)
		if err != nil {
			return false, fmt.Errorf("error checking if network is in use by %T: %w", typeAndField.Type, err)
		}
		if err != nil || ok {
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
			&source.Kind{Type: &networkingv1alpha1.NetworkInterface{}},
			r.enqueueByNetworkInterface(),
		).
		Watches(
			&source.Kind{Type: &networkingv1alpha1.LoadBalancer{}},
			r.enqueueByLoadBalancer(),
		).
		Watches(
			&source.Kind{Type: &networkingv1alpha1.NATGateway{}},
			r.enqueueByNATGateway(),
		).
		Complete(r)
}

func (r *NetworkProtectionReconciler) enqueueByNetworkInterface() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
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
	return handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
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
	return handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
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
