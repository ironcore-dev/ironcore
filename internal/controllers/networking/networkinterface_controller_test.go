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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

var _ = Describe("NetworkInterfaceReconciler", func() {
	ctx := SetupContext()
	ns, _ := SetupTest()
	const networkHandle = "foo"

	It("should reconcile the ips and update the status", func() {
		By("creating a network")
		network := &networkingv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "network-",
			},
			Spec: networkingv1alpha1.NetworkSpec{
				Handle: networkHandle,
			},
		}
		Expect(k8sClient.Create(ctx, network)).To(Succeed())

		By("patching the network to be ready")
		baseNetwork := network.DeepCopy()
		network.Status.State = networkingv1alpha1.NetworkStateAvailable
		Expect(k8sClient.Status().Patch(ctx, network, client.MergeFrom(baseNetwork))).To(Succeed())

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
		Eventually(func(g Gomega) networkingv1alpha1.NetworkInterfaceStatus {
			Expect(k8sClient.Get(ctx, nicKey, nic)).To(Succeed())
			return nic.Status
		}).Should(SatisfyAll(
			HaveField("State", networkingv1alpha1.NetworkInterfaceStateAvailable),
			HaveField("NetworkHandle", networkHandle),
			HaveField("IPs", []commonv1alpha1.IP{commonv1alpha1.MustParseIP("10.0.0.1")}),
		))
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
			Spec: networkingv1alpha1.NetworkSpec{
				Handle: networkHandle,
			},
		}
		Expect(k8sClient.Create(ctx, network)).To(Succeed())

		By("patching the network to be ready")
		baseNetwork := network.DeepCopy()
		network.Status.State = networkingv1alpha1.NetworkStateAvailable
		Expect(k8sClient.Status().Patch(ctx, network, client.MergeFrom(baseNetwork))).To(Succeed())

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
		prefixKey := client.ObjectKey{Namespace: ns.Name, Name: networkingv1alpha1.NetworkInterfaceIPSourceEphemeralPrefixName(nic.Name, 0)}
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
			g.Expect(prefix.Status.Phase).To(Equal(ipamv1alpha1.PrefixPhaseAllocated))
		}).Should(Succeed())

		By("waiting for the network interface to be available and report the correct ips")
		Eventually(Object(nic)).Should(HaveField("Status", SatisfyAll(
			HaveField("State", networkingv1alpha1.NetworkInterfaceStateAvailable),
			HaveField("IPs", []commonv1alpha1.IP{commonv1alpha1.MustParseIP("10.0.0.1")}),
		)))
	})

	It("should create and manage ephemeral virtual ips", func() {
		By("creating a network")
		network := &networkingv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "network-",
			},
			Spec: networkingv1alpha1.NetworkSpec{
				Handle: networkHandle,
			},
		}
		Expect(k8sClient.Create(ctx, network)).To(Succeed())

		By("patching the network to be ready")
		baseNetwork := network.DeepCopy()
		network.Status.State = networkingv1alpha1.NetworkStateAvailable
		Expect(k8sClient.Status().Patch(ctx, network, client.MergeFrom(baseNetwork))).To(Succeed())

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
		Eventually(Object(nic)).Should(HaveField("Status", SatisfyAll(
			HaveField("State", networkingv1alpha1.NetworkInterfaceStateAvailable),
			HaveField("IPs", []commonv1alpha1.IP{commonv1alpha1.MustParseIP("192.168.178.1")}),
			HaveField("VirtualIP", commonv1alpha1.MustParseNewIP("10.0.0.1")),
		)))
	})
})
