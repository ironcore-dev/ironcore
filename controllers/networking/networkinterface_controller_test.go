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
	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	ipamv1alpha1 "github.com/onmetal/onmetal-api/apis/ipam/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/apis/networking/v1alpha1"
	"github.com/onmetal/onmetal-api/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("NetworkInterfaceReconciler", func() {
	ctx := testutils.SetupContext()
	ns := SetupTest(ctx)

	It("should reconcile the ips and update the status", func() {
		By("creating a network")
		network := &networkingv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "network-",
			},
		}
		Expect(k8sClient.Create(ctx, network)).To(Succeed())

		By("creating a network interface")
		nic := &networkingv1alpha1.NetworkInterface{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "nic-",
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

		By("waiting for the network interface to report the correct ips")
		nicKey := client.ObjectKeyFromObject(nic)
		Eventually(func(g Gomega) []commonv1alpha1.IP {
			Expect(k8sClient.Get(ctx, nicKey, nic)).To(Succeed())
			return nic.Status.IPs
		}).Should(Equal([]commonv1alpha1.IP{commonv1alpha1.MustParseIP("10.0.0.1")}))
	})

	It("should create ephemeral prefixes and report their IPs once allocated", func() {
		By("creating a root prefix")
		rootPrefix := &ipamv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "nic-",
			},
			Spec: ipamv1alpha1.PrefixSpec{
				IPFamily: corev1.IPv4Protocol,
				Prefix:   commonv1alpha1.MustParseNewIPPrefix("10.0.0.0/24"),
			},
		}
		Expect(k8sClient.Create(ctx, rootPrefix)).To(Succeed())

		By("creating a network")
		network := &networkingv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "network-",
			},
		}
		Expect(k8sClient.Create(ctx, network)).To(Succeed())

		By("creating a network interface")
		nic := &networkingv1alpha1.NetworkInterface{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "nic-",
			},
			Spec: networkingv1alpha1.NetworkInterfaceSpec{
				NetworkRef: corev1.LocalObjectReference{Name: network.Name},
				IPFamilies: []corev1.IPFamily{
					corev1.IPv4Protocol,
				},
				IPs: []networkingv1alpha1.IPSource{
					{
						Ephemeral: &networkingv1alpha1.EphemeralPrefixSource{
							PrefixTemplate: &ipamv1alpha1.PrefixTemplateSpec{
								Spec: ipamv1alpha1.PrefixSpec{
									IPFamily:  corev1.IPv4Protocol,
									ParentRef: &corev1.LocalObjectReference{Name: rootPrefix.Name},
									Prefix:    commonv1alpha1.MustParseNewIPPrefix("10.0.0.1/32"),
								},
							},
						},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, nic)).To(Succeed())

		By("waiting for the prefix to be created with the correct ips and become ready")
		prefixKey := client.ObjectKey{Namespace: ns.Name, Name: NetworkInterfaceEphemeralIPName(nic.Name, 0)}
		Eventually(func(g Gomega) {
			prefix := &ipamv1alpha1.Prefix{}
			err := k8sClient.Get(ctx, prefixKey, prefix)
			Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(metav1.IsControlledBy(prefix, nic)).To(BeTrue(), "ephemeral prefix is not controlled by network interface: %#v", prefix)
			g.Expect(prefix.Spec).To(Equal(ipamv1alpha1.PrefixSpec{
				IPFamily:  corev1.IPv4Protocol,
				ParentRef: &corev1.LocalObjectReference{Name: rootPrefix.Name},
				Prefix:    commonv1alpha1.MustParseNewIPPrefix("10.0.0.1/32"),
			}))
			g.Expect(ipamv1alpha1.GetPrefixReadiness(prefix)).To(Equal(ipamv1alpha1.ReadinessSucceeded))
		}).Should(Succeed())

		By("waiting for the network interface to report the correct ips")
		nicKey := client.ObjectKeyFromObject(nic)
		Eventually(func(g Gomega) []commonv1alpha1.IP {
			Expect(k8sClient.Get(ctx, nicKey, nic)).To(Succeed())
			return nic.Status.IPs
		}).Should(Equal([]commonv1alpha1.IP{commonv1alpha1.MustParseIP("10.0.0.1")}))
	})

	It("should create and manage ephemeral virtual ips", func() {
		By("creating a network")
		network := &networkingv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "network-",
			},
		}
		Expect(k8sClient.Create(ctx, network)).To(Succeed())

		By("creating a network interface")
		nic := &networkingv1alpha1.NetworkInterface{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "nic-",
			},
			Spec: networkingv1alpha1.NetworkInterfaceSpec{
				NetworkRef: corev1.LocalObjectReference{Name: network.Name},
				IPFamilies: []corev1.IPFamily{
					corev1.IPv4Protocol,
				},
				IPs: []networkingv1alpha1.IPSource{
					{Value: commonv1alpha1.MustParseNewIP("192.168.178.1")},
				},
				VirtualIP: &networkingv1alpha1.VirtualIPSource{
					Ephemeral: &networkingv1alpha1.EphemeralVirtualIPSource{
						VirtualIPTemplate: &networkingv1alpha1.VirtualIPTemplateSpec{
							Spec: networkingv1alpha1.VirtualIPSpec{
								Type:     networkingv1alpha1.VirtualIPTypePublic,
								IPFamily: corev1.IPv4Protocol,
							},
						},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, nic)).To(Succeed())

		By("waiting for the virtual ip to exist and be bound")
		virtualIP := &networkingv1alpha1.VirtualIP{}
		virtualIPKey := client.ObjectKeyFromObject(nic)
		Eventually(func(g Gomega) {
			err := k8sClient.Get(ctx, virtualIPKey, virtualIP)
			Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(metav1.IsControlledBy(virtualIP, nic)).To(BeTrue(), "virtual ip is not owned by nic: %#v", virtualIP)
			g.Expect(virtualIP.Spec).To(Equal(networkingv1alpha1.VirtualIPSpec{
				Type:      networkingv1alpha1.VirtualIPTypePublic,
				IPFamily:  corev1.IPv4Protocol,
				TargetRef: &commonv1alpha1.LocalUIDReference{Name: virtualIP.Name, UID: nic.UID},
			}))
			g.Expect(virtualIP.Status.Phase).To(Equal(networkingv1alpha1.VirtualIPPhaseBound))
		}).Should(Succeed())

		By("updating the virtual ip ip")
		baseVirtualIP := virtualIP.DeepCopy()
		virtualIP.Status.IP = commonv1alpha1.MustParseNewIP("10.0.0.1")
		Expect(k8sClient.Status().Patch(ctx, virtualIP, client.MergeFrom(baseVirtualIP))).To(Succeed())

		By("waiting for the virtual ip to be reported in the network interface status")
		nicKey := client.ObjectKeyFromObject(nic)
		Eventually(func() *commonv1alpha1.IP {
			Expect(k8sClient.Get(ctx, nicKey, nic)).To(Succeed())
			return nic.Status.VirtualIP
		}).Should(Equal(commonv1alpha1.MustParseNewIP("10.0.0.1")))
	})
})
