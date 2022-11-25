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
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	"github.com/onmetal/onmetal-api/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("NatGatewayReconciler", func() {
	ctx := testutils.SetupContext()
	ns := SetupTest(ctx)

	It("should reconcile the prefix and routing destinations", func() {
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
				Type:                     "",
				IPFamilies:               nil,
				IPs:                      nil,
				NetworkRef:               corev1.LocalObjectReference{},
				NetworkInterfaceSelector: nil,
				PortsPerNetworkInterface: nil,
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

			g.Expect(natGatewayRouting.Destinations).To(Equal([]commonv1alpha1.LocalUIDReference{
				{Name: nic.Name, UID: nic.UID},
			}))
		}).Should(Succeed())
	})
})
