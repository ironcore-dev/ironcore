// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	volumebrokerv1alpha1 "github.com/ironcore-dev/ironcore/broker/volumebroker/api/v1alpha1"
	irimeta "github.com/ironcore-dev/ironcore/iri/apis/meta/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("CreateVolumeSnapshot", func() {
	ns, srv := SetupTest()
	volumeClass := SetupVolumeClass()

	var (
		volume *storagev1alpha1.Volume
	)

	BeforeEach(func() {

		volume = &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns.Name,
				Name:      "test-volume",
				Labels: map[string]string{
					volumebrokerv1alpha1.ManagerLabel: volumebrokerv1alpha1.VolumeBrokerManager,
					volumebrokerv1alpha1.CreatedLabel: "true",
				},
			},
			Spec: storagev1alpha1.VolumeSpec{
				VolumeClassRef: &corev1.LocalObjectReference{
					Name: volumeClass.Name,
				},
				Resources: corev1alpha1.ResourceList{
					corev1alpha1.ResourceStorage: resource.MustParse("1Gi"),
				},
			},
			Status: storagev1alpha1.VolumeStatus{
				State:    storagev1alpha1.VolumeStateAvailable,
				VolumeID: "test-volume",
			},
		}
		Expect(k8sClient.Create(context.Background(), volume)).To(Succeed())
	})

	AfterEach(func() {
		Expect(k8sClient.Delete(context.Background(), volume)).To(Succeed())
	})

	It("should create a volume snapshot", func(ctx SpecContext) {
		By("creating a volume snapshot")
		req := &iri.CreateVolumeSnapshotRequest{
			VolumeSnapshot: &iri.VolumeSnapshot{
				Metadata: &irimeta.ObjectMetadata{
					Labels: map[string]string{
						"test-label": "test-value",
					},
				},
				Spec: &iri.VolumeSnapshotSpec{
					VolumeId: volume.Status.VolumeID,
				},
			},
		}

		res, err := srv.CreateVolumeSnapshot(ctx, req)
		Expect(err).NotTo(HaveOccurred())
		Expect(res).NotTo(BeNil())
		Expect(res.VolumeSnapshot).NotTo(BeNil())
		Expect(res.VolumeSnapshot.Spec.VolumeId).To(Equal(volume.Status.VolumeID))

		By("getting the ironcore volume snapshot")
		var volumeSnapshotList storagev1alpha1.VolumeSnapshotList
		Expect(k8sClient.List(ctx, &volumeSnapshotList, client.InNamespace(ns.Name))).To(Succeed())
		Expect(volumeSnapshotList.Items).To(HaveLen(1))

		volumeSnapshot := volumeSnapshotList.Items[0]
		Expect(volumeSnapshot.Spec.VolumeRef.Name).To(Equal(volume.Name))
		Expect(volumeSnapshot.Labels).To(HaveKeyWithValue(volumebrokerv1alpha1.ManagerLabel, volumebrokerv1alpha1.VolumeBrokerManager))
		Expect(volumeSnapshot.Labels).To(HaveKeyWithValue(volumebrokerv1alpha1.CreatedLabel, "true"))
	})

	It("should return error if volume ID is empty", func(ctx SpecContext) {
		By("creating a volume snapshot with empty volume ID")
		req := &iri.CreateVolumeSnapshotRequest{
			VolumeSnapshot: &iri.VolumeSnapshot{
				Metadata: &irimeta.ObjectMetadata{},
				Spec: &iri.VolumeSnapshotSpec{
					VolumeId: "",
				},
			},
		}

		res, err := srv.CreateVolumeSnapshot(ctx, req)
		Expect(err).To(HaveOccurred())
		Expect(res).To(BeNil())
		Expect(status.Code(err)).To(Equal(codes.InvalidArgument))
	})

	It("should return error if volume is not found", func(ctx SpecContext) {
		By("creating a volume snapshot with non-existent volume ID")
		req := &iri.CreateVolumeSnapshotRequest{
			VolumeSnapshot: &iri.VolumeSnapshot{
				Metadata: &irimeta.ObjectMetadata{},
				Spec: &iri.VolumeSnapshotSpec{
					VolumeId: "non-existent-volume-id",
				},
			},
		}

		res, err := srv.CreateVolumeSnapshot(ctx, req)
		Expect(err).To(HaveOccurred())
		Expect(res).To(BeNil())
		Expect(status.Code(err)).To(Equal(codes.NotFound))
	})

})
