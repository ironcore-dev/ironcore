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
	onmetalapivalidation "github.com/onmetal/onmetal-api/onmetal-apiserver/api/validation"
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

func validateVolumeSpec(spec *storage.VolumeSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if spec.VolumeClassRef == (corev1.LocalObjectReference{}) {
		allErrs = append(allErrs, field.Required(fldPath.Child("volumeClassRef"), "must specify a volume class ref"))
	}
	for _, msg := range apivalidation.NameIsDNSLabel(spec.VolumeClassRef.Name, false) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("volumeClassRef").Child("name"), spec.VolumeClassRef.Name, msg))
	}

	allErrs = append(allErrs, metav1validation.ValidateLabels(spec.VolumePoolSelector, fldPath.Child("volumePoolSelector"))...)

	if spec.VolumePoolRef != nil {
		for _, msg := range ValidateVolumePoolName(spec.VolumePoolRef.Name, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("volumePoolRef").Child("name"), spec.VolumePoolRef.Name, msg))
		}
	}

	if spec.Unclaimable {
		if spec.ClaimRef != nil {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("claimRef"), "cannot specify unclaimable and claimRef"))
		}
	} else {
		if spec.ClaimRef != nil {
			for _, msg := range apivalidation.NameIsDNSLabel(spec.ClaimRef.Name, false) {
				allErrs = append(allErrs, field.Invalid(fldPath.Child("claimRef").Child("name"), spec.ClaimRef.Name, msg))
			}
		}
	}

	if storageValue, ok := spec.Resources[corev1.ResourceStorage]; ok {
		allErrs = append(allErrs, onmetalapivalidation.ValidatePositiveQuantity(storageValue, fldPath.Child("resources").Key(string(corev1.ResourceStorage)))...)
	}

	if spec.ImagePullSecretRef != nil {
		for _, msg := range apivalidation.NameIsDNSLabel(spec.ImagePullSecretRef.Name, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("imagePullSecretRef").Child("name"), spec.ImagePullSecretRef.Name, msg))
		}
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
