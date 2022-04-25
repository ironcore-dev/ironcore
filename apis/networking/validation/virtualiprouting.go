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
	"k8s.io/apimachinery/pkg/util/validation/field"
)

type virtualIPRoutingSubsetKey struct {
	IP  commonv1alpha1.IP
	Ref networking.LocalUIDReference
}
type virtualIPRoutingSubsetKeySet map[virtualIPRoutingSubsetKey]struct{}

// ValidateVirtualIPRouting validates a virtual ip object.
func ValidateVirtualIPRouting(virtualIPRouting *networking.VirtualIPRouting) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessor(virtualIPRouting, true, apivalidation.NameIsDNSLabel, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateVirtualIPRoutingSubsets(virtualIPRouting, field.NewPath("subsets"))...)

	return allErrs
}

func validateVirtualIPRoutingSubsets(virtualIPRouting *networking.VirtualIPRouting, subsetsField *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	seen := make(virtualIPRoutingSubsetKeySet)
	for idx := range virtualIPRouting.Subsets {
		subset := &virtualIPRouting.Subsets[idx]

		allErrs = append(allErrs, validateVirtualIPRoutingSubset(subset, subsetsField.Index(idx))...)

		key := virtualIPRoutingSubsetKey{IP: subset.IP, Ref: subset.TargetRef}
		if _, ok := seen[key]; ok {
			allErrs = append(allErrs, field.Duplicate(subsetsField.Index(idx), subset))
		}
		seen[key] = struct{}{}
	}

	return allErrs
}

func validateVirtualIPRoutingSubset(subset *networking.VirtualIPRoutingSubset, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, validation.ValidateIP(subset.IP.Family(), subset.IP, fldPath.Child("ip"))...)

	if subset.TargetRef.Name == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("targetRef", "name"), "must specify network interface ref name"))
	}
	for _, msg := range apivalidation.NameIsDNSLabel(subset.TargetRef.Name, false) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("targetRef", "name"), subset.TargetRef.Name, msg))
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
