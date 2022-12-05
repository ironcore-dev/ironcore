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
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	machinebrokerv1alpha1 "github.com/onmetal/onmetal-api/broker/machinebroker/api/v1alpha1"
	"github.com/onmetal/onmetal-api/broker/machinebroker/apiutils"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	orimeta "github.com/onmetal/onmetal-api/ori/apis/meta/v1alpha1"
	"github.com/onmetal/onmetal-api/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("CreateVolume", func() {
	ctx := testutils.SetupContext()
	ns, srv := SetupTest(ctx)

	It("Should correctly create a volume", func() {
		By("creating a volume")
		var (
			annotations = map[string]string{
				"annotation-key": "annotation-value",
			}
			labels = map[string]string{
				"label-key": "label-value",
			}
			attributes = map[string]string{
				"attribute1": "value1",
			}
			secretData = map[string][]byte{
				"secret": []byte("dta"),
			}
		)
		const (
			driver = "my-driver"
			handle = "my-handle"
		)

		res, err := srv.CreateVolume(ctx, &ori.CreateVolumeRequest{
			Volume: &ori.Volume{
				Metadata: &orimeta.ObjectMetadata{
					Annotations: annotations,
					Labels:      labels,
				},
				Spec: &ori.VolumeSpec{
					Driver:     driver,
					Handle:     handle,
					Attributes: attributes,
					SecretData: secretData,
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		By("inspecting the resulting volume's metadata")
		Expect(res.Volume.Metadata.Id).NotTo(BeEmpty())
		Expect(res.Volume.Metadata.Labels).To(Equal(labels))
		Expect(res.Volume.Metadata.Annotations).To(Equal(annotations))
		Expect(res.Volume.Metadata.CreatedAt).NotTo(BeZero())
		Expect(res.Volume.Metadata.DeletedAt).To(BeZero())

		By("inspecting the resulting volume's spec")
		Expect(res.Volume.Spec.Driver).To(Equal(driver))
		Expect(res.Volume.Spec.Handle).To(Equal(handle))
		Expect(res.Volume.Spec.Attributes).To(Equal(attributes))
		Expect(res.Volume.Spec.SecretData).To(Equal(secretData))

		By("getting the onmetal volume")
		onmetalVolume := &storagev1alpha1.Volume{}
		onmetalVolumeKey := client.ObjectKey{Namespace: ns.Name, Name: res.Volume.Metadata.Id}
		Expect(k8sClient.Get(ctx, onmetalVolumeKey, onmetalVolume)).To(Succeed())

		By("inspecting the onmetal volume metadata")
		Expect(apiutils.GetObjectMetadata(onmetalVolume)).To(Equal(res.Volume.Metadata))
		Expect(apiutils.IsCreated(onmetalVolume)).To(BeTrue(), "onmetal volume should be marked as created")
		Expect(apiutils.IsManagedBy(onmetalVolume, machinebrokerv1alpha1.MachineBrokerManager)).To(BeTrue(), "onmetal volume should be managed by machine broker")

		By("inspecting the onmetal volume spec")
		Expect(onmetalVolume.Spec).To(Equal(storagev1alpha1.VolumeSpec{}))

		By("inspecting the onmetal volume status")
		Expect(onmetalVolume.Status.State).To(Equal(storagev1alpha1.VolumeStateAvailable))
		Expect(onmetalVolume.Status.Access).NotTo(BeNil())
		Expect(onmetalVolume.Status.Access.Driver).To(Equal(driver))
		Expect(onmetalVolume.Status.Access.Handle).To(Equal(handle))
		Expect(onmetalVolume.Status.Access.VolumeAttributes).To(Equal(attributes))
		Expect(onmetalVolume.Status.Access.SecretRef).NotTo(BeNil())

		By("getting the onmetal volume secret")
		secret := &corev1.Secret{}
		secretKey := client.ObjectKey{Namespace: ns.Name, Name: onmetalVolume.Status.Access.SecretRef.Name}
		Expect(k8sClient.Get(ctx, secretKey, secret)).To(Succeed())

		By("inspecting the secret metadata")
		Expect(apiutils.GetPurpose(secret)).To(Equal(machinebrokerv1alpha1.VolumeAccessPurpose))
		Expect(metav1.IsControlledBy(secret, onmetalVolume)).To(BeTrue(), "secret should be controlled by volume")

		By("inspecting the secret data")
		Expect(secret.Data).To(Equal(secretData))
	})
})
