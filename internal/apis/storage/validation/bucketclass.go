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
	onmetalapivalidation "github.com/onmetal/onmetal-api/internal/api/validation"
	"github.com/onmetal/onmetal-api/internal/apis/core"
	"github.com/onmetal/onmetal-api/internal/apis/storage"
	"k8s.io/apimachinery/pkg/api/resource"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func ValidateBucketClass(bucketClass *storage.BucketClass) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessor(bucketClass, false, apivalidation.NameIsDNSLabel, field.NewPath("metadata"))...)

	allErrs = append(allErrs, validateBucketClassCapabilities(bucketClass.Capabilities, field.NewPath("capabilities"))...)

	return allErrs
}

func validateBucketClassCapabilities(capabilities core.ResourceList, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	tps := capabilities.Name(core.ResourceTPS, resource.DecimalSI)
	allErrs = append(allErrs, onmetalapivalidation.ValidatePositiveQuantity(*tps, fldPath.Key(string(core.ResourceTPS)))...)

	iops := capabilities.Name(core.ResourceIOPS, resource.DecimalSI)
	allErrs = append(allErrs, onmetalapivalidation.ValidatePositiveQuantity(*iops, fldPath.Key(string(core.ResourceIOPS)))...)

	return allErrs
}

func ValidateBucketClassUpdate(newBucketClass, oldBucketClass *storage.BucketClass) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessorUpdate(newBucketClass, oldBucketClass, field.NewPath("metadata"))...)
	allErrs = append(allErrs, onmetalapivalidation.ValidateImmutableField(newBucketClass.Capabilities, oldBucketClass.Capabilities, field.NewPath("capabilities"))...)
	allErrs = append(allErrs, ValidateBucketClass(newBucketClass)...)

	return allErrs
}
