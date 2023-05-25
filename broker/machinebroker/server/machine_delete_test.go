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
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	orimeta "github.com/onmetal/onmetal-api/ori/apis/meta/v1alpha1"
	machinepoolletv1alpha1 "github.com/onmetal/onmetal-api/poollet/machinepoollet/api/v1alpha1"
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

		By("deleting the machine")
		machineID := res.Machine.Metadata.Id
		Expect(srv.DeleteMachine(ctx, &ori.DeleteMachineRequest{
			MachineId: machineID,
		})).Error().NotTo(HaveOccurred())

		By("listing for onmetal machines in the namespace")
		onmetalMachineList := &computev1alpha1.MachineList{}
		Expect(k8sClient.List(ctx, onmetalMachineList, client.InNamespace(ns.Name))).To(Succeed())

		By("asserting there are no onmetal machines in the returned list")
		Expect(onmetalMachineList.Items).To(BeEmpty())

		By("listing for onmetal network interfaces in the namespace")
		onmetalNicList := &networkingv1alpha1.NetworkInterfaceList{}
		Expect(k8sClient.List(ctx, onmetalNicList, client.InNamespace(ns.Name))).To(Succeed())

		By("asserting there is a single onmetal network interface with an owner reference to the onmetal machine")
		Expect(onmetalNicList.Items).To(ConsistOf(
			HaveField("ObjectMeta.OwnerReferences", ConsistOf(MatchFields(IgnoreExtras, Fields{
				"APIVersion": Equal(computev1alpha1.SchemeGroupVersion.String()),
				"Kind":       Equal("Machine"),
				"Name":       Equal(machineID),
			}))),
		))

		By("listing for onmetal volumes in the namespace")
		onmetalVolumeList := &storagev1alpha1.VolumeList{}
		Expect(k8sClient.List(ctx, onmetalVolumeList, client.InNamespace(ns.Name))).To(Succeed())

		By("asserting there is a single onmetal volume with an owner reference to the onmetal machine")
		Expect(onmetalVolumeList.Items).To(ConsistOf(
			HaveField("ObjectMeta.OwnerReferences", ConsistOf(MatchFields(IgnoreExtras, Fields{
				"APIVersion": Equal(computev1alpha1.SchemeGroupVersion.String()),
				"Kind":       Equal("Machine"),
				"Name":       Equal(machineID),
			}))),
		))
	})
})
