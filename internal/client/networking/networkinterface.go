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
	NetworkInterfacePrefixNamesField        = "networkinterface-prefix-names"
	NetworkInterfaceVirtualIPNamesField     = "networkinterface-virtual-ip-names"
	NetworkInterfaceSpecNetworkRefNameField = "spec.networkRef.name"
	NetworkInterfaceSpecMachineRefNameField = "spec.machineRef.name"
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

func SetupNetworkInterfaceSpecMachineRefNameField(ctx context.Context, indexer client.FieldIndexer) error {
	return indexer.IndexField(ctx, &networkingv1alpha1.NetworkInterface{}, NetworkInterfaceSpecMachineRefNameField, func(obj client.Object) []string {
		nic := obj.(*networkingv1alpha1.NetworkInterface)

		machineRef := nic.Spec.MachineRef
		if machineRef == nil {
			return []string{""}
		}

		return []string{machineRef.Name}
	})
}
