// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	ironcorevalidation "github.com/ironcore-dev/ironcore/internal/api/validation"
	"github.com/ironcore-dev/ironcore/internal/apis/networking"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ValidateNetwork validates a network object.
func ValidateNetwork(network *networking.Network) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessor(network, true, apivalidation.NameIsDNSLabel, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateNetworkSpec(network.Namespace, network.Name, &network.Spec, field.NewPath("spec"))...)

	return allErrs
}

func validateNetworkSpec(namespace, name string, spec *networking.NetworkSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	seenNames := sets.New[string]()
	seenPeeringNetworkKeys := sets.New[client.ObjectKey]()

	for i, peering := range spec.Peerings {
		fldPath := fldPath.Child("peerings").Index(i)
		if seenNames.Has(peering.Name) {
			allErrs = append(allErrs, field.Duplicate(fldPath.Child("name"), peering.Name))
		} else {
			seenNames.Insert(peering.Name)
		}

		peeringNetworkNamespace := peering.NetworkRef.Namespace
		if peeringNetworkNamespace == "" {
			peeringNetworkNamespace = namespace
		}

		peeringNetworkKey := client.ObjectKey{Namespace: peeringNetworkNamespace, Name: peering.NetworkRef.Name}

		if name != "" && (client.ObjectKey{Namespace: namespace, Name: name}) == peeringNetworkKey {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("networkRef"), "cannot peer a network with itself"))
		} else if seenPeeringNetworkKeys.Has(peeringNetworkKey) {
			allErrs = append(allErrs, field.Duplicate(fldPath.Child("networkRef"), peering.NetworkRef))
		} else {
			seenPeeringNetworkKeys.Insert(peeringNetworkKey)
		}

		allErrs = append(allErrs, validateNetworkPeering(peering, fldPath)...)
	}

	return allErrs
}

func validateNetworkPeering(peering networking.NetworkPeering, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	for _, msg := range apivalidation.NameIsDNSLabel(peering.Name, false) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("name"), peering.Name, msg))
	}

	networkRef := peering.NetworkRef
	if networkRef.Namespace != "" {
		for _, msg := range apivalidation.NameIsDNSLabel(networkRef.Namespace, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("networkRef", "namespace"), networkRef.Namespace, msg))
		}
	}
	for _, msg := range apivalidation.NameIsDNSLabel(networkRef.Name, false) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("networkRef", "name"), networkRef.Name, msg))
	}

	return allErrs
}

// ValidateNetworkUpdate validates a Network object before an update.
func ValidateNetworkUpdate(newNetwork, oldNetwork *networking.Network) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessorUpdate(newNetwork, oldNetwork, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateNetworkSpecUpdate(&newNetwork.Spec, &oldNetwork.Spec, field.NewPath("spec"))...)
	allErrs = append(allErrs, ValidateNetwork(newNetwork)...)

	return allErrs
}

func validateNetworkSpecUpdate(newSpec, oldSpec *networking.NetworkSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if oldSpec.ProviderID != "" {
		allErrs = append(allErrs, ironcorevalidation.ValidateImmutableField(newSpec.ProviderID, oldSpec.ProviderID, fldPath.Child("providerID"))...)
	}

	return allErrs
}
