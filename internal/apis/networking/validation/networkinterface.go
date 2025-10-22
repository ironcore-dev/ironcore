// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"fmt"

	ironcorevalidation "github.com/ironcore-dev/ironcore/internal/api/validation"
	"github.com/ironcore-dev/ironcore/internal/apis/ipam"
	ipamvalidation "github.com/ironcore-dev/ironcore/internal/apis/ipam/validation"
	"github.com/ironcore-dev/ironcore/internal/apis/networking"
	corev1 "k8s.io/api/core/v1"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateNetworkInterface validates a network interface object.
func ValidateNetworkInterface(networkInterface *networking.NetworkInterface) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessor(networkInterface, true, apivalidation.NameIsDNSSubdomain, field.NewPath("metadata"))...)
	allErrs = append(allErrs, ValidateNetworkInterfaceSpec(&networkInterface.Spec, &networkInterface.ObjectMeta, field.NewPath("spec"))...)

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

func ValidateNetworkInterfaceSpec(spec *networking.NetworkInterfaceSpec, nicMeta *metav1.ObjectMeta, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if spec.NetworkRef == (corev1.LocalObjectReference{}) {
		allErrs = append(allErrs, field.Required(fldPath.Child("networkRef"), "must specify a network ref"))
	}

	for _, msg := range apivalidation.NameIsDNSSubdomain(spec.NetworkRef.Name, false) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("networkRef").Child("name"), spec.NetworkRef.Name, msg))
	}

	if spec.MachineRef != nil {
		for _, msg := range apivalidation.NameIsDNSLabel(spec.MachineRef.Name, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("machineRef").Child("name"), spec.MachineRef.Name, msg))
		}
	}

	allErrs = append(allErrs, ironcorevalidation.ValidateIPFamilies(spec.IPFamilies, fldPath.Child("ipFamilies"))...)

	if len(spec.IPFamilies) != len(spec.IPs) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("ips"), spec.IPFamilies, "ip families must match ips"))
	}

	allErrs = append(allErrs, validateNetworkInterfaceIPSources(spec.IPs, spec.IPFamilies, nicMeta, fldPath.Child("ips"))...)
	allErrs = append(allErrs, validateNetworkInterfacePrefixSources(spec.Prefixes, nicMeta, fldPath.Child("prefixes"))...)

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

func validateNetworkInterfacePrefixSources(prefixSources []networking.PrefixSource, nicMeta *metav1.ObjectMeta, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	for i, src := range prefixSources {
		allErrs = append(allErrs, validatePrefixSource(src, i, nicMeta, fldPath.Index(i))...)
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

func ValidateIPPrefixTemplate(template *ipam.PrefixTemplateSpec, ipFamily corev1.IPFamily, fldPath *field.Path) field.ErrorList {
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

func ValidatePrefixPrefixTemplate(template *ipam.PrefixTemplateSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if template == nil {
		allErrs = append(allErrs, field.Required(fldPath, "must specify template"))
	} else {
		allErrs = append(allErrs, ipamvalidation.ValidatePrefixTemplateSpec(template, fldPath)...)
	}

	return allErrs
}

func validateEphemeralPrefixSource(ipFamily corev1.IPFamily, source *networking.EphemeralPrefixSource, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, ValidateIPPrefixTemplate(source.PrefixTemplate, ipFamily, fldPath.Child("prefixTemplate"))...)

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

	oldSpecCopy.ProviderID = newSpec.ProviderID
	oldSpecCopy.IPs = newSpec.IPs
	oldSpecCopy.Prefixes = newSpec.Prefixes
	oldSpecCopy.MachineRef = newSpec.MachineRef
	oldSpecCopy.VirtualIP = newSpec.VirtualIP
	allErrs = append(allErrs, ironcorevalidation.ValidateImmutableFieldWithDiff(newSpecCopy, oldSpecCopy, fldPath)...)

	return allErrs
}
