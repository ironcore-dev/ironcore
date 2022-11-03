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

	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	onmetalapivalidation "github.com/onmetal/onmetal-api/onmetal-apiserver/api/validation"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func ValidateIPPrefix(ipFamily corev1.IPFamily, ipPrefix commonv1alpha1.IPPrefix, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if !ipPrefix.IsValid() {
		allErrs = append(allErrs, field.Invalid(fldPath, ipPrefix, "must specify a valid prefix"))
	} else {
		if !onmetalapivalidation.IsSupportedIPFamily(ipFamily) {
			allErrs = append(allErrs, field.Invalid(fldPath, ipPrefix, "cannot determine ip family for prefix"))
		} else if ipFamily != ipPrefix.IP().Family() {
			allErrs = append(allErrs, field.Invalid(fldPath, ipPrefix, fmt.Sprintf("expected ip family %s but got %s", ipFamily, ipPrefix.IP().Family())))
		}
	}

	return allErrs
}

func ValidateIP(ipFamily corev1.IPFamily, ip commonv1alpha1.IP, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if !ip.IsValid() {
		allErrs = append(allErrs, field.Invalid(fldPath, ip, "must specify a valid ip"))
	} else {
		if !onmetalapivalidation.IsSupportedIPFamily(ipFamily) {
			allErrs = append(allErrs, field.Invalid(fldPath, ip, "cannot determine ip family for ipo"))
		} else if ipFamily != ip.Family() {
			allErrs = append(allErrs, field.Invalid(fldPath, ip, fmt.Sprintf("expected ip family %s but got %s", ipFamily, ip.Family())))
		}
	}

	return allErrs
}

func ValidatePrefixLength(ipFamily corev1.IPFamily, prefixLength int32, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if prefixLength <= 0 {
		allErrs = append(allErrs, field.Invalid(fldPath, prefixLength, "has to be > 0"))
	} else {
		switch ipFamily {
		case corev1.IPv4Protocol:
			if prefixLength > 32 {
				allErrs = append(allErrs, field.Invalid(fldPath, prefixLength, "too large prefix length for IPv4 (max 32)"))
			}
		case corev1.IPv6Protocol:
			if prefixLength > 128 {
				allErrs = append(allErrs, field.Invalid(fldPath, prefixLength, "too large prefix length for IPv6 (max 128)"))
			}
		default:
			allErrs = append(allErrs, field.Invalid(fldPath, prefixLength, "no ip family specified"))
		}
	}

	return allErrs
}
