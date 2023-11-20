// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package networking

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
