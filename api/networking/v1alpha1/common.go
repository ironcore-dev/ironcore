// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	ipamv1alpha1 "github.com/ironcore-dev/ironcore/api/ipam/v1alpha1"
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
	PrefixTemplate *ipamv1alpha1.PrefixTemplateSpec `json:"prefixTemplate,omitempty"`
}

// EphemeralVirtualIPSource contains the definition to create an ephemeral (i.e. coupled to the lifetime of the
// surrounding object) VirtualIP.
type EphemeralVirtualIPSource struct {
	// VirtualIPTemplate is the template for the VirtualIP.
	VirtualIPTemplate *VirtualIPTemplateSpec `json:"virtualIPTemplate,omitempty"`
}
