// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"fmt"

	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	ironcorevalidation "github.com/ironcore-dev/ironcore/internal/api/validation"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func ValidateIPPrefix(ipFamily corev1.IPFamily, ipPrefix commonv1alpha1.IPPrefix, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if !ipPrefix.IsValid() {
		allErrs = append(allErrs, field.Invalid(fldPath, ipPrefix, "must specify a valid prefix"))
	} else {
		if !ironcorevalidation.IsSupportedIPFamily(ipFamily) {
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
		if !ironcorevalidation.IsSupportedIPFamily(ipFamily) {
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
