// Copyright 2023 IronCore authors
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
	ironcorevalidation "github.com/ironcore-dev/ironcore/internal/api/validation"
	"github.com/ironcore-dev/ironcore/internal/apis/core"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func ValidateResourceQuota(resourceQuota *core.ResourceQuota) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessor(resourceQuota, true, apivalidation.NameIsDNSLabel, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateResourceQuotaSpec(&resourceQuota.Spec, field.NewPath("spec"))...)

	return allErrs
}

func validateResourceQuotaSpec(spec *core.ResourceQuotaSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	for name, quantity := range spec.Hard {
		allErrs = append(allErrs, ironcorevalidation.ValidateNonNegativeQuantity(quantity, fldPath.Child("hard").Key(string(name)))...)
	}

	return allErrs
}

func ValidateResourceQuotaUpdate(newResourceQuota, oldResourceQuota *core.ResourceQuota) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessorUpdate(newResourceQuota, oldResourceQuota, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateResourceQuotaSpecUpdate(&newResourceQuota.Spec, &oldResourceQuota.Spec, field.NewPath("spec"))...)

	return allErrs
}

func validateResourceQuotaSpecUpdate(newSpec, oldSpec *core.ResourceQuotaSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	return allErrs
}

func ValidateResourceQuotaStatus(status *core.ResourceQuotaStatus, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	return allErrs
}

func ValidateResourceQuotaStatusUpdate(newResourceQuota, oldResourceQuota *core.ResourceQuota) field.ErrorList {
	var allErrs field.ErrorList

	statusField := field.NewPath("status")

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessorUpdate(newResourceQuota, oldResourceQuota, field.NewPath("metadata"))...)
	allErrs = append(allErrs, ValidateResourceQuotaStatus(&newResourceQuota.Status, statusField)...)

	return allErrs
}
