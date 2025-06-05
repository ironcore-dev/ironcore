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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

var _ = Describe("AttachVolume", func() {
	ns, srv := SetupTest()
	machineClass := SetupMachineClass()

	It("should correctly attach a volume", func(ctx SpecContext) {
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

		By("attaching a volume")
		Expect(srv.AttachVolume(ctx, &iri.AttachVolumeRequest{
			MachineId: machineID,
			Volume: &iri.Volume{
				Name:   "my-volume",
				Device: "oda",
				Connection: &iri.VolumeConnection{
					Driver: "ceph",
					Handle: "mycephvolume",
					Attributes: map[string]string{
						"foo": "bar",
					},
					SecretData: map[string][]byte{
						"key": []byte("supersecret"),
					},
					EffectiveStorageBytes: 2000,
				},
			},
		})).Error().ShouldNot(HaveOccurred())

		By("getting the ironcore machine")
		ironcoreMachine := &computev1alpha1.Machine{}
		ironcoreMachineKey := client.ObjectKey{Namespace: ns.Name, Name: machineID}
		Expect(k8sClient.Get(ctx, ironcoreMachineKey, ironcoreMachine)).To(Succeed())

		By("inspecting the ironcore machine's volumes")
		Expect(ironcoreMachine.Spec.Volumes).To(ConsistOf(MatchAllFields(Fields{
			"Name":   Equal("my-volume"),
			"Device": PointTo(Equal("oda")),
			"VolumeSource": MatchFields(IgnoreExtras, Fields{
				"VolumeRef": PointTo(MatchAllFields(Fields{
					"Name": Not(BeEmpty()),
				})),
			}),
		})))

		By("getting the corresponding ironcore volume")
		volume := &storagev1alpha1.Volume{}
		volumeName := ironcoreMachine.Spec.Volumes[0].VolumeRef.Name
		volumeKey := client.ObjectKey{Namespace: ns.Name, Name: volumeName}
		Expect(k8sClient.Get(ctx, volumeKey, volume)).To(Succeed())

		By("inspecting the ironcore volume")
		Expect(volume.Status.Access).To(PointTo(MatchAllFields(Fields{
			"SecretRef": PointTo(MatchAllFields(Fields{
				"Name": Not(BeEmpty()),
			})),
			"Driver": Equal("ceph"),
			"Handle": Equal("mycephvolume"),
			"VolumeAttributes": Equal(map[string]string{
				"foo": "bar",
			}),
		})))

		By("inspecting the effective storage resource is set for volume")
		Expect(volume.Status.Resources.Storage().String()).Should(Equal("2k"))

		By("fetching the corresponding ironcore volume access secret")
		secret := &corev1.Secret{}
		secretName := volume.Status.Access.SecretRef.Name
		secretKey := client.ObjectKey{Namespace: ns.Name, Name: secretName}
		Expect(k8sClient.Get(ctx, secretKey, secret)).To(Succeed())

		By("inspecting the ironcore volume access secret")
		Expect(metav1.IsControlledBy(secret, volume)).To(BeTrue(), "secret should be controlled by volume")
		Expect(secret.Type).To(Equal(storagev1alpha1.SecretTypeVolumeAuth))
		Expect(secret.Data).To(Equal(map[string][]byte{"key": []byte("supersecret")}))
	})

	It("should correctly attach an encrypted volume", func(ctx SpecContext) {
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

		By("attaching a volume")
		Expect(srv.AttachVolume(ctx, &iri.AttachVolumeRequest{
			MachineId: machineID,
			Volume: &iri.Volume{
				Name:   "my-volume",
				Device: "oda",
				Connection: &iri.VolumeConnection{
					Driver: "ceph",
					Handle: "mycephvolume",
					Attributes: map[string]string{
						"foo": "bar",
					},
					SecretData: map[string][]byte{
						"key": []byte("supersecret"),
					},
					EncryptionData: map[string][]byte{
						"encryption": []byte("supersecret2"),
					},
				},
			},
		})).Error().ShouldNot(HaveOccurred())

		By("getting the ironcore machine")
		ironcoreMachine := &computev1alpha1.Machine{}
		ironcoreMachineKey := client.ObjectKey{Namespace: ns.Name, Name: machineID}
		Expect(k8sClient.Get(ctx, ironcoreMachineKey, ironcoreMachine)).To(Succeed())

		By("inspecting the ironcore machine's volumes")
		Expect(ironcoreMachine.Spec.Volumes).To(ConsistOf(MatchAllFields(Fields{
			"Name":   Equal("my-volume"),
			"Device": PointTo(Equal("oda")),
			"VolumeSource": MatchFields(IgnoreExtras, Fields{
				"VolumeRef": PointTo(MatchAllFields(Fields{
					"Name": Not(BeEmpty()),
				})),
			}),
		})))

		By("getting the corresponding ironcore volume")
		volume := &storagev1alpha1.Volume{}
		volumeName := ironcoreMachine.Spec.Volumes[0].VolumeRef.Name
		volumeKey := client.ObjectKey{Namespace: ns.Name, Name: volumeName}
		Expect(k8sClient.Get(ctx, volumeKey, volume)).To(Succeed())

		By("inspecting the ironcore volume")
		Expect(volume).To(SatisfyAll(
			HaveField("Spec.Encryption.SecretRef.Name", Not(BeEmpty())),
			HaveField("Status.Access.SecretRef.Name", Not(BeEmpty())),
			HaveField("Status.Access.Driver", Equal("ceph")),
			HaveField("Status.Access.Handle", Equal("mycephvolume")),
			HaveField("Status.Access.VolumeAttributes", Equal(map[string]string{
				"foo": "bar",
			})),
		))

		By("fetching the corresponding ironcore volume encryption secret")
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      volume.Spec.Encryption.SecretRef.Name,
				Namespace: ns.Name,
			},
		}
		Expect(Object(secret)()).To(SatisfyAll(
			HaveField("Type", Equal(storagev1alpha1.SecretTypeVolumeEncryption)),
			HaveField("Data", Equal(map[string][]byte{"encryption": []byte("supersecret2")})),
			Satisfy(func(o *corev1.Secret) bool { return metav1.IsControlledBy(o, volume) }),
		))
	})
})
