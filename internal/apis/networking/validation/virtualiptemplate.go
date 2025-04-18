// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	ironcorevalidation "github.com/ironcore-dev/ironcore/internal/api/validation"
	"github.com/ironcore-dev/ironcore/internal/apis/networking"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1validation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var allowedPrefixTemplateObjectMetaFields = sets.New(
	"Annotations",
	"Labels",
)

func validateVirtualIPTemplateSpecMetadata(objMeta *metav1.ObjectMeta, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, metav1validation.ValidateLabels(objMeta.Labels, fldPath.Child("labels"))...)
	allErrs = append(allErrs, apivalidation.ValidateAnnotations(objMeta.Annotations, fldPath.Child("annotations"))...)
	allErrs = append(allErrs, ironcorevalidation.ValidateFieldAllowList(*objMeta, allowedPrefixTemplateObjectMetaFields, "cannot be set for an ephemeral prefix", fldPath)...)

	return allErrs
}

// ValidateVirtualIPTemplateSpec validates the spec of a virtual ip template.
func ValidateVirtualIPTemplateSpec(spec *networking.VirtualIPTemplateSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, validateVirtualIPTemplateSpecMetadata(&spec.ObjectMeta, fldPath.Child("metadata"))...)
	allErrs = append(allErrs, validateVirtualIPSpec(&spec.Spec.VirtualIPSpec, fldPath.Child("spec"))...)

	return allErrs
}
