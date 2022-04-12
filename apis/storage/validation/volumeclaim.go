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
	corev1 "k8s.io/api/core/v1"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1validation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func ValidateVolumeClaim(volumeClaim *storage.VolumeClaim) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessor(volumeClaim, true, apivalidation.NameIsDNSLabel, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateVolumeClaimSpec(&volumeClaim.Spec, field.NewPath("spec"))...)

	return allErrs
}

func validateVolumeClaimSpec(volumeClaimSpec *storage.VolumeClaimSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if volumeClaimSpec.VolumeClassRef == (corev1.LocalObjectReference{}) {
		allErrs = append(allErrs, field.Required(fldPath.Child("volumeClassRef"), "must specify a volume class ref"))
	}
	for _, msg := range apivalidation.NameIsDNSLabel(volumeClaimSpec.VolumeClassRef.Name, false) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("volumeClassRef").Child("name"), volumeClaimSpec.VolumeClassRef.Name, msg))
	}

	allErrs = append(allErrs, metav1validation.ValidateLabelSelector(volumeClaimSpec.Selector, fldPath.Child("selector"))...)

	if volumeClaimSpec.VolumeRef.Name != "" {
		for _, msg := range apivalidation.NameIsDNSLabel(volumeClaimSpec.VolumeRef.Name, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("volumeRef").Child("name"), volumeClaimSpec.VolumeRef.Name, msg))
		}
	}

	storageValue, ok := volumeClaimSpec.Resources[corev1.ResourceStorage]
	if !ok {
		allErrs = append(allErrs, field.Required(fldPath.Child("resources").Key(string(corev1.ResourceStorage)), ""))
	} else {
		allErrs = append(allErrs, ValidatePositiveQuantity(storageValue, fldPath.Child("resources").Key(string(corev1.ResourceStorage)))...)
	}

	return allErrs
}

func ValidateVolumeClaimUpdate(newVolumeClaim, oldVolumeClaim *storage.VolumeClaim) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessorUpdate(newVolumeClaim, oldVolumeClaim, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateVolumeClaimSpecUpdate(&newVolumeClaim.Spec, &oldVolumeClaim.Spec, field.NewPath("spec"))...)
	allErrs = append(allErrs, ValidateVolumeClaim(newVolumeClaim)...)

	return allErrs
}

func validateVolumeClaimSpecUpdate(newSpec, oldSpec *storage.VolumeClaimSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	newSpecCopy := newSpec.DeepCopy()
	oldSpecCopy := oldSpec.DeepCopy()

	if oldSpec.VolumeRef.Name == "" {
		oldSpecCopy.VolumeRef.Name = newSpecCopy.VolumeRef.Name
	}

	allErrs = append(allErrs, ValidateImmutableWithDiff(newSpecCopy, oldSpecCopy, fldPath)...)

	return allErrs
}
