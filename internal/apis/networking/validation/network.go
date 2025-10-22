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

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessor(network, true, apivalidation.NameIsDNSSubdomain, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateNetworkSpec(network.Namespace, network.Name, &network.Spec, field.NewPath("spec"))...)

	return allErrs
}

func validateNetworkSpec(namespace, name string, spec *networking.NetworkSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	seenNames := sets.New[string]()
	seenPeeringNetworkKeys := sets.New[client.ObjectKey]()

	seenPrefixNames := sets.New[string]()
	seenPeeringPrefixKeys := sets.New[client.ObjectKey]()

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

		for j, prefix := range peering.Prefixes {
			fldPath := fldPath.Child("prefixes").Index(j)

			if seenPrefixNames.Has(prefix.Name) {
				allErrs = append(allErrs, field.Duplicate(fldPath.Child("name"), prefix.Name))
			} else {
				seenPrefixNames.Insert(prefix.Name)
			}

			if peeringPrefix := prefix.Prefix; peeringPrefix != nil {
				if !peeringPrefix.IsValid() {
					allErrs = append(allErrs, field.Forbidden(fldPath.Child("prefix"), "must be a valid IP range"))
				}
			}

			peeringPrefixKey := client.ObjectKey{Namespace: namespace, Name: prefix.PrefixRef.Name}

			if seenPeeringPrefixKeys.Has(peeringPrefixKey) {
				allErrs = append(allErrs, field.Duplicate(fldPath.Child("prefixRef"), prefix.PrefixRef))
			} else {
				seenPeeringPrefixKeys.Insert(peeringPrefixKey)
			}

			allErrs = append(allErrs, validatePeeringPrefix(prefix, fldPath)...)
		}

		allErrs = append(allErrs, validateNetworkPeering(peering, fldPath)...)
	}

	seenPeeringClaimRefKeys := sets.New[client.ObjectKey]()

	for i, peeringClaimRef := range spec.PeeringClaimRefs {
		fldPath := fldPath.Child("incomingPeerings").Index(i)

		peeringClaimRefNamespace := peeringClaimRef.Namespace
		if peeringClaimRefNamespace == "" {
			peeringClaimRefNamespace = namespace
		}

		peeringClaimRefkKey := client.ObjectKey{Namespace: peeringClaimRefNamespace, Name: peeringClaimRef.Name}

		if name != "" && (client.ObjectKey{Namespace: namespace, Name: name}) == peeringClaimRefkKey {
			allErrs = append(allErrs, field.Forbidden(fldPath, "cannot claim itself"))
		} else if seenPeeringClaimRefKeys.Has(peeringClaimRefkKey) {
			allErrs = append(allErrs, field.Duplicate(fldPath, peeringClaimRef))
		} else {
			seenPeeringClaimRefKeys.Insert(peeringClaimRefkKey)
		}

		allErrs = append(allErrs, validatePeeringClaimRef(peeringClaimRef, fldPath)...)
	}

	return allErrs
}

func validateNetworkPeering(peering networking.NetworkPeering, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	for _, msg := range apivalidation.NameIsDNSSubdomain(peering.Name, false) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("name"), peering.Name, msg))
	}

	networkRef := peering.NetworkRef
	if networkRef.Namespace != "" {
		for _, msg := range apivalidation.NameIsDNSLabel(networkRef.Namespace, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("networkRef", "namespace"), networkRef.Namespace, msg))
		}
	}
	for _, msg := range apivalidation.NameIsDNSSubdomain(networkRef.Name, false) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("networkRef", "name"), networkRef.Name, msg))
	}

	return allErrs
}

func validatePeeringClaimRef(peeringClaimRef networking.NetworkPeeringClaimRef, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if len(peeringClaimRef.Name) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("name"), "name is required"))
	} else {
		for _, msg := range apivalidation.NameIsDNSSubdomain(peeringClaimRef.Name, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("name"), peeringClaimRef.Name, msg))
		}
	}

	if peeringClaimRef.Namespace != "" {
		for _, msg := range apivalidation.NameIsDNSSubdomain(peeringClaimRef.Namespace, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("namespace"), peeringClaimRef.Namespace, msg))
		}
	}
	return allErrs
}

func validatePeeringPrefix(prefix networking.PeeringPrefix, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	for _, msg := range apivalidation.NameIsDNSLabel(prefix.Name, false) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("name"), prefix.Name, msg))
	}

	prefixRef := prefix.PrefixRef
	if prefixRef.Name != "" {
		for _, msg := range apivalidation.NameIsDNSLabel(prefixRef.Name, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("prefixRef", "name"), prefixRef.Name, msg))
		}
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
