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

	commonv1alpha1validation "github.com/onmetal/onmetal-api/apis/common/v1alpha1/validation"
	"github.com/onmetal/onmetal-api/apis/networking"
	corev1 "k8s.io/api/core/v1"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func ValidateNetworkInterfaceBinding(networkInterfaceBinding *networking.NetworkInterfaceBinding) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessor(networkInterfaceBinding, true, apivalidation.NameIsDNSLabel, field.NewPath("metadata"))...)

	seenIPFamilies := make(map[corev1.IPFamily]struct{})
	for i, ip := range networkInterfaceBinding.IPs {
		allErrs = append(allErrs, commonv1alpha1validation.ValidateIP(ip.Family(), ip, field.NewPath("ips").Index(i))...)

		if _, ok := seenIPFamilies[ip.Family()]; ok {
			allErrs = append(allErrs, field.Forbidden(field.NewPath("ips").Index(i), fmt.Sprintf("duplicate ip family %q", ip.Family())))
		}
		seenIPFamilies[ip.Family()] = struct{}{}
	}

	if virtualIPRef := networkInterfaceBinding.VirtualIPRef; virtualIPRef != nil {
		for _, msg := range apivalidation.NameIsDNSLabel(virtualIPRef.Name, false) {
			allErrs = append(allErrs, field.Invalid(field.NewPath("virtualIPRef", "name"), virtualIPRef.Name, msg))
		}
	}

	return allErrs
}

func ValidateNetworkInterfaceBindingUpdate(newNetworkInterfaceBinding, oldNetworkInterfaceBinding *networking.NetworkInterfaceBinding) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessorUpdate(newNetworkInterfaceBinding, oldNetworkInterfaceBinding, field.NewPath("metadata"))...)
	allErrs = append(allErrs, ValidateNetworkInterfaceBinding(newNetworkInterfaceBinding)...)

	return allErrs
}
