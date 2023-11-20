// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"fmt"

	"github.com/ironcore-dev/ironcore/internal/apis/core"
	corev1 "k8s.io/api/core/v1"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1validation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"

	ironcorevalidation "github.com/ironcore-dev/ironcore/internal/api/validation"
	"github.com/ironcore-dev/ironcore/internal/apis/storage"
)

func ValidateVolume(volume *storage.Volume) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessor(volume, true, apivalidation.NameIsDNSLabel, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateVolumeSpec(&volume.Spec, field.NewPath("spec"))...)

	return allErrs
}

func validateVolumeSpec(spec *storage.VolumeSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if volumeClassRef := spec.VolumeClassRef; volumeClassRef != nil {
		for _, msg := range apivalidation.NameIsDNSLabel(spec.VolumeClassRef.Name, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("volumeClassRef").Child("name"), spec.VolumeClassRef.Name, msg))
		}

		allErrs = append(allErrs, metav1validation.ValidateLabels(spec.VolumePoolSelector, fldPath.Child("volumePoolSelector"))...)

		if spec.VolumePoolRef != nil {
			for _, msg := range ValidateVolumePoolName(spec.VolumePoolRef.Name, false) {
				allErrs = append(allErrs, field.Invalid(fldPath.Child("volumePoolRef").Child("name"), spec.VolumePoolRef.Name, msg))
			}
		}

		storageValue, ok := spec.Resources[core.ResourceStorage]
		if ok {
			allErrs = append(allErrs, ironcorevalidation.ValidatePositiveQuantity(storageValue, fldPath.Child("resources").Key(string(corev1.ResourceStorage)))...)
		} else {
			allErrs = append(allErrs, field.Required(fldPath.Child("resources").Key(string(core.ResourceStorage)), fmt.Sprintf("must specify %s", core.ResourceStorage)))
		}

		if spec.ImagePullSecretRef != nil {
			for _, msg := range apivalidation.NameIsDNSLabel(spec.ImagePullSecretRef.Name, false) {
				allErrs = append(allErrs, field.Invalid(fldPath.Child("imagePullSecretRef").Child("name"), spec.ImagePullSecretRef.Name, msg))
			}
		}
	} else {
		if spec.VolumePoolSelector != nil {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("volumePoolSelector"), "must not specify if volume class is empty"))
		}

		if spec.VolumePoolRef != nil {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("volumePoolRef"), "must not specify if volume class is empty"))
		}

		if spec.Resources != nil {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("resources"), "must not specify if volume class is empty"))
		}

		if spec.Image != "" {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("image"), "must not specify if volume class is empty"))
		}

		if spec.ImagePullSecretRef != nil {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("imagePullSecretRef"), "must not specify if volume class is empty"))
		}

		if spec.Tolerations != nil {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("tolerations"), "must not specify if volume class is empty"))
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

	if spec.Encryption != nil {
		for _, msg := range apivalidation.NameIsDNSLabel(spec.Encryption.SecretRef.Name, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("encryption").Child("secretRef").Child("name"), spec.Encryption.SecretRef.Name, msg))
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

	allErrs = append(allErrs, ironcorevalidation.ValidateImmutableField(newSpec.VolumeClassRef, oldSpec.VolumeClassRef, fldPath.Child("volumeClassRef"))...)
	allErrs = append(allErrs, ironcorevalidation.ValidateSetOnceField(newSpec.VolumePoolRef, oldSpec.VolumePoolRef, fldPath.Child("volumePoolRef"))...)
	allErrs = append(allErrs, ironcorevalidation.ValidateImmutableField(newSpec.Encryption, oldSpec.Encryption, fldPath.Child("encryption"))...)

	return allErrs
}
