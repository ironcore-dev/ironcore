// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package controllers_test

import (
	"context"

	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
	testingvolume "github.com/ironcore-dev/ironcore/iri/testing/volume"
	poolletutils "github.com/ironcore-dev/ironcore/poollet/common/utils"
	volumepoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/volumepoollet/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/proto"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

var _ = Describe("VolumeSnapshotController", func() {
	ns, vp, vc, _, srv := SetupTest()

	It("should create a volume snapshot", func(ctx SpecContext) {
		volumeSize := int64(1024 * 1024 * 1024)

		By("creating a volume")
		volume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{
				VolumeClassRef: &corev1.LocalObjectReference{Name: vc.Name},
				VolumePoolRef:  &corev1.LocalObjectReference{Name: vp.Name},
				Resources: corev1alpha1.ResourceList{
					corev1alpha1.ResourceStorage: resource.MustParse("1Gi"),
				},
			},
		}
		Expect(k8sClient.Create(ctx, volume)).To(Succeed())
		DeferCleanup(expectVolumeDeleted, volume)

		By("waiting for the runtime to report the volume")
		Eventually(srv).Should(HaveField("Volumes", HaveLen(1)))

		By("setting the volume status in the runtime")
		_, iriVolume := GetSingleMapEntry(srv.Volumes)
		iriVolume = &testingvolume.FakeVolume{Volume: proto.Clone(iriVolume.Volume).(*iri.Volume)}
		iriVolume.Status.State = iri.VolumeState_VOLUME_AVAILABLE
		iriVolume.Status.Access = &iri.VolumeAccess{
			Driver: "test-driver",
			Handle: "test-volume-id",
		}
		srv.SetVolumes([]*testingvolume.FakeVolume{iriVolume})

		By("waiting for the volume status to be updated")
		Eventually(Object(volume)).Should(HaveField("Status.State", Equal(storagev1alpha1.VolumeStateAvailable)))

		By("creating a volume snapshot")
		volumeSnapshot := &storagev1alpha1.VolumeSnapshot{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "volume-snapshot-",
			},
			Spec: storagev1alpha1.VolumeSnapshotSpec{
				VolumeRef: &corev1.LocalObjectReference{Name: volume.Name},
			},
		}
		Expect(k8sClient.Create(ctx, volumeSnapshot)).To(Succeed())
		DeferCleanup(expectVolumeSnapshotDeleted, volumeSnapshot)

		By("waiting for the volume snapshot to have a finalizer")
		Eventually(Object(volumeSnapshot)).Should(HaveField("Finalizers", ContainElement(volumepoolletv1alpha1.VolumeSnapshotFinalizer)))

		By("waiting for the runtime to report the volume snapshot")
		Eventually(srv).Should(SatisfyAll(
			HaveField("VolumeSnapshots", HaveLen(1)),
		))

		By("setting the volume snapshot status to ready in the runtime")
		_, iriVolumeSnapshot := GetSingleMapEntry(srv.VolumeSnapshots)
		iriVolumeSnapshot = &testingvolume.FakeVolumeSnapshot{VolumeSnapshot: proto.Clone(iriVolumeSnapshot.VolumeSnapshot).(*iri.VolumeSnapshot)}
		iriVolumeSnapshot.Status.State = iri.VolumeSnapshotState_VOLUME_SNAPSHOT_READY
		iriVolumeSnapshot.Status.Size = volumeSize
		srv.SetVolumeSnapshots([]*testingvolume.FakeVolumeSnapshot{iriVolumeSnapshot})

		By("waiting for the volume snapshot status to be updated")
		expectedSnapshotID := poolletutils.MakeID(testingvolume.FakeRuntimeName, iriVolumeSnapshot.Metadata.Id)
		Eventually(Object(volumeSnapshot)).Should(SatisfyAll(
			HaveField("Status.SnapshotID", Equal(expectedSnapshotID.String())),
			HaveField("Status.State", Equal(storagev1alpha1.VolumeSnapshotStateReady)),
			HaveField("Status.Size", Not(BeNil())),
			HaveField("Status.Size.Value()", Equal(volumeSize)),
		))
	})

	It("should delete a volume snapshot", func(ctx SpecContext) {
		volumeSize := int64(1024 * 1024 * 1024)

		By("creating a volume")
		volume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{
				VolumeClassRef: &corev1.LocalObjectReference{Name: vc.Name},
				VolumePoolRef:  &corev1.LocalObjectReference{Name: vp.Name},
				Resources: corev1alpha1.ResourceList{
					corev1alpha1.ResourceStorage: resource.MustParse("1Gi"),
				},
			},
		}
		Expect(k8sClient.Create(ctx, volume)).To(Succeed())
		DeferCleanup(expectVolumeDeleted, volume)

		By("waiting for the runtime to report the volume")
		Eventually(srv).Should(HaveField("Volumes", HaveLen(1)))

		By("setting the volume status in the runtime")
		_, iriVolume := GetSingleMapEntry(srv.Volumes)
		iriVolume = &testingvolume.FakeVolume{Volume: proto.Clone(iriVolume.Volume).(*iri.Volume)}
		iriVolume.Status.State = iri.VolumeState_VOLUME_AVAILABLE
		iriVolume.Status.Access = &iri.VolumeAccess{
			Driver: "test-driver",
			Handle: "test-volume-id",
		}
		srv.SetVolumes([]*testingvolume.FakeVolume{iriVolume})

		By("waiting for the volume status to be updated")
		Eventually(Object(volume)).Should(HaveField("Status.State", Equal(storagev1alpha1.VolumeStateAvailable)))

		By("creating a volume snapshot")
		volumeSnapshot := &storagev1alpha1.VolumeSnapshot{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "volume-snapshot-",
			},
			Spec: storagev1alpha1.VolumeSnapshotSpec{
				VolumeRef: &corev1.LocalObjectReference{Name: volume.Name},
			},
		}
		Expect(k8sClient.Create(ctx, volumeSnapshot)).To(Succeed())

		By("waiting for the volume snapshot to have a finalizer")
		Eventually(Object(volumeSnapshot)).Should(HaveField("Finalizers", ContainElement(volumepoolletv1alpha1.VolumeSnapshotFinalizer)))

		By("waiting for the runtime to report the volume snapshot")
		Eventually(srv).Should(HaveField("VolumeSnapshots", HaveLen(1)))

		By("setting the volume snapshot status to ready in the runtime")
		_, iriVolumeSnapshot := GetSingleMapEntry(srv.VolumeSnapshots)
		iriVolumeSnapshot = &testingvolume.FakeVolumeSnapshot{VolumeSnapshot: proto.Clone(iriVolumeSnapshot.VolumeSnapshot).(*iri.VolumeSnapshot)}
		iriVolumeSnapshot.Status.State = iri.VolumeSnapshotState_VOLUME_SNAPSHOT_READY
		iriVolumeSnapshot.Status.Size = volumeSize
		srv.SetVolumeSnapshots([]*testingvolume.FakeVolumeSnapshot{iriVolumeSnapshot})

		By("waiting for the volume snapshot status to be updated")
		expectedSnapshotID := poolletutils.MakeID(testingvolume.FakeRuntimeName, iriVolumeSnapshot.Metadata.Id)
		Eventually(Object(volumeSnapshot)).Should(HaveField("Status.SnapshotID", Equal(expectedSnapshotID.String())))

		By("deleting the volume snapshot")
		Expect(k8sClient.Delete(ctx, volumeSnapshot)).To(Succeed())

		By("waiting for the volume snapshot to be deleted from the runtime")
		Eventually(srv).Should(HaveField("VolumeSnapshots", HaveLen(0)))

		By("waiting for the volume snapshot to be deleted from the cluster")
		Eventually(Get(volumeSnapshot)).Should(Satisfy(apierrors.IsNotFound))
	})

	It("should not create volume snapshot if referenced volume is not available", func(ctx SpecContext) {
		By("creating a volume")
		volume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{
				VolumeClassRef: &corev1.LocalObjectReference{Name: vc.Name},
				VolumePoolRef:  &corev1.LocalObjectReference{Name: vp.Name},
				Resources: corev1alpha1.ResourceList{
					corev1alpha1.ResourceStorage: resource.MustParse("1Gi"),
				},
			},
		}
		Expect(k8sClient.Create(ctx, volume)).To(Succeed())
		DeferCleanup(expectVolumeDeleted, volume)

		By("waiting for the runtime to report the volume")
		Eventually(srv).Should(HaveField("Volumes", HaveLen(1)))

		By("setting the volume status to pending in the runtime")
		_, iriVolume := GetSingleMapEntry(srv.Volumes)
		iriVolume.Status.State = iri.VolumeState_VOLUME_PENDING

		By("creating a volume snapshot")
		volumeSnapshot := &storagev1alpha1.VolumeSnapshot{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "volume-snapshot-",
			},
			Spec: storagev1alpha1.VolumeSnapshotSpec{
				VolumeRef: &corev1.LocalObjectReference{Name: volume.Name},
			},
		}
		Expect(k8sClient.Create(ctx, volumeSnapshot)).To(Succeed())
		DeferCleanup(expectVolumeSnapshotDeleted, volumeSnapshot)

		By("waiting for the volume snapshot to have a finalizer")
		Eventually(Object(volumeSnapshot)).Should(HaveField("Finalizers", ContainElement(volumepoolletv1alpha1.VolumeSnapshotFinalizer)))

		By("ensuring no volume snapshot is created in the runtime")
		Consistently(srv).Should(HaveField("VolumeSnapshots", HaveLen(0)))
	})

	It("should not create volume snapshot if referenced volume does not exist", func(ctx SpecContext) {
		By("creating a volume snapshot with non-existent volume reference")
		volumeSnapshot := &storagev1alpha1.VolumeSnapshot{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "volume-snapshot-",
			},
			Spec: storagev1alpha1.VolumeSnapshotSpec{
				VolumeRef: &corev1.LocalObjectReference{Name: "non-existent-volume"},
			},
		}
		Expect(k8sClient.Create(ctx, volumeSnapshot)).To(Succeed())
		DeferCleanup(expectVolumeSnapshotDeleted, volumeSnapshot)

		By("waiting for the volume snapshot to have a finalizer")
		Eventually(Object(volumeSnapshot)).Should(HaveField("Finalizers", ContainElement(volumepoolletv1alpha1.VolumeSnapshotFinalizer)))

		By("ensuring no volume snapshot is created in the runtime")
		Consistently(srv).Should(HaveField("VolumeSnapshots", HaveLen(0)))
	})
})

func expectVolumeSnapshotDeleted(volumeSnapshot *storagev1alpha1.VolumeSnapshot) {
	Expect(k8sClient.Delete(context.Background(), volumeSnapshot)).To(Succeed())
	Eventually(Get(volumeSnapshot)).Should(Satisfy(apierrors.IsNotFound))
}
