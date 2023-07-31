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
	. "github.com/onmetal/onmetal-api/utils/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"

	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	ipamv1alpha1 "github.com/onmetal/onmetal-api/api/ipam/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
)

var _ = Describe("LoadBalancerReconciler", func() {
	ns, _ := SetupTest()

	It("should reconcile the prefix and routing destinations", func(ctx SpecContext) {
		By("creating a network")
		network := &networkingv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "network-",
			},
		}
		Expect(k8sClient.Create(ctx, network)).To(Succeed())

		By("setting the network to be available")
		Eventually(UpdateStatus(network, func() {
			network.Status.State = networkingv1alpha1.NetworkStateAvailable
		})).Should(Succeed())

		By("creating a load balancer")
		loadBalancer := &networkingv1alpha1.LoadBalancer{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "load-balancer-",
			},
			Spec: networkingv1alpha1.LoadBalancerSpec{
				Type: networkingv1alpha1.LoadBalancerTypePublic,
				IPFamilies: []corev1.IPFamily{
					corev1.IPv4Protocol,
				},
				NetworkRef: corev1.LocalObjectReference{Name: network.Name},
				NetworkInterfaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"foo": "bar"},
				},
			},
		}
		Expect(k8sClient.Create(ctx, loadBalancer)).To(Succeed())

		By("waiting for the load balancer routing to exist with no destinations")
		loadBalancerRouting := &networkingv1alpha1.LoadBalancerRouting{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: loadBalancer.Namespace,
				Name:      loadBalancer.Name,
			},
		}
		Eventually(Object(loadBalancerRouting)).Should(SatisfyAll(
			BeControlledBy(loadBalancer),
			HaveField("Destinations", BeEmpty()),
		))

		By("creating a network interface")
		nic := &networkingv1alpha1.NetworkInterface{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "nic-",
				Labels:       map[string]string{"foo": "bar"},
			},
			Spec: networkingv1alpha1.NetworkInterfaceSpec{
				NetworkRef: corev1.LocalObjectReference{Name: network.Name},
				IPFamilies: []corev1.IPFamily{
					corev1.IPv4Protocol,
				},
				IPs: []networkingv1alpha1.IPSource{
					{
						Value: commonv1alpha1.MustParseNewIP("10.0.0.1"),
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, nic)).To(Succeed())

		By("setting the network interface provider ID")
		Eventually(UpdateStatus(nic, func() {
			nic.Status.ProviderID = "my://provider-id"
		})).Should(Succeed())

		By("waiting for the network interface to be available and report IPs")
		Eventually(Object(nic)).Should(HaveField("Status", SatisfyAll(
			HaveField("State", networkingv1alpha1.NetworkInterfaceStateAvailable),
			HaveField("IPs", []commonv1alpha1.IP{commonv1alpha1.MustParseIP("10.0.0.1")}),
		)))

		By("waiting for the load balancer routing to be updated")
		Eventually(Object(loadBalancerRouting)).Should(SatisfyAll(
			HaveField("NetworkRef", commonv1alpha1.LocalUIDReference{
				Name: network.Name,
				UID:  network.UID,
			}),
			HaveField("Destinations", []networkingv1alpha1.LoadBalancerDestination{
				{
					IP: commonv1alpha1.MustParseIP("10.0.0.1"),
					TargetRef: &networkingv1alpha1.LoadBalancerTargetRef{
						Name:       nic.Name,
						UID:        nic.UID,
						ProviderID: "my://provider-id",
					},
				},
			}),
		))
	})

	It("should allocate internal IPs for internal load balancers", func(ctx SpecContext) {
		By("creating a network")
		network := &networkingv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "network-",
			},
		}
		Expect(k8sClient.Create(ctx, network)).To(Succeed())

		By("creating a prefix")
		rootPrefix := &ipamv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "lb-",
			},
			Spec: ipamv1alpha1.PrefixSpec{
				IPFamily: corev1.IPv4Protocol,
				Prefix:   commonv1alpha1.MustParseNewIPPrefix("10.0.0.0/24"),
			},
		}
		Expect(k8sClient.Create(ctx, rootPrefix)).To(Succeed())

		By("creating an internal load balancer")
		loadBalancer := &networkingv1alpha1.LoadBalancer{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "internal-lb-",
			},
			Spec: networkingv1alpha1.LoadBalancerSpec{
				Type:       networkingv1alpha1.LoadBalancerTypeInternal,
				NetworkRef: corev1.LocalObjectReference{Name: network.Name},
				IPFamilies: []corev1.IPFamily{corev1.IPv4Protocol},
				IPs: []networkingv1alpha1.IPSource{
					{
						Ephemeral: &networkingv1alpha1.EphemeralPrefixSource{
							PrefixTemplate: &ipamv1alpha1.PrefixTemplateSpec{
								Spec: ipamv1alpha1.PrefixSpec{
									IPFamily:  corev1.IPv4Protocol,
									ParentRef: &corev1.LocalObjectReference{Name: rootPrefix.Name},
								},
							},
						},
					},
				},
				NetworkInterfaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"foo": "bar"},
				},
			},
		}
		Expect(k8sClient.Create(ctx, loadBalancer)).To(Succeed())

		By("waiting for the prefix to be created with the correct spec and become ready")
		prefix := &ipamv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: loadBalancer.Namespace,
				Name:      networkingv1alpha1.LoadBalancerIPIPAMPrefixName(loadBalancer.Name, 0),
			},
		}
		Eventually(Object(prefix)).Should(SatisfyAll(
			HaveField("Spec.IPFamily", corev1.IPv4Protocol),
			HaveField("Spec.ParentRef", &corev1.LocalObjectReference{Name: rootPrefix.Name}),
			HaveField("Spec.PrefixLength", int32(32)),
			HaveField("Spec.Prefix", commonv1alpha1.MustParseNewIPPrefix("10.0.0.0/32")),
			HaveField("Status.Phase", ipamv1alpha1.PrefixPhaseAllocated),
		))

		By("asserting it get's an internal IP")
		Eventually(Object(loadBalancer)).Should(HaveField("Status.IPs", ContainElements(*commonv1alpha1.MustParseNewIP("10.0.0.0"))))
	})
})
