// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("DetachVolume", func() {
	ns, srv := SetupTest()
	machineClass := SetupMachineClass()

	It("should correctly detach a volume", func(ctx SpecContext) {
		By("creating a machine with a volume")
		createMachineRes, err := srv.CreateMachine(ctx, &iri.CreateMachineRequest{
			Machine: &iri.Machine{
				Spec: &iri.MachineSpec{
					Power: iri.Power_POWER_ON,
					Class: machineClass.Name,
					Volumes: []*iri.Volume{{
						Name:   "root",
						Device: "oda",
						LocalDisk: &iri.LocalDisk{
							Image: &iri.ImageSpec{
								Image: "example.org/foo:latest",
							},
						},
					},
						{
							Name:   "my-volume",
							Device: "odb",
							Connection: &iri.VolumeConnection{
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
		Expect(srv.DetachVolume(ctx, &iri.DetachVolumeRequest{
			MachineId: machineID,
			Name:      "my-volume",
		})).Error().NotTo(HaveOccurred())

		By("getting the ironcore machine")
		ironcoreMachine := &computev1alpha1.Machine{}
		ironcoreMachineKey := client.ObjectKey{Namespace: ns.Name, Name: machineID}
		Expect(k8sClient.Get(ctx, ironcoreMachineKey, ironcoreMachine)).To(Succeed())

		By("inspecting the ironcore machine's volumes")
		Expect(ironcoreMachine.Spec.Volumes).To(ContainElement(computev1alpha1.Volume{
			Name:   "root",
			Device: ptr.To("oda"),
			VolumeSource: computev1alpha1.VolumeSource{
				LocalDisk: &computev1alpha1.LocalDiskVolumeSource{
					Image: "example.org/foo:latest",
				},
			},
		}))

		By("listing for any ironcore volume in the namespace")
		volumeList := &storagev1alpha1.VolumeList{}
		Expect(k8sClient.List(ctx, volumeList, client.InNamespace(ns.Name))).To(Succeed())

		By("asserting the list to be empty")
		Expect(volumeList.Items).To(BeEmpty())

		By("listing for any ironcore secret in the namespace")
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
