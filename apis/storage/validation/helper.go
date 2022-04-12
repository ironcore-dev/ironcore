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
	"fmt"
	"reflect"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/resource"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

const (
	isNegativeErrorMsg    = apivalidation.IsNegativeErrorMsg
	isNotPositiveErrorMsg = `must be greater than zero`
)

func ValidateNonNegativeQuantity(value resource.Quantity, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList
	if value.Cmp(resource.Quantity{}) < 0 {
		allErrs = append(allErrs, field.Invalid(fldPath, value.String(), isNegativeErrorMsg))
	}
	return allErrs
}

func ValidatePositiveQuantity(value resource.Quantity, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList
	if value.Cmp(resource.Quantity{}) <= 0 {
		allErrs = append(allErrs, field.Invalid(fldPath, value.String(), isNotPositiveErrorMsg))
	}
	return allErrs
}

func ValidateSetOnceField(newVal, oldVal interface{}, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList
	if !reflect.ValueOf(oldVal).IsZero() {
		allErrs = append(allErrs, apivalidation.ValidateImmutableField(newVal, oldVal, fldPath)...)
	}
	return allErrs
}

func ValidateImmutable(newVal, oldVal interface{}, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList
	if !equality.Semantic.DeepEqual(oldVal, newVal) {
		allErrs = append(allErrs, field.Invalid(fldPath, newVal, apivalidation.FieldImmutableErrorMsg))
	}
	return allErrs
}

func ValidateImmutableWithDiff(newVal, oldVal interface{}, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList
	if !equality.Semantic.DeepEqual(oldVal, newVal) {
		diff := cmp.Diff(oldVal, newVal)
		allErrs = append(allErrs, field.Invalid(fldPath, newVal, fmt.Sprintf("%s\n%s", apivalidation.FieldImmutableErrorMsg, diff)))
	}
	return allErrs
}
