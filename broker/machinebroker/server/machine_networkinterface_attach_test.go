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
	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("AttachNetworkInterface", func() {
	ns, srv := SetupTest()
	machineClass := SetupMachineClass()

	It("should correctly attach a network interface", func(ctx SpecContext) {
		By("creating a machine")
		createMachineRes, err := srv.CreateMachine(ctx, &ori.CreateMachineRequest{
			Machine: &ori.Machine{
				Spec: &ori.MachineSpec{
					Power: ori.Power_POWER_ON,
					Image: &ori.ImageSpec{
						Image: "example.org/foo:latest",
					},
					Class: machineClass.Name,
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())
		machineID := createMachineRes.Machine.Metadata.Id

		By("attaching a network interface")
		Expect(srv.AttachNetworkInterface(ctx, &ori.AttachNetworkInterfaceRequest{
			MachineId: machineID,
			NetworkInterface: &ori.NetworkInterface{
				Name:      "my-nic",
				NetworkId: "network-id",
				Ips:       []string{"10.0.0.1"},
			},
		})).Error().NotTo(HaveOccurred())

		By("getting the onmetal machine")
		onmetalMachine := &computev1alpha1.Machine{}
		onmetalMachineKey := client.ObjectKey{Namespace: ns.Name, Name: machineID}
		Expect(k8sClient.Get(ctx, onmetalMachineKey, onmetalMachine)).To(Succeed())

		By("inspecting the onmetal machine's network interfaces")
		Expect(onmetalMachine.Spec.NetworkInterfaces).To(ConsistOf(MatchAllFields(Fields{
			"Name": Equal("my-nic"),
			"NetworkInterfaceSource": MatchFields(IgnoreExtras, Fields{
				"NetworkInterfaceRef": PointTo(MatchAllFields(Fields{
					"Name": Not(BeEmpty()),
				})),
			}),
		})))

		By("getting the corresponding onmetal network interface")
		nic := &networkingv1alpha1.NetworkInterface{}
		nicName := onmetalMachine.Spec.NetworkInterfaces[0].NetworkInterfaceRef.Name
		nicKey := client.ObjectKey{Namespace: ns.Name, Name: nicName}
		Expect(k8sClient.Get(ctx, nicKey, nic)).To(Succeed())

		By("inspecting the onmetal network interface")
		Expect(nic.Spec.IPs).To(Equal([]networkingv1alpha1.IPSource{
			{Value: commonv1alpha1.MustParseNewIP("10.0.0.1")},
		}))

		By("getting the referenced onmetal network")
		network := &networkingv1alpha1.Network{}
		networkKey := client.ObjectKey{Namespace: ns.Name, Name: nic.Spec.NetworkRef.Name}
		Expect(k8sClient.Get(ctx, networkKey, network)).To(Succeed())

		By("inspecting the onmetal network")
		Expect(network.Spec).To(Equal(networkingv1alpha1.NetworkSpec{
			Handle: "network-id",
		}))
		Expect(network.Status).To(Equal(networkingv1alpha1.NetworkStatus{
			State: networkingv1alpha1.NetworkStateAvailable,
		}))
	})
})
