// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
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
		createMachineRes, err := srv.CreateMachine(ctx, &iri.CreateMachineRequest{
			Machine: &iri.Machine{
				Spec: &iri.MachineSpec{
					Power: iri.Power_POWER_ON,
					Image: &iri.ImageSpec{
						Image: "example.org/foo:latest",
					},
					Class: machineClass.Name,
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())
		machineID := createMachineRes.Machine.Metadata.Id

		By("attaching a network interface")
		Expect(srv.AttachNetworkInterface(ctx, &iri.AttachNetworkInterfaceRequest{
			MachineId: machineID,
			NetworkInterface: &iri.NetworkInterface{
				Name:      "my-nic",
				NetworkId: "network-id",
				Ips:       []string{"10.0.0.1"},
			},
		})).Error().NotTo(HaveOccurred())

		By("getting the ironcore machine")
		ironcoreMachine := &computev1alpha1.Machine{}
		ironcoreMachineKey := client.ObjectKey{Namespace: ns.Name, Name: machineID}
		Expect(k8sClient.Get(ctx, ironcoreMachineKey, ironcoreMachine)).To(Succeed())

		By("inspecting the ironcore machine's network interfaces")
		Expect(ironcoreMachine.Spec.NetworkInterfaces).To(ConsistOf(MatchAllFields(Fields{
			"Name": Equal("my-nic"),
			"NetworkInterfaceSource": MatchFields(IgnoreExtras, Fields{
				"NetworkInterfaceRef": PointTo(MatchAllFields(Fields{
					"Name": Not(BeEmpty()),
				})),
			}),
		})))

		By("getting the corresponding ironcore network interface")
		nic := &networkingv1alpha1.NetworkInterface{}
		nicName := ironcoreMachine.Spec.NetworkInterfaces[0].NetworkInterfaceRef.Name
		nicKey := client.ObjectKey{Namespace: ns.Name, Name: nicName}
		Expect(k8sClient.Get(ctx, nicKey, nic)).To(Succeed())

		By("inspecting the ironcore network interface")
		Expect(nic.Spec.IPs).To(Equal([]networkingv1alpha1.IPSource{
			{Value: commonv1alpha1.MustParseNewIP("10.0.0.1")},
		}))

		By("getting the referenced ironcore network")
		network := &networkingv1alpha1.Network{}
		networkKey := client.ObjectKey{Namespace: ns.Name, Name: nic.Spec.NetworkRef.Name}
		Expect(k8sClient.Get(ctx, networkKey, network)).To(Succeed())

		By("inspecting the ironcore network")
		Expect(network.Spec).To(Equal(networkingv1alpha1.NetworkSpec{
			ProviderID: "network-id",
		}))
		Expect(network.Status).To(Equal(networkingv1alpha1.NetworkStatus{
			State: networkingv1alpha1.NetworkStateAvailable,
		}))
	})

	It("should correctly re-create a network in case it has been removed", func(ctx SpecContext) {
		By("creating a machine")
		createMachineRes, err := srv.CreateMachine(ctx, &iri.CreateMachineRequest{
			Machine: &iri.Machine{
				Spec: &iri.MachineSpec{
					Power: iri.Power_POWER_ON,
					Image: &iri.ImageSpec{
						Image: "example.org/foo:latest",
					},
					Class: machineClass.Name,
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())
		machineID := createMachineRes.Machine.Metadata.Id

		By("attaching a network interface")
		Expect(srv.AttachNetworkInterface(ctx, &iri.AttachNetworkInterfaceRequest{
			MachineId: machineID,
			NetworkInterface: &iri.NetworkInterface{
				Name:      "my-nic",
				NetworkId: "network-id",
				Ips:       []string{"10.0.0.1"},
			},
		})).Error().NotTo(HaveOccurred())

		By("getting the ironcore machine")
		ironcoreMachine := &computev1alpha1.Machine{}
		ironcoreMachineKey := client.ObjectKey{Namespace: ns.Name, Name: machineID}
		Expect(k8sClient.Get(ctx, ironcoreMachineKey, ironcoreMachine)).To(Succeed())

		By("inspecting the ironcore machine's network interfaces")
		Expect(ironcoreMachine.Spec.NetworkInterfaces).To(ConsistOf(MatchAllFields(Fields{
			"Name": Equal("my-nic"),
			"NetworkInterfaceSource": MatchFields(IgnoreExtras, Fields{
				"NetworkInterfaceRef": PointTo(MatchAllFields(Fields{
					"Name": Not(BeEmpty()),
				})),
			}),
		})))

		By("getting the corresponding ironcore network interface")
		nic := &networkingv1alpha1.NetworkInterface{}
		nicName := ironcoreMachine.Spec.NetworkInterfaces[0].NetworkInterfaceRef.Name
		nicKey := client.ObjectKey{Namespace: ns.Name, Name: nicName}
		Expect(k8sClient.Get(ctx, nicKey, nic)).To(Succeed())

		By("inspecting the ironcore network interface")
		Expect(nic.Spec.IPs).To(Equal([]networkingv1alpha1.IPSource{
			{Value: commonv1alpha1.MustParseNewIP("10.0.0.1")},
		}))

		By("getting the referenced ironcore network")
		network := &networkingv1alpha1.Network{}
		networkKey := client.ObjectKey{Namespace: ns.Name, Name: nic.Spec.NetworkRef.Name}
		Expect(k8sClient.Get(ctx, networkKey, network)).To(Succeed())

		By("inspecting the ironcore network")
		Expect(network.Spec).To(Equal(networkingv1alpha1.NetworkSpec{
			ProviderID: "network-id",
		}))
		Expect(network.Status).To(Equal(networkingv1alpha1.NetworkStatus{
			State: networkingv1alpha1.NetworkStateAvailable,
		}))

		By("detaching the network interface")
		Expect(srv.DetachNetworkInterface(ctx, &iri.DetachNetworkInterfaceRequest{
			MachineId: machineID,
			Name:      "my-nic",
		})).Error().NotTo(HaveOccurred())

		By("deleting the network")
		Expect(k8sClient.Delete(ctx, network)).To(Succeed())

		By("re-attaching a network interface")
		Expect(srv.AttachNetworkInterface(ctx, &iri.AttachNetworkInterfaceRequest{
			MachineId: machineID,
			NetworkInterface: &iri.NetworkInterface{
				Name:      "my-nic",
				NetworkId: "network-id",
				Ips:       []string{"10.0.0.1"},
			},
		})).Error().NotTo(HaveOccurred())

		By("inspecting the ironcore network again")
		Expect(network.Spec).To(Equal(networkingv1alpha1.NetworkSpec{
			ProviderID: "network-id",
		}))
		Expect(network.Status).To(Equal(networkingv1alpha1.NetworkStatus{
			State: networkingv1alpha1.NetworkStateAvailable,
		}))
	})
})
