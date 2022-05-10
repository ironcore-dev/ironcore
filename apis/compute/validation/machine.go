/*
 * Copyright (c) 2022 by the OnMetal authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package validation

import (
	onmetalapivalidation "github.com/onmetal/onmetal-api/api/validation"
	"github.com/onmetal/onmetal-api/apis/compute"
	corev1 "k8s.io/api/core/v1"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1validation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateMachine validates a Machine object.
func ValidateMachine(machine *compute.Machine) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessor(machine, true, apivalidation.NameIsDNSLabel, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateMachineSpec(&machine.Spec, field.NewPath("spec"))...)

	return allErrs
}

// ValidateMachineUpdate validates a Machine object before an update.
func ValidateMachineUpdate(newMachine, oldMachine *compute.Machine) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessorUpdate(newMachine, oldMachine, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateMachineSpecUpdate(&newMachine.Spec, &oldMachine.Spec, newMachine.DeletionTimestamp != nil, field.NewPath("spec"))...)
	allErrs = append(allErrs, ValidateMachine(newMachine)...)

	return allErrs
}

// validateMachineSpec validates the spec of a Machine object.
func validateMachineSpec(machineSpec *compute.MachineSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if machineSpec.MachineClassRef == (corev1.LocalObjectReference{}) {
		allErrs = append(allErrs, field.Required(fldPath.Child("machineClassRef"), "must specify a machine class ref"))
	}

	for _, msg := range apivalidation.NameIsDNSLabel(machineSpec.MachineClassRef.Name, false) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("machineClassRef").Child("name"), machineSpec.MachineClassRef.Name, msg))
	}

	if machineSpec.MachinePoolRef != nil {
		for _, msg := range apivalidation.NameIsDNSLabel(machineSpec.MachinePoolRef.Name, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("machinePoolRef").Child("name"), machineSpec.MachinePoolRef.Name, msg))
		}
	}

	if machineSpec.IgnitionRef != nil && machineSpec.IgnitionRef.Name != "" {
		for _, msg := range apivalidation.NameIsDNSLabel(machineSpec.IgnitionRef.Name, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("ignitionRef").Child("name"), machineSpec.IgnitionRef.Name, msg))
		}
	}

	for i, vol := range machineSpec.Volumes {
		for _, msg := range apivalidation.NameIsDNSLabel(vol.Name, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("volume").Index(i).Child("name"), vol.Name, msg))
		}
		if vol.VolumeClaimRef != nil && vol.VolumeClaimRef.Name != "" {
			for _, msg := range apivalidation.NameIsDNSLabel(vol.VolumeClaimRef.Name, false) {
				allErrs = append(allErrs, field.Invalid(fldPath.Child("volume").Index(i).Child("volumeClaimRef").Child("name"), vol.VolumeClaimRef.Name, msg))
			}
		}
	}

	if machineSpec.Image == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("image"), "must specify an image"))
	}

	allErrs = append(allErrs, metav1validation.ValidateLabels(machineSpec.MachinePoolSelector, fldPath.Child("machinePoolSelector"))...)

	return allErrs
}

// validateMachineSpecUpdate validates the spec of a Machine object before an update.
func validateMachineSpecUpdate(new, old *compute.MachineSpec, deletionTimestampSet bool, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, onmetalapivalidation.ValidateImmutableField(new.Image, old.Image, fldPath.Child("image"))...)
	allErrs = append(allErrs, onmetalapivalidation.ValidateImmutableField(new.MachineClassRef, old.MachineClassRef, fldPath.Child("machineClassRef"))...)
	allErrs = append(allErrs, onmetalapivalidation.ValidateSetOnceField(new.MachinePoolRef, old.MachinePoolRef, fldPath.Child("machinePoolRef"))...)

	return allErrs
}
