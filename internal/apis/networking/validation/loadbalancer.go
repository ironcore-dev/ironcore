// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"fmt"

	ironcorevalidation "github.com/ironcore-dev/ironcore/internal/api/validation"
	"github.com/ironcore-dev/ironcore/internal/apis/networking"
	corev1 "k8s.io/api/core/v1"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1validation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateLoadBalancer validates an LoadBalancer object.
func ValidateLoadBalancer(loadBalancer *networking.LoadBalancer) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessor(loadBalancer, true, apivalidation.NameIsDNSLabel, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateLoadBalancerSpec(&loadBalancer.Spec, &loadBalancer.ObjectMeta, field.NewPath("spec"))...)

	return allErrs
}

func validateLoadBalancerSpec(spec *networking.LoadBalancerSpec, lbMeta *metav1.ObjectMeta, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, validateLoadBalancerType(spec.Type, fldPath.Child("type"))...)

	allErrs = append(allErrs, ironcorevalidation.ValidateIPFamilies(spec.IPFamilies, fldPath.Child("ipFamilies"))...)

	allErrs = append(allErrs, validateNetworkInterfaceIPSources(spec.IPs, spec.IPFamilies, lbMeta, fldPath.Child("ips"))...)

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
	networking.LoadBalancerTypeInternal,
)

func validateLoadBalancerType(loadBalancerType networking.LoadBalancerType, fldPath *field.Path) field.ErrorList {
	return ironcorevalidation.ValidateEnum(supportedLoadBalancerTypes, loadBalancerType, fldPath, "must specify type")
}

func validateLoadBalancerPort(port networking.LoadBalancerPort, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if port.Protocol != nil {
		allErrs = append(allErrs, ironcorevalidation.ValidateProtocol(*port.Protocol, fldPath.Child("protocol"))...)
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
	allErrs = append(allErrs, validateLoadBalancerSpecUpdate(&newLoadBalancer.Spec, &oldLoadBalancer.Spec, field.NewPath("spec"))...)
	allErrs = append(allErrs, ValidateLoadBalancer(newLoadBalancer)...)

	return allErrs
}

// validateLoadBalancerSpecUpdate validates the spec of a loadBalancer object before an update.
func validateLoadBalancerSpecUpdate(newSpec, oldSpec *networking.LoadBalancerSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, ironcorevalidation.ValidateImmutableField(newSpec.NetworkRef, oldSpec.NetworkRef, fldPath.Child("networkRef"))...)

	return allErrs
}
