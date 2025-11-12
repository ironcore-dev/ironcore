// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	irimeta "github.com/ironcore-dev/ironcore/iri/apis/meta/v1alpha1"
	machinepoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/machinepoollet/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ListMachines", func() {
	_, srv := SetupTest()
	machineClass := SetupMachineClass()

	It("should correctly list machines", func(ctx SpecContext) {
		By("creating multiple machines")
		const noOfMachines = 4

		machines := make([]any, noOfMachines)
		for i := 0; i < noOfMachines; i++ {
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

						Volumes: []*iri.Volume{{
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
			machines[i] = res.Machine
		}

		By("listing the machines")
		Expect(srv.ListMachines(ctx, &iri.ListMachinesRequest{})).To(HaveField("Machines", ConsistOf(machines...)))
	})
})
