// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package networking

import (
	"github.com/ironcore-dev/ironcore/internal/apis/ipam"
)

const (
	// NetworkPluginsGroup is the system rbac group all network plugins are in.
	NetworkPluginsGroup = "networking.ironcore.dev:system:networkplugins"

	// NetworkPluginUserNamePrefix is the prefix all network plugin users should have.
	NetworkPluginUserNamePrefix = "networking.ironcore.dev:system:networkplugin:"
)

// NetworkPluginCommonName constructs the common name for a certificate of a network plugin user.
func NetworkPluginCommonName(name string) string {
	return NetworkPluginUserNamePrefix + name
}

// EphemeralPrefixSource contains the definition to create an ephemeral (i.e. coupled to the lifetime of the
// surrounding object) Prefix.
type EphemeralPrefixSource struct {
	// PrefixTemplate is the template for the Prefix.
	PrefixTemplate *ipam.PrefixTemplateSpec
}

// EphemeralVirtualIPSource contains the definition to create an ephemeral (i.e. coupled to the lifetime of the
// surrounding object) VirtualIP.
type EphemeralVirtualIPSource struct {
	// VirtualIPTemplate is the template for the VirtualIP.
	VirtualIPTemplate *VirtualIPTemplateSpec
}
