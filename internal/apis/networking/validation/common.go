// Copyright 2023 OnMetal authors
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

	"github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	commonvalidation "github.com/onmetal/onmetal-api/internal/apis/common/validation"
	"github.com/onmetal/onmetal-api/internal/apis/networking"
	corev1 "k8s.io/api/core/v1"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func validateIPSource(ipSource networking.IPSource, idx int, ipFamily corev1.IPFamily, objectMeta *metav1.ObjectMeta, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	var numSources int
	if ip := ipSource.Value; ip.IsValid() {
		numSources++
		allErrs = append(allErrs, commonvalidation.ValidateIP(ipFamily, *ip, fldPath.Child("value"))...)
	}
	if ephemeral := ipSource.Ephemeral; ephemeral != nil {
		if numSources > 0 {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("ephemeral"), ephemeral, "cannot specify multiple ip sources"))
		} else {
			numSources++
			allErrs = append(allErrs, validateEphemeralPrefixSource(ipFamily, ephemeral, fldPath.Child("ephemeral"))...)
			if objectMeta != nil && objectMeta.Name != "" {
				prefixName := v1alpha1.NetworkInterfaceIPIPAMPrefixName(objectMeta.Name, idx)
				for _, msg := range apivalidation.NameIsDNSLabel(prefixName, false) {
					allErrs = append(allErrs, field.Invalid(fldPath, prefixName, fmt.Sprintf("resulting prefix name %q is invalid: %s", prefixName, msg)))
				}
			}
		}
	}
	if numSources == 0 {
		allErrs = append(allErrs, field.Invalid(fldPath, ipSource, "must specify an ip source"))
	}

	return allErrs
}

func validatePrefixSource(src networking.PrefixSource, idx int, objectMeta *metav1.ObjectMeta, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	var numSources int
	if prefix := src.Value; prefix.IsValid() {
		numSources++
		if !prefix.IsValid() {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("value"), prefix, "must specify valid prefix"))
		}
	}
	if ephemeral := src.Ephemeral; ephemeral != nil {
		if numSources > 0 {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("ephemeral"), ephemeral, "cannot specify multiple ip sources"))
		} else {
			numSources++
			allErrs = append(allErrs, ValidatePrefixPrefixTemplate(ephemeral.PrefixTemplate, fldPath.Child("ephemeral"))...)
			if objectMeta != nil && objectMeta.Name != "" {
				prefixName := networking.NetworkInterfacePrefixIPAMPrefixName(objectMeta.Name, idx)
				for _, msg := range apivalidation.NameIsDNSLabel(prefixName, false) {
					allErrs = append(allErrs, field.Invalid(fldPath, prefixName, fmt.Sprintf("resulting prefix name %q is invalid: %s", prefixName, msg)))
				}
			}
		}
	}
	if numSources == 0 {
		allErrs = append(allErrs, field.Invalid(fldPath, src, "must specify a prefix source"))
	}

	return allErrs
}
