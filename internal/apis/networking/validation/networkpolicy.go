// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"fmt"

	ironcorevalidation "github.com/ironcore-dev/ironcore/internal/api/validation"
	"github.com/ironcore-dev/ironcore/internal/apis/core"
	"github.com/ironcore-dev/ironcore/internal/apis/networking"
	"go4.org/netipx"
	corev1 "k8s.io/api/core/v1"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1validation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func ValidateNetworkPolicy(networkPolicy *networking.NetworkPolicy) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessor(networkPolicy, true, apivalidation.NameIsDNSLabel, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateNetworkPolicySpec(&networkPolicy.Spec, field.NewPath("spec"))...)

	return allErrs
}

func validateNetworkPolicySpec(spec *networking.NetworkPolicySpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if spec.NetworkRef == (corev1.LocalObjectReference{}) {
		allErrs = append(allErrs, field.Required(fldPath.Child("networkRef"), "must specify a network ref"))
	} else {
		for _, msg := range apivalidation.NameIsDNSLabel(spec.NetworkRef.Name, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("networkRef").Child("name"), spec.NetworkRef.Name, msg))
		}
	}

	allErrs = append(allErrs, metav1validation.ValidateLabelSelector(&spec.NetworkInterfaceSelector, metav1validation.LabelSelectorValidationOptions{}, fldPath.Child("networkInterfaceSelector"))...)

	for i := range spec.Ingress {
		ingressRule := &spec.Ingress[i]
		fldPath := fldPath.Child("ingress").Index(i)
		allErrs = append(allErrs, validateNetworkPolicyIngressRule(ingressRule, fldPath)...)
	}

	for i := range spec.Egress {
		egressRule := &spec.Egress[i]
		fldPath := fldPath.Child("egress").Index(i)
		allErrs = append(allErrs, validateNetworkPolicyEgressRule(egressRule, fldPath)...)
	}

	if len(spec.PolicyTypes) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("policyTypes"), "must specify policyTypes"))
	} else {
		allErrs = append(allErrs, validatePolicyTypes(spec.PolicyTypes, fldPath.Child("policyTypes"))...)
	}

	return allErrs
}

var supportedIngressObjectSelectorKinds = sets.New[string](
	"NetworkInterface",
	"LoadBalancer",
	"VirtualIP",
)

func validateNetworkPolicyIngressRule(rule *networking.NetworkPolicyIngressRule, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	for i := range rule.From {
		from := &rule.From[i]
		fldPath := fldPath.Child("from").Index(i)
		allErrs = append(allErrs, validateNetworkPolicyPeer(from, supportedIngressObjectSelectorKinds, fldPath)...)
	}

	for i := range rule.Ports {
		port := &rule.Ports[i]
		fldPath := fldPath.Child("ports").Index(i)
		allErrs = append(allErrs, validateNetworkPolicyPort(port, fldPath)...)
	}

	return allErrs
}

var supportedEgressObjectSelectorKinds = sets.New[string](
	"NetworkInterface",
	"LoadBalancer",
	"VirtualIP",
)

func validateNetworkPolicyEgressRule(rule *networking.NetworkPolicyEgressRule, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	for i := range rule.To {
		to := &rule.To[i]
		fldPath := fldPath.Child("to").Index(i)
		allErrs = append(allErrs, validateNetworkPolicyPeer(to, supportedEgressObjectSelectorKinds, fldPath)...)
	}

	for i := range rule.Ports {
		port := &rule.Ports[i]
		fldPath := fldPath.Child("ports").Index(i)
		allErrs = append(allErrs, validateNetworkPolicyPort(port, fldPath)...)
	}

	return allErrs
}

func validateNetworkPolicyPort(port *networking.NetworkPolicyPort, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if port.Port != 0 {
		if issues := validation.IsValidPortNum(int(port.Port)); len(issues) > 0 {
			for _, issue := range issues {
				allErrs = append(allErrs, field.Invalid(fldPath.Child("port"), port.Port, issue))
			}
		}
	}

	if endPort := port.EndPort; endPort != nil {
		if port.Port == 0 {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("endPort"), "must not specify endPort without port"))
		} else {
			if *endPort < port.Port {
				allErrs = append(allErrs, field.Forbidden(fldPath.Child("endPort"), fmt.Sprintf("endPort %d must not be smaller than port %d", *endPort, port.Port)))
			}
		}
	}

	if protocol := port.Protocol; protocol != nil {
		allErrs = append(allErrs, ironcorevalidation.ValidateProtocol(*protocol, fldPath.Child("protocol"))...)
	}

	return allErrs
}

func validateNetworkPolicyPeer(peer *networking.NetworkPolicyPeer, supportedObjectSelectorKinds sets.Set[string], fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	var numPeers int

	if peer.ObjectSelector != nil {
		if numPeers > 0 {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("ipBlock"), "cannot specify multiple peers"))
		} else {
			numPeers++
			allErrs = append(allErrs, validateNetworkPolicyPeerObjectSelector(peer.ObjectSelector, supportedObjectSelectorKinds, fldPath.Child("objectSelector"))...)
		}
	}

	if peer.IPBlock != nil {
		if numPeers > 0 {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("ipBlock"), "cannot specify multiple peers"))
		} else {
			numPeers++ //nolint:ineffassign
			allErrs = append(allErrs, validateIPBlock(peer.IPBlock, fldPath.Child("ipBlock"))...)
		}
	}

	return allErrs
}

func validateNetworkPolicyPeerObjectSelector(sel *core.ObjectSelector, allowedKinds sets.Set[string], fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, ironcorevalidation.ValidateEnum(allowedKinds, sel.Kind, fldPath.Child("kind"), "must specify kind")...)
	allErrs = append(allErrs, metav1validation.ValidateLabelSelector(&sel.LabelSelector, metav1validation.LabelSelectorValidationOptions{}, fldPath)...)

	return allErrs
}

func validateIPBlock(ipBlock *networking.IPBlock, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if !ipBlock.CIDR.IsValid() {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("cidr"), ipBlock.CIDR, "must specify valid cidr"))
	} else {
		var bldr netipx.IPSetBuilder
		bldr.AddPrefix(ipBlock.CIDR.Prefix)
		ipSet, _ := bldr.IPSet()

		for i, except := range ipBlock.Except {
			fldPath := fldPath.Child("except").Index(i)
			if !except.IsValid() {
				allErrs = append(allErrs, field.Invalid(fldPath, except, "must specify valid except value"))
			} else {
				if !ipSet.ContainsPrefix(except.Prefix) {
					allErrs = append(allErrs,
						field.Forbidden(fldPath, fmt.Sprintf("cidr %s does not contain except %s",
							ipBlock.CIDR, except)),
					)
				}
			}
		}
	}

	return allErrs
}

var supportedPolicyTypes = sets.New(
	networking.PolicyTypeIngress,
	networking.PolicyTypeEgress,
)

func validatePolicyType(policyType networking.PolicyType, fldPath *field.Path) field.ErrorList {
	return ironcorevalidation.ValidateEnum(supportedPolicyTypes, policyType, fldPath, "must specify type")
}

func validatePolicyTypes(policyTypes []networking.PolicyType, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	seen := sets.New[networking.PolicyType]()

	for i := range policyTypes {
		policyType := policyTypes[i]
		fldPath := fldPath.Index(i)
		allErrs = append(allErrs, validatePolicyType(policyType, fldPath)...)
		if seen.Has(policyType) {
			allErrs = append(allErrs, field.Duplicate(fldPath, policyType))
		} else {
			seen.Insert(policyType)
		}
	}

	return allErrs
}

// ValidateNetworkPolicyUpdate validates a NetworkPolicy object before an update.
func ValidateNetworkPolicyUpdate(newNetworkPolicy, oldNetworkPolicy *networking.NetworkPolicy) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessorUpdate(newNetworkPolicy, oldNetworkPolicy, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateNetworkPolicySpecUpdate(&newNetworkPolicy.Spec, &oldNetworkPolicy.Spec, field.NewPath("spec"))...)
	allErrs = append(allErrs, ValidateNetworkPolicy(newNetworkPolicy)...)

	return allErrs
}

// validateNetworkPolicySpecUpdate validates the spec of a networkPolicy object before an update.
func validateNetworkPolicySpecUpdate(newSpec, oldSpec *networking.NetworkPolicySpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, ironcorevalidation.ValidateImmutableField(newSpec.NetworkRef, oldSpec.NetworkRef, fldPath.Child("networkRef"))...)

	return allErrs
}
