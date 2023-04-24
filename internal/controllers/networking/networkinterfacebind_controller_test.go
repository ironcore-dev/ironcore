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
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	. "github.com/onmetal/onmetal-api/utils/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("NetworkInterfaceBindReconciler", func() {
	ctx := SetupContext()
	ns, machineClass := SetupTest()

	It("should reconcile the binding phase", func() {
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

		By("waiting for the phase to be unbound")
		nicKey := client.ObjectKeyFromObject(nic)
		Eventually(func() networkingv1alpha1.NetworkInterfacePhase {
			Expect(k8sClient.Get(ctx, nicKey, nic)).To(Succeed())
			return nic.Status.Phase
		}).Should(Equal(networkingv1alpha1.NetworkInterfacePhaseUnbound))

		By("creating a machine")
		machine := &computev1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "machine-",
			},
			Spec: computev1alpha1.MachineSpec{
				MachineClassRef: corev1.LocalObjectReference{Name: machineClass.Name},
				Image:           "my-image:latest",
				NetworkInterfaces: []computev1alpha1.NetworkInterface{
					{
						Name: "interface",
						NetworkInterfaceSource: computev1alpha1.NetworkInterfaceSource{
							NetworkInterfaceRef: &corev1.LocalObjectReference{
								Name: nic.Name,
							},
						},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, machine)).To(Succeed())

		By("waiting for the assignment and the phase to be bound")
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, nicKey, nic)).To(Succeed())

			g.Expect(nic.Spec.MachineRef).To(Equal(&commonv1alpha1.LocalUIDReference{Name: machine.Name, UID: machine.UID}))
			g.Expect(nic.Status.Phase).To(Equal(networkingv1alpha1.NetworkInterfacePhaseBound))
		}).Should(Succeed())

		By("deleting the machine")
		Expect(k8sClient.Delete(ctx, machine)).To(Succeed())

		By("waiting for the assignment and the phase to be unbound")
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, nicKey, nic)).To(Succeed())

			g.Expect(nic.Spec.MachineRef).To(BeNil())
			g.Expect(nic.Status.Phase).To(Equal(networkingv1alpha1.NetworkInterfacePhaseUnbound))
		}).Should(Succeed())
	})
})
