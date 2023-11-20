// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import "fmt"

// NetworkInterfaceVirtualIPName returns the name of a VirtualIP for a NetworkInterface VirtualIPSource.
func NetworkInterfaceVirtualIPName(nicName string, vipSource VirtualIPSource) string {
	switch {
	case vipSource.VirtualIPRef != nil:
		return vipSource.VirtualIPRef.Name
	case vipSource.Ephemeral != nil:
		return nicName
	default:
		return ""
	}
}

// NetworkInterfaceIPIPAMPrefixName returns the name of a Prefix for a network interface ephemeral prefix.
func NetworkInterfaceIPIPAMPrefixName(nicName string, idx int) string {
	return fmt.Sprintf("%s-%d", nicName, idx)
}

// NetworkInterfacePrefixIPAMPrefixName returns the name of a Prefix for a network interface ephemeral prefix.
func NetworkInterfacePrefixIPAMPrefixName(nicName string, idx int) string {
	return fmt.Sprintf("%s-pf-%d", nicName, idx)
}

// LoadBalancerIPIPAMPrefixName returns the name of a Prefix for a network interface ephemeral prefix.
func LoadBalancerIPIPAMPrefixName(loadBalancerName string, idx int) string {
	return fmt.Sprintf("%s-%d", loadBalancerName, idx)
}

// NetworkInterfacePrefixNames returns the name of all ipam prefixes the network interface references.
func NetworkInterfacePrefixNames(nic *NetworkInterface) []string {
	var names []string

	for i, nicIP := range nic.Spec.IPs {
		if nicIP.Ephemeral == nil {
			continue
		}

		names = append(names, NetworkInterfaceIPIPAMPrefixName(nic.Name, i))
	}

	for i, nicPrefix := range nic.Spec.Prefixes {
		if nicPrefix.Ephemeral == nil {
			continue
		}

		names = append(names, NetworkInterfacePrefixIPAMPrefixName(nic.Name, i))
	}

	return names
}

// LoadBalancerPrefixNames returns the name of all ipam prefixes the load balancer references.
func LoadBalancerPrefixNames(loadBalancer *LoadBalancer) []string {
	var names []string

	for i, loadBalancerIP := range loadBalancer.Spec.IPs {
		if loadBalancerIP.Ephemeral == nil {
			continue
		}

		names = append(names, LoadBalancerIPIPAMPrefixName(loadBalancer.Name, i))
	}

	return names
}
