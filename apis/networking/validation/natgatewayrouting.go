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
	"github.com/onmetal/onmetal-api/apis/networking"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateNATGatewayRouting validates an NATGatewayRouting object.
func ValidateNATGatewayRouting(natGatewayRouting *networking.NATGatewayRouting) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessor(natGatewayRouting, true, apivalidation.NameIsDNSLabel, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateNATGatewayRouting(natGatewayRouting)...)

	return allErrs
}

func validateNATGatewayRouting(natGatewayRouting *networking.NATGatewayRouting) field.ErrorList {
	var allErrs field.ErrorList

	seenNames := sets.NewString()
	for idx := range natGatewayRouting.Destinations {
		destination := &natGatewayRouting.Destinations[idx]
		destinationPath := field.NewPath("destinations").Index(idx)
		if seenNames.Has(destination.Name) {
			allErrs = append(allErrs, field.Duplicate(destinationPath, destination))
		} else {
			seenNames.Insert(destination.Name)
			for _, msg := range apivalidation.NameIsDNSLabel(destination.Name, false) {
				allErrs = append(allErrs, field.Invalid(destinationPath.Child("name"), destination.Name, msg))
			}
		}
	}

	return allErrs
}

// ValidateNATGatewayRoutingUpdate validates a NATGatewayRouting object before an update.
func ValidateNATGatewayRoutingUpdate(newNATGatewayRouting, oldNATGatewayRouting *networking.NATGatewayRouting) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessorUpdate(newNATGatewayRouting, oldNATGatewayRouting, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateNATGatewayRoutingUpdate(newNATGatewayRouting, oldNATGatewayRouting)...)
	allErrs = append(allErrs, ValidateNATGatewayRouting(newNATGatewayRouting)...)

	return allErrs
}

// validateNATGatewayRoutingUpdate validates the spec of a natGatewayRouting object before an update.
func validateNATGatewayRoutingUpdate(new, old *networking.NATGatewayRouting) field.ErrorList {
	var allErrs field.ErrorList

	return allErrs
}
