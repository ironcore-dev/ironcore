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

package server_test

import (
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
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
		createMachineRes, err := srv.CreateMachine(ctx, &ori.CreateMachineRequest{
			Machine: &ori.Machine{
				Spec: &ori.MachineSpec{
					Power: ori.Power_POWER_ON,
					Image: &ori.ImageSpec{
						Image: "example.org/foo:latest",
					},
					Class: machineClass.Name,
					NetworkInterfaces: []*ori.NetworkInterface{
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
		Expect(srv.DetachNetworkInterface(ctx, &ori.DetachNetworkInterfaceRequest{
			MachineId: machineID,
			Name:      "my-nic",
		})).Error().NotTo(HaveOccurred())

		By("getting the onmetal machine")
		onmetalMachine := &computev1alpha1.Machine{}
		onmetalMachineKey := client.ObjectKey{Namespace: ns.Name, Name: machineID}
		Expect(k8sClient.Get(ctx, onmetalMachineKey, onmetalMachine)).To(Succeed())

		By("inspecting the onmetal machine's network interfaces")
		Expect(onmetalMachine.Spec.NetworkInterfaces).To(BeEmpty())

		By("listing for any onmetal network interfaces in the namespace")
		nicList := &networkingv1alpha1.NetworkInterfaceList{}
		Expect(k8sClient.List(ctx, nicList, client.InNamespace(ns.Name))).To(Succeed())

		By("asserting the list to be empty")
		Expect(nicList.Items).To(BeEmpty())

		By("listing all onmetal networks in the namespace")
		networkList := &networkingv1alpha1.NetworkList{}
		Expect(k8sClient.List(ctx, networkList, client.InNamespace(ns.Name))).To(Succeed())

		By("asserting the list contains a single onmetal network with an owner reference to a network interface")
		Expect(networkList.Items).To(ConsistOf(
			HaveField("ObjectMeta.OwnerReferences", ConsistOf(MatchFields(IgnoreExtras, Fields{
				"APIVersion": Equal(networkingv1alpha1.SchemeGroupVersion.String()),
				"Kind":       Equal("NetworkInterface"),
			}))),
		))
	})
})
