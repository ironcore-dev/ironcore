/*
 * Copyright (c) 2022 by the OnMetal authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package validation

import (
	"fmt"

	onmetalapivalidation "github.com/onmetal/onmetal-api/internal/api/validation"
	"github.com/onmetal/onmetal-api/internal/apis/networking"
	corev1 "k8s.io/api/core/v1"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1validation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateLoadBalancer validates an LoadBalancer object.
func ValidateLoadBalancer(loadBalancer *networking.LoadBalancer) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessor(loadBalancer, true, apivalidation.NameIsDNSLabel, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateLoadBalancerSpec(&loadBalancer.Spec, field.NewPath("spec"))...)

	return allErrs
}

func validateLoadBalancerSpec(spec *networking.LoadBalancerSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, validateLoadBalancerType(spec.Type, fldPath.Child("type"))...)

	allErrs = append(allErrs, onmetalapivalidation.ValidateIPFamilies(spec.IPFamilies, fldPath.Child("ipFamilies"))...)

	for _, msg := range apivalidation.NameIsDNSLabel(spec.NetworkRef.Name, false) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("networkRef").Child("name"), spec.NetworkRef.Name, msg))
	}

	allErrs = append(allErrs, metav1validation.ValidateLabelSelector(spec.NetworkInterfaceSelector, metav1validation.LabelSelectorValidationOptions{}, fldPath.Child("networkInterfaceSelector"))...)

	var (
		portRangesByProtocol = make(map[corev1.Protocol][][2]int32)
	)
	for i, port := range spec.Ports {
		portFldPath := fldPath.Child("ports").Index(i)
		portRange := getLoadBalancerPortRange(port)
		protocol := getLoadBalancerProtocol(port.Protocol)
		portRanges := portRangesByProtocol[protocol]

		for _, existingPortRange := range portRanges {
			allErrs = append(allErrs, validateLoadBalancerPort(port, portFldPath)...)
			if portRangesOverlap(portRange, existingPortRange) {
				allErrs = append(allErrs, field.Forbidden(portFldPath, fmt.Sprintf("port range %v overlaps with port range %v", portRange, existingPortRange)))
			}
		}

		portRangesByProtocol[protocol] = append(portRanges, portRange)
	}

	return allErrs
}

func getLoadBalancerProtocol(protocol *corev1.Protocol) corev1.Protocol {
	if protocol != nil {
		return *protocol
	}
	return corev1.ProtocolTCP
}

func getLoadBalancerPortRange(port networking.LoadBalancerPort) [2]int32 {
	if endPort := port.EndPort; endPort != nil {
		return [2]int32{port.Port, *endPort}
	}
	return [2]int32{port.Port, port.Port}
}

func portRangesOverlap(x, y [2]int32) bool {
	return x[0] <= y[1] && y[0] <= x[1]
}

var supportedLoadBalancerTypes = sets.New(
	networking.LoadBalancerTypePublic,
)

func validateLoadBalancerType(loadBalancerType networking.LoadBalancerType, fldPath *field.Path) field.ErrorList {
	return onmetalapivalidation.ValidateEnum(supportedLoadBalancerTypes, loadBalancerType, fldPath, "must specify type")
}

func validateLoadBalancerPort(port networking.LoadBalancerPort, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if port.Protocol != nil {
		allErrs = append(allErrs, onmetalapivalidation.ValidateProtocol(*port.Protocol, fldPath.Child("protocol"))...)
	}

	for _, msg := range validation.IsValidPortNum(int(port.Port)) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("port"), port.Port, msg))
	}

	if port.EndPort != nil {
		for _, msg := range validation.IsValidPortNum(int(*port.EndPort)) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("endPort"), *port.EndPort, msg))
		}
		if *port.EndPort < port.Port {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("endPort"), fmt.Sprintf("endPort %d must be >= port %d", *port.EndPort, port.Port)))
		}
	}

	return allErrs
}

// ValidateLoadBalancerUpdate validates a LoadBalancer object before an update.
func ValidateLoadBalancerUpdate(newLoadBalancer, oldLoadBalancer *networking.LoadBalancer) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessorUpdate(newLoadBalancer, oldLoadBalancer, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateLoadBalancerSpecPrefixUpdate(&newLoadBalancer.Spec, &oldLoadBalancer.Spec, field.NewPath("spec"))...)
	allErrs = append(allErrs, ValidateLoadBalancer(newLoadBalancer)...)

	return allErrs
}

// validateLoadBalancerSpecPrefixUpdate validates the spec of a loadBalancer object before an update.
func validateLoadBalancerSpecPrefixUpdate(newSpec, oldSpec *networking.LoadBalancerSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, onmetalapivalidation.ValidateImmutableField(newSpec.NetworkRef, oldSpec.NetworkRef, fldPath.Child("networkRef"))...)

	return allErrs
}
