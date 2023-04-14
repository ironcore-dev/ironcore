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

	corevalidation "github.com/onmetal/onmetal-api/internal/apis/core/validation"
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
		allErrs = append(allErrs, corevalidation.ValidateIP(ipFamily, *ip, fldPath.Child("value"))...)
	}
	if ephemeral := ipSource.Ephemeral; ephemeral != nil {
		if numSources > 0 {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("ephemeral"), ephemeral, "cannot specify multiple ip sources"))
		} else {
			numSources++
			allErrs = append(allErrs, validateEphemeralPrefixSource(ipFamily, ephemeral, fldPath.Child("ephemeral"))...)
			if objectMeta != nil && objectMeta.Name != "" {
				prefixName := fmt.Sprintf("%s-%d", objectMeta.Name, idx)
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
