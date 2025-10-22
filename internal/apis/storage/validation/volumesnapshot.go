// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	ironcorevalidation "github.com/ironcore-dev/ironcore/internal/api/validation"
	"github.com/ironcore-dev/ironcore/internal/apis/storage"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func ValidateVolumeSnapshot(volumeSnapshot *storage.VolumeSnapshot) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessor(volumeSnapshot, true, apivalidation.NameIsDNSSubdomain, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateVolumeSnapshotSpec(&volumeSnapshot.Spec, field.NewPath("spec"))...)

	return allErrs
}

func validateVolumeSnapshotSpec(spec *storage.VolumeSnapshotSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if spec.VolumeRef != nil {
		for _, msg := range apivalidation.NameIsDNSSubdomain(spec.VolumeRef.Name, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("volumeRef").Child("name"), spec.VolumeRef.Name, msg))
		}
	}

	return allErrs
}

func ValidateVolumeSnapshotUpdate(newVolumeSnapshot, oldVolumeSnapshot *storage.VolumeSnapshot) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessorUpdate(newVolumeSnapshot, oldVolumeSnapshot, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateVolumeSnapshotSpecUpdate(&newVolumeSnapshot.Spec, &oldVolumeSnapshot.Spec, field.NewPath("spec"))...)
	allErrs = append(allErrs, ValidateVolumeSnapshot(newVolumeSnapshot)...)

	return allErrs
}

func validateVolumeSnapshotSpecUpdate(newSpec, oldSpec *storage.VolumeSnapshotSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, ironcorevalidation.ValidateImmutableField(newSpec.VolumeRef, oldSpec.VolumeRef, fldPath.Child("volumeRef"))...)

	return allErrs
}
