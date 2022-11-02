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

package apivalidation

import (
	"fmt"
	"reflect"
	"sort"
	"unicode"
	"unicode/utf8"

	"github.com/google/go-cmp/cmp"
	"github.com/onmetal/controller-utils/set"
	"github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	"github.com/onmetal/onmetal-api/equality"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

const (
	isNegativeErrorMsg    = validation.IsNegativeErrorMsg
	isNotPositiveErrorMsg = `must be greater than zero`
)

func ValidatePowerOfTwo(value int64, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList
	if value <= 0 || value&(value-1) != 0 {
		allErrs = append(allErrs, field.Invalid(fldPath, value, fmt.Sprintf("%d is not a power of 2", value)))
	}
	return allErrs
}

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
		allErrs = append(allErrs, ValidateImmutableField(newVal, oldVal, fldPath)...)
	}
	return allErrs
}

func ValidateImmutableField(newVal, oldVal interface{}, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList
	if !equality.Semantic.DeepEqual(oldVal, newVal) {
		allErrs = append(allErrs, field.Forbidden(fldPath, validation.FieldImmutableErrorMsg))
	}
	return allErrs
}

func ValidateImmutableFieldWithDiff(newVal, oldVal interface{}, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList
	if !equality.Semantic.DeepEqual(oldVal, newVal) {
		diff := cmp.Diff(oldVal, newVal, cmp.Comparer(v1alpha1.EqualIPs), cmp.Comparer(v1alpha1.EqualIPPrefixes))
		allErrs = append(allErrs, field.Forbidden(fldPath, fmt.Sprintf("%s\n%s", validation.FieldImmutableErrorMsg, diff)))
	}
	return allErrs
}

func ValidateEnum[E comparable](allowed set.Set[E], value E, fldPath *field.Path, requiredDetail string) field.ErrorList {
	var allErrs field.ErrorList
	var zero E
	if value == zero && !allowed.Has(zero) {
		allErrs = append(allErrs, field.Required(fldPath, requiredDetail))
	} else if !allowed.Has(value) {
		validValues := make([]string, 0, allowed.Len())
		for item := range allowed {
			validValues = append(validValues, fmt.Sprintf("%v", item))
		}
		sort.Strings(validValues)

		allErrs = append(allErrs, field.NotSupported(fldPath, value, validValues))
	}
	return allErrs
}

// ValidateFieldAllowList checks that only allowed fields are set.
// The value must be a struct (not a pointer to a struct!).
func ValidateFieldAllowList(value interface{}, allowedFields sets.String, errorText string, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	reflectType, reflectValue := reflect.TypeOf(value), reflect.ValueOf(value)
	for i := 0; i < reflectType.NumField(); i++ {
		f := reflectType.Field(i)
		if allowedFields.Has(f.Name) {
			continue
		}

		// Compare the value of this field to its zero value to determine if it has been set
		if !equality.Semantic.DeepEqual(reflectValue.Field(i).Interface(), reflect.Zero(f.Type).Interface()) {
			r, n := utf8.DecodeRuneInString(f.Name)
			lcName := string(unicode.ToLower(r)) + f.Name[n:]
			allErrs = append(allErrs, field.Forbidden(fldPath.Child(lcName), errorText))
		}
	}

	return allErrs
}
