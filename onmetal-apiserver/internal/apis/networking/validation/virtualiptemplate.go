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
	onmetalapivalidation "github.com/onmetal/onmetal-api/onmetal-apiserver/internal/api/validation"
	"github.com/onmetal/onmetal-api/onmetal-apiserver/internal/apis/networking"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1validation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var allowedPrefixTemplateObjectMetaFields = sets.NewString(
	"Annotations",
	"Labels",
)

func validateVirtualIPTemplateSpecMetadata(objMeta *metav1.ObjectMeta, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, metav1validation.ValidateLabels(objMeta.Labels, fldPath.Child("labels"))...)
	allErrs = append(allErrs, apivalidation.ValidateAnnotations(objMeta.Annotations, fldPath.Child("annotations"))...)
	allErrs = append(allErrs, onmetalapivalidation.ValidateFieldAllowList(*objMeta, allowedPrefixTemplateObjectMetaFields, "cannot be set for an ephemeral prefix", fldPath)...)

	return allErrs
}

// ValidateVirtualIPTemplateSpec validates the spec of a virtual ip template.
func ValidateVirtualIPTemplateSpec(spec *networking.VirtualIPTemplateSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, validateVirtualIPTemplateSpecMetadata(&spec.ObjectMeta, fldPath.Child("metadata"))...)
	allErrs = append(allErrs, validateVirtualIPSpec(&spec.Spec, fldPath.Child("spec"))...)

	return allErrs
}
