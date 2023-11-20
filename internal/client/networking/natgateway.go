// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package networking

import (
	"context"

	"github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const NATGatewayNetworkNameField = "natgateway-network-name"

func SetupNATGatewayNetworkNameFieldIndexer(ctx context.Context, indexer client.FieldIndexer) error {
	return indexer.IndexField(ctx, &v1alpha1.NATGateway{}, NATGatewayNetworkNameField, func(obj client.Object) []string {
		natGateway := obj.(*v1alpha1.NATGateway)
		return []string{natGateway.Spec.NetworkRef.Name}
	})
}
