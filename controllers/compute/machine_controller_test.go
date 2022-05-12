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

package compute

import (
	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/apis/networking/v1alpha1"
	"github.com/onmetal/onmetal-api/controllers/shared"
	"github.com/onmetal/onmetal-api/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("MachineReconciler", func() {
	ctx := testutils.SetupContext()
	ns := SetupTest(ctx)

	It("should manage ephemeral network interfaces", func() {
		By("creating a network")
		network := &networkingv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "network-",
			},
		}
		Expect(k8sClient.Create(ctx, network)).To(Succeed())

		By("creating a machine")
		machine := &computev1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "machine-",
			},
			Spec: computev1alpha1.MachineSpec{
				MachineClassRef: corev1.LocalObjectReference{Name: "machine-class"},
				Image:           "my-image:latest",
				NetworkInterfaces: []computev1alpha1.NetworkInterface{
					{
						Name: "interface",
						NetworkInterfaceSource: computev1alpha1.NetworkInterfaceSource{
							Ephemeral: &computev1alpha1.EphemeralNetworkInterfaceSource{
								NetworkInterfaceTemplate: &networkingv1alpha1.NetworkInterfaceTemplateSpec{
									Spec: networkingv1alpha1.NetworkInterfaceSpec{
										NetworkRef: corev1.LocalObjectReference{Name: network.Name},
										IPs:        []networkingv1alpha1.IPSource{{Value: commonv1alpha1.MustParseNewIP("10.0.0.1")}},
									},
								},
							},
						},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, machine)).To(Succeed())

		By("waiting for the network interface to exist")
		nic := &networkingv1alpha1.NetworkInterface{}
		nicKey := client.ObjectKey{Namespace: ns.Name, Name: shared.MachineEphemeralNetworkInterfaceName(machine.Name, "interface")}
		Eventually(func(g Gomega) {
			err := k8sClient.Get(ctx, nicKey, nic)
			Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(metav1.IsControlledBy(nic, machine)).To(BeTrue(), "network interface is not controlled by machine: %#v", nic.OwnerReferences)
			g.Expect(nic.Spec.MachineRef).To(Equal(&commonv1alpha1.LocalUIDReference{
				Name: machine.Name,
				UID:  machine.UID,
			}))
		}).Should(Succeed())
	})
})
