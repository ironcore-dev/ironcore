// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package networking

import (
	"context"

	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	NetworkInterfacePrefixNamesField        = "networkinterface-prefix-names"
	NetworkInterfaceVirtualIPNamesField     = "networkinterface-virtual-ip-names"
	NetworkInterfaceSpecNetworkRefNameField = "spec.networkRef.name"
)

func SetupNetworkInterfacePrefixNamesFieldIndexer(ctx context.Context, indexer client.FieldIndexer) error {
	return indexer.IndexField(ctx, &networkingv1alpha1.NetworkInterface{}, NetworkInterfacePrefixNamesField, func(obj client.Object) []string {
		nic := obj.(*networkingv1alpha1.NetworkInterface)
		return networkingv1alpha1.NetworkInterfacePrefixNames(nic)
	})
}

func SetupNetworkInterfaceVirtualIPNameFieldIndexer(ctx context.Context, indexer client.FieldIndexer) error {
	return indexer.IndexField(ctx, &networkingv1alpha1.NetworkInterface{}, NetworkInterfaceVirtualIPNamesField, func(obj client.Object) []string {
		nic := obj.(*networkingv1alpha1.NetworkInterface)

		virtualIP := nic.Spec.VirtualIP
		if virtualIP == nil {
			return []string{""}
		}

		virtualIPName := networkingv1alpha1.NetworkInterfaceVirtualIPName(nic.Name, *nic.Spec.VirtualIP)
		if virtualIPName == "" {
			return []string{""}
		}

		return []string{virtualIPName}
	})
}

func SetupNetworkInterfaceNetworkNameFieldIndexer(ctx context.Context, indexer client.FieldIndexer) error {
	return indexer.IndexField(ctx, &networkingv1alpha1.NetworkInterface{}, NetworkInterfaceSpecNetworkRefNameField, func(obj client.Object) []string {
		nic := obj.(*networkingv1alpha1.NetworkInterface)
		return []string{nic.Spec.NetworkRef.Name}
	})
}
