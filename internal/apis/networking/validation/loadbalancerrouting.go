// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	commonvalidation "github.com/ironcore-dev/ironcore/internal/apis/common/validation"
	"github.com/ironcore-dev/ironcore/internal/apis/networking"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateLoadBalancerRouting validates a LoadBalancerRouting object.
func ValidateLoadBalancerRouting(loadBalancerRouting *networking.LoadBalancerRouting) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessor(loadBalancerRouting, true, apivalidation.NameIsDNSSubdomain, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateLoadBalancerRouting(loadBalancerRouting)...)

	return allErrs
}

func validateLoadBalancerRouting(loadBalancerRouting *networking.LoadBalancerRouting) field.ErrorList {
	var allErrs field.ErrorList

	destinationsField := field.NewPath("destinations")
	for idx := range loadBalancerRouting.Destinations {
		fldPath := destinationsField.Index(idx)
		destination := &loadBalancerRouting.Destinations[idx]

		allErrs = append(allErrs, commonvalidation.ValidateIP(destination.IP.Family(), destination.IP, fldPath.Child("ip"))...)

		if targetRef := destination.TargetRef; targetRef != nil {
			for _, msg := range apivalidation.NameIsDNSSubdomain(targetRef.Name, false) {
				allErrs = append(allErrs, field.Invalid(fldPath.Child("targetRef", "name"), targetRef.Name, msg))
			}
		}
	}

	return allErrs
}

// ValidateLoadBalancerRoutingUpdate validates a LoadBalancerRouting object before an update.
func ValidateLoadBalancerRoutingUpdate(newLoadBalancerRouting, oldLoadBalancerRouting *networking.LoadBalancerRouting) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessorUpdate(newLoadBalancerRouting, oldLoadBalancerRouting, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateLoadBalancerRoutingUpdate(newLoadBalancerRouting, oldLoadBalancerRouting)...)
	allErrs = append(allErrs, ValidateLoadBalancerRouting(newLoadBalancerRouting)...)

	return allErrs
}

// validateLoadBalancerRoutingUpdate validates the spec of a loadBalancerRouting object before an update.
func validateLoadBalancerRoutingUpdate(new, old *networking.LoadBalancerRouting) field.ErrorList {
	var allErrs field.ErrorList

	return allErrs
}
