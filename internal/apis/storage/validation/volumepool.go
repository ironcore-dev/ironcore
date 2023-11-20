// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	ironcorevalidation "github.com/ironcore-dev/ironcore/internal/api/validation"
	"github.com/ironcore-dev/ironcore/internal/apis/storage"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var ValidateVolumePoolName = apivalidation.NameIsDNSSubdomain

func ValidateVolumePool(volumePool *storage.VolumePool) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessor(volumePool, false, ValidateVolumePoolName, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateVolumePoolSpec(&volumePool.Spec, field.NewPath("spec"))...)

	return allErrs
}

func validateVolumePoolSpec(volumePoolSpec *storage.VolumePoolSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	return allErrs
}

func ValidateVolumePoolUpdate(newVolumePool, oldVolumePool *storage.VolumePool) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessorUpdate(newVolumePool, oldVolumePool, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateVolumePoolSpecUpdate(&newVolumePool.Spec, &oldVolumePool.Spec, field.NewPath("spec"))...)
	allErrs = append(allErrs, ValidateVolumePool(newVolumePool)...)

	return allErrs
}

func validateVolumePoolSpecUpdate(newSpec, oldSpec *storage.VolumePoolSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if oldSpec.ProviderID != "" {
		allErrs = append(allErrs, ironcorevalidation.ValidateImmutableField(newSpec.ProviderID, oldSpec.ProviderID, fldPath.Child("providerID"))...)
	}

	return allErrs
}
