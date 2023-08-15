// Copyright 2023 OnMetal authors
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

	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
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
