// Copyright 2022 IronCore authors
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
	ori "github.com/ironcore-dev/ironcore/ori/apis/machine/v1alpha1"
	orimeta "github.com/ironcore-dev/ironcore/ori/apis/meta/v1alpha1"
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
			res, err := srv.CreateMachine(ctx, &ori.CreateMachineRequest{
				Machine: &ori.Machine{
					Metadata: &orimeta.ObjectMetadata{
						Labels: map[string]string{
							machinepoolletv1alpha1.MachineUIDLabel: "foobar",
						},
					},
					Spec: &ori.MachineSpec{
						Power: ori.Power_POWER_ON,
						Image: &ori.ImageSpec{
							Image: "example.org/foo:latest",
						},
						Class: machineClass.Name,
						NetworkInterfaces: []*ori.NetworkInterface{
							{
								Name:      "primary-nic",
								NetworkId: "network-id",
								Ips:       []string{"10.0.0.1"},
							},
						},
						Volumes: []*ori.Volume{
							{
								Name:   "primary-volume",
								Device: "oda",
								Connection: &ori.VolumeConnection{
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
		Expect(srv.ListMachines(ctx, &ori.ListMachinesRequest{})).To(HaveField("Machines", ConsistOf(machines...)))
	})
})
