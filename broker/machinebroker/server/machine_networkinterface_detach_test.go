// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("DetachNetworkInterface", func() {
	ns, srv := SetupTest()
	machineClass := SetupMachineClass()

	It("should correctly detach a network interface", func(ctx SpecContext) {
		By("creating a machine")
		createMachineRes, err := srv.CreateMachine(ctx, &iri.CreateMachineRequest{
			Machine: &iri.Machine{
				Spec: &iri.MachineSpec{
					Power: iri.Power_POWER_ON,
					Image: &iri.ImageSpec{
						Image: "example.org/foo:latest",
					},
					Class: machineClass.Name,
					NetworkInterfaces: []*iri.NetworkInterface{
						{
							Name:      "my-nic",
							NetworkId: "network-id",
							Ips:       []string{"10.0.0.1"},
						},
					},
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())
		machineID := createMachineRes.Machine.Metadata.Id

		By("detaching the network interface")
		Expect(srv.DetachNetworkInterface(ctx, &iri.DetachNetworkInterfaceRequest{
			MachineId: machineID,
			Name:      "my-nic",
		})).Error().NotTo(HaveOccurred())

		By("getting the ironcore machine")
		ironcoreMachine := &computev1alpha1.Machine{}
		ironcoreMachineKey := client.ObjectKey{Namespace: ns.Name, Name: machineID}
		Expect(k8sClient.Get(ctx, ironcoreMachineKey, ironcoreMachine)).To(Succeed())

		By("inspecting the ironcore machine's network interfaces")
		Expect(ironcoreMachine.Spec.NetworkInterfaces).To(BeEmpty())

		By("listing for any ironcore network interfaces in the namespace")
		nicList := &networkingv1alpha1.NetworkInterfaceList{}
		Expect(k8sClient.List(ctx, nicList, client.InNamespace(ns.Name))).To(Succeed())

		By("asserting the list to be empty")
		Expect(nicList.Items).To(BeEmpty())

		By("listing all ironcore networks in the namespace")
		networkList := &networkingv1alpha1.NetworkList{}
		Expect(k8sClient.List(ctx, networkList, client.InNamespace(ns.Name))).To(Succeed())

		By("asserting the list contains a single ironcore network with an owner reference to a network interface")
		Expect(networkList.Items).To(ConsistOf(
			HaveField("ObjectMeta.OwnerReferences", ConsistOf(MatchFields(IgnoreExtras, Fields{
				"APIVersion": Equal(networkingv1alpha1.SchemeGroupVersion.String()),
				"Kind":       Equal("NetworkInterface"),
			}))),
		))
	})
})
