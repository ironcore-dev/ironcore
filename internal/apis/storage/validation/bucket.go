// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	ironcorevalidation "github.com/ironcore-dev/ironcore/internal/api/validation"
	"github.com/ironcore-dev/ironcore/internal/apis/storage"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1validation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func ValidateBucket(bucket *storage.Bucket) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessor(bucket, true, apivalidation.NameIsDNSLabel, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateBucketSpec(&bucket.Spec, field.NewPath("spec"))...)

	return allErrs
}

func validateBucketSpec(spec *storage.BucketSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if bucketClassRef := spec.BucketClassRef; bucketClassRef != nil {
		for _, msg := range apivalidation.NameIsDNSLabel(spec.BucketClassRef.Name, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("bucketClassRef").Child("name"), spec.BucketClassRef.Name, msg))
		}

		allErrs = append(allErrs, metav1validation.ValidateLabels(spec.BucketPoolSelector, fldPath.Child("bucketPoolSelector"))...)

		if spec.BucketPoolRef != nil {
			for _, msg := range ValidateBucketPoolName(spec.BucketPoolRef.Name, false) {
				allErrs = append(allErrs, field.Invalid(fldPath.Child("bucketPoolRef").Child("name"), spec.BucketPoolRef.Name, msg))
			}
		}

	} else {
		if spec.BucketPoolSelector != nil {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("bucketPoolSelector"), "must not specify if bucket class is empty"))
		}

		if spec.BucketPoolRef != nil {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("bucketPoolRef"), "must not specify if bucket class is empty"))
		}

		if spec.Tolerations != nil {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("tolerations"), "must not specify if bucket class is empty"))
		}
	}

	return allErrs
}

func ValidateBucketUpdate(newBucket, oldBucket *storage.Bucket) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessorUpdate(newBucket, oldBucket, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateBucketSpecUpdate(&newBucket.Spec, &oldBucket.Spec, field.NewPath("spec"))...)
	allErrs = append(allErrs, ValidateBucket(newBucket)...)

	return allErrs
}

func validateBucketSpecUpdate(newSpec, oldSpec *storage.BucketSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, ironcorevalidation.ValidateImmutableField(newSpec.BucketClassRef, oldSpec.BucketClassRef, fldPath.Child("bucketClassRef"))...)
	allErrs = append(allErrs, ironcorevalidation.ValidateSetOnceField(newSpec.BucketPoolRef, oldSpec.BucketPoolRef, fldPath.Child("bucketPoolRef"))...)

	return allErrs
}
