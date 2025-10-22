// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	ironcorevalidation "github.com/ironcore-dev/ironcore/internal/api/validation"
	"github.com/ironcore-dev/ironcore/internal/apis/compute"
	"github.com/ironcore-dev/ironcore/internal/apis/core"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateMachineClass validates a MachineClass object.
func ValidateMachineClass(machineClass *compute.MachineClass) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessor(machineClass, false, apivalidation.NameIsDNSSubdomain, field.NewPath("metadata"))...)

	allErrs = append(allErrs, validateMachineClassCapabilities(machineClass.Capabilities, field.NewPath("capabilities"))...)

	return allErrs
}

func validateMachineClassCapabilities(capabilities core.ResourceList, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	cpu := capabilities.CPU()
	allErrs = append(allErrs, ironcorevalidation.ValidatePositiveQuantity(*cpu, fldPath.Key(string(core.ResourceCPU)))...)

	memory := capabilities.Memory()
	allErrs = append(allErrs, ironcorevalidation.ValidatePositiveQuantity(*memory, fldPath.Key(string(core.ResourceMemory)))...)

	return allErrs
}

// ValidateMachineClassUpdate validates a MachineClass object before an update.
func ValidateMachineClassUpdate(newMachineClass, oldMachineClass *compute.MachineClass) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessorUpdate(newMachineClass, oldMachineClass, field.NewPath("metadata"))...)
	allErrs = append(allErrs, ironcorevalidation.ValidateImmutableField(newMachineClass.Capabilities, oldMachineClass.Capabilities, field.NewPath("capabilities"))...)
	allErrs = append(allErrs, ValidateMachineClass(newMachineClass)...)

	return allErrs
}
