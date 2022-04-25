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
	onmetalapivalidation "github.com/onmetal/onmetal-api/api/validation"
	"github.com/onmetal/onmetal-api/apis/networking"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1validation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateVirtualIP validates a virtual ip object.
func ValidateVirtualIP(virtualIP *networking.VirtualIP) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessor(virtualIP, true, apivalidation.NameIsDNSLabel, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateVirtualIPSpec(&virtualIP.Spec, field.NewPath("spec"))...)

	return allErrs
}

// ValidateVirtualIPUpdate validates a VirtualIP object before an update.
func ValidateVirtualIPUpdate(newVirtualIP, oldVirtualIP *networking.VirtualIP) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessorUpdate(newVirtualIP, oldVirtualIP, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateVirtualIPSpecUpdate(&newVirtualIP.Spec, &oldVirtualIP.Spec, field.NewPath("spec"))...)
	allErrs = append(allErrs, ValidateVirtualIP(newVirtualIP)...)

	return allErrs
}

var supportedVirtualIPTypes = sets.NewString(
	string(networking.VirtualIPTypePublic),
)

func validateVirtualIPSpec(spec *networking.VirtualIPSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, onmetalapivalidation.ValidateStringSetEnum(supportedVirtualIPTypes, string(spec.Type), fldPath.Child("type"), "must specify type")...)
	allErrs = append(allErrs, onmetalapivalidation.ValidateIPFamily(spec.IPFamily, fldPath.Child("ipFamily"))...)
	allErrs = append(allErrs, metav1validation.ValidateLabelSelector(spec.NetworkInterfaceSelector, fldPath.Child("networkInterfaceSelector"))...)

	return allErrs
}

// validateVirtualIPSpecUpdate validates the spec of a VirtualIP object before an update.
func validateVirtualIPSpecUpdate(newSpec, oldSpec *networking.VirtualIPSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	newSpecCopy := newSpec.DeepCopy()
	oldSpecCopy := oldSpec.DeepCopy()

	oldSpecCopy.NetworkInterfaceSelector = newSpec.NetworkInterfaceSelector
	allErrs = append(allErrs, onmetalapivalidation.ValidateImmutableFieldWithDiff(newSpecCopy, oldSpecCopy, fldPath)...)

	return allErrs
}
