// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"fmt"

	"github.com/ironcore-dev/ironcore/internal/admission/plugin/machinevolumedevices/device"
	ironcorevalidation "github.com/ironcore-dev/ironcore/internal/api/validation"
	"github.com/ironcore-dev/ironcore/internal/apis/compute"
	"github.com/ironcore-dev/ironcore/internal/apis/networking"
	networkvalidation "github.com/ironcore-dev/ironcore/internal/apis/networking/validation"
	"github.com/ironcore-dev/ironcore/internal/apis/storage"
	storagevalidation "github.com/ironcore-dev/ironcore/internal/apis/storage/validation"
	corev1 "k8s.io/api/core/v1"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1validation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"k8s.io/apimachinery/pkg/util/sets"
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

	seenVolumeNames := sets.NewString()
	newVolumeNameIndexMap := map[string]int{}
	for _, vol := range oldMachine.Spec.Volumes {
		seenVolumeNames.Insert(vol.Name)
	}
	for index, vol := range newMachine.Spec.Volumes {
		if !seenVolumeNames.Has(vol.Name) {
			newVolumeNameIndexMap[vol.Name] = index
		}
	}
	for _, vol := range newMachine.Status.Volumes {
		if i, ok := newVolumeNameIndexMap[vol.Name]; ok {
			allErrs = append(allErrs, field.Duplicate(field.NewPath("spec.volume").Index(i).Child("name"), vol.Name))
		}
	}

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessorUpdate(newMachine, oldMachine, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateMachineSpecUpdate(&newMachine.Spec, &oldMachine.Spec, field.NewPath("spec"))...)
	allErrs = append(allErrs, ValidateMachine(newMachine)...)

	return allErrs
}

var supportedMachinePowers = sets.New(
	compute.PowerOn,
	compute.PowerOff,
)

func validateMachinePower(power compute.Power, fldPath *field.Path) field.ErrorList {
	return ironcorevalidation.ValidateEnum(supportedMachinePowers, power, fldPath, "must specify machine power")
}

// validateMachineSpec validates the spec of a Machine object.
func validateMachineSpec(machineSpec *compute.MachineSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if machineSpec.MachineClassRef == (corev1.LocalObjectReference{}) {
		allErrs = append(allErrs, field.Required(fldPath.Child("machineClassRef"), "must specify a machine class ref"))
	}

	for _, msg := range apivalidation.NameIsDNSSubdomain(machineSpec.MachineClassRef.Name, false) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("machineClassRef").Child("name"), machineSpec.MachineClassRef.Name, msg))
	}

	if machineSpec.MachinePoolRef != nil {
		for _, msg := range ValidateMachinePoolName(machineSpec.MachinePoolRef.Name, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("machinePoolRef").Child("name"), machineSpec.MachinePoolRef.Name, msg))
		}
	}

	allErrs = append(allErrs, validateMachinePower(machineSpec.Power, fldPath.Child("power"))...)

	if machineSpec.IgnitionRef != nil && machineSpec.IgnitionRef.Name != "" {
		for _, msg := range apivalidation.NameIsDNSSubdomain(machineSpec.IgnitionRef.Name, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("ignitionRef").Child("name"), machineSpec.IgnitionRef.Name, msg))
		}
	}

	if machineSpec.ImagePullSecretRef != nil {
		for _, msg := range apivalidation.NameIsDNSSubdomain(machineSpec.ImagePullSecretRef.Name, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("imagePullSecretRef").Child("name"), machineSpec.ImagePullSecretRef.Name, msg))
		}
	}

	seenNames := sets.NewString()
	seenDevices := sets.NewString()
	for i, vol := range machineSpec.Volumes {
		if seenNames.Has(vol.Name) {
			allErrs = append(allErrs, field.Duplicate(fldPath.Child("volume").Index(i).Child("name"), vol.Name))
		} else {
			seenNames.Insert(vol.Name)
		}
		if seenDevices.Has(vol.Device) {
			allErrs = append(allErrs, field.Duplicate(fldPath.Child("volume").Index(i).Child("device"), vol.Device))
		} else {
			seenDevices.Insert(vol.Device)
		}
		allErrs = append(allErrs, validateVolume(&vol, fldPath.Child("volume").Index(i))...)
	}

	allErrs = append(allErrs, metav1validation.ValidateLabels(machineSpec.MachinePoolSelector, fldPath.Child("machinePoolSelector"))...)

	seenNwiNames := sets.NewString()
	for i, nwi := range machineSpec.NetworkInterfaces {
		if seenNwiNames.Has(nwi.Name) {
			allErrs = append(allErrs, field.Duplicate(fldPath.Child("networkInterface").Index(i).Child("name"), nwi.Name))
		} else {
			seenNwiNames.Insert(nwi.Name)
		}
		allErrs = append(allErrs, validateNetworkInterface(&nwi, fldPath.Child("networkInterface").Index(i))...)
	}

	return allErrs
}

func validateNetworkInterface(networkInterface *compute.NetworkInterface, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	for _, msg := range apivalidation.NameIsDNSSubdomain(networkInterface.Name, false) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("name"), networkInterface.Name, msg))
	}

	allErrs = append(allErrs, validateNetworkInterfaceSource(&networkInterface.NetworkInterfaceSource, fldPath)...)

	return allErrs
}

func validateNetworkInterfaceSource(source *compute.NetworkInterfaceSource, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList
	var numDefs int
	if source.NetworkInterfaceRef != nil {
		numDefs++
		for _, msg := range apivalidation.NameIsDNSSubdomain(source.NetworkInterfaceRef.Name, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("networkInterfaceRef").Child("name"), source.NetworkInterfaceRef.Name, msg))
		}
	}
	if source.Ephemeral != nil {
		if numDefs > 0 {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("ephemeral"), "must only specify one networkInterface source"))
		} else {
			numDefs++
			allErrs = append(allErrs, validateEphemeralNetworkInterface(source.Ephemeral, fldPath.Child("ephemeral"))...)
		}
	}
	if numDefs == 0 {
		allErrs = append(allErrs, field.Invalid(fldPath, source, "must specify at least one networkInterface source"))
	}
	return allErrs
}

func validateEphemeralNetworkInterface(source *compute.EphemeralNetworkInterfaceSource, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if source.NetworkInterfaceTemplate == nil {
		allErrs = append(allErrs, field.Required(fldPath.Child("NetworkInterfaceTemplate"), "must specify networkInterface template "))
	} else {
		allErrs = append(allErrs, validateNetworkInterfaceTemplateSpecForMachine(source.NetworkInterfaceTemplate, fldPath.Child("networkInterfaceTemplate"))...)
	}

	return allErrs
}

func validateNetworkInterfaceTemplateSpecForMachine(template *networking.NetworkInterfaceTemplateSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if template == nil {
		allErrs = append(allErrs, field.Required(fldPath, ""))
	} else {
		allErrs = append(allErrs, networkvalidation.ValidateNetworkInterfaceSpec(&template.Spec, &template.ObjectMeta, fldPath)...)
	}

	return allErrs
}

func validateVolume(volume *compute.Volume, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	for _, msg := range apivalidation.NameIsDNSSubdomain(volume.Name, false) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("name"), volume.Name, msg))
	}

	if _, _, err := device.ParseName(volume.Device); err != nil {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("device"), volume.Device, fmt.Sprintf("invalid device name: %v", err)))
	}

	allErrs = append(allErrs, validateVolumeSource(&volume.VolumeSource, fldPath)...)

	return allErrs
}

func validateVolumeSource(source *compute.VolumeSource, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	var numDefs int
	if source.VolumeRef != nil {
		numDefs++
		for _, msg := range apivalidation.NameIsDNSSubdomain(source.VolumeRef.Name, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("volumeRef").Child("name"), source.VolumeRef.Name, msg))
		}
	}
	if source.EmptyDisk != nil {
		if numDefs > 0 {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("emptyDisk"), "must only specify one volume source"))
		} else {
			numDefs++
			allErrs = append(allErrs, validateEmptyDiskVolumeSource(source.EmptyDisk, fldPath.Child("emptyDisk"))...)
		}
	}
	if source.Ephemeral != nil {
		if numDefs > 0 {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("ephemeral"), "must only specify one volume source"))
		} else {
			numDefs++
			allErrs = append(allErrs, validateEphemeralVolumeSource(source.Ephemeral, fldPath.Child("ephemeral"))...)
		}
	}
	if numDefs == 0 {
		allErrs = append(allErrs, field.Invalid(fldPath, source, "must specify at least one volume source"))
	}

	return allErrs
}

func validateEmptyDiskVolumeSource(source *compute.EmptyDiskVolumeSource, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if sizeLimit := source.SizeLimit; sizeLimit != nil {
		allErrs = append(allErrs, ironcorevalidation.ValidateNonNegativeQuantity(*sizeLimit, fldPath.Child("sizeLimit"))...)
	}

	return allErrs
}

func validateEphemeralVolumeSource(source *compute.EphemeralVolumeSource, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if source.VolumeTemplate == nil {
		allErrs = append(allErrs, field.Required(fldPath.Child("volumeTemplate"), "must specify volume template"))
	} else {
		allErrs = append(allErrs, validateVolumeTemplateSpecForMachine(source.VolumeTemplate, fldPath.Child("volumeTemplate"))...)
	}

	return allErrs
}

func validateVolumeTemplateSpecForMachine(template *storage.VolumeTemplateSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if template == nil {
		allErrs = append(allErrs, field.Required(fldPath, ""))
	} else {
		allErrs = append(allErrs, storagevalidation.ValidateVolumeTemplateSpec(template, fldPath)...)

		if template.Spec.ClaimRef != nil {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("spec", "claimRef"), "may not specify claimRef"))
		}
		if template.Spec.Unclaimable {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("spec", "unclaimable"), "may not specify unclaimable"))
		}
	}

	return allErrs
}

// validateMachineSpecUpdate validates the spec of a Machine object before an update.
func validateMachineSpecUpdate(new, old *compute.MachineSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, ironcorevalidation.ValidateImmutableField(new.Image, old.Image, fldPath.Child("image"))...)
	allErrs = append(allErrs, ironcorevalidation.ValidateImmutableField(new.MachineClassRef, old.MachineClassRef, fldPath.Child("machineClassRef"))...)
	allErrs = append(allErrs, ironcorevalidation.ValidateSetOnceField(new.MachinePoolRef, old.MachinePoolRef, fldPath.Child("machinePoolRef"))...)

	return allErrs
}
