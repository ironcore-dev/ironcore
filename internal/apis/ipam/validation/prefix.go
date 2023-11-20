// Copyright 2022 IronCore authors
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

	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	ironcorevalidation "github.com/ironcore-dev/ironcore/internal/api/validation"
	commonvalidation "github.com/ironcore-dev/ironcore/internal/apis/common/validation"
	"github.com/ironcore-dev/ironcore/internal/apis/ipam"
	"go4.org/netipx"
	corev1 "k8s.io/api/core/v1"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1validation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func ValidatePrefix(prefix *ipam.Prefix) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessor(prefix, true, apivalidation.NameIsDNSLabel, field.NewPath("metadata"))...)
	allErrs = append(allErrs, ValidatePrefixSpec(&prefix.Spec, field.NewPath("spec"))...)

	return allErrs
}

func validateIPFamilyAndOptionalPrefixAndLength(ipFamily corev1.IPFamily, prefix *commonv1alpha1.IPPrefix, prefixLength int32, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, ironcorevalidation.ValidateIPFamily(ipFamily, fldPath.Child("ipFamily"))...)

	if prefix != nil {
		allErrs = append(allErrs, commonvalidation.ValidateIPPrefix(ipFamily, *prefix, fldPath.Child("prefix"))...)
	}
	if prefixLength != 0 {
		allErrs = append(allErrs, commonvalidation.ValidatePrefixLength(ipFamily, prefixLength, fldPath.Child("prefixLength"))...)
	}

	return allErrs
}

func validateOptionalRef(ref *corev1.LocalObjectReference, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if ref == nil {
		return allErrs
	}

	if ref.Name == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("name"), "must specify name"))
	}

	for _, msg := range apivalidation.NameIsDNSLabel(ref.Name, false) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("name"), ref.Name, msg))
	}
	return allErrs
}

func ValidatePrefixSpec(spec *ipam.PrefixSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if spec.Prefix != nil {
		allErrs = append(allErrs, commonvalidation.ValidateIPPrefix(spec.IPFamily, *spec.Prefix, fldPath.Child("prefix"))...)
	}
	allErrs = append(allErrs, validateIPFamilyAndOptionalPrefixAndLength(spec.IPFamily, spec.Prefix, spec.PrefixLength, fldPath)...)

	if spec.IsRoot() {
		if spec.PrefixLength != 0 {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("prefixLength"), spec.PrefixLength, "cannot specify prefixLength for a root prefix"))
		}
		if spec.Prefix == nil {
			allErrs = append(allErrs, field.Required(fldPath.Child("prefix"), "must specify for root prefix"))
		}
	} else {
		var numDefs int
		if spec.Prefix != nil {
			numDefs++
		}
		if spec.PrefixLength > 0 {
			numDefs++
			if spec.Prefix.IsValid() && spec.PrefixLength != int32(spec.Prefix.Bits()) {
				allErrs = append(allErrs, field.Invalid(fldPath.Child("prefixLength"), spec.PrefixLength, fmt.Sprintf("does not match prefix %s", spec.Prefix)))
			}
		}
		if numDefs == 0 {
			allErrs = append(allErrs, field.Invalid(fldPath, spec, "must specify either prefix or prefixLength"))
		}

		allErrs = append(allErrs, validateOptionalRef(spec.ParentRef, fldPath.Child("parentRef"))...)
		allErrs = append(allErrs, metav1validation.ValidateLabelSelector(spec.ParentSelector, metav1validation.LabelSelectorValidationOptions{}, fldPath.Child("parentSelector"))...)
	}

	return allErrs
}

func ValidatePrefixUpdate(newPrefix, oldPrefix *ipam.Prefix) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessorUpdate(newPrefix, oldPrefix, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validatePrefixSpecUpdate(&newPrefix.Spec, &oldPrefix.Spec, field.NewPath("spec"))...)
	allErrs = append(allErrs, ValidatePrefix(newPrefix)...)

	return allErrs
}

func validatePrefixSpecUpdate(newSpec, oldSpec *ipam.PrefixSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	newSpecCopy := newSpec.DeepCopy()
	oldSpecCopy := oldSpec.DeepCopy()

	// Allow setting Prefix, PrefixLength and ParentRef exactly once.
	if oldSpec.Prefix == nil {
		oldSpecCopy.Prefix = newSpecCopy.Prefix
	}
	// We allow setting PrefixLength for symmetry with setting Prefix.
	// ValidatePrefixSpec will guard against any mismatch with prefix.
	if oldSpec.PrefixLength == 0 {
		oldSpecCopy.PrefixLength = newSpecCopy.PrefixLength
	}
	// ParentRef can only be set if the prefix is not a root prefix.
	if oldSpec.ParentRef == nil && !oldSpec.IsRoot() {
		oldSpecCopy.ParentRef = newSpecCopy.ParentRef
	}

	allErrs = append(allErrs, ironcorevalidation.ValidateImmutableFieldWithDiff(newSpecCopy, oldSpecCopy, fldPath)...)

	return allErrs
}

func ValidatePrefixStatus(status *ipam.PrefixStatus, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	phase := status.Phase
	switch phase {
	case ipam.PrefixPhaseAllocated:
		var bldr netipx.IPSetBuilder

		var (
			seenFamilies = sets.NewString()
			overlapSeen  bool
		)
		for i, used := range status.Used {
			usedField := fldPath.Child("used").Index(i)
			allErrs = append(allErrs, commonvalidation.ValidateIPPrefix(used.IP().Family(), used, usedField)...)
			seenFamilies.Insert(string(used.IP().Family()))

			if !overlapSeen {
				ipSet, _ := bldr.IPSet()
				var usedBldr netipx.IPSetBuilder
				usedBldr.AddPrefix(used.Prefix)
				usedSet, _ := usedBldr.IPSet()

				overlapSeen = ipSet.Overlaps(usedSet)
				bldr.AddSet(usedSet)
			}
		}
		if seenFamilies.Len() > 1 {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("used"), status.Used, fmt.Sprintf("different bit lengths among prefixes: %v", seenFamilies.List())))
		}
		if overlapSeen {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("used"), status.Used, "overlapping used addresses"))
		}
	default:
		if len(status.Used) > 0 {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("used"), status.Used, "must not specify used if the prefix is not succeeded"))
		}
	}

	return allErrs
}

func ValidatePrefixStatusUpdate(newPrefix, oldPrefix *ipam.Prefix) field.ErrorList {
	var allErrs field.ErrorList

	statusField := field.NewPath("status")

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessorUpdate(newPrefix, oldPrefix, field.NewPath("metadata"))...)
	allErrs = append(allErrs, ValidatePrefixStatus(&newPrefix.Status, statusField)...)

	newPhase := newPrefix.Status.Phase
	oldPhase := oldPrefix.Status.Phase

	if oldPhase == ipam.PrefixPhaseAllocated && oldPhase != newPhase {
		allErrs = append(allErrs, field.Forbidden(field.NewPath("status", "phase"), "may not update from allocated to non-allocated phase"))
	}

	if newPhase == ipam.PrefixPhaseAllocated {
		if !newPrefix.Spec.Prefix.IsValid() {
			allErrs = append(allErrs, field.Required(field.NewPath("spec", "prefix"), "must specify prefix if allocated"))
		}

		if newPrefix.Spec.ParentSelector != nil && newPrefix.Spec.ParentRef == nil {
			allErrs = append(allErrs, field.Required(field.NewPath("spec", "parentRef"), "must specify parentRef when allocating sub-prefix"))
		}
	}

	// We only have to validate that used / reserved prefixes are contained in the allocated prefix.
	// ValidatePrefixStatus already validates that they're non-overlapping.
	if prefix := newPrefix.Spec.Prefix; prefix.IsValid() {
		var bldr netipx.IPSetBuilder
		bldr.AddPrefix(prefix.Prefix)
		set, _ := bldr.IPSet()

		for i, used := range newPrefix.Status.Used {
			if !set.ContainsPrefix(used.Prefix) {
				allErrs = append(allErrs, field.Forbidden(statusField.Child("used").Index(i), fmt.Sprintf("not contained in prefix %s", prefix)))
			}
		}
	}

	return allErrs
}
