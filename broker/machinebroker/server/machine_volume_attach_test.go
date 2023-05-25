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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("AttachVolume", func() {
	ns, srv := SetupTest()
	machineClass := SetupMachineClass()

	It("should correctly attach a volume", func(ctx SpecContext) {
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

		By("attaching a volume")
		Expect(srv.AttachVolume(ctx, &ori.AttachVolumeRequest{
			MachineId: machineID,
			Volume: &ori.Volume{
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
		})).Error().ShouldNot(HaveOccurred())

		By("getting the onmetal machine")
		onmetalMachine := &computev1alpha1.Machine{}
		onmetalMachineKey := client.ObjectKey{Namespace: ns.Name, Name: machineID}
		Expect(k8sClient.Get(ctx, onmetalMachineKey, onmetalMachine)).To(Succeed())

		By("inspecting the onmetal machine's volumes")
		Expect(onmetalMachine.Spec.Volumes).To(ConsistOf(MatchAllFields(Fields{
			"Name":   Equal("my-volume"),
			"Device": PointTo(Equal("oda")),
			"VolumeSource": MatchFields(IgnoreExtras, Fields{
				"VolumeRef": PointTo(MatchAllFields(Fields{
					"Name": Not(BeEmpty()),
				})),
			}),
		})))

		By("getting the corresponding onmetal volume")
		volume := &storagev1alpha1.Volume{}
		volumeName := onmetalMachine.Spec.Volumes[0].VolumeRef.Name
		volumeKey := client.ObjectKey{Namespace: ns.Name, Name: volumeName}
		Expect(k8sClient.Get(ctx, volumeKey, volume)).To(Succeed())

		By("inspecting the onmetal volume")
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

		By("fetching the corresponding onmetal volume access secret")
		secret := &corev1.Secret{}
		secretName := volume.Status.Access.SecretRef.Name
		secretKey := client.ObjectKey{Namespace: ns.Name, Name: secretName}
		Expect(k8sClient.Get(ctx, secretKey, secret)).To(Succeed())

		By("inspecting the onmetal volume access secret")
		Expect(metav1.IsControlledBy(secret, volume)).To(BeTrue(), "secret should be controlled by volume")
		Expect(secret.Type).To(Equal(storagev1alpha1.SecretTypeVolumeAuth))
		Expect(secret.Data).To(Equal(map[string][]byte{"key": []byte("supersecret")}))
	})
})
