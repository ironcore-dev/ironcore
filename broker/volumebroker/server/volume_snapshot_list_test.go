// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	volumebrokerv1alpha1 "github.com/ironcore-dev/ironcore/broker/volumebroker/api/v1alpha1"
	"github.com/ironcore-dev/ironcore/broker/volumebroker/apiutils"
	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("ListVolumeSnapshots", func() {
	ns, srv := SetupTest()
	volumeClass := SetupVolumeClass()

	var (
		volume          *storagev1alpha1.Volume
		volumeSnapshot1 *storagev1alpha1.VolumeSnapshot
		volumeSnapshot2 *storagev1alpha1.VolumeSnapshot
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

		volumeSnapshot1 = &storagev1alpha1.VolumeSnapshot{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns.Name,
				Name:      "test-volume-snapshot-1",
				Labels: map[string]string{
					volumebrokerv1alpha1.ManagerLabel: volumebrokerv1alpha1.VolumeBrokerManager,
					volumebrokerv1alpha1.CreatedLabel: "true",
					"test-label":                      "test-value-1",
				},
				Annotations: map[string]string{
					volumebrokerv1alpha1.AnnotationsAnnotation: "{}",
					volumebrokerv1alpha1.LabelsAnnotation:      `{"test-label":"test-value-1"}`,
				},
			},
			Spec: storagev1alpha1.VolumeSnapshotSpec{
				VolumeRef: &corev1.LocalObjectReference{
					Name: volume.Name,
				},
			},
			Status: storagev1alpha1.VolumeSnapshotStatus{
				SnapshotID: "test-snapshot-id-1",
				State:      storagev1alpha1.VolumeSnapshotStateReady,
				Size:       resource.NewQuantity(1024*1024*1024, resource.DecimalSI), // 1Gi
			},
		}
		Expect(k8sClient.Create(context.Background(), volumeSnapshot1)).To(Succeed())
		Expect(apiutils.PatchCreated(context.Background(), k8sClient, volumeSnapshot1)).To(Succeed())

		volumeSnapshot2 = &storagev1alpha1.VolumeSnapshot{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns.Name,
				Name:      "test-volume-snapshot-2",
				Labels: map[string]string{
					volumebrokerv1alpha1.ManagerLabel: volumebrokerv1alpha1.VolumeBrokerManager,
					volumebrokerv1alpha1.CreatedLabel: "true",
					"test-label":                      "test-value-2",
				},
				Annotations: map[string]string{
					volumebrokerv1alpha1.AnnotationsAnnotation: "{}",
					volumebrokerv1alpha1.LabelsAnnotation:      `{"test-label":"test-value-2"}`,
				},
			},
			Spec: storagev1alpha1.VolumeSnapshotSpec{
				VolumeRef: &corev1.LocalObjectReference{
					Name: volume.Name,
				},
			},
			Status: storagev1alpha1.VolumeSnapshotStatus{
				SnapshotID: "test-snapshot-id-2",
				State:      storagev1alpha1.VolumeSnapshotStatePending,
			},
		}
		Expect(k8sClient.Create(context.Background(), volumeSnapshot2)).To(Succeed())
		Expect(apiutils.PatchCreated(context.Background(), k8sClient, volumeSnapshot2)).To(Succeed())
	})

	AfterEach(func() {
		Expect(k8sClient.Delete(context.Background(), volumeSnapshot2)).To(Succeed())
		Expect(k8sClient.Delete(context.Background(), volumeSnapshot1)).To(Succeed())
		Expect(k8sClient.Delete(context.Background(), volume)).To(Succeed())
	})

	It("should list all volume snapshots", func(ctx SpecContext) {
		By("listing all volume snapshots")
		req := &iri.ListVolumeSnapshotsRequest{}

		res, err := srv.ListVolumeSnapshots(ctx, req)
		Expect(err).NotTo(HaveOccurred())
		Expect(res).NotTo(BeNil())
		Expect(res.VolumeSnapshots).To(HaveLen(2))

		By("verifying the first volume snapshot")
		snapshot1 := res.VolumeSnapshots[0]
		Expect(snapshot1.Spec.VolumeId).To(Equal(volume.Status.VolumeID))
		Expect(snapshot1.Status.State).To(Equal(iri.VolumeSnapshotState_VOLUME_SNAPSHOT_READY))
		Expect(snapshot1.Status.Size).To(Equal(int64(1024 * 1024 * 1024)))

		By("verifying the second volume snapshot")
		snapshot2 := res.VolumeSnapshots[1]
		Expect(snapshot2.Spec.VolumeId).To(Equal(volume.Status.VolumeID))
		Expect(snapshot2.Status.State).To(Equal(iri.VolumeSnapshotState_VOLUME_SNAPSHOT_PENDING))
		Expect(snapshot2.Status.Size).To(Equal(int64(0)))
	})

	It("should filter volume snapshots by ID", func(ctx SpecContext) {
		By("filtering volume snapshots by ID")
		req := &iri.ListVolumeSnapshotsRequest{
			Filter: &iri.VolumeSnapshotFilter{
				Id: volumeSnapshot1.Name,
			},
		}

		res, err := srv.ListVolumeSnapshots(ctx, req)
		Expect(err).NotTo(HaveOccurred())
		Expect(res).NotTo(BeNil())
		Expect(res.VolumeSnapshots).To(HaveLen(1))

		snapshot := res.VolumeSnapshots[0]
		Expect(snapshot.Spec.VolumeId).To(Equal(volume.Status.VolumeID))
		Expect(snapshot.Status.State).To(Equal(iri.VolumeSnapshotState_VOLUME_SNAPSHOT_READY))
	})

	It("should filter volume snapshots by label selector", func(ctx SpecContext) {
		By("filtering volume snapshots by label selector")
		req := &iri.ListVolumeSnapshotsRequest{
			Filter: &iri.VolumeSnapshotFilter{
				LabelSelector: map[string]string{
					"test-label": "test-value-1",
				},
			},
		}

		res, err := srv.ListVolumeSnapshots(ctx, req)
		Expect(err).NotTo(HaveOccurred())
		Expect(res).NotTo(BeNil())
		Expect(res.VolumeSnapshots).To(HaveLen(1))

		snapshot := res.VolumeSnapshots[0]
		Expect(snapshot.Spec.VolumeId).To(Equal(volume.Status.VolumeID))
		Expect(snapshot.Status.State).To(Equal(iri.VolumeSnapshotState_VOLUME_SNAPSHOT_READY))
	})

	It("should return empty list for non-existent volume snapshot ID", func(ctx SpecContext) {
		By("filtering by non-existent volume snapshot ID")
		req := &iri.ListVolumeSnapshotsRequest{
			Filter: &iri.VolumeSnapshotFilter{
				Id: "non-existent-snapshot",
			},
		}

		res, err := srv.ListVolumeSnapshots(ctx, req)
		Expect(err).NotTo(HaveOccurred())
		Expect(res).NotTo(BeNil())
		Expect(res.VolumeSnapshots).To(BeEmpty())
	})

	It("should return empty list for non-matching label selector", func(ctx SpecContext) {
		By("filtering by non-matching label selector")
		req := &iri.ListVolumeSnapshotsRequest{
			Filter: &iri.VolumeSnapshotFilter{
				LabelSelector: map[string]string{
					"non-existent-label": "non-existent-value",
				},
			},
		}

		res, err := srv.ListVolumeSnapshots(ctx, req)
		Expect(err).NotTo(HaveOccurred())
		Expect(res).NotTo(BeNil())
		Expect(res.VolumeSnapshots).To(BeEmpty())
	})

})
