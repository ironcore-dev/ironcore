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
	onmetalapivalidation "github.com/onmetal/onmetal-api/api/validation"
	"github.com/onmetal/onmetal-api/apis/storage"
	corev1 "k8s.io/api/core/v1"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1validation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func ValidateVolume(volume *storage.Volume) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessor(volume, true, apivalidation.NameIsDNSLabel, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateVolumeSpec(&volume.Spec, field.NewPath("spec"))...)

	return allErrs
}

func validateVolumeSpec(volumeSpec *storage.VolumeSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if volumeSpec.VolumeClassRef == (corev1.LocalObjectReference{}) {
		allErrs = append(allErrs, field.Required(fldPath.Child("volumeClassRef"), "must specify a volume class ref"))
	}
	for _, msg := range apivalidation.NameIsDNSLabel(volumeSpec.VolumeClassRef.Name, false) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("volumeClassRef").Child("name"), volumeSpec.VolumeClassRef.Name, msg))
	}

	allErrs = append(allErrs, metav1validation.ValidateLabels(volumeSpec.VolumePoolSelector, fldPath.Child("volumePoolSelector"))...)

	if volumeSpec.VolumePoolRef.Name != "" {
		for _, msg := range apivalidation.NameIsDNSLabel(volumeSpec.VolumePoolRef.Name, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("volumePoolRef").Child("name"), volumeSpec.VolumePoolRef.Name, msg))
		}
	}

	if volumeSpec.ClaimRef.Name != "" {
		for _, msg := range apivalidation.NameIsDNSLabel(volumeSpec.ClaimRef.Name, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("claimRef").Child("name"), volumeSpec.ClaimRef.Name, msg))
		}
	}

	storageValue, ok := volumeSpec.Resources[corev1.ResourceStorage]
	if !ok {
		allErrs = append(allErrs, field.Required(fldPath.Child("resources").Key(string(corev1.ResourceStorage)), ""))
	} else {
		allErrs = append(allErrs, onmetalapivalidation.ValidatePositiveQuantity(storageValue, fldPath.Child("resources").Key(string(corev1.ResourceStorage)))...)
	}

	return allErrs
}

func ValidateVolumeUpdate(newVolume, oldVolume *storage.Volume) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessorUpdate(newVolume, oldVolume, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateVolumeSpecUpdate(&newVolume.Spec, &oldVolume.Spec, field.NewPath("spec"))...)
	allErrs = append(allErrs, ValidateVolume(newVolume)...)

	return allErrs
}

func validateVolumeSpecUpdate(newSpec, oldSpec *storage.VolumeSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, onmetalapivalidation.ValidateImmutableField(newSpec.VolumeClassRef, oldSpec.VolumeClassRef, fldPath.Child("volumeClassRef"))...)
	allErrs = append(allErrs, onmetalapivalidation.ValidateSetOnceField(newSpec.VolumePoolRef, oldSpec.VolumePoolRef, fldPath.Child("volumePoolRef"))...)

	return allErrs
}
