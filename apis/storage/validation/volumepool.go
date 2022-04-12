// Copyright 2022 OnMetal authors
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
	"github.com/onmetal/onmetal-api/apis/storage"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func ValidateVolumePool(volumePool *storage.VolumePool) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessor(volumePool, false, apivalidation.NameIsDNSLabel, field.NewPath("metadata"))...)
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
		allErrs = append(allErrs, ValidateImmutable(newSpec.ProviderID, oldSpec.ProviderID, fldPath.Child("providerID"))...)
	}

	return allErrs
}
