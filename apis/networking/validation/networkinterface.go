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
	"github.com/onmetal/onmetal-api/apis/networking"
	corev1 "k8s.io/api/core/v1"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var supportedServiceIPFamily = sets.NewString(string(corev1.IPv4Protocol), string(corev1.IPv6Protocol))

// ValidateNetworkInterface validates a network interface object.
func ValidateNetworkInterface(networkInterface *networking.NetworkInterface) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessor(networkInterface, true, apivalidation.NameIsDNSLabel, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateNetworkInterfaceSpec(&networkInterface.Spec, field.NewPath("spec"))...)

	return allErrs
}

// ValidateNetworkInterfaceUpdate validates a NetworkInterface object before an update.
func ValidateNetworkInterfaceUpdate(newNetworkInterface, oldNetworkInterface *networking.NetworkInterface) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessorUpdate(newNetworkInterface, oldNetworkInterface, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateNetworkInterfaceSpecUpdate(&newNetworkInterface.Spec, &oldNetworkInterface.Spec, newNetworkInterface.DeletionTimestamp != nil, field.NewPath("spec"))...)
	allErrs = append(allErrs, ValidateNetworkInterface(newNetworkInterface)...)

	return allErrs
}

func validateNetworkInterfaceSpec(networkInterfaceSpec *networking.NetworkInterfaceSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if networkInterfaceSpec.NetworkRef == (corev1.LocalObjectReference{}) {
		allErrs = append(allErrs, field.Required(fldPath.Child("networkRef"), "must specify a network ref"))
	}

	for _, msg := range apivalidation.NameIsDNSLabel(networkInterfaceSpec.NetworkRef.Name, false) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("networkRef").Child("name"), networkInterfaceSpec.NetworkRef.Name, msg))
	}

	if networkInterfaceSpec.MachineRef.Name != "" {
		for _, msg := range apivalidation.NameIsDNSLabel(networkInterfaceSpec.MachineRef.Name, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("machineRef").Child("name"), networkInterfaceSpec.MachineRef.Name, msg))
		}
	}

	if len(networkInterfaceSpec.IPFamilies) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("ipFamilies"), "must provide at least one ip family"))
	}

	// ipfamilies stand alone validation
	// must be either IPv4 or IPv6
	seen := sets.String{}
	for i, ipFamily := range networkInterfaceSpec.IPFamilies {
		if !supportedServiceIPFamily.Has(string(ipFamily)) {
			allErrs = append(allErrs, field.NotSupported(fldPath.Child("ipFamilies").Index(i), ipFamily, supportedServiceIPFamily.List()))
		}
		// no duplicate check also ensures that ipfamilies is dualstacked, in any order
		if seen.Has(string(ipFamily)) {
			allErrs = append(allErrs, field.Duplicate(fldPath.Child("ipFamilies").Index(i), ipFamily))
		}
		seen.Insert(string(ipFamily))
	}

	for i, ipSource := range networkInterfaceSpec.IPs {
		if ipSource.EphemeralPrefix != nil && ipSource.EphemeralPrefix.PrefixTemplate != nil {
			allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessor(ipSource.EphemeralPrefix.PrefixTemplate,
				true, apivalidation.NameIsDNSLabel, fldPath.Child("ips").Index(i).Child("ephemeralPrefix").Child("metadata"))...)
		}
		// TODO: validate PrefixSpec once new Prefix is merged
	}

	return allErrs
}

// validateNetworkInterfaceSpecUpdate validates the spec of a NetworkInterface object before an update.
func validateNetworkInterfaceSpecUpdate(new, old *networking.NetworkInterfaceSpec, deletionTimestampSet bool, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateImmutableField(new.NetworkRef, old.NetworkRef, fldPath.Child("networkRef"))...)

	return allErrs
}

func IsValidIPFamily(family corev1.IPFamily) bool {
	return family == corev1.IPv4Protocol || family == corev1.IPv6Protocol
}
