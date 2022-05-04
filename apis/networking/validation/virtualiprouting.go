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
	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	"github.com/onmetal/onmetal-api/apis/common/v1alpha1/validation"
	"github.com/onmetal/onmetal-api/apis/networking"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateVirtualIPRouting validates a virtual ip object.
func ValidateVirtualIPRouting(virtualIPRouting *networking.VirtualIPRouting) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessor(virtualIPRouting, true, apivalidation.NameIsDNSLabel, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateVirtualIPRoutingSubsets(virtualIPRouting, field.NewPath("subsets"))...)

	return allErrs
}

func validateVirtualIPRoutingSubsets(virtualIPRouting *networking.VirtualIPRouting, subsetsField *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	seenNetworkNames := sets.NewString()
	for idx := range virtualIPRouting.Subsets {
		subset := &virtualIPRouting.Subsets[idx]

		allErrs = append(allErrs, validateVirtualIPRoutingSubset(subset, subsetsField.Index(idx))...)

		if seenNetworkNames.Has(subset.NetworkRef.Name) {
			allErrs = append(allErrs, field.Duplicate(subsetsField.Index(idx), subset))
		}
	}

	return allErrs
}

func validateVirtualIPRoutingSubset(subset *networking.VirtualIPRoutingSubset, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if subset.NetworkRef.Name == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("networkRef", "name"), "must specify network name"))
	} else {
		for _, msg := range apivalidation.NameIsDNSLabel(subset.NetworkRef.Name, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("networkRef", "name"), subset.NetworkRef.Name, msg))
		}
	}

	if len(subset.Targets) == 0 {
		allErrs = append(allErrs, field.Invalid(fldPath, subset, "may not specify no targets"))
	} else {
		type targetKey struct {
			name string
			ip   commonv1alpha1.IP
		}
		seenTargetKeys := make(map[targetKey]struct{})
		for idx := range subset.Targets {
			target := &subset.Targets[idx]
			allErrs = append(allErrs, validateVirtualIPRoutingSubsetTarget(target, fldPath.Child("targets").Index(idx))...)
			key := targetKey{target.Name, target.IP}
			if _, ok := seenTargetKeys[key]; ok {
				allErrs = append(allErrs, field.Duplicate(fldPath.Child("targets").Index(idx), target))
			} else {
				seenTargetKeys[key] = struct{}{}
			}
		}
	}
	return allErrs
}

func validateVirtualIPRoutingSubsetTarget(target *networking.VirtualIPRoutingSubsetTarget, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, validation.ValidateIP(target.IP.Family(), target.IP, fldPath.Child("ip"))...)

	if target.Name == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("name"), "must specify target name"))
	} else {
		for _, msg := range apivalidation.NameIsDNSLabel(target.Name, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("name"), target.Name, msg))
		}
	}

	return allErrs
}

// ValidateVirtualIPRoutingUpdate validates a VirtualIPRouting object before an update.
func ValidateVirtualIPRoutingUpdate(newVirtualIPRouting, oldVirtualIPRouting *networking.VirtualIPRouting) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessorUpdate(newVirtualIPRouting, oldVirtualIPRouting, field.NewPath("metadata"))...)
	allErrs = append(allErrs, ValidateVirtualIPRouting(newVirtualIPRouting)...)

	return allErrs
}
