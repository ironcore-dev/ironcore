// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package networking

import (
	"context"

	"github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const NetworkSpecPeeringClaimRefNamesField = "network-spec-peering-claim-ref-names"

func SetupNetworkSpecPeeringClaimRefNamesFieldIndexer(ctx context.Context, indexer client.FieldIndexer) error {
	return indexer.IndexField(ctx, &v1alpha1.Network{}, NetworkSpecPeeringClaimRefNamesField, func(obj client.Object) []string {
		network := obj.(*v1alpha1.Network)
		peeringClaimRefNames := make([]string, 0, len(network.Spec.PeeringClaimRefs))
		for _, peeringClaimRef := range network.Spec.PeeringClaimRefs {
			peeringClaimRefNames = append(peeringClaimRefNames, peeringClaimRef.Name)
		}

		if len(peeringClaimRefNames) == 0 {
			return []string{""}
		}
		return peeringClaimRefNames
	})
}
