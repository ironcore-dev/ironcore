// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"fmt"

	ironcorevalidation "github.com/ironcore-dev/ironcore/internal/api/validation"
	commonvalidation "github.com/ironcore-dev/ironcore/internal/apis/common/validation"
	"github.com/ironcore-dev/ironcore/internal/apis/ipam"
	"github.com/ironcore-dev/ironcore/utils/equality"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1validation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func ValidatePrefixAllocation(prefixAllocation *ipam.PrefixAllocation) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessor(prefixAllocation, true, apivalidation.NameIsDNSLabel, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validatePrefixAllocationSpec(&prefixAllocation.Spec, field.NewPath("spec"))...)

	return allErrs
}

func validatePrefixAllocationSpec(spec *ipam.PrefixAllocationSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, validateIPFamilyAndOptionalPrefixAndLength(spec.IPFamily, spec.Prefix, spec.PrefixLength, fldPath)...)
	allErrs = append(allErrs, validateOptionalRef(spec.PrefixRef, fldPath.Child("prefixRef"))...)
	allErrs = append(allErrs, metav1validation.ValidateLabelSelector(spec.PrefixSelector, metav1validation.LabelSelectorValidationOptions{}, fldPath.Child("prefixSelector"))...)

	var numRequests int
	if spec.Prefix != nil {
		numRequests++
	}
	if spec.PrefixLength > 0 {
		if numRequests > 0 {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("prefixLength"), "must not specify more than 1 request"))
		} else {
			numRequests++
		}
	}
	if numRequests == 0 {
		allErrs = append(allErrs, field.Invalid(fldPath, spec, "must specify a request"))
	}

	var numSources int
	if spec.PrefixRef != nil {
		numSources++
	}
	if spec.PrefixSelector != nil {
		numSources++
	}
	if numSources == 0 {
		allErrs = append(allErrs, field.Invalid(fldPath, spec, "must specify a source"))
	}

	return allErrs
}

func ValidatePrefixAllocationUpdate(newPrefixAllocation, oldPrefixAllocation *ipam.PrefixAllocation) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessorUpdate(newPrefixAllocation, oldPrefixAllocation, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validatePrefixAllocationSpecUpdate(&newPrefixAllocation.Spec, &oldPrefixAllocation.Spec, field.NewPath("spec"))...)
	allErrs = append(allErrs, ValidatePrefixAllocation(newPrefixAllocation)...)

	return allErrs
}

func validatePrefixAllocationSpecUpdate(newSpec, oldSpec *ipam.PrefixAllocationSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	newSpecCopy := newSpec.DeepCopy()
	oldSpecCopy := oldSpec.DeepCopy()

	if oldSpec.PrefixRef == nil {
		oldSpecCopy.PrefixRef = newSpecCopy.PrefixRef
	}

	allErrs = append(allErrs, ironcorevalidation.ValidateImmutableFieldWithDiff(newSpecCopy, oldSpecCopy, fldPath)...)

	return allErrs
}

func ValidatePrefixAllocationStatus(status *ipam.PrefixAllocationStatus, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if status.Prefix != nil {
		allErrs = append(allErrs, commonvalidation.ValidateIPPrefix(status.Prefix.IP().Family(), *status.Prefix, fldPath.Child("prefix"))...)
	}

	switch status.Phase {
	case ipam.PrefixAllocationPhaseAllocated:
		if status.Prefix == nil {
			allErrs = append(allErrs, field.Required(fldPath.Child("prefix"), "must specify prefix when allocated"))
		}
	default:
		if status.Prefix != nil {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("prefix"), "must not specify a prefix when not allocated"))
		}
	}

	return allErrs
}

func ValidatePrefixAllocationStatusUpdate(newPrefixAllocation, oldPrefixAllocation *ipam.PrefixAllocation) field.ErrorList {
	var allErrs field.ErrorList

	statusField := field.NewPath("status")

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessorUpdate(newPrefixAllocation, oldPrefixAllocation, field.NewPath("metadata"))...)
	allErrs = append(allErrs, ValidatePrefixAllocationStatus(&newPrefixAllocation.Status, statusField)...)

	newPhase := newPrefixAllocation.Status.Phase
	oldPhase := oldPrefixAllocation.Status.Phase
	if (oldPhase == ipam.PrefixAllocationPhaseFailed || oldPhase == ipam.PrefixAllocationPhaseAllocated) && newPhase != oldPhase {
		allErrs = append(allErrs, field.Forbidden(field.NewPath("status", "phase"), "must not set failed / allocated allocation to non-failed / non-allocated"))
	}

	statusPrefixField := statusField.Child("prefix")
	if newStatusPrefix := newPrefixAllocation.Status.Prefix; newStatusPrefix != nil {
		allErrs = append(allErrs, commonvalidation.ValidateIPPrefix(newPrefixAllocation.Spec.IPFamily, *newStatusPrefix, statusPrefixField)...)

		if newSpecPrefix := newPrefixAllocation.Spec.Prefix; newSpecPrefix.IsValid() {
			if !equality.Semantic.DeepEqual(newStatusPrefix, newSpecPrefix) {
				allErrs = append(allErrs, field.Forbidden(statusPrefixField, fmt.Sprintf("does not match spec prefix %s", newSpecPrefix)))
			}
		}

		if newSpecPrefixLength := newPrefixAllocation.Spec.PrefixLength; newSpecPrefixLength > 0 {
			if int32(newStatusPrefix.Bits()) != newSpecPrefixLength {
				allErrs = append(allErrs, field.Forbidden(statusPrefixField, fmt.Sprintf("does not match spec prefix length %d", newSpecPrefixLength)))
			}
		}

		if newPrefixAllocation.Spec.PrefixRef == nil {
			allErrs = append(allErrs, field.Forbidden(statusPrefixField, "spec.prefixRef needs to be set first"))
		}
	}

	return allErrs
}
