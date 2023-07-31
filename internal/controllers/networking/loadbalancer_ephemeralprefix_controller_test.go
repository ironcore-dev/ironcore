/*
 * Copyright (c) 2021 by the OnMetal authors.
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
package networking

import (
	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	ipamv1alpha1 "github.com/onmetal/onmetal-api/api/ipam/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	. "github.com/onmetal/onmetal-api/utils/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

var _ = Describe("LoadBalancerEphemeralPrefix", func() {
	ns := SetupNamespace(&k8sClient)

	It("should manage ephemeral IP prefixes for a load balancer", func(ctx SpecContext) {
		By("creating a load balancer that requires a prefix")
		loadBalancer := &networkingv1alpha1.LoadBalancer{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "load-balancer-",
			},
			Spec: networkingv1alpha1.LoadBalancerSpec{
				Type:       networkingv1alpha1.LoadBalancerTypeInternal,
				NetworkRef: corev1.LocalObjectReference{Name: "my-network"},
				IPs: []networkingv1alpha1.IPSource{
					{
						Ephemeral: &networkingv1alpha1.EphemeralPrefixSource{
							PrefixTemplate: &ipamv1alpha1.PrefixTemplateSpec{
								Spec: ipamv1alpha1.PrefixSpec{
									IPFamily: corev1.IPv4Protocol,
									Prefix:   commonv1alpha1.MustParseNewIPPrefix("10.0.0.1/32"),
								},
							},
						},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, loadBalancer)).To(Succeed())

		By("waiting for the prefix to exist")
		prefix := &ipamv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns.Name,
				Name:      networkingv1alpha1.LoadBalancerIPIPAMPrefixName(loadBalancer.Name, 0),
			},
		}
		Eventually(Object(prefix)).Should(SatisfyAll(
			BeControlledBy(loadBalancer),
			HaveField("Spec", ipamv1alpha1.PrefixSpec{
				IPFamily: corev1.IPv4Protocol,
				Prefix:   commonv1alpha1.MustParseNewIPPrefix("10.0.0.1/32"),
			}),
		))
	})

	It("should delete undesired prefixes for a load balancer", MustPassRepeatedly(10), func(ctx SpecContext) {
		By("creating a load balancer")
		loadBalancer := &networkingv1alpha1.LoadBalancer{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "load-balancer-",
			},
			Spec: networkingv1alpha1.LoadBalancerSpec{
				Type:       networkingv1alpha1.LoadBalancerTypeInternal,
				NetworkRef: corev1.LocalObjectReference{Name: "my-network"},
				IPs:        []networkingv1alpha1.IPSource{{Value: commonv1alpha1.MustParseNewIP("10.0.0.1")}},
			},
		}
		Expect(k8sClient.Create(ctx, loadBalancer)).To(Succeed())

		By("creating an undesired prefix")
		prefix := &ipamv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "undesired-prefix-",
			},
			Spec: ipamv1alpha1.PrefixSpec{
				IPFamily: corev1.IPv4Protocol,
				Prefix:   commonv1alpha1.MustParseNewIPPrefix("10.0.0.1/32"),
			},
		}
		Expect(ctrl.SetControllerReference(loadBalancer, prefix, k8sClient.Scheme())).To(Succeed())
		Expect(k8sClient.Create(ctx, prefix)).To(Succeed())

		By("waiting for the prefix to be marked for deletion")
		Eventually(Get(prefix)).Should(Satisfy(apierrors.IsNotFound))
	})
})
