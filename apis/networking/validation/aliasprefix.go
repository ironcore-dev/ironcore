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
	"github.com/onmetal/onmetal-api/apis/ipam"
	ipamvalidation "github.com/onmetal/onmetal-api/apis/ipam/validation"
	"github.com/onmetal/onmetal-api/apis/networking"
	corev1 "k8s.io/api/core/v1"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1validation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateAliasPrefix validates an AliasPrefix object.
func ValidateAliasPrefix(aliasPrefix *networking.AliasPrefix) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessor(aliasPrefix, true, apivalidation.NameIsDNSLabel, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateAliasPrefixSpec(&aliasPrefix.Spec, field.NewPath("spec"))...)

	return allErrs
}

func validateAliasPrefixSpec(spec *networking.AliasPrefixSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if spec.NetworkRef == (corev1.LocalObjectReference{}) {
		allErrs = append(allErrs, field.Required(fldPath.Child("networkRef"), "must specify a network ref"))
	}

	for _, msg := range apivalidation.NameIsDNSLabel(spec.NetworkRef.Name, false) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("networkRef").Child("name"), spec.NetworkRef.Name, msg))
	}

	allErrs = append(allErrs, metav1validation.ValidateLabelSelector(spec.NetworkInterfaceSelector, fldPath.Child("networkInterfaceSelector"))...)
	allErrs = append(allErrs, validateAliasPrefixSources(spec.Prefix, fldPath.Child("prefix"))...)

	return allErrs
}

func validateAliasPrefixSources(prefixSource networking.PrefixSource, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if ephemeralPrefix := prefixSource.EphemeralPrefix; ephemeralPrefix != nil {
		allErrs = append(allErrs, validateEphemeralAliasPrefixSource(prefixSource.EphemeralPrefix, fldPath.Child("ephemeralPrefix"))...)
	}

	allErrs = append(allErrs, validateValuePrefixSource(prefixSource.Value, fldPath.Child("value"))...)

	return allErrs
}

func validateEphemeralAliasPrefixSource(source *networking.EphemeralPrefixSource, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, ValidatePrefixTemplateForAliasPrefix(source.PrefixTemplateSpec, fldPath)...)

	return allErrs
}

func ValidatePrefixTemplateForAliasPrefix(template *ipam.PrefixTemplateSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if template == nil {
		allErrs = append(allErrs, field.Required(fldPath, ""))
	} else {
		allErrs = append(allErrs, ipamvalidation.ValidatePrefixTemplateSpec(template, fldPath)...)
	}

	return allErrs
}

func validateValuePrefixSource(value *commonv1alpha1.IPPrefix, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if value != nil && !value.IsValid() {
		allErrs = append(allErrs, field.Invalid(fldPath, value, "must specify a valid prefix"))
	}

	return allErrs
}

// ValidateAliasPrefixUpdate validates a AliasPrefix object before an update.
func ValidateAliasPrefixUpdate(newAliasPrefix, oldAliasPrefix *networking.AliasPrefix) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessorUpdate(newAliasPrefix, oldAliasPrefix, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateAliasPrefixUpdate(newAliasPrefix, oldAliasPrefix, field.NewPath("spec"))...)
	allErrs = append(allErrs, ValidateAliasPrefix(newAliasPrefix)...)

	return allErrs
}

// validateAliasPrefixUpdate validates the spec of a aliasPrefix object before an update.
func validateAliasPrefixUpdate(new, old *networking.AliasPrefix, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, onmetalapivalidation.ValidateImmutableField(new, old, fldPath.Child("networkRef"))...)

	return allErrs
}
