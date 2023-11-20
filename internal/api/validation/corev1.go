// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package apivalidation

import (
	"github.com/ironcore-dev/ironcore/internal/apis/storage"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var supportedIPFamilies = sets.New(
	corev1.IPv4Protocol,
	corev1.IPv6Protocol,
)

var supportedResizePolicies = sets.New(
	storage.ResizePolicyStatic,
	storage.ResizePolicyExpandOnly,
)

func IsSupportedIPFamily(ipFamily corev1.IPFamily) bool {
	return supportedIPFamilies.Has(ipFamily)
}

func ValidateIPFamily(ipFamily corev1.IPFamily, fldPath *field.Path) field.ErrorList {
	return ValidateEnum(supportedIPFamilies, ipFamily, fldPath, "must specify ipFamily")
}

func ValidateResizePolicy(policy storage.ResizePolicy, fldPath *field.Path) field.ErrorList {
	return ValidateEnum(supportedResizePolicies, policy, fldPath, "must specify resizePolicy")
}

func ValidateIPFamilies(ipFamilies []corev1.IPFamily, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if len(ipFamilies) == 0 {
		allErrs = append(allErrs, field.Required(fldPath, "must specify ip families"))
		return allErrs
	}

	seen := sets.NewString()
	for i, ipFamily := range ipFamilies {
		allErrs = append(allErrs, ValidateIPFamily(ipFamily, fldPath.Index(i))...)

		if seen.Has(string(ipFamily)) {
			allErrs = append(allErrs, field.Duplicate(fldPath.Index(i), ipFamily))
		}
		seen.Insert(string(ipFamily))
	}

	return allErrs
}

var supportedProtocols = sets.New(
	corev1.ProtocolTCP,
	corev1.ProtocolUDP,
	corev1.ProtocolSCTP,
)

func ValidateProtocol(protocol corev1.Protocol, fldPath *field.Path) field.ErrorList {
	return ValidateEnum(supportedProtocols, protocol, fldPath, "must specify protocol")
}
