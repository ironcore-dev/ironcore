// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/client-go/applyconfigurations/networking/v1alpha1"
	typednetworkingv1alpha1 "github.com/ironcore-dev/ironcore/client-go/ironcore/versioned/typed/networking/v1alpha1"
	gentype "k8s.io/client-go/gentype"
)

// fakeNetworkPolicies implements NetworkPolicyInterface
type fakeNetworkPolicies struct {
	*gentype.FakeClientWithListAndApply[*v1alpha1.NetworkPolicy, *v1alpha1.NetworkPolicyList, *networkingv1alpha1.NetworkPolicyApplyConfiguration]
	Fake *FakeNetworkingV1alpha1
}

func newFakeNetworkPolicies(fake *FakeNetworkingV1alpha1, namespace string) typednetworkingv1alpha1.NetworkPolicyInterface {
	return &fakeNetworkPolicies{
		gentype.NewFakeClientWithListAndApply[*v1alpha1.NetworkPolicy, *v1alpha1.NetworkPolicyList, *networkingv1alpha1.NetworkPolicyApplyConfiguration](
			fake.Fake,
			namespace,
			v1alpha1.SchemeGroupVersion.WithResource("networkpolicies"),
			v1alpha1.SchemeGroupVersion.WithKind("NetworkPolicy"),
			func() *v1alpha1.NetworkPolicy { return &v1alpha1.NetworkPolicy{} },
			func() *v1alpha1.NetworkPolicyList { return &v1alpha1.NetworkPolicyList{} },
			func(dst, src *v1alpha1.NetworkPolicyList) { dst.ListMeta = src.ListMeta },
			func(list *v1alpha1.NetworkPolicyList) []*v1alpha1.NetworkPolicy {
				return gentype.ToPointerSlice(list.Items)
			},
			func(list *v1alpha1.NetworkPolicyList, items []*v1alpha1.NetworkPolicy) {
				list.Items = gentype.FromPointerSlice(items)
			},
		),
		fake,
	}
}
