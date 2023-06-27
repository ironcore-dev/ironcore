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
	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	"github.com/onmetal/onmetal-api/utils/generic"
	. "github.com/onmetal/onmetal-api/utils/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

var _ = Describe("NATGatewayReconciler", func() {
	ns := SetupNamespace(&k8sClient)

	It("should reconcile the NAT gateway and routing destinations", func(ctx SpecContext) {
		By("creating a network")
		network := &networkingv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "network-",
			},
		}
		Expect(k8sClient.Create(ctx, network)).To(Succeed())

		By("creating a nat gateway")
		const portsPerNetworkInterface = int32(2)
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
				NetworkRef: corev1.LocalObjectReference{
					Name: network.Name,
				},
				NetworkInterfaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"foo": "bar"},
				},
				PortsPerNetworkInterface: generic.Pointer(portsPerNetworkInterface),
			},
		}
		Expect(k8sClient.Create(ctx, natGateway)).To(Succeed())

		By("waiting for the nat gateway routing to exist with no destinations")
		natGatewayRouting := &networkingv1alpha1.NATGatewayRouting{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns.Name,
				Name:      natGateway.Name,
			},
		}
		Eventually(Object(natGatewayRouting)).Should(SatisfyAll(
			Satisfy(func(ngwr *networkingv1alpha1.NATGatewayRouting) bool {
				return metav1.IsControlledBy(natGatewayRouting, natGateway)
			}),
			HaveField("Destinations", BeEmpty()),
		))

		natGatewayBase := natGateway.DeepCopy()
		natGateway.Status.IPs = []networkingv1alpha1.NATGatewayIPStatus{
			{IP: commonv1alpha1.MustParseIP("192.168.178.1")},
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
		Eventually(Object(natGatewayRouting)).Should(SatisfyAll(
			HaveField("NetworkRef", commonv1alpha1.LocalUIDReference{
				Name: network.Name,
				UID:  network.UID,
			}),
			HaveField("Destinations", ConsistOf(networkingv1alpha1.NATGatewayDestination{
				Name: nic.Name,
				UID:  nic.UID,
				IPs: []networkingv1alpha1.NATGatewayDestinationIP{
					{
						IP:      commonv1alpha1.MustParseIP("192.168.178.1"),
						Port:    MinEphemeralPort,
						EndPort: 1025,
					},
				},
			})),
		))

		By("deleting a network interface")
		Expect(k8sClient.Delete(ctx, nic)).To(Succeed())

		By("waiting for the nat gateway routing to be updated")
		Eventually(Object(natGatewayRouting)).Should(SatisfyAll(
			HaveField("NetworkRef", commonv1alpha1.LocalUIDReference{
				Name: network.Name,
				UID:  network.UID,
			}),
			HaveField("Destinations", BeEmpty()),
		))

		By("waiting for nat gateway status to be updated")
		Eventually(Object(natGateway)).Should(HaveField("Status.PortsUsed", HaveValue(BeEquivalentTo(0))))
	})

	It("should reconcile the nat gateway and routing destinations not enough ports", func(ctx SpecContext) {
		const portsPerNetworkInterface = int32(1024)
		By("creating a network")
		network := &networkingv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "network-",
			},
		}
		Expect(k8sClient.Create(ctx, network)).To(Succeed())

		By("creating a NAT gateway")
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
		natGatewayRouting := &networkingv1alpha1.NATGatewayRouting{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns.Name,
				Name:      natGateway.Name,
			},
		}
		Eventually(Object(natGatewayRouting)).Should(SatisfyAll(
			Satisfy(func(ngwr *networkingv1alpha1.NATGatewayRouting) bool {
				return metav1.IsControlledBy(natGatewayRouting, natGateway)
			}),
			HaveField("Destinations", BeEmpty()),
		))

		natGatewayBase := natGateway.DeepCopy()
		natGateway.Status.IPs = []networkingv1alpha1.NATGatewayIPStatus{
			{IP: commonv1alpha1.MustParseIP("192.168.178.1")},
			{IP: commonv1alpha1.MustParseIP("192.168.178.2")},
		}
		Expect(k8sClient.Patch(ctx, natGateway, client.MergeFrom(natGatewayBase))).To(Succeed())

		By("creating network interfaces to exhaust the nat gateway's IPs")

		const noOfNetworkInterfaces = 64
		for i := 0; i < noOfNetworkInterfaces; i++ {
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
		Eventually(Object(natGatewayRouting)).Should(HaveField("Destinations", HaveLen(noOfNetworkInterfaces)))

		By("checking if ports are in correct range")
		for _, destination := range natGatewayRouting.Destinations {
			Expect(destination.IPs).To(HaveLen(1))
			Expect(destination.IPs[0].Port >= MinEphemeralPort).To(BeTrue())
			Expect(destination.IPs[0].EndPort <= MaxEphemeralPort).To(BeTrue())
		}

		By("creating more network interfaces that cannot be allocated in the nat gateway anymore")
		for i := 0; i < noOfNetworkInterfaces; i++ {
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
		Eventually(Object(natGatewayRouting)).Should(HaveField("Destinations", HaveLen(126)))

		By("waiting for nat gateway status to be updated")
		Eventually(Object(natGateway)).Should(HaveField("Status.PortsUsed", HaveValue(BeEquivalentTo(NoOfEphemeralPorts*2))))
	})

	It("should update the NAT gateway routing when the ips change", func(ctx SpecContext) {
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
				NetworkRef: corev1.LocalObjectReference{
					Name: network.Name,
				},
				NetworkInterfaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"foo": "bar"},
				},
				PortsPerNetworkInterface: generic.Pointer(networkingv1alpha1.DefaultPortsPerNetworkInterface),
			},
		}
		Expect(k8sClient.Create(ctx, natGateway)).To(Succeed())

		By("adding an IP to the NAT gateway")
		natGatewayBase := natGateway.DeepCopy()
		natGateway.Status.IPs = []networkingv1alpha1.NATGatewayIPStatus{
			{IP: commonv1alpha1.MustParseIP("192.168.178.1")},
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

		By("waiting for the nat gateway routing to exist targeting the network interface")
		natGatewayRouting := &networkingv1alpha1.NATGatewayRouting{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns.Name,
				Name:      natGateway.Name,
			},
		}
		Eventually(Object(natGatewayRouting)).Should(SatisfyAll(
			HaveField("NetworkRef", commonv1alpha1.LocalUIDReference{
				Name: network.Name,
				UID:  network.UID,
			}),
			HaveField("Destinations", ConsistOf(networkingv1alpha1.NATGatewayDestination{
				Name: nic.Name,
				UID:  nic.UID,
				IPs: []networkingv1alpha1.NATGatewayDestinationIP{
					{
						IP:      commonv1alpha1.MustParseIP("192.168.178.1"),
						Port:    MinEphemeralPort,
						EndPort: 3071,
					},
				},
			})),
		))

		By("updating the NAT gateway IP")
		natGatewayBase = natGateway.DeepCopy()
		natGateway.Status.IPs = []networkingv1alpha1.NATGatewayIPStatus{
			{IP: commonv1alpha1.MustParseIP("22.22.22.1")},
		}
		Expect(k8sClient.Patch(ctx, natGateway, client.MergeFrom(natGatewayBase))).To(Succeed())

		By("waiting for the NAT gateway routing to be updated")
		Eventually(Object(natGatewayRouting)).Should(SatisfyAll(
			HaveField("NetworkRef", commonv1alpha1.LocalUIDReference{
				Name: network.Name,
				UID:  network.UID,
			}),
			HaveField("Destinations", ConsistOf(networkingv1alpha1.NATGatewayDestination{
				Name: nic.Name,
				UID:  nic.UID,
				IPs: []networkingv1alpha1.NATGatewayDestinationIP{
					{
						IP:      commonv1alpha1.MustParseIP("22.22.22.1"),
						Port:    MinEphemeralPort,
						EndPort: 3071,
					},
				},
			})),
		))
	})
})
