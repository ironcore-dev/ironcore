// Copyright 2022 OnMetal authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package validation

import (
	onmetalapivalidation "github.com/onmetal/onmetal-api/api/validation"
	"github.com/onmetal/onmetal-api/apis/networking"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func ValidateVirtualIPClaim(vipClaim *networking.VirtualIPClaim) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessor(vipClaim, true, apivalidation.NameIsDNSLabel, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateVirtualIPClaimSpec(&vipClaim.Spec, field.NewPath("spec"))...)

	return allErrs
}

func validateVirtualIPClaimSpec(spec *networking.VirtualIPClaimSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, validateVirtualIPType(spec.Type, fldPath.Child("type"))...)
	allErrs = append(allErrs, onmetalapivalidation.ValidateIPFamily(spec.IPFamily, fldPath.Child("ipFamily"))...)

	if vipRef := spec.VirtualIPRef; vipRef != nil {
		for _, msg := range apivalidation.NameIsDNSLabel(vipRef.Name, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("virtualIPRef", "name"), vipRef.Name, msg))
		}
	}

	return allErrs
}

func ValidateVirtualIPClaimUpdate(newVipClaim, oldVipClaim *networking.VirtualIPClaim) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessorUpdate(newVipClaim, oldVipClaim, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateVirtualIPClaimSpecUpdate(&newVipClaim.Spec, &oldVipClaim.Spec, field.NewPath("spec"))...)
	allErrs = append(allErrs, ValidateVirtualIPClaim(newVipClaim)...)

	return allErrs
}

func validateVirtualIPClaimSpecUpdate(newSpec, oldSpec *networking.VirtualIPClaimSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	newSpecCopy := newSpec.DeepCopy()
	oldSpecCopy := oldSpec.DeepCopy()

	if oldSpec.VirtualIPRef == nil {
		oldSpecCopy.VirtualIPRef = newSpecCopy.VirtualIPRef
	}

	allErrs = append(allErrs, onmetalapivalidation.ValidateImmutableFieldWithDiff(newSpecCopy, oldSpecCopy, fldPath)...)

	return allErrs
}
