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
	"github.com/onmetal/onmetal-api/internal/apis/storage"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var ValidateBucketPoolName = apivalidation.NameIsDNSSubdomain

func ValidateBucketPool(bucketPool *storage.BucketPool) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessor(bucketPool, false, ValidateBucketPoolName, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateBucketPoolSpec(&bucketPool.Spec, field.NewPath("spec"))...)

	return allErrs
}

func validateBucketPoolSpec(bucketPoolSpec *storage.BucketPoolSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	return allErrs
}

func ValidateBucketPoolUpdate(newBucketPool, oldBucketPool *storage.BucketPool) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessorUpdate(newBucketPool, oldBucketPool, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateBucketPoolSpecUpdate(&newBucketPool.Spec, &oldBucketPool.Spec, field.NewPath("spec"))...)
	allErrs = append(allErrs, ValidateBucketPool(newBucketPool)...)

	return allErrs
}

func validateBucketPoolSpecUpdate(newSpec, oldSpec *storage.BucketPoolSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if oldSpec.ProviderID != "" {
		allErrs = append(allErrs, onmetalapivalidation.ValidateImmutableField(newSpec.ProviderID, oldSpec.ProviderID, fldPath.Child("providerID"))...)
	}

	return allErrs
}
