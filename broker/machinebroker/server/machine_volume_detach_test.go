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
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("DetachVolume", func() {
	ns, srv := SetupTest()
	machineClass := SetupMachineClass()

	It("should correctly detach a volume", func(ctx SpecContext) {
		By("creating a machine with a volume")
		createMachineRes, err := srv.CreateMachine(ctx, &ori.CreateMachineRequest{
			Machine: &ori.Machine{
				Spec: &ori.MachineSpec{
					Power: ori.Power_POWER_ON,
					Image: &ori.ImageSpec{
						Image: "example.org/foo:latest",
					},
					Class: machineClass.Name,
					Volumes: []*ori.Volume{
						{
							Name:   "my-volume",
							Device: "oda",
							Connection: &ori.VolumeConnection{
								Driver: "ceph",
								Handle: "mycephvolume",
								Attributes: map[string]string{
									"foo": "bar",
								},
								SecretData: map[string][]byte{
									"key": []byte("supersecret"),
								},
							},
						},
					},
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())
		machineID := createMachineRes.Machine.Metadata.Id

		By("detaching the volume")
		Expect(srv.DetachVolume(ctx, &ori.DetachVolumeRequest{
			MachineId: machineID,
			Name:      "my-volume",
		})).Error().NotTo(HaveOccurred())

		By("getting the onmetal machine")
		onmetalMachine := &computev1alpha1.Machine{}
		onmetalMachineKey := client.ObjectKey{Namespace: ns.Name, Name: machineID}
		Expect(k8sClient.Get(ctx, onmetalMachineKey, onmetalMachine)).To(Succeed())

		By("inspecting the onmetal machine's volumes")
		Expect(onmetalMachine.Spec.Volumes).To(BeEmpty())

		By("listing for any onmetal volume in the namespace")
		volumeList := &storagev1alpha1.VolumeList{}
		Expect(k8sClient.List(ctx, volumeList, client.InNamespace(ns.Name))).To(Succeed())

		By("asserting the list to be empty")
		Expect(volumeList.Items).To(BeEmpty())

		By("listing for any onmetal secret in the namespace")
		secretList := &corev1.SecretList{}
		Expect(k8sClient.List(ctx, secretList,
			client.InNamespace(ns.Name),
			client.MatchingFields{"type": string(storagev1alpha1.SecretTypeVolumeAuth)},
		)).To(Succeed())

		By("asserting the list contains a single secret with an owner reference to a volume")
		Expect(secretList.Items).To(ConsistOf(
			HaveField("ObjectMeta.OwnerReferences", ConsistOf(MatchFields(IgnoreExtras, Fields{
				"APIVersion": Equal(storagev1alpha1.SchemeGroupVersion.String()),
				"Kind":       Equal("Volume"),
			}))),
		))
	})
})
