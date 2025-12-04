// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	brokerutils "github.com/ironcore-dev/ironcore/broker/common/utils"
	volumebrokerv1alpha1 "github.com/ironcore-dev/ironcore/broker/volumebroker/api/v1alpha1"
	irimeta "github.com/ironcore-dev/ironcore/iri/apis/meta/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
	poolletutils "github.com/ironcore-dev/ironcore/poollet/common/utils"
	volumepoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/volumepoollet/api/v1alpha1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
					Annotations: map[string]string{
						volumepoolletv1alpha1.IRIVolumeGenerationAnnotation: "1",
						volumepoolletv1alpha1.VolumeGenerationAnnotation:    "1",
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
			poolletutils.DownwardAPILabel(volumepoolletv1alpha1.VolumeDownwardAPIPrefix, "root-volume-uid"): "foobar",
			volumebrokerv1alpha1.CreatedLabel: "true",
			volumebrokerv1alpha1.ManagerLabel: volumebrokerv1alpha1.VolumeBrokerManager,
		}))
		encodedIRIAnnotations, err := brokerutils.EncodeAnnotationsAnnotation(map[string]string{
			volumepoolletv1alpha1.IRIVolumeGenerationAnnotation: "1",
			volumepoolletv1alpha1.VolumeGenerationAnnotation:    "1",
		},
		)
		Expect(err).NotTo(HaveOccurred())
		encodedIRILabels, err := brokerutils.EncodeLabelsAnnotation(map[string]string{
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

	It("should correctly create a volume from snapshot", func(ctx SpecContext) {
		By("creating a volume snapshot")
		volumeSnapshot := &storagev1alpha1.VolumeSnapshot{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns.Name,
				Name:      "test-snapshot",
			},
			Spec: storagev1alpha1.VolumeSnapshotSpec{
				VolumeRef: &corev1.LocalObjectReference{Name: "source-volume"},
			},
			Status: storagev1alpha1.VolumeSnapshotStatus{
				State:      storagev1alpha1.VolumeSnapshotStateReady,
				SnapshotID: "test-snapshot-id",
			},
		}
		Expect(k8sClient.Create(ctx, volumeSnapshot)).To(Succeed())

		By("creating a volume with snapshot data source")
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
					VolumeDataSource: &iri.VolumeDataSource{
						SnapshotDataSource: &iri.SnapshotDataSource{
							SnapshotId: "test-snapshot",
						},
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

		By("verifying the volume has the correct snapshot reference")
		Expect(ironcoreVolume.Spec.VolumeDataSource.VolumeSnapshotRef).NotTo(BeNil())
		Expect(ironcoreVolume.Spec.VolumeDataSource.VolumeSnapshotRef.Name).To(Equal("test-snapshot"))
	})

})
