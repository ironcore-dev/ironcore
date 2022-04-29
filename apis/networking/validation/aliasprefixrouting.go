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
	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	"github.com/onmetal/onmetal-api/apis/networking"
	corev1 "k8s.io/api/core/v1"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateAliasPrefixRouting validates an AliasPrefixRouting object.
func ValidateAliasPrefixRouting(aliasPrefixRouting *networking.AliasPrefixRouting) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessor(aliasPrefixRouting, true, apivalidation.NameIsDNSLabel, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateAliasPrefixRouting(aliasPrefixRouting)...)
	allErrs = append(allErrs, validateAliasPrefixRoutingSubsets(aliasPrefixRouting, field.NewPath("subsets"))...)

	return allErrs
}

func validateAliasPrefixRouting(aliasPrefixRouting *networking.AliasPrefixRouting) field.ErrorList {
	var allErrs field.ErrorList

	if aliasPrefixRouting.NetworkRef == (corev1.LocalObjectReference{}) {
		allErrs = append(allErrs, field.Required(field.NewPath("networkRef"), "must specify a network ref"))
	} else {
		for _, msg := range apivalidation.NameIsDNSLabel(aliasPrefixRouting.NetworkRef.Name, false) {
			allErrs = append(allErrs, field.Invalid(field.NewPath("networkRef").Child("name"), aliasPrefixRouting.NetworkRef.Name, msg))
		}
	}

	return allErrs
}

func validateAliasPrefixRoutingSubsets(aliasPrefixRouting *networking.AliasPrefixRouting, subsetsField *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	seen := make(map[commonv1alpha1.LocalUIDReference]struct{})
	for idx := range aliasPrefixRouting.Subsets {
		subset := &aliasPrefixRouting.Subsets[idx]

		allErrs = append(allErrs, validateAliasPrefixRoutingSubset(subset, subsetsField.Index(idx))...)

		key := subset.TargetRef
		if _, ok := seen[key]; ok {
			allErrs = append(allErrs, field.Duplicate(subsetsField.Index(idx), subset))
		}
		seen[key] = struct{}{}
	}

	return allErrs
}

func validateAliasPrefixRoutingSubset(subset *networking.AliasPrefixSubset, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if subset.TargetRef.Name == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("targetRef", "name"), "must specify network interface ref name"))
	} else {
		for _, msg := range apivalidation.NameIsDNSLabel(subset.TargetRef.Name, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("targetRef", "name"), subset.TargetRef.Name, msg))
		}
	}

	return allErrs
}

// ValidateAliasPrefixRoutingUpdate validates a AliasPrefixRouting object before an update.
func ValidateAliasPrefixRoutingUpdate(newAliasPrefixRouting, oldAliasPrefixRouting *networking.AliasPrefixRouting) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessorUpdate(newAliasPrefixRouting, oldAliasPrefixRouting, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateAliasPrefixRoutingUpdate(newAliasPrefixRouting, oldAliasPrefixRouting)...)
	allErrs = append(allErrs, ValidateAliasPrefixRouting(newAliasPrefixRouting)...)

	return allErrs
}

// validateAliasPrefixRoutingUpdate validates the spec of a aliasPrefixRouting object before an update.
func validateAliasPrefixRoutingUpdate(new, old *networking.AliasPrefixRouting) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, onmetalapivalidation.ValidateImmutableField(new, old, field.NewPath("networkRef"))...)

	return allErrs
}
