/*
 * Copyright (c) 2022 by the OnMetal authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package v1alpha1

import (
	ipamv1alpha1 "github.com/onmetal/onmetal-api/apis/ipam/v1alpha1"
)

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
