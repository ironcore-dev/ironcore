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
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func ValidateVolumeClass(volumeClass *storage.VolumeClass) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessor(volumeClass, false, apivalidation.NameIsDNSLabel, field.NewPath("metadata"))...)

	return allErrs
}

func ValidateVolumeClassUpdate(newVolumeClass, oldVolumeClass *storage.VolumeClass) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessorUpdate(newVolumeClass, oldVolumeClass, field.NewPath("metadata"))...)
	allErrs = append(allErrs, onmetalapivalidation.ValidateImmutableField(newVolumeClass.Capabilities, oldVolumeClass.Capabilities, field.NewPath("capabilities"))...)
	allErrs = append(allErrs, ValidateVolumeClass(newVolumeClass)...)

	return allErrs
}
