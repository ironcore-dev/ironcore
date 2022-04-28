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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	ipFamilyToPrefixLength = map[corev1.IPFamily]int32{
		corev1.IPv4Protocol: 32,
		corev1.IPv6Protocol: 128,
	}
)

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
}

func SetDefaults_NetworkInterfaceSpec(spec *NetworkInterfaceSpec) {
	if len(spec.IPFamilies) > 0 {
		if len(spec.IPFamilies) == len(spec.IPs) {
			for i, ip := range spec.IPs {
				if ip.EphemeralPrefix != nil {
					if ip.EphemeralPrefix.PrefixTemplateSpec != nil {
						ephemeralPrefixSpec := &ip.EphemeralPrefix.PrefixTemplateSpec.Spec

						if ephemeralPrefixSpec.IPFamily == "" {
							ephemeralPrefixSpec.IPFamily = spec.IPFamilies[i]
						}
					}
				}
			}
		}
	} else if len(spec.IPs) > 0 {
		for _, ip := range spec.IPs {
			switch {
			case ip.Value != nil:
				spec.IPFamilies = append(spec.IPFamilies, ip.Value.Family())
			case ip.EphemeralPrefix != nil && ip.EphemeralPrefix.PrefixTemplateSpec != nil:
				spec.IPFamilies = append(spec.IPFamilies, ip.EphemeralPrefix.PrefixTemplateSpec.Spec.IPFamily)
			}
		}
	}

	for _, ip := range spec.IPs {
		if ip.EphemeralPrefix != nil && ip.EphemeralPrefix.PrefixTemplateSpec != nil {
			templateSpec := ip.EphemeralPrefix.PrefixTemplateSpec.Spec
			if templateSpec.Prefix == nil && templateSpec.PrefixLength == 0 {
				templateSpec.PrefixLength = ipFamilyToPrefixLength[templateSpec.IPFamily]
			}
		}
	}
}
