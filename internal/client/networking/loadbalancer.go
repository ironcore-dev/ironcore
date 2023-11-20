// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package networking

import (
	"context"

	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	LoadBalancerPrefixNamesField = "loadbalancer-prefix-names"

	LoadBalancerNetworkNameField = "loadbalancer-network-name"
)

func SetupLoadBalancerPrefixNamesFieldIndexer(ctx context.Context, indexer client.FieldIndexer) error {
	return indexer.IndexField(ctx, &networkingv1alpha1.LoadBalancer{}, LoadBalancerPrefixNamesField, func(obj client.Object) []string {
		loadBalancer := obj.(*networkingv1alpha1.LoadBalancer)
		return networkingv1alpha1.LoadBalancerPrefixNames(loadBalancer)
	})
}

func SetupLoadBalancerNetworkNameFieldIndexer(ctx context.Context, indexer client.FieldIndexer) error {
	return indexer.IndexField(ctx, &networkingv1alpha1.LoadBalancer{}, LoadBalancerNetworkNameField, func(obj client.Object) []string {
		loadBalancer := obj.(*networkingv1alpha1.LoadBalancer)
		return []string{loadBalancer.Spec.NetworkRef.Name}
	})
}
