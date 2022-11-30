// Copyright 2022 OnMetal authors
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

package networking

import (
	"net/netip"

	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	"github.com/onmetal/onmetal-api/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("NatGatewayReconciler", func() {
	ctx := testutils.SetupContext()
	ns := SetupTest(ctx)

	It("should reconcile the natgateway and routing destinations", func() {
		By("creating a network")
		network := &networkingv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "network-",
			},
		}
		Expect(k8sClient.Create(ctx, network)).To(Succeed())

		By("creating a nat gateway")
		natGateway := &networkingv1alpha1.NATGateway{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "nat-gateway-",
			},
			Spec: networkingv1alpha1.NATGatewaySpec{
				Type: networkingv1alpha1.NATGatewayTypePublic,
				IPFamilies: []corev1.IPFamily{
					corev1.IPv4Protocol,
				},
				IPs: []networkingv1alpha1.NATGatewayIP{
					{
						Name: "ip1",
					},
				},
				NetworkRef: corev1.LocalObjectReference{
					Name: network.Name,
				},
				NetworkInterfaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"foo": "bar"},
				},
				PortsPerNetworkInterface: pointer.Int32(2),
			},
		}
		Expect(k8sClient.Create(ctx, natGateway)).To(Succeed())

		By("waiting for the nat gateway routing to exist with no destinations")
		natGatewayKey := client.ObjectKeyFromObject(natGateway)
		natGatewayRouting := &networkingv1alpha1.NATGatewayRouting{}
		Eventually(func(g Gomega) {
			err := k8sClient.Get(ctx, natGatewayKey, natGatewayRouting)
			Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(metav1.IsControlledBy(natGatewayRouting, natGateway)).To(BeTrue(), "nat gateway routing is not controlled by nat gateway: %#v", natGatewayRouting.OwnerReferences)
			g.Expect(natGatewayRouting.Destinations).To(BeEmpty())
		}).Should(Succeed())

		natGatewayBase := natGateway.DeepCopy()
		ip := netip.IPv4Unspecified()
		for _, v := range natGateway.Spec.IPs {
			natGateway.Status.IPs = append(natGateway.Status.IPs, networkingv1alpha1.NATGatewayIPStatus{
				Name: v.Name,
				IP: commonv1alpha1.IP{
					Addr: ip.Next(),
				},
			})
		}
		Expect(k8sClient.Patch(ctx, natGateway, client.MergeFrom(natGatewayBase))).To(Succeed())

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

		By("waiting for the nat gateway routing to be updated")
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, natGatewayKey, natGatewayRouting)).To(Succeed())

			g.Expect(natGatewayRouting.NetworkRef.Name).To(BeEquivalentTo(network.Name))
			g.Expect(natGatewayRouting.NetworkRef.UID).To(BeEquivalentTo(network.UID))

			g.Expect(natGatewayRouting.Destinations).To(HaveLen(1))

			g.Expect(natGatewayRouting.Destinations[0].Name).To(BeEquivalentTo(nic.Name))
			g.Expect(natGatewayRouting.Destinations[0].UID).To(BeEquivalentTo(nic.UID))
			g.Expect(natGatewayRouting.Destinations[0].IPs).To(HaveLen(1))
			g.Expect(natGatewayRouting.Destinations[0].IPs[0].IP).To(BeEquivalentTo(natGateway.Status.IPs[0].IP))
		}).Should(Succeed())

		By("deleting a network interface")
		Expect(k8sClient.Delete(ctx, nic)).To(Succeed())

		By("waiting for the nat gateway routing to be updated")
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, natGatewayKey, natGatewayRouting)).To(Succeed())

			g.Expect(natGatewayRouting.NetworkRef.Name).To(BeEquivalentTo(network.Name))
			g.Expect(natGatewayRouting.NetworkRef.UID).To(BeEquivalentTo(network.UID))

			g.Expect(natGatewayRouting.Destinations).To(HaveLen(0))
		}).Should(Succeed())

		By("waiting for natgateway status to be updated")
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, natGatewayKey, natGateway)).To(Succeed())
			g.Expect(natGateway.Status.PortsUsed).To(BeEquivalentTo(pointer.Int32(0)))
		}).Should(Succeed())
	})

	It("should reconcile the nattateway and routing destinations, with to little ports", func() {
		portsPerNetworkInterface := int32(1024)
		By("creating a network")
		network := &networkingv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "network-",
			},
		}
		Expect(k8sClient.Create(ctx, network)).To(Succeed())

		By("creating a nat gateway")
		natGateway := &networkingv1alpha1.NATGateway{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "nat-gateway-",
			},
			Spec: networkingv1alpha1.NATGatewaySpec{
				//ToDo
				Type: networkingv1alpha1.NATGatewayTypePublic,
				IPFamilies: []corev1.IPFamily{
					corev1.IPv4Protocol,
				},
				IPs: []networkingv1alpha1.NATGatewayIP{
					{
						Name: "ip1",
					},
					{
						Name: "ip2",
					},
				},
				NetworkRef: corev1.LocalObjectReference{
					Name: network.Name,
				},
				NetworkInterfaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"foo": "bar"},
				},
				PortsPerNetworkInterface: pointer.Int32(portsPerNetworkInterface),
			},
		}
		Expect(k8sClient.Create(ctx, natGateway)).To(Succeed())

		By("waiting for the nat gateway routing to exist with no destinations")
		natGatewayKey := client.ObjectKeyFromObject(natGateway)
		natGatewayRouting := &networkingv1alpha1.NATGatewayRouting{}
		Eventually(func(g Gomega) {
			err := k8sClient.Get(ctx, natGatewayKey, natGatewayRouting)
			Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(metav1.IsControlledBy(natGatewayRouting, natGateway)).To(BeTrue(), "nat gateway routing is not controlled by nat gateway: %#v", natGatewayRouting.OwnerReferences)
			g.Expect(natGatewayRouting.Destinations).To(BeEmpty())
		}).Should(Succeed())

		natGatewayBase := natGateway.DeepCopy()
		ip := netip.IPv4Unspecified()
		for _, v := range natGateway.Spec.IPs {
			ip = ip.Next()
			natGateway.Status.IPs = append(natGateway.Status.IPs, networkingv1alpha1.NATGatewayIPStatus{
				Name: v.Name,
				IP: commonv1alpha1.IP{
					Addr: ip,
				},
			})
		}
		Expect(k8sClient.Patch(ctx, natGateway, client.MergeFrom(natGatewayBase))).To(Succeed())

		By("creating a network interfaces")

		nics := 64
		for i := 0; i < nics; i++ {
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
		}

		By("waiting for the nat gateway routing to be updated")
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, natGatewayKey, natGatewayRouting)).To(Succeed())
			g.Expect(natGatewayRouting.Destinations).To(HaveLen(nics))
		}).Should(Succeed())

		By("creating more network interfaces that cannot be allocated in the nat gateway anymore")
		for i := 0; i < nics; i++ {
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
		}

		totalSlots := int(((MaxEphemeralPort-MinEphemeralPort)/portsPerNetworkInterface)+1) * 2
		By("waiting for the nat gateway routing to be updated")
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, natGatewayKey, natGatewayRouting)).To(Succeed())
			g.Expect(natGatewayRouting.Destinations).To(HaveLen(2 * nics))
			var assigned, unassigned int
			for _, v := range natGatewayRouting.Destinations {
				if len(v.IPs) == 0 {
					unassigned++
				} else {
					assigned++
				}
			}
			g.Expect(assigned).To(BeEquivalentTo(totalSlots))
			g.Expect(unassigned).To(BeEquivalentTo(nics*2 - totalSlots))
		}).Should(Succeed())

		By("waiting for natgateway status to be updated")
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, natGatewayKey, natGateway)).To(Succeed())
			g.Expect(natGateway.Status.PortsUsed).To(BeEquivalentTo(pointer.Int32((MaxEphemeralPort - MinEphemeralPort + 1) * 2)))
		}).Should(Succeed())

	})
})
