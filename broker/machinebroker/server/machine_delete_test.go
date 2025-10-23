// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	irimeta "github.com/ironcore-dev/ironcore/iri/apis/meta/v1alpha1"
	machinepoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/machinepoollet/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("DeleteMachine", func() {
	ns, srv := SetupTest()
	machineClass := SetupMachineClass()

	It("should correctly delete a machine", func(ctx SpecContext) {
		By("creating a machine")
		res, err := srv.CreateMachine(ctx, &iri.CreateMachineRequest{
			Machine: &iri.Machine{
				Metadata: &irimeta.ObjectMetadata{
					Labels: map[string]string{
						machinepoolletv1alpha1.MachineUIDLabel: "foobar",
					},
				},
				Spec: &iri.MachineSpec{
					Power: iri.Power_POWER_ON,
					Class: machineClass.Name,

					NetworkInterfaces: []*iri.NetworkInterface{
						{
							Name:      "primary-nic",
							NetworkId: "network-id",
							Ips:       []string{"10.0.0.1"},
						},
					},
					Volumes: []*iri.Volume{
						{
							Name: "root",
							LocalDisk: &iri.LocalDisk{
								Image: &iri.ImageSpec{
									Image: "example.org/foo:latest",
								},
							},
						},
						{
							Name:   "primary-volume",
							Device: "oda",
							Connection: &iri.VolumeConnection{
								Driver:     "ceph",
								Handle:     "ceph-volume",
								Attributes: map[string]string{"foo": "bar"},
								SecretData: map[string][]byte{"super": []byte("secret")},
							},
						},
					},
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(res).NotTo(BeNil())

		By("deleting the machine")
		machineID := res.Machine.Metadata.Id
		Expect(srv.DeleteMachine(ctx, &iri.DeleteMachineRequest{
			MachineId: machineID,
		})).Error().NotTo(HaveOccurred())

		By("listing for ironcore machines in the namespace")
		ironcoreMachineList := &computev1alpha1.MachineList{}
		Expect(k8sClient.List(ctx, ironcoreMachineList, client.InNamespace(ns.Name))).To(Succeed())

		By("asserting there are no ironcore machines in the returned list")
		Expect(ironcoreMachineList.Items).To(BeEmpty())

		By("listing for ironcore network interfaces in the namespace")
		ironcoreNicList := &networkingv1alpha1.NetworkInterfaceList{}
		Expect(k8sClient.List(ctx, ironcoreNicList, client.InNamespace(ns.Name))).To(Succeed())

		By("asserting there is a single ironcore network interface with an owner reference to the ironcore machine")
		Expect(ironcoreNicList.Items).To(ConsistOf(
			HaveField("ObjectMeta.OwnerReferences", ConsistOf(MatchFields(IgnoreExtras, Fields{
				"APIVersion": Equal(computev1alpha1.SchemeGroupVersion.String()),
				"Kind":       Equal("Machine"),
				"Name":       Equal(machineID),
			}))),
		))

		By("listing for ironcore volumes in the namespace")
		ironcoreVolumeList := &storagev1alpha1.VolumeList{}
		Expect(k8sClient.List(ctx, ironcoreVolumeList, client.InNamespace(ns.Name))).To(Succeed())

		By("asserting there is a single ironcore volume with an owner reference to the ironcore machine")
		Expect(ironcoreVolumeList.Items).To(ConsistOf(
			HaveField("ObjectMeta.OwnerReferences", ConsistOf(MatchFields(IgnoreExtras, Fields{
				"APIVersion": Equal(computev1alpha1.SchemeGroupVersion.String()),
				"Kind":       Equal("Machine"),
				"Name":       Equal(machineID),
			}))),
		))
	})
})
