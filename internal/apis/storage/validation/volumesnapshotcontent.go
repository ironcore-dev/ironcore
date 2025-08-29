// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	ironcorevalidation "github.com/ironcore-dev/ironcore/internal/api/validation"
	"github.com/ironcore-dev/ironcore/internal/apis/storage"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func ValidateVolumeSnapshotContent(volumeSnapshotContent *storage.VolumeSnapshotContent) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessor(volumeSnapshotContent, true, apivalidation.NameIsDNSSubdomain, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateVolumeSnapshotContentSpec(&volumeSnapshotContent.Spec, field.NewPath("spec"))...)

	return allErrs
}

func validateVolumeSnapshotContentSpec(spec *storage.VolumeSnapshotContentSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if spec.Source != nil {
		if len(spec.Source.VolumeSnapshotHandle) == 0 {
			allErrs = append(allErrs, field.Required(fldPath.Child("source", "snapshotHandle"), "must specify volume snapshot handle"))
		}
	}

	if spec.VolumeSnapshotRef != nil {
		for _, msg := range apivalidation.NameIsDNSLabel(spec.VolumeSnapshotRef.Name, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("volumeSnapshotRef").Child("name"), spec.VolumeSnapshotRef.Name, msg))
		}
	}

	return allErrs
}

func ValidateVolumeSnapshotContentUpdate(newVolumeSnapshotContent, oldVolumeSnapshotContent *storage.VolumeSnapshotContent) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessorUpdate(newVolumeSnapshotContent, oldVolumeSnapshotContent, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateVolumeSnapshotContentSpecUpdate(&newVolumeSnapshotContent.Spec, &oldVolumeSnapshotContent.Spec, field.NewPath("spec"))...)
	allErrs = append(allErrs, ValidateVolumeSnapshotContent(newVolumeSnapshotContent)...)

	return allErrs
}

func validateVolumeSnapshotContentSpecUpdate(newSpec, oldSpec *storage.VolumeSnapshotContentSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, ironcorevalidation.ValidateImmutableField(newSpec.Source, oldSpec.Source, fldPath.Child("source"))...)
	allErrs = append(allErrs, ironcorevalidation.ValidateImmutableField(newSpec.VolumeSnapshotRef, oldSpec.VolumeSnapshotRef, fldPath.Child("volumeSnapshotRef"))...)

	return allErrs
}
