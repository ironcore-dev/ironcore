// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	"github.com/ironcore-dev/ironcore/broker/machinebroker/apiutils"
	volumebrokerv1alpha1 "github.com/ironcore-dev/ironcore/broker/volumebroker/api/v1alpha1"
	irimeta "github.com/ironcore-dev/ironcore/iri/apis/meta/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
	volumepoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/volumepoollet/api/v1alpha1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("CreateVolume", func() {
	ns, srv := SetupTest()
	volumeClass := SetupVolumeClass()

	It("should correctly create a volume", func(ctx SpecContext) {
		By("creating a volume")
		res, err := srv.CreateVolume(ctx, &iri.CreateVolumeRequest{
			Volume: &iri.Volume{
				Metadata: &irimeta.ObjectMetadata{
					Labels: map[string]string{
						volumepoolletv1alpha1.VolumeUIDLabel: "foobar",
					},
				},
				Spec: &iri.VolumeSpec{
					Class: volumeClass.Name,
					Resources: &iri.VolumeResources{
						StorageBytes: 100,
					},
				},
			},
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(res).NotTo(BeNil())

		By("getting the ironcore volume")
		ironcoreVolume := &storagev1alpha1.Volume{}
		ironcoreVolumeKey := client.ObjectKey{Namespace: ns.Name, Name: res.Volume.Metadata.Id}
		Expect(k8sClient.Get(ctx, ironcoreVolumeKey, ironcoreVolume)).To(Succeed())

		By("inspecting the ironcore volume")
		Expect(ironcoreVolume.Labels).To(Equal(map[string]string{
			volumebrokerv1alpha1.CreatedLabel: "true",
			volumebrokerv1alpha1.ManagerLabel: volumebrokerv1alpha1.VolumeBrokerManager,
		}))
		encodedIRIAnnotations, err := apiutils.EncodeAnnotationsAnnotation(nil)
		Expect(err).NotTo(HaveOccurred())
		encodedIRILabels, err := apiutils.EncodeLabelsAnnotation(map[string]string{
			volumepoolletv1alpha1.VolumeUIDLabel: "foobar",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(ironcoreVolume.Annotations).To(Equal(map[string]string{
			volumebrokerv1alpha1.AnnotationsAnnotation: encodedIRIAnnotations,
			volumebrokerv1alpha1.LabelsAnnotation:      encodedIRILabels,
		}))
		Expect(ironcoreVolume.Spec.VolumeClassRef.Name).To(Equal(volumeClass.Name))
		Expect(ironcoreVolume.Spec.Resources).To(HaveLen(1))
	})
})
