// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	ironcorevalidation "github.com/ironcore-dev/ironcore/internal/api/validation"
	"github.com/ironcore-dev/ironcore/internal/apis/networking"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
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

var supportedVirtualIPTypes = sets.New(
	networking.VirtualIPTypePublic,
)

func validateVirtualIPType(virtualIPType networking.VirtualIPType, fldPath *field.Path) field.ErrorList {
	return ironcorevalidation.ValidateEnum(supportedVirtualIPTypes, virtualIPType, fldPath, "must specify type")
}

func validateVirtualIPSpec(spec *networking.VirtualIPSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, validateVirtualIPType(spec.Type, fldPath.Child("type"))...)
	allErrs = append(allErrs, ironcorevalidation.ValidateIPFamily(spec.IPFamily, fldPath.Child("ipFamily"))...)

	if targetRef := spec.TargetRef; targetRef != nil {
		for _, msg := range apivalidation.NameIsDNSSubdomain(targetRef.Name, false) {
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
	allErrs = append(allErrs, ironcorevalidation.ValidateImmutableFieldWithDiff(newSpecCopy, oldSpecCopy, fldPath)...)

	return allErrs
}
