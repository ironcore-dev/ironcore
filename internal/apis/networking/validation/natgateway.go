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
	onmetalapivalidation "github.com/onmetal/onmetal-api/internal/api/validation"
	"github.com/onmetal/onmetal-api/internal/apis/networking"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1validation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateNATGateway validates an NATGateway object.
func ValidateNATGateway(natGateway *networking.NATGateway) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessor(natGateway, true, apivalidation.NameIsDNSLabel, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateNATGatewaySpec(&natGateway.Spec, field.NewPath("spec"))...)

	return allErrs
}

func validateNATGatewaySpec(spec *networking.NATGatewaySpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, validateNATGatewayType(spec.Type, fldPath.Child("type"))...)

	allErrs = append(allErrs, onmetalapivalidation.ValidateIPFamilies(spec.IPFamilies, fldPath.Child("ipFamilies"))...)

	seenIPNames := sets.New[string]()
	for i, ip := range spec.IPs {
		ipFieldPath := fldPath.Child("ips").Index(i)
		if seenIPNames.Has(ip.Name) {
			allErrs = append(allErrs, field.Duplicate(ipFieldPath.Child("name"), ip.Name))
		} else {
			seenIPNames.Insert(ip.Name)
		}
		allErrs = append(allErrs, validateNATGatewayIP(ip, ipFieldPath)...)
	}

	for _, msg := range apivalidation.NameIsDNSLabel(spec.NetworkRef.Name, false) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("networkRef").Child("name"), spec.NetworkRef.Name, msg))
	}

	if spec.PortsPerNetworkInterface != nil {
		allErrs = append(allErrs, onmetalapivalidation.ValidatePowerOfTwo(int64(*spec.PortsPerNetworkInterface), fldPath.Child("portsPerNetworkInterface"))...)
	}

	allErrs = append(allErrs, metav1validation.ValidateLabelSelector(spec.NetworkInterfaceSelector, metav1validation.LabelSelectorValidationOptions{}, fldPath.Child("networkInterfaceSelector"))...)

	return allErrs
}

var supportedNATGatewayTypes = sets.New(
	networking.NATGatewayTypePublic,
)

func validateNATGatewayType(natGatewayType networking.NATGatewayType, fldPath *field.Path) field.ErrorList {
	return onmetalapivalidation.ValidateEnum(supportedNATGatewayTypes, natGatewayType, fldPath, "must specify type")
}

func validateNATGatewayIP(ip networking.NATGatewayIP, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	for _, msg := range apivalidation.NameIsDNSLabel(ip.Name, false) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("name"), ip.Name, msg))
	}

	return allErrs
}

// ValidateNATGatewayUpdate validates a NATGateway object before an update.
func ValidateNATGatewayUpdate(newNATGateway, oldNATGateway *networking.NATGateway) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessorUpdate(newNATGateway, oldNATGateway, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateNATGatewaySpecPrefixUpdate(&newNATGateway.Spec, &oldNATGateway.Spec, field.NewPath("spec"))...)
	allErrs = append(allErrs, ValidateNATGateway(newNATGateway)...)

	return allErrs
}

// validateNATGatewaySpecPrefixUpdate validates the spec of a natGateway object before an update.
func validateNATGatewaySpecPrefixUpdate(newSpec, oldSpec *networking.NATGatewaySpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, onmetalapivalidation.ValidateImmutableField(newSpec.NetworkRef, oldSpec.NetworkRef, fldPath.Child("networkRef"))...)

	return allErrs
}
