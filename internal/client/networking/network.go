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
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	NetworkPeeringKeysField = "network-peering-keys"
)

func NetworkPeeringKey(network *networkingv1alpha1.Network) string {
	return client.ObjectKeyFromObject(network).String()
}

func SetupNetworkPeeringKeysFieldIndexer(ctx context.Context, indexer client.FieldIndexer) error {
	return indexer.IndexField(ctx, &networkingv1alpha1.Network{}, NetworkPeeringKeysField, func(obj client.Object) []string {
		network := obj.(*networkingv1alpha1.Network)
		peerings := sets.New[string]()
		for _, peering := range network.Spec.Peerings {
			ref := peering.NetworkRef
			refNamespace := ref.Namespace
			if refNamespace == "" {
				refNamespace = network.Namespace
			}
			refKey := client.ObjectKey{Namespace: refNamespace, Name: ref.Name}

			peerings.Insert(refKey.String())
		}
		return peerings.UnsortedList()
	})
}
