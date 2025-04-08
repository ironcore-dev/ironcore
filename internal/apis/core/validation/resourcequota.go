// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	ironcorevalidation "github.com/ironcore-dev/ironcore/internal/api/validation"
	"github.com/ironcore-dev/ironcore/internal/apis/core"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/sets"
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
	hardResourceNames := sets.KeySet(spec.Hard)

	for name := range hardResourceNames {
		allErrs = append(allErrs, ironcorevalidation.ValidateResourceName(name, fldPath.Child("hard").Key(string(name)))...)
	}

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
