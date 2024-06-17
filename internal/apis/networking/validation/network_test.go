// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	"github.com/ironcore-dev/ironcore/internal/apis/networking"
	. "github.com/ironcore-dev/ironcore/internal/testutils/validation"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Network", func() {
	DescribeTable("ValidateNetwork",
		func(network *networking.Network, match types.GomegaMatcher) {
			errList := ValidateNetwork(network)
			Expect(errList).To(match)
		},
		Entry("missing name",
			&networking.Network{},
			ContainElement(RequiredField("metadata.name")),
		),
		Entry("missing namespace",
			&networking.Network{ObjectMeta: metav1.ObjectMeta{Name: "foo"}},
			ContainElement(RequiredField("metadata.namespace")),
		),
		Entry("bad name",
			&networking.Network{ObjectMeta: metav1.ObjectMeta{Name: "foo*"}},
			ContainElement(InvalidField("metadata.name")),
		),
		Entry("peering references itself",
			&networking.Network{
				ObjectMeta: metav1.ObjectMeta{Namespace: "ns", Name: "foo"},
				Spec: networking.NetworkSpec{
					Peerings: []networking.NetworkPeering{
						{
							Name: "peering",
							NetworkRef: networking.NetworkPeeringNetworkRef{
								Name: "foo",
							},
						},
					},
				},
			},
			ContainElement(ForbiddenField("spec.peerings[0].networkRef")),
		),
		Entry("duplicate peering name",
			&networking.Network{
				Spec: networking.NetworkSpec{
					Peerings: []networking.NetworkPeering{
						{Name: "peering"},
						{Name: "peering"},
					},
				},
			},
			ContainElement(DuplicateField("spec.peerings[1].name")),
		),
		Entry("duplicate network ref",
			&networking.Network{
				Spec: networking.NetworkSpec{
					Peerings: []networking.NetworkPeering{
						{NetworkRef: networking.NetworkPeeringNetworkRef{Name: "bar"}},
						{NetworkRef: networking.NetworkPeeringNetworkRef{Name: "bar"}},
					},
				},
			},
			ContainElement(DuplicateField("spec.peerings[1].networkRef")),
		),
		Entry("peering claim references itself",
			&networking.Network{
				ObjectMeta: metav1.ObjectMeta{Namespace: "ns", Name: "foo"},
				Spec: networking.NetworkSpec{
					PeeringClaimRefs: []networking.NetworkPeeringClaimRef{
						{
							Name:      "foo",
							Namespace: "ns",
						},
					},
				},
			},
			ContainElement(ForbiddenField("spec.incomingPeerings[0]")),
		),
		Entry("duplicate peering claim ref",
			&networking.Network{
				Spec: networking.NetworkSpec{
					PeeringClaimRefs: []networking.NetworkPeeringClaimRef{
						{Name: "bar"},
						{Name: "bar"},
					},
				},
			},
			ContainElement(DuplicateField("spec.incomingPeerings[1]")),
		),
		Entry("invalid peering claim ref name",
			&networking.Network{
				Spec: networking.NetworkSpec{
					PeeringClaimRefs: []networking.NetworkPeeringClaimRef{
						{Name: "bar*"},
					},
				},
			},
			ContainElement(InvalidField("spec.incomingPeerings[0].name")),
		),
		Entry("invalid peering claim ref namespace",
			&networking.Network{
				Spec: networking.NetworkSpec{
					PeeringClaimRefs: []networking.NetworkPeeringClaimRef{
						{Namespace: "ns*"},
					},
				},
			},
			ContainElements(InvalidField("spec.incomingPeerings[0].namespace"),
				RequiredField("spec.incomingPeerings[0].name")),
		),
		Entry("duplicate peering prefix name",
			&networking.Network{
				Spec: networking.NetworkSpec{
					Peerings: []networking.NetworkPeering{{
						Prefixes: []networking.PeeringPrefix{
							{Name: "peering"},
							{Name: "peering"},
						}},
					},
				},
			},
			ContainElement(DuplicateField("spec.peerings[0].prefixes[1].name")),
		),
		Entry("bad peering prefix cidr",
			&networking.Network{
				Spec: networking.NetworkSpec{
					Peerings: []networking.NetworkPeering{{
						Prefixes: []networking.PeeringPrefix{
							{Prefix: &commonv1alpha1.IPPrefix{}},
						}},
					},
				},
			},
			ContainElement(ForbiddenField("spec.peerings[0].prefixes[0].prefix")),
		),
		Entry("duplicate peering prefix ref",
			&networking.Network{
				Spec: networking.NetworkSpec{
					Peerings: []networking.NetworkPeering{{
						Prefixes: []networking.PeeringPrefix{
							{PrefixRef: networking.PeeringPrefixRef{Name: "foo"}},
							{PrefixRef: networking.PeeringPrefixRef{Name: "foo"}},
						}},
					},
				},
			},
			ContainElement(DuplicateField("spec.peerings[0].prefixes[1].prefixRef")),
		),
	)

	DescribeTable("ValidateNetworkUpdate",
		func(newNetwork, oldNetwork *networking.Network, match types.GomegaMatcher) {
			errList := ValidateNetworkUpdate(newNetwork, oldNetwork)
			Expect(errList).To(match)
		},
		Entry("immutable providerID if set",
			&networking.Network{
				Spec: networking.NetworkSpec{
					ProviderID: "foo",
				},
			},
			&networking.Network{
				Spec: networking.NetworkSpec{
					ProviderID: "bar",
				},
			},
			ContainElement(ImmutableField("spec.providerID")),
		),
		Entry("mutable providerID if not set",
			&networking.Network{
				Spec: networking.NetworkSpec{
					ProviderID: "foo",
				},
			},
			&networking.Network{},
			Not(ContainElement(ImmutableField("spec.providerID"))),
		),
	)
})
