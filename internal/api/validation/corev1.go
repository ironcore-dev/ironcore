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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var supportedIPFamilies = sets.New(
	corev1.IPv4Protocol,
	corev1.IPv6Protocol,
)

func IsSupportedIPFamily(ipFamily corev1.IPFamily) bool {
	return supportedIPFamilies.Has(ipFamily)
}

func ValidateIPFamily(ipFamily corev1.IPFamily, fldPath *field.Path) field.ErrorList {
	return ValidateEnum(supportedIPFamilies, ipFamily, fldPath, "must specify ipFamily")
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
