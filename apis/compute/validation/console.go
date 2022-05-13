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
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateConsole validates a Console object.
func ValidateConsole(console *compute.Console) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessor(console, true, apivalidation.NameIsDNSLabel, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateConsoleSpec(&console.Spec, field.NewPath("spec"))...)

	return allErrs
}

// ValidateConsoleUpdate validates a Console object before an update.
func ValidateConsoleUpdate(newConsole, oldConsole *compute.Console) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessorUpdate(newConsole, oldConsole, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateConsoleSpecUpdate(&newConsole.Spec, &oldConsole.Spec, newConsole.DeletionTimestamp != nil, field.NewPath("spec"))...)
	allErrs = append(allErrs, ValidateConsole(newConsole)...)

	return allErrs
}

// validateConsoleSpec validates the spec of a Console object.
func validateConsoleSpec(consoleSpec *compute.ConsoleSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if consoleSpec.MachineRef == (corev1.LocalObjectReference{}) {
		allErrs = append(allErrs, field.Required(fldPath.Child("machineRef"), "must specify a machine ref"))
	}

	for _, msg := range apivalidation.NameIsDNSLabel(consoleSpec.MachineRef.Name, false) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("machineRef").Child("name"), consoleSpec.MachineRef.Name, msg))
	}

	return allErrs
}

// validateConsoleSpecUpdate validates the spec of a Console object before an update.
func validateConsoleSpecUpdate(new, old *compute.ConsoleSpec, deletionTimestampSet bool, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, onmetalapivalidation.ValidateImmutableField(new.MachineRef, old.MachineRef, fldPath.Child("machineRef"))...)

	return allErrs
}
