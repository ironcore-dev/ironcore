// Copyright 2023 IronCore authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1_test

import (
	ipamv1alpha1 "github.com/ironcore-dev/ironcore/api/ipam/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	. "github.com/ironcore-dev/ironcore/internal/apis/networking/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Defaults", func() {
	Describe("SetDefaults_NetworkInterfaceSpec", func() {
		It("should default the prefix length of ephemeral ips depending on the ip family", func() {
			spec := &networkingv1alpha1.NetworkInterfaceSpec{
				IPFamilies: []corev1.IPFamily{corev1.IPv4Protocol, corev1.IPv6Protocol},
				IPs: []networkingv1alpha1.IPSource{
					{
						Ephemeral: &networkingv1alpha1.EphemeralPrefixSource{
							PrefixTemplate: &ipamv1alpha1.PrefixTemplateSpec{
								Spec: ipamv1alpha1.PrefixSpec{
									IPFamily:  corev1.IPv4Protocol,
									ParentRef: &corev1.LocalObjectReference{Name: "parent-v4"},
								},
							},
						},
					},
					{
						Ephemeral: &networkingv1alpha1.EphemeralPrefixSource{
							PrefixTemplate: &ipamv1alpha1.PrefixTemplateSpec{
								Spec: ipamv1alpha1.PrefixSpec{
									IPFamily:  corev1.IPv6Protocol,
									ParentRef: &corev1.LocalObjectReference{Name: "parent-v6"},
								},
							},
						},
					},
				},
			}
			SetDefaults_NetworkInterfaceSpec(spec)

			Expect(spec.IPs).To(Equal([]networkingv1alpha1.IPSource{
				{
					Ephemeral: &networkingv1alpha1.EphemeralPrefixSource{
						PrefixTemplate: &ipamv1alpha1.PrefixTemplateSpec{
							Spec: ipamv1alpha1.PrefixSpec{
								IPFamily:     corev1.IPv4Protocol,
								ParentRef:    &corev1.LocalObjectReference{Name: "parent-v4"},
								PrefixLength: 32,
							},
						},
					},
				},
				{
					Ephemeral: &networkingv1alpha1.EphemeralPrefixSource{
						PrefixTemplate: &ipamv1alpha1.PrefixTemplateSpec{
							Spec: ipamv1alpha1.PrefixSpec{
								IPFamily:     corev1.IPv6Protocol,
								ParentRef:    &corev1.LocalObjectReference{Name: "parent-v6"},
								PrefixLength: 128,
							},
						},
					},
				},
			}))
		})
	})
})
