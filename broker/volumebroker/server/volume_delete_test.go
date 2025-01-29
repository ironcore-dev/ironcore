// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	irimeta "github.com/ironcore-dev/ironcore/iri/apis/meta/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
	volumepoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/volumepoollet/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("DeleteVolume", func() {
	ns, srv := SetupTest()
	volumeClass := SetupVolumeClass()

	It("should correctly delete a volume", func(ctx SpecContext) {
		By("creating a volume")
		createRes, err := srv.CreateVolume(ctx, &iri.CreateVolumeRequest{
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
		Expect(createRes).NotTo(BeNil())

		By("deleting the volume")
		deleteRes, err := srv.DeleteVolume(ctx, &iri.DeleteVolumeRequest{
			VolumeId: createRes.Volume.Metadata.Id,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(deleteRes).NotTo(BeNil())

		By("verifying the volume is deleted")
		ironcoreVolume := &storagev1alpha1.Volume{}
		ironcoreVolumeKey := client.ObjectKey{Namespace: ns.Name, Name: createRes.Volume.Metadata.Id}
		err = k8sClient.Get(ctx, ironcoreVolumeKey, ironcoreVolume)
		Expect(apierrors.IsNotFound(err)).To(BeTrue())
	})
})
