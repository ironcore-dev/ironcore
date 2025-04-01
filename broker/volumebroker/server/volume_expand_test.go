// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	irimeta "github.com/ironcore-dev/ironcore/iri/apis/meta/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
	volumepoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/volumepoollet/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("ExpandVolume", func() {
	ns, srv := SetupTest()

	It("should correctly Expand a volume", func(ctx SpecContext) {
		By("creating a volume class with resize policy expandonly")
		volumeClass := &storagev1alpha1.VolumeClass{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "volume-class-",
			},
			Capabilities: corev1alpha1.ResourceList{
				corev1alpha1.ResourceIOPS: resource.MustParse("250Mi"),
				corev1alpha1.ResourceTPS:  resource.MustParse("1500"),
			},
			ResizePolicy: storagev1alpha1.ResizePolicyExpandOnly,
		}
		Expect(k8sClient.Create(ctx, volumeClass)).To(Succeed(), "failed to create test volume class")
		DeferCleanup(k8sClient.Delete, volumeClass)

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

		By("checking the volume storage size")
		ironcoreVolume := &storagev1alpha1.Volume{}
		ironcoreVolumeKey := client.ObjectKey{Namespace: ns.Name, Name: createRes.Volume.Metadata.Id}
		Expect(k8sClient.Get(ctx, ironcoreVolumeKey, ironcoreVolume)).To(Succeed())

		Expect(ironcoreVolume.Spec.Resources.Storage().String()).To(Equal("100"))

		By("expanding the volume size")
		expandRes, err := srv.ExpandVolume(ctx, &iri.ExpandVolumeRequest{
			VolumeId: createRes.Volume.Metadata.Id,
			Resources: &iri.VolumeResources{
				StorageBytes: 200,
			},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(expandRes).NotTo(BeNil())

		By("verifying the volume storage is expanded")
		Expect(k8sClient.Get(ctx, ironcoreVolumeKey, ironcoreVolume)).To(Succeed())
		Expect(ironcoreVolume.Spec.Resources.Storage().String()).To(Equal("200"))

	})
})
