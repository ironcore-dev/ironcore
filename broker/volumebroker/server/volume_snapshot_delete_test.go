// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	volumebrokerv1alpha1 "github.com/ironcore-dev/ironcore/broker/volumebroker/api/v1alpha1"
	irimeta "github.com/ironcore-dev/ironcore/iri/apis/meta/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
	volumepoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/volumepoollet/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Describe("DeleteVolumeSnapshot", func() {
	ns, srv := SetupTest()
	volumeClass := SetupVolumeClass()

	It("should correctly delete a volume snapshot", func(ctx SpecContext) {
		By("creating a volume")
		createVolumeRes, err := srv.CreateVolume(ctx, &iri.CreateVolumeRequest{
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
		Expect(createVolumeRes).NotTo(BeNil())

		By("creating a volume snapshot")
		volumeSnapshot := &storagev1alpha1.VolumeSnapshot{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns.Name,
				Name:      "test-volume-snapshot",
				Labels: map[string]string{
					volumebrokerv1alpha1.ManagerLabel: volumebrokerv1alpha1.VolumeBrokerManager,
					volumebrokerv1alpha1.CreatedLabel: "true",
				},
			},
			Spec: storagev1alpha1.VolumeSnapshotSpec{
				VolumeRef: &corev1.LocalObjectReference{
					Name: createVolumeRes.Volume.Metadata.Id,
				},
			},
		}
		Expect(k8sClient.Create(ctx, volumeSnapshot)).To(Succeed())

		By("deleting the volume snapshot")
		deleteRes, err := srv.DeleteVolumeSnapshot(ctx, &iri.DeleteVolumeSnapshotRequest{
			VolumeSnapshotId: volumeSnapshot.Name,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(deleteRes).NotTo(BeNil())

		By("verifying the volume snapshot is deleted")
		ironcoreVolumeSnapshot := &storagev1alpha1.VolumeSnapshot{}
		ironcoreVolumeSnapshotKey := client.ObjectKey{Namespace: ns.Name, Name: volumeSnapshot.Name}
		err = k8sClient.Get(ctx, ironcoreVolumeSnapshotKey, ironcoreVolumeSnapshot)
		Expect(apierrors.IsNotFound(err)).To(BeTrue())
	})

	It("should return error if volume snapshot is not found", func(ctx SpecContext) {
		By("attempting to delete a non-existent volume snapshot")
		req := &iri.DeleteVolumeSnapshotRequest{
			VolumeSnapshotId: "non-existent-snapshot",
		}

		res, err := srv.DeleteVolumeSnapshot(ctx, req)
		Expect(err).To(HaveOccurred())
		Expect(res).To(BeNil())
		Expect(status.Code(err)).To(Equal(codes.NotFound))
	})

})
