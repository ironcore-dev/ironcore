// Copyright 2022 OnMetal authors
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

package client

import (
	"context"

	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	ipamv1alpha1 "github.com/onmetal/onmetal-api/api/ipam/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	MachineSpecNetworkInterfaceNamesField = "machine-spec-network-interface-names"
	MachineSpecVolumeNamesField           = "machine-spec-volume-names"

	NetworkInterfaceVirtualIPNames = "networkinterface-virtual-ip-names"
	NetworkInterfaceNetworkName    = "networkinterface-network-name"

	PrefixSpecIPFamilyField           = "prefix-spec-ipfamily"
	PrefixAllocationSpecIPFamilyField = "prefixallocation-spec-ipfamily"
)

func SetupMachineSpecNetworkInterfaceNamesFieldIndexer(ctx context.Context, indexer client.FieldIndexer) error {
	return indexer.IndexField(ctx, &computev1alpha1.Machine{}, MachineSpecNetworkInterfaceNamesField, func(obj client.Object) []string {
		machine := obj.(*computev1alpha1.Machine)
		if names := computev1alpha1.MachineNetworkInterfaceNames(machine); len(names) > 0 {
			return names
		}
		return []string{""}
	})
}

func SetupMachineSpecVolumeNamesFieldIndexer(ctx context.Context, indexer client.FieldIndexer) error {
	return indexer.IndexField(ctx, &computev1alpha1.Machine{}, MachineSpecVolumeNamesField, func(obj client.Object) []string {
		machine := obj.(*computev1alpha1.Machine)
		if names := computev1alpha1.MachineVolumeNames(machine); len(names) > 0 {
			return names
		}
		return []string{""}
	})
}

func SetupNetworkInterfaceVirtualIPNameFieldIndexer(ctx context.Context, indexer client.FieldIndexer) error {
	return indexer.IndexField(ctx, &networkingv1alpha1.NetworkInterface{}, NetworkInterfaceVirtualIPNames, func(obj client.Object) []string {
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
	return indexer.IndexField(ctx, &networkingv1alpha1.NetworkInterface{}, NetworkInterfaceNetworkName, func(obj client.Object) []string {
		nic := obj.(*networkingv1alpha1.NetworkInterface)
		return []string{nic.Spec.NetworkRef.Name}
	})
}

func SetupPrefixSpecIPFamilyFieldIndexer(ctx context.Context, indexer client.FieldIndexer) error {
	return indexer.IndexField(ctx, &ipamv1alpha1.Prefix{}, PrefixSpecIPFamilyField, func(obj client.Object) []string {
		prefix := obj.(*ipamv1alpha1.Prefix)
		return []string{string(prefix.Spec.IPFamily)}
	})
}

func SetupPrefixAllocationSpecIPFamilyFieldIndexer(ctx context.Context, indexer client.FieldIndexer) error {
	return indexer.IndexField(ctx, &ipamv1alpha1.PrefixAllocation{}, PrefixAllocationSpecIPFamilyField, func(obj client.Object) []string {
		prefixAllocation := obj.(*ipamv1alpha1.PrefixAllocation)
		return []string{string(prefixAllocation.Spec.IPFamily)}
	})
}
