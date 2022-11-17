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

package validation

import (
	"fmt"

	onmetalapivalidation "github.com/onmetal/onmetal-api/onmetal-apiserver/internal/api/validation"
	commonvalidation "github.com/onmetal/onmetal-api/onmetal-apiserver/internal/apis/common/validation"
	"github.com/onmetal/onmetal-api/onmetal-apiserver/internal/apis/ipam"
	ipamvalidation "github.com/onmetal/onmetal-api/onmetal-apiserver/internal/apis/ipam/validation"
	"github.com/onmetal/onmetal-api/onmetal-apiserver/internal/apis/networking"
	corev1 "k8s.io/api/core/v1"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateNetworkInterface validates a network interface object.
func ValidateNetworkInterface(networkInterface *networking.NetworkInterface) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessor(networkInterface, true, apivalidation.NameIsDNSLabel, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateNetworkInterfaceSpec(&networkInterface.Spec, &networkInterface.ObjectMeta, field.NewPath("spec"))...)

	return allErrs
}

// ValidateNetworkInterfaceUpdate validates a NetworkInterface object before an update.
func ValidateNetworkInterfaceUpdate(newNetworkInterface, oldNetworkInterface *networking.NetworkInterface) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessorUpdate(newNetworkInterface, oldNetworkInterface, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateNetworkInterfaceSpecUpdate(&newNetworkInterface.Spec, &oldNetworkInterface.Spec, field.NewPath("spec"))...)
	allErrs = append(allErrs, ValidateNetworkInterface(newNetworkInterface)...)

	return allErrs
}

func validateNetworkInterfaceSpec(spec *networking.NetworkInterfaceSpec, nicMeta *metav1.ObjectMeta, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if spec.NetworkRef == (corev1.LocalObjectReference{}) {
		allErrs = append(allErrs, field.Required(fldPath.Child("networkRef"), "must specify a network ref"))
	}

	for _, msg := range apivalidation.NameIsDNSLabel(spec.NetworkRef.Name, false) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("networkRef").Child("name"), spec.NetworkRef.Name, msg))
	}

	if spec.MachineRef != nil {
		for _, msg := range apivalidation.NameIsDNSLabel(spec.MachineRef.Name, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("machineRef").Child("name"), spec.MachineRef.Name, msg))
		}
	}

	allErrs = append(allErrs, onmetalapivalidation.ValidateIPFamilies(spec.IPFamilies, fldPath.Child("ipFamilies"))...)

	if len(spec.IPFamilies) != len(spec.IPs) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("ips"), spec.IPFamilies, "ip families must match ips"))
	}

	allErrs = append(allErrs, validateNetworkInterfaceIPSources(spec.IPs, spec.IPFamilies, nicMeta, fldPath.Child("ips"))...)

	if virtualIP := spec.VirtualIP; virtualIP != nil {
		allErrs = append(allErrs, validateVirtualIPSource(virtualIP, fldPath.Child("virtualIP"))...)
	}

	return allErrs
}

func validateNetworkInterfaceIPSources(ipSources []networking.IPSource, ipFamilies []corev1.IPFamily, nicMeta *metav1.ObjectMeta, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	for i, ip := range ipSources {
		var ipFamily corev1.IPFamily
		if i < len(ipFamilies) {
			ipFamily = ipFamilies[i]
		}

		allErrs = append(allErrs, validateIPSource(ip, i, ipFamily, nicMeta, fldPath.Index(i))...)
	}

	return allErrs
}

func validateIPSource(ipSource networking.IPSource, idx int, ipFamily corev1.IPFamily, nicMeta *metav1.ObjectMeta, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	var numSources int
	if ip := ipSource.Value; ip.IsValid() {
		numSources++
		allErrs = append(allErrs, commonvalidation.ValidateIP(ipFamily, *ip, fldPath.Child("value"))...)
	}
	if ephemeral := ipSource.Ephemeral; ephemeral != nil {
		if numSources > 0 {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("ephemeral"), ephemeral, "cannot specify multiple ip sources"))
		} else {
			numSources++
			allErrs = append(allErrs, validateEphemeralPrefixSource(ipFamily, ephemeral, fldPath.Child("ephemeral"))...)
			if nicMeta != nil && nicMeta.Name != "" {
				prefixName := fmt.Sprintf("%s-%d", nicMeta.Name, idx)
				for _, msg := range apivalidation.NameIsDNSLabel(prefixName, false) {
					allErrs = append(allErrs, field.Invalid(fldPath, prefixName, fmt.Sprintf("resulting prefix name %q is invalid: %s", prefixName, msg)))
				}
			}
		}
	}
	if numSources == 0 {
		allErrs = append(allErrs, field.Invalid(fldPath, ipSource, "must specify an ip source"))
	}

	return allErrs
}

func validateVirtualIPSource(vipSource *networking.VirtualIPSource, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	var numSources int
	if vipRef := vipSource.VirtualIPRef; vipRef != nil {
		numSources++
		for _, msg := range apivalidation.NameIsDNSLabel(vipRef.Name, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("virtualIPRef", "name"), vipRef.Name, msg))
		}
	}
	if ephemeral := vipSource.Ephemeral; ephemeral != nil {
		if numSources > 0 {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("ephemeral"), ephemeral, "cannot specify multiple sources"))
		} else {
			numSources++
			allErrs = append(allErrs, validateEphemeralVirtualIPSource(ephemeral, fldPath.Child("ephemeral"))...)
		}
	}
	if numSources == 0 {
		allErrs = append(allErrs, field.Invalid(fldPath, vipSource, "must specify a virtual ip source"))
	}

	return allErrs
}

var ipFamilyToBits = map[corev1.IPFamily]int32{
	corev1.IPv4Protocol: 32,
	corev1.IPv6Protocol: 128,
}

func ValidatePrefixTemplateForNetworkInterface(template *ipam.PrefixTemplateSpec, ipFamily corev1.IPFamily, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if template == nil {
		allErrs = append(allErrs, field.Required(fldPath, ""))
	} else {
		allErrs = append(allErrs, ipamvalidation.ValidatePrefixTemplateSpec(template, fldPath)...)

		spec := template.Spec
		specField := fldPath.Child("spec")
		if spec.IPFamily != ipFamily {
			allErrs = append(allErrs, field.Forbidden(specField.Child("ipFamily"), fmt.Sprintf("has to match network interface ip family %q", ipFamily)))
		}

		if prefix := spec.Prefix; prefix != nil {
			if !prefix.IsSingleIP() {
				allErrs = append(allErrs, field.Forbidden(specField.Child("prefix"), "must be a single IP"))
			}
		}
		if prefixLength := spec.PrefixLength; prefixLength != 0 {
			if prefixLength != ipFamilyToBits[ipFamily] {
				allErrs = append(allErrs, field.Forbidden(specField.Child("prefixLength"), "must be a single IP"))
			}
		}
	}

	return allErrs
}

func validateEphemeralPrefixSource(ipFamily corev1.IPFamily, source *networking.EphemeralPrefixSource, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, ValidatePrefixTemplateForNetworkInterface(source.PrefixTemplate, ipFamily, fldPath.Child("prefixTemplate"))...)

	return allErrs
}

func ValidateVirtualIPTemplateForNetworkInterface(vipTemplateSpec *networking.VirtualIPTemplateSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if vipTemplateSpec == nil {
		allErrs = append(allErrs, field.Required(fldPath, ""))
	} else {
		allErrs = append(allErrs, ValidateVirtualIPTemplateSpec(vipTemplateSpec, fldPath)...)
	}

	return allErrs
}

func validateEphemeralVirtualIPSource(source *networking.EphemeralVirtualIPSource, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, ValidateVirtualIPTemplateForNetworkInterface(source.VirtualIPTemplate, fldPath.Child("virtualIPTemplate"))...)

	return allErrs
}

// validateNetworkInterfaceSpecUpdate validates the spec of a NetworkInterface object before an update.
func validateNetworkInterfaceSpecUpdate(newSpec, oldSpec *networking.NetworkInterfaceSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	newSpecCopy := newSpec.DeepCopy()
	oldSpecCopy := oldSpec.DeepCopy()

	oldSpecCopy.MachineRef = newSpec.MachineRef
	allErrs = append(allErrs, onmetalapivalidation.ValidateImmutableFieldWithDiff(newSpecCopy, oldSpecCopy, fldPath)...)

	return allErrs
}
