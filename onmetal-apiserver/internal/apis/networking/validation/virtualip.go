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
	"github.com/onmetal/controller-utils/set"
	onmetalapivalidation "github.com/onmetal/onmetal-api/onmetal-apiserver/internal/api/validation"
	"github.com/onmetal/onmetal-api/onmetal-apiserver/internal/apis/networking"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
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

var supportedVirtualIPTypes = set.New(
	networking.VirtualIPTypePublic,
)

func validateVirtualIPType(virtualIPType networking.VirtualIPType, fldPath *field.Path) field.ErrorList {
	return onmetalapivalidation.ValidateEnum(supportedVirtualIPTypes, virtualIPType, fldPath, "must specify type")
}

func validateVirtualIPSpec(spec *networking.VirtualIPSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, validateVirtualIPType(spec.Type, fldPath.Child("type"))...)
	allErrs = append(allErrs, onmetalapivalidation.ValidateIPFamily(spec.IPFamily, fldPath.Child("ipFamily"))...)

	if targetRef := spec.TargetRef; targetRef != nil {
		for _, msg := range apivalidation.NameIsDNSLabel(targetRef.Name, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("targetRef", "name"), targetRef.Name, msg))
		}
	}

	return allErrs
}

// validateVirtualIPSpecUpdate validates the spec of a VirtualIP object before an update.
func validateVirtualIPSpecUpdate(newSpec, oldSpec *networking.VirtualIPSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	newSpecCopy := newSpec.DeepCopy()
	oldSpecCopy := oldSpec.DeepCopy()

	oldSpecCopy.TargetRef = newSpec.TargetRef
	allErrs = append(allErrs, onmetalapivalidation.ValidateImmutableFieldWithDiff(newSpecCopy, oldSpecCopy, fldPath)...)

	return allErrs
}
