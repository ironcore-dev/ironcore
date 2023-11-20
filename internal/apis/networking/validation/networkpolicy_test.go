// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	"github.com/ironcore-dev/ironcore/internal/apis/core"
	"github.com/ironcore-dev/ironcore/internal/apis/networking"
	. "github.com/ironcore-dev/ironcore/internal/testutils/validation"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("NetworkPolicy", func() {
	DescribeTable("ValidateNetworkPolicy",
		func(networkPolicy *networking.NetworkPolicy, match types.GomegaMatcher) {
			errList := ValidateNetworkPolicy(networkPolicy)
			Expect(errList).To(match)
		},
		Entry("missing name",
			&networking.NetworkPolicy{},
			ContainElement(RequiredField("metadata.name")),
		),
		Entry("missing namespace",
			&networking.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Name: "foo"}},
			ContainElement(RequiredField("metadata.namespace")),
		),
		Entry("bad name",
			&networking.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Name: "foo*"}},
			ContainElement(InvalidField("metadata.name")),
		),
		Entry("no network ref",
			&networking.NetworkPolicy{},
			ContainElement(RequiredField("spec.networkRef")),
		),
		Entry("invalid network ref name",
			&networking.NetworkPolicy{
				Spec: networking.NetworkPolicySpec{
					NetworkRef: corev1.LocalObjectReference{Name: "foo*"},
				},
			},
			ContainElement(InvalidField("spec.networkRef.name")),
		),
		Entry("invalid ingress port",
			&networking.NetworkPolicy{
				Spec: networking.NetworkPolicySpec{
					Ingress: []networking.NetworkPolicyIngressRule{
						{
							Ports: []networking.NetworkPolicyPort{
								{Port: -10},
							},
						},
					},
				},
			},
			ContainElement(InvalidField("spec.ingress[0].ports[0].port")),
		),
		Entry("not supported ingress peer object selector",
			&networking.NetworkPolicy{
				Spec: networking.NetworkPolicySpec{
					Ingress: []networking.NetworkPolicyIngressRule{
						{
							From: []networking.NetworkPolicyPeer{
								{
									ObjectSelector: &core.ObjectSelector{
										Kind: "Invalid",
									},
								},
							},
						},
					},
				},
			},
			ContainElement(NotSupportedField("spec.ingress[0].from[0].objectSelector.kind")),
		),
		Entry("multiple network policy ingress peer sources in a peer",
			&networking.NetworkPolicy{
				Spec: networking.NetworkPolicySpec{
					Ingress: []networking.NetworkPolicyIngressRule{
						{
							From: []networking.NetworkPolicyPeer{
								{
									ObjectSelector: &core.ObjectSelector{
										Kind: "LoadBalancer",
									},
									IPBlock: &networking.IPBlock{
										CIDR: commonv1alpha1.MustParseIPPrefix("10.0.0.0/16"),
									},
								},
							},
						},
					},
				},
			},
			ContainElement(ForbiddenField("spec.ingress[0].from[0].ipBlock")),
		),
		Entry("ip block except not contained in cidr",
			&networking.NetworkPolicy{
				Spec: networking.NetworkPolicySpec{
					Ingress: []networking.NetworkPolicyIngressRule{
						{
							From: []networking.NetworkPolicyPeer{
								{
									IPBlock: &networking.IPBlock{
										CIDR: commonv1alpha1.MustParseIPPrefix("10.0.0.0/16"),
										Except: []commonv1alpha1.IPPrefix{
											commonv1alpha1.MustParseIPPrefix("10.1.0.0/16"),
										},
									},
								},
							},
						},
					},
				},
			},
			ContainElement(ForbiddenField("spec.ingress[0].from[0].ipBlock.except[0]")),
		),
	)

	DescribeTable("ValidateNetworkPolicyUpdate",
		func(newNetworkPolicy, oldNetworkPolicy *networking.NetworkPolicy, match types.GomegaMatcher) {
			errList := ValidateNetworkPolicyUpdate(newNetworkPolicy, oldNetworkPolicy)
			Expect(errList).To(match)
		},
		Entry("immutable networkRef",
			&networking.NetworkPolicy{
				Spec: networking.NetworkPolicySpec{
					NetworkRef: corev1.LocalObjectReference{Name: "foo"},
				},
			},
			&networking.NetworkPolicy{
				Spec: networking.NetworkPolicySpec{
					NetworkRef: corev1.LocalObjectReference{Name: "bar"},
				},
			},
			ContainElement(ForbiddenField("spec.networkRef")),
		),
	)
})
