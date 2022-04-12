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
	"fmt"

	"github.com/onmetal/onmetal-api/apis/ipam"
	"github.com/onmetal/onmetal-api/equality"
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

	allErrs = append(allErrs, validateIPFamilyPrefixAndLength(spec.IPFamily, spec.Prefix, spec.PrefixLength, fldPath)...)
	allErrs = append(allErrs, validateOptionalRef(spec.PrefixRef, fldPath.Child("prefixRef"))...)
	allErrs = append(allErrs, metav1validation.ValidateLabelSelector(spec.PrefixSelector, fldPath.Child("prefixSelector"))...)

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

	allErrs = append(allErrs, ValidateImmutableWithDiff(newSpecCopy, oldSpecCopy, fldPath)...)

	return allErrs
}

func ValidatePrefixAllocationStatus(status *ipam.PrefixAllocationStatus, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, validateOptionalPrefix(status.Prefix, fldPath.Child("prefix"))...)

	readiness, _ := ipam.GetPrefixAllocationConditionsReadinessAndIndex(status.Conditions)
	switch readiness {
	case ipam.ReadinessSucceeded:
		if status.Prefix == nil {
			allErrs = append(allErrs, field.Required(fldPath.Child("prefix"), "must specify prefix when succeeded"))
		}
	default:
		if status.Prefix != nil {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("prefix"), "must not specify a prefix when not succeeded"))
		}
	}

	return allErrs
}

func ValidatePrefixAllocationStatusUpdate(newPrefixAllocation, oldPrefixAllocation *ipam.PrefixAllocation) field.ErrorList {
	var allErrs field.ErrorList

	statusField := field.NewPath("status")

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessorUpdate(newPrefixAllocation, oldPrefixAllocation, field.NewPath("metadata"))...)
	allErrs = append(allErrs, ValidatePrefixAllocationStatus(&newPrefixAllocation.Status, statusField)...)

	conditionsField := statusField.Child("conditions")

	newReadiness, newReadyIdx := ipam.GetPrefixAllocationConditionsReadinessAndIndex(newPrefixAllocation.Status.Conditions)
	oldReadiness, _ := ipam.GetPrefixAllocationConditionsReadinessAndIndex(oldPrefixAllocation.Status.Conditions)
	if oldReadiness.Terminal() && newReadiness != oldReadiness {
		if newReadyIdx < 0 {
			allErrs = append(allErrs, field.Required(conditionsField.Index(0), "terminal ready condition is missing"))
		} else {
			allErrs = append(allErrs, field.Forbidden(conditionsField.Index(newReadyIdx), "may not change terminal ready condition"))
		}
	}

	statusPrefixField := statusField.Child("prefix")
	allErrs = append(allErrs, validateOptionalPrefixAndIPFamily(newPrefixAllocation.Spec.IPFamily, newPrefixAllocation.Status.Prefix, statusPrefixField)...)
	if newStatusPrefix := newPrefixAllocation.Status.Prefix; newStatusPrefix != nil {
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
