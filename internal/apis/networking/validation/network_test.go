// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
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
