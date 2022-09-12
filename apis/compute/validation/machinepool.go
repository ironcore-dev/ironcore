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
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var ValidateMachinePoolName = apivalidation.NameIsDNSSubdomain

func ValidateMachinePool(machinePool *compute.MachinePool) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessor(machinePool, false, ValidateMachinePoolName, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateMachinePoolSpec(&machinePool.Spec, field.NewPath("spec"))...)

	return allErrs
}

func validateMachinePoolSpec(machinePoolSpec *compute.MachinePoolSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	return allErrs
}

func ValidateMachinePoolUpdate(newMachinePool, oldMachinePool *compute.MachinePool) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessorUpdate(newMachinePool, oldMachinePool, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateMachinePoolSpecUpdate(&newMachinePool.Spec, &oldMachinePool.Spec, field.NewPath("spec"))...)
	allErrs = append(allErrs, ValidateMachinePool(newMachinePool)...)

	return allErrs
}

func validateMachinePoolSpecUpdate(newSpec, oldSpec *compute.MachinePoolSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if oldSpec.ProviderID != "" {
		allErrs = append(allErrs, onmetalapivalidation.ValidateImmutableField(newSpec.ProviderID, oldSpec.ProviderID, fldPath.Child("providerID"))...)
	}

	return allErrs
}
