// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	"context"

	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

var _ = Describe("UpdateVolume", func() {
	ns, srv := SetupTest()
	machineClass := SetupMachineClass()

	It("should update volume size and connection information", func(ctx context.Context) {
		By("creating a machine with the volume")
		createMachineRes, err := srv.CreateMachine(ctx, &iri.CreateMachineRequest{
			Machine: &iri.Machine{
				Spec: &iri.MachineSpec{
					Power: iri.Power_POWER_ON,
					Image: &iri.ImageSpec{
						Image: "example.org/foo:latest",
					},
					Class: machineClass.Name,
					Volumes: []*iri.Volume{
						{
							Name:   "primary",
							Device: "oda",
							Connection: &iri.VolumeConnection{
								Driver:                "test",
								Handle:                "testhandle",
								EffectiveStorageBytes: resource.NewQuantity(1*1024*1024*1024, resource.BinarySI).Value(),
							},
						},
					},
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())
		machineID := createMachineRes.Machine.Metadata.Id

		By("getting the ironcore machine")
		ironcoreMachine := &computev1alpha1.Machine{}
		ironcoreMachineKey := client.ObjectKey{Namespace: ns.Name, Name: machineID}
		Expect(k8sClient.Get(ctx, ironcoreMachineKey, ironcoreMachine)).To(Succeed())

		By("getting the corresponding ironcore volume")
		volume := &storagev1alpha1.Volume{}
		volumeName := ironcoreMachine.Spec.Volumes[0].VolumeRef.Name
		volumeKey := client.ObjectKey{Namespace: ns.Name, Name: volumeName}
		Expect(k8sClient.Get(ctx, volumeKey, volume)).To(Succeed())

		By("calling UpdateVolume with new size")
		_, err = srv.UpdateVolume(ctx, &iri.UpdateVolumeRequest{
			MachineId: machineID,
			Volume: &iri.Volume{
				Name:   "primary",
				Device: "oda",
				Connection: &iri.VolumeConnection{
					Driver:                "test",
					Handle:                "testhandle",
					EffectiveStorageBytes: resource.NewQuantity(2*1024*1024*1024, resource.BinarySI).Value(),
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		By("verifying the ironcore volume status is updated with new size")
		Eventually(Object(volume)).Should(SatisfyAll(
			HaveField("Status.Resources", HaveKeyWithValue(
				corev1alpha1.ResourceStorage,
				Equal(resource.MustParse("2Gi")),
			)),
			HaveField("Status.Access", Equal(&storagev1alpha1.VolumeAccess{
				Driver: "test",
				Handle: "testhandle",
			})),
		))

		By("verifying machine volume is updated")
		Eventually(Object(ironcoreMachine)).Should(HaveField("Spec.Volumes", ConsistOf(MatchFields(IgnoreExtras, Fields{
			"Name":   Equal("primary"),
			"Device": Equal(ptr.To("oda")),
			"VolumeSource": Equal(computev1alpha1.VolumeSource{
				VolumeRef: &corev1.LocalObjectReference{Name: volume.Name},
			}),
		}))))
	})
})
