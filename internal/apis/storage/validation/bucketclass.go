// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	ironcorevalidation "github.com/ironcore-dev/ironcore/internal/api/validation"
	"github.com/ironcore-dev/ironcore/internal/apis/core"
	"github.com/ironcore-dev/ironcore/internal/apis/storage"
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
	allErrs = append(allErrs, ironcorevalidation.ValidatePositiveQuantity(*tps, fldPath.Key(string(core.ResourceTPS)))...)

	iops := capabilities.Name(core.ResourceIOPS, resource.DecimalSI)
	allErrs = append(allErrs, ironcorevalidation.ValidatePositiveQuantity(*iops, fldPath.Key(string(core.ResourceIOPS)))...)

	return allErrs
}

func ValidateBucketClassUpdate(newBucketClass, oldBucketClass *storage.BucketClass) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessorUpdate(newBucketClass, oldBucketClass, field.NewPath("metadata"))...)
	allErrs = append(allErrs, ironcorevalidation.ValidateImmutableField(newBucketClass.Capabilities, oldBucketClass.Capabilities, field.NewPath("capabilities"))...)
	allErrs = append(allErrs, ValidateBucketClass(newBucketClass)...)

	return allErrs
}
