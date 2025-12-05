// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package controllers_test

import (
	"fmt"

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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

const (
	MonitorsKey = "monitors"
	ImageKey    = "image"
	UserIDKey   = "userID"
	UserKeyKey  = "userKey"
)

var _ = Describe("VolumeController", func() {
	ns, vp, vc, expandableVc, srv := SetupTest()

	It("should create a basic volume", func(ctx SpecContext) {
		size := resource.MustParse("10Mi")
		volumeMonitors := "test-monitors"
		volumeImage := "test-image"
		volumeId := "test-id"
		volumeUser := "test-user"
		volumeDriver := "test"
		volumeHandle := "testhandle"

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
					corev1alpha1.ResourceStorage: size,
				},
			},
		}
		Expect(k8sClient.Create(ctx, volume)).To(Succeed())
		DeferCleanup(expectVolumeDeleted, volume)

		By("waiting for the runtime to report the volume")
		Eventually(srv).Should(SatisfyAll(
			HaveField("Volumes", HaveLen(1)),
		))

		_, iriVolume := GetSingleMapEntry(srv.Volumes)

		Expect(iriVolume.Spec.Image).To(Equal(""))
		Expect(iriVolume.Spec.Class).To(Equal(vc.Name))
		Expect(iriVolume.Spec.Encryption).To(BeNil())
		Expect(iriVolume.Spec.Resources.StorageBytes).To(Equal(size.Value()))
		Expect(iriVolume.Metadata.Annotations).To(Equal(map[string]string{
			volumepoolletv1alpha1.IRIVolumeGenerationAnnotation: "0",
			volumepoolletv1alpha1.VolumeGenerationAnnotation:    "0",
		}))

		iriVolume = &testingvolume.FakeVolume{Volume: proto.Clone(iriVolume.Volume).(*iri.Volume)}
		iriVolume.Status.Access = &iri.VolumeAccess{
			Driver: volumeDriver,
			Handle: volumeHandle,
			Attributes: map[string]string{
				MonitorsKey: volumeMonitors,
				ImageKey:    volumeImage,
			},
			SecretData: map[string][]byte{
				UserIDKey:  []byte(volumeId),
				UserKeyKey: []byte(volumeUser),
			},
		}
		iriVolume.Status.State = iri.VolumeState_VOLUME_AVAILABLE
		iriVolume.Status.Resources = iriVolume.Spec.Resources
		srv.SetVolumes([]*testingvolume.FakeVolume{iriVolume})

		By("Waiting for the ironcore volume Status to be up-to-date")
		expectedVolumeID := poolletutils.MakeID(testingvolume.FakeRuntimeName, iriVolume.Metadata.Id)
		Eventually(Object(volume)).Should(SatisfyAll(
			HaveField("Status.State", storagev1alpha1.VolumeStateAvailable),
			HaveField("Status.VolumeID", expectedVolumeID.String()),
			HaveField("Status.Access.Driver", volumeDriver),
			HaveField("Status.Access.SecretRef", Not(BeNil())),
			HaveField("Status.Access.VolumeAttributes", HaveKeyWithValue(MonitorsKey, volumeMonitors)),
			HaveField("Status.Access.VolumeAttributes", HaveKeyWithValue(ImageKey, volumeImage)),
		))

		Expect(volume.Status.Resources.Storage().Value()).Should(Equal(size.Value()))

		accessSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns.Name,
				Name:      volume.Status.Access.SecretRef.Name,
			}}
		Eventually(Object(accessSecret)).Should(SatisfyAll(
			HaveField("Data", HaveKeyWithValue(UserKeyKey, []byte(volumeUser))),
			HaveField("Data", HaveKeyWithValue(UserIDKey, []byte(volumeId))),
		))

	})

	It("should create a volume with encryption secret", func(ctx SpecContext) {
		size := resource.MustParse("99Mi")

		encryptionDataKey := "encryptionKey"
		encryptionData := "test-data"

		By("creating a volume encryption secret")
		encryptionSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "encryption-",
			},
			Data: map[string][]byte{
				encryptionDataKey: []byte(encryptionData),
			},
		}
		Expect(k8sClient.Create(ctx, encryptionSecret)).To(Succeed())

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
					corev1alpha1.ResourceStorage: size,
				},
				Encryption: &storagev1alpha1.VolumeEncryption{
					SecretRef: corev1.LocalObjectReference{Name: encryptionSecret.Name},
				},
			},
		}
		Expect(k8sClient.Create(ctx, volume)).To(Succeed())
		DeferCleanup(expectVolumeDeleted, volume)

		By("waiting for the runtime to report the volume")
		Eventually(srv).Should(SatisfyAll(
			HaveField("Volumes", HaveLen(1)),
		))

		_, iriVolume := GetSingleMapEntry(srv.Volumes)

		Expect(iriVolume.Spec.Image).To(Equal(""))
		Expect(iriVolume.Spec.Class).To(Equal(vc.Name))
		Expect(iriVolume.Spec.Resources.StorageBytes).To(Equal(size.Value()))
		Expect(iriVolume.Spec.Encryption.SecretData).NotTo(HaveKeyWithValue(encryptionDataKey, encryptionData))

	})

	It("should expand a volume", func(ctx SpecContext) {
		size := resource.MustParse("100Mi")
		newSize := resource.MustParse("200Mi")

		By("creating a volume")
		volume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{
				VolumeClassRef: &corev1.LocalObjectReference{Name: expandableVc.Name},
				VolumePoolRef:  &corev1.LocalObjectReference{Name: vp.Name},
				Resources: corev1alpha1.ResourceList{
					corev1alpha1.ResourceStorage: size,
				},
			},
		}
		Expect(k8sClient.Create(ctx, volume)).To(Succeed())
		DeferCleanup(expectVolumeDeleted, volume)

		By("waiting for the runtime to report the volume")
		Eventually(srv).Should(SatisfyAll(
			HaveField("Volumes", HaveLen(1)),
		))

		_, iriVolume := GetSingleMapEntry(srv.Volumes)

		Expect(iriVolume.Spec.Image).To(Equal(""))
		Expect(iriVolume.Spec.Class).To(Equal(expandableVc.Name))
		Expect(iriVolume.Spec.Resources.StorageBytes).To(Equal(size.Value()))

		By("update increasing the storage resource")
		baseVolume := volume.DeepCopy()
		volume.Spec.Resources = corev1alpha1.ResourceList{
			corev1alpha1.ResourceStorage: newSize,
		}
		Expect(k8sClient.Patch(ctx, volume, client.MergeFrom(baseVolume))).To(Succeed())

		By("waiting for the runtime to report the volume")
		Eventually(srv).Should(SatisfyAll(
			HaveField("Volumes", HaveLen(1)),
		))

		Eventually(func() int64 {
			_, iriVolume = GetSingleMapEntry(srv.Volumes)
			return iriVolume.Spec.Resources.StorageBytes
		}).Should(Equal(newSize.Value()))

		By("inspecting the effective storage resource is set for volume")
		Eventually(func() int64 {
			return iriVolume.Status.Resources.StorageBytes
		}).Should(Equal(newSize.Value()))

		Eventually(Object(volume)).Should(SatisfyAll(
			HaveField("Status.Resources", Not(BeNil())),
		))
		Expect(volume.Status.Resources.Storage().Value()).Should(Equal(newSize.Value()))
	})

	It("should create a volume from snapshot", func(ctx SpecContext) {
		size := resource.MustParse("10Mi")

		By("creating a source volume")
		sourceVolume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "source-volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{
				VolumeClassRef: &corev1.LocalObjectReference{Name: vc.Name},
				VolumePoolRef:  &corev1.LocalObjectReference{Name: vp.Name},
				Resources: corev1alpha1.ResourceList{
					corev1alpha1.ResourceStorage: size,
				},
			},
		}
		Expect(k8sClient.Create(ctx, sourceVolume)).To(Succeed())
		DeferCleanup(expectVolumeDeleted, sourceVolume)

		By("waiting for the runtime to report the source volume")
		Eventually(srv).Should(SatisfyAll(
			HaveField("Volumes", HaveLen(1)),
		))

		By("setting the source volume status in the runtime")
		_, iriSourceVolume := GetSingleMapEntry(srv.Volumes)
		iriSourceVolume = &testingvolume.FakeVolume{Volume: proto.Clone(iriSourceVolume.Volume).(*iri.Volume)}
		iriSourceVolume.Status.State = iri.VolumeState_VOLUME_AVAILABLE
		iriSourceVolume.Status.Access = &iri.VolumeAccess{
			Driver: "test-driver",
			Handle: "test-volume-id",
		}
		srv.SetVolumes([]*testingvolume.FakeVolume{iriSourceVolume})

		By("waiting for the source volume status to be updated")
		Eventually(Object(sourceVolume)).Should(HaveField("Status.State", Equal(storagev1alpha1.VolumeStateAvailable)))

		By("creating a volume snapshot")
		volumeSnapshot := &storagev1alpha1.VolumeSnapshot{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "snapshot-",
			},
			Spec: storagev1alpha1.VolumeSnapshotSpec{
				VolumeRef: &corev1.LocalObjectReference{Name: sourceVolume.Name},
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

		By("setting the volume snapshot status in the runtime")
		_, iriVolumeSnapshot := GetSingleMapEntry(srv.VolumeSnapshots)
		iriVolumeSnapshot = &testingvolume.FakeVolumeSnapshot{VolumeSnapshot: proto.Clone(iriVolumeSnapshot.VolumeSnapshot).(*iri.VolumeSnapshot)}
		iriVolumeSnapshot.Status.State = iri.VolumeSnapshotState_VOLUME_SNAPSHOT_READY
		iriVolumeSnapshot.Status.Size = size.Value()
		srv.SetVolumeSnapshots([]*testingvolume.FakeVolumeSnapshot{iriVolumeSnapshot})

		By("waiting for the volume snapshot status to be updated")
		expectedSnapshotID := poolletutils.MakeID(testingvolume.FakeRuntimeName, iriVolumeSnapshot.Metadata.Id)
		Eventually(Object(volumeSnapshot)).Should(SatisfyAll(
			HaveField("Status.SnapshotID", Equal(expectedSnapshotID.String())),
			HaveField("Status.State", Equal(storagev1alpha1.VolumeSnapshotStateReady)),
		))

		By("creating a volume from snapshot")
		volume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{
				VolumeClassRef: &corev1.LocalObjectReference{Name: vc.Name},
				VolumePoolRef:  &corev1.LocalObjectReference{Name: vp.Name},
				Resources: corev1alpha1.ResourceList{
					corev1alpha1.ResourceStorage: size,
				},
				VolumeDataSource: storagev1alpha1.VolumeDataSource{
					VolumeSnapshotRef: &corev1.LocalObjectReference{Name: volumeSnapshot.Name},
				},
			},
		}
		Expect(k8sClient.Create(ctx, volume)).To(Succeed())
		DeferCleanup(expectVolumeDeleted, volume)

		By("waiting for the volume to be created in the runtime")
		Eventually(srv).Should(SatisfyAll(
			HaveField("Volumes", HaveLen(2)),
		))

		By("verifying the volume has the correct data source")
		var iriVolume *iri.Volume
		for _, iriVol := range srv.Volumes {
			if iriVol.Spec.VolumeDataSource != nil && iriVol.Spec.VolumeDataSource.SnapshotDataSource != nil {
				iriVolume = iriVol.Volume
				break
			}
		}
		Expect(iriVolume).NotTo(BeNil())
		Expect(iriVolume.Spec.VolumeDataSource).NotTo(BeNil())
		Expect(iriVolume.Spec.VolumeDataSource.SnapshotDataSource).NotTo(BeNil())
		Expect(iriVolume.Spec.VolumeDataSource.SnapshotDataSource.SnapshotId).To(Equal(volumeSnapshot.Name))

		By("verifying the volume is created")
		Expect(volume).NotTo(BeNil())
	})

	It("should retry volume creation when snapshot becomes ready", func(ctx SpecContext) {
		size := resource.MustParse("10Mi")

		By("creating a source volume")
		sourceVolume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "source-volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{
				VolumeClassRef: &corev1.LocalObjectReference{Name: vc.Name},
				VolumePoolRef:  &corev1.LocalObjectReference{Name: vp.Name},
				Resources: corev1alpha1.ResourceList{
					corev1alpha1.ResourceStorage: size,
				},
			},
		}
		Expect(k8sClient.Create(ctx, sourceVolume)).To(Succeed())
		DeferCleanup(expectVolumeDeleted, sourceVolume)

		By("waiting for the runtime to report the source volume")
		Eventually(srv).Should(SatisfyAll(
			HaveField("Volumes", HaveLen(1)),
		))

		By("setting the source volume status in the runtime")
		_, iriSourceVolume := GetSingleMapEntry(srv.Volumes)
		iriSourceVolume = &testingvolume.FakeVolume{Volume: proto.Clone(iriSourceVolume.Volume).(*iri.Volume)}
		iriSourceVolume.Status.State = iri.VolumeState_VOLUME_AVAILABLE
		iriSourceVolume.Status.Access = &iri.VolumeAccess{
			Driver: "test-driver",
			Handle: "test-volume-id",
		}
		srv.SetVolumes([]*testingvolume.FakeVolume{iriSourceVolume})

		By("waiting for the source volume status to be updated")
		Eventually(Object(sourceVolume)).Should(HaveField("Status.State", Equal(storagev1alpha1.VolumeStateAvailable)))

		By("creating a volume snapshot")
		volumeSnapshot := &storagev1alpha1.VolumeSnapshot{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "snapshot-",
			},
			Spec: storagev1alpha1.VolumeSnapshotSpec{
				VolumeRef: &corev1.LocalObjectReference{Name: sourceVolume.Name},
			},
		}
		Expect(k8sClient.Create(ctx, volumeSnapshot)).To(Succeed())
		DeferCleanup(expectVolumeSnapshotDeleted, volumeSnapshot)

		By("waiting for the volume snapshot to have a finalizer")
		Eventually(Object(volumeSnapshot)).Should(HaveField("Finalizers", ContainElement(volumepoolletv1alpha1.VolumeSnapshotFinalizer)))

		By("waiting for the runtime to report the volume snapshot")
		Eventually(srv).Should(HaveField("VolumeSnapshots", HaveLen(1)))

		By("creating a volume from snapshot (while snapshot is still pending)")
		volume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{
				VolumeClassRef: &corev1.LocalObjectReference{Name: vc.Name},
				VolumePoolRef:  &corev1.LocalObjectReference{Name: vp.Name},
				Resources: corev1alpha1.ResourceList{
					corev1alpha1.ResourceStorage: size,
				},
				VolumeDataSource: storagev1alpha1.VolumeDataSource{
					VolumeSnapshotRef: &corev1.LocalObjectReference{Name: volumeSnapshot.Name},
				},
			},
		}
		Expect(k8sClient.Create(ctx, volume)).To(Succeed())
		DeferCleanup(expectVolumeDeleted, volume)

		By("verifying the volume is not created in runtime yet (snapshot still pending)")
		Consistently(srv).Should(SatisfyAll(HaveField("Volumes", HaveLen(1))), "Only the source volume exists")

		By("setting the volume snapshot status to ready in the runtime")
		_, iriVolumeSnapshot := GetSingleMapEntry(srv.VolumeSnapshots)
		iriVolumeSnapshot = &testingvolume.FakeVolumeSnapshot{VolumeSnapshot: proto.Clone(iriVolumeSnapshot.VolumeSnapshot).(*iri.VolumeSnapshot)}
		iriVolumeSnapshot.Status.State = iri.VolumeSnapshotState_VOLUME_SNAPSHOT_READY
		iriVolumeSnapshot.Status.Size = size.Value()
		srv.SetVolumeSnapshots([]*testingvolume.FakeVolumeSnapshot{iriVolumeSnapshot})

		By("waiting for the volume snapshot status to be updated")
		expectedSnapshotID := poolletutils.MakeID(testingvolume.FakeRuntimeName, iriVolumeSnapshot.Metadata.Id)
		Eventually(Object(volumeSnapshot)).Should(SatisfyAll(
			HaveField("Status.SnapshotID", Equal(expectedSnapshotID.String())),
			HaveField("Status.State", Equal(storagev1alpha1.VolumeSnapshotStateReady)),
		))

		By("waiting for the runtime to report the volume")
		Eventually(srv).Should(SatisfyAll(HaveField("Volumes", HaveLen(2))))

		By("verifying the volume has the correct data source")
		var iriVolume *iri.Volume
		for _, iriVol := range srv.Volumes {
			if iriVol.Spec.VolumeDataSource != nil && iriVol.Spec.VolumeDataSource.SnapshotDataSource != nil {
				iriVolume = iriVol.Volume
				break
			}
		}
		Expect(iriVolume).NotTo(BeNil())
		Expect(iriVolume.Spec.VolumeDataSource).NotTo(BeNil())
		Expect(iriVolume.Spec.VolumeDataSource.SnapshotDataSource).NotTo(BeNil())
		Expect(iriVolume.Spec.VolumeDataSource.SnapshotDataSource.SnapshotId).To(Equal(volumeSnapshot.Name))
	})

	It("should inherit encryption when creating volume from snapshot of encrypted volume", func(ctx SpecContext) {
		size := resource.MustParse("10Mi")
		encryptionDataKey := "encryptionKey"
		encryptionData := "test-encryption-data"

		By("creating an encryption secret for the source volume")
		encryptionSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "encryption-",
			},
			Data: map[string][]byte{
				encryptionDataKey: []byte(encryptionData),
			},
		}
		Expect(k8sClient.Create(ctx, encryptionSecret)).To(Succeed())

		By("creating an encrypted source volume")
		sourceVolume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "source-volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{
				VolumeClassRef: &corev1.LocalObjectReference{Name: vc.Name},
				VolumePoolRef:  &corev1.LocalObjectReference{Name: vp.Name},
				Resources: corev1alpha1.ResourceList{
					corev1alpha1.ResourceStorage: size,
				},
				Encryption: &storagev1alpha1.VolumeEncryption{
					SecretRef: corev1.LocalObjectReference{Name: encryptionSecret.Name},
				},
			},
		}
		Expect(k8sClient.Create(ctx, sourceVolume)).To(Succeed())
		DeferCleanup(expectVolumeDeleted, sourceVolume)

		By("waiting for the runtime to report the source volume")
		Eventually(srv).Should(SatisfyAll(
			HaveField("Volumes", HaveLen(1)),
		))

		By("setting the source volume status in the runtime")
		_, iriSourceVolume := GetSingleMapEntry(srv.Volumes)
		iriSourceVolume = &testingvolume.FakeVolume{Volume: proto.Clone(iriSourceVolume.Volume).(*iri.Volume)}
		iriSourceVolume.Status.State = iri.VolumeState_VOLUME_AVAILABLE
		iriSourceVolume.Status.Access = &iri.VolumeAccess{
			Driver: "test-driver",
			Handle: "test-volume-id",
		}
		srv.SetVolumes([]*testingvolume.FakeVolume{iriSourceVolume})

		By("waiting for the source volume status to be updated")
		Eventually(Object(sourceVolume)).Should(HaveField("Status.State", Equal(storagev1alpha1.VolumeStateAvailable)))

		By("creating a volume snapshot")
		volumeSnapshot := &storagev1alpha1.VolumeSnapshot{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "snapshot-",
			},
			Spec: storagev1alpha1.VolumeSnapshotSpec{
				VolumeRef: &corev1.LocalObjectReference{Name: sourceVolume.Name},
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

		By("setting the volume snapshot status in the runtime")
		_, iriVolumeSnapshot := GetSingleMapEntry(srv.VolumeSnapshots)
		iriVolumeSnapshot = &testingvolume.FakeVolumeSnapshot{VolumeSnapshot: proto.Clone(iriVolumeSnapshot.VolumeSnapshot).(*iri.VolumeSnapshot)}
		iriVolumeSnapshot.Status.State = iri.VolumeSnapshotState_VOLUME_SNAPSHOT_READY
		iriVolumeSnapshot.Status.Size = size.Value()
		srv.SetVolumeSnapshots([]*testingvolume.FakeVolumeSnapshot{iriVolumeSnapshot})

		By("waiting for the volume snapshot status to be updated")
		expectedSnapshotID := poolletutils.MakeID(testingvolume.FakeRuntimeName, iriVolumeSnapshot.Metadata.Id)
		Eventually(Object(volumeSnapshot)).Should(SatisfyAll(
			HaveField("Status.SnapshotID", Equal(expectedSnapshotID.String())),
			HaveField("Status.State", Equal(storagev1alpha1.VolumeSnapshotStateReady)),
		))

		By("creating a volume from snapshot (without explicit encryption)")
		volume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{
				VolumeClassRef: &corev1.LocalObjectReference{Name: vc.Name},
				VolumePoolRef:  &corev1.LocalObjectReference{Name: vp.Name},
				Resources: corev1alpha1.ResourceList{
					corev1alpha1.ResourceStorage: size,
				},
				VolumeDataSource: storagev1alpha1.VolumeDataSource{
					VolumeSnapshotRef: &corev1.LocalObjectReference{Name: volumeSnapshot.Name},
				},
			},
		}
		Expect(k8sClient.Create(ctx, volume)).To(Succeed())
		DeferCleanup(expectVolumeDeleted, volume)

		By("waiting for the volume to be created in the runtime")
		Eventually(srv).Should(SatisfyAll(
			HaveField("Volumes", HaveLen(2)),
		))

		By("verifying the volume has the correct data source")
		var iriVolume *iri.Volume
		for _, iriVol := range srv.Volumes {
			if iriVol.Spec.VolumeDataSource != nil && iriVol.Spec.VolumeDataSource.SnapshotDataSource != nil {
				iriVolume = iriVol.Volume
				break
			}
		}
		Expect(iriVolume).NotTo(BeNil())
		Expect(iriVolume.Spec.VolumeDataSource).NotTo(BeNil())
		Expect(iriVolume.Spec.VolumeDataSource.SnapshotDataSource).NotTo(BeNil())
		Expect(iriVolume.Spec.VolumeDataSource.SnapshotDataSource.SnapshotId).To(Equal(volumeSnapshot.Name))

		By("verifying encryption is inherited")
		Expect(iriVolume.Spec.Encryption).NotTo(BeNil())
		Expect(iriVolume.Spec.Encryption.SecretData).To(HaveKeyWithValue(encryptionDataKey, []byte(encryptionData)))
	})

	It("should reject explicit encryption when creating volume from snapshot of encrypted volume", func(ctx SpecContext) {
		size := resource.MustParse("10Mi")
		sourceEncryptionDataKey := "sourceEncryptionKey"
		sourceEncryptionData := "source-encryption-data"
		userEncryptionDataKey := "userEncryptionKey"
		userEncryptionData := "user-encryption-data"

		By("creating an encryption secret for the source volume")
		sourceEncryptionSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "source-encryption-",
			},
			Data: map[string][]byte{
				sourceEncryptionDataKey: []byte(sourceEncryptionData),
			},
		}
		Expect(k8sClient.Create(ctx, sourceEncryptionSecret)).To(Succeed())

		By("creating an encrypted source volume")
		sourceVolume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "source-volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{
				VolumeClassRef: &corev1.LocalObjectReference{Name: vc.Name},
				VolumePoolRef:  &corev1.LocalObjectReference{Name: vp.Name},
				Resources: corev1alpha1.ResourceList{
					corev1alpha1.ResourceStorage: size,
				},
				Encryption: &storagev1alpha1.VolumeEncryption{
					SecretRef: corev1.LocalObjectReference{Name: sourceEncryptionSecret.Name},
				},
			},
		}
		Expect(k8sClient.Create(ctx, sourceVolume)).To(Succeed())
		DeferCleanup(expectVolumeDeleted, sourceVolume)

		By("waiting for the runtime to report the source volume")
		Eventually(srv).Should(SatisfyAll(
			HaveField("Volumes", HaveLen(1)),
		))

		By("setting the source volume status in the runtime")
		_, iriSourceVolume := GetSingleMapEntry(srv.Volumes)
		iriSourceVolume = &testingvolume.FakeVolume{Volume: proto.Clone(iriSourceVolume.Volume).(*iri.Volume)}
		iriSourceVolume.Status.State = iri.VolumeState_VOLUME_AVAILABLE
		iriSourceVolume.Status.Access = &iri.VolumeAccess{
			Driver: "test-driver",
			Handle: "test-volume-id",
		}
		srv.SetVolumes([]*testingvolume.FakeVolume{iriSourceVolume})

		By("waiting for the source volume status to be updated")
		Eventually(Object(sourceVolume)).Should(HaveField("Status.State", Equal(storagev1alpha1.VolumeStateAvailable)))

		By("creating a volume snapshot")
		volumeSnapshot := &storagev1alpha1.VolumeSnapshot{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "snapshot-",
			},
			Spec: storagev1alpha1.VolumeSnapshotSpec{
				VolumeRef: &corev1.LocalObjectReference{Name: sourceVolume.Name},
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

		By("setting the volume snapshot status in the runtime")
		_, iriVolumeSnapshot := GetSingleMapEntry(srv.VolumeSnapshots)
		iriVolumeSnapshot = &testingvolume.FakeVolumeSnapshot{VolumeSnapshot: proto.Clone(iriVolumeSnapshot.VolumeSnapshot).(*iri.VolumeSnapshot)}
		iriVolumeSnapshot.Status.State = iri.VolumeSnapshotState_VOLUME_SNAPSHOT_READY
		iriVolumeSnapshot.Status.Size = size.Value()
		srv.SetVolumeSnapshots([]*testingvolume.FakeVolumeSnapshot{iriVolumeSnapshot})

		By("waiting for the volume snapshot status to be updated")
		expectedSnapshotID := poolletutils.MakeID(testingvolume.FakeRuntimeName, iriVolumeSnapshot.Metadata.Id)
		Eventually(Object(volumeSnapshot)).Should(SatisfyAll(
			HaveField("Status.SnapshotID", Equal(expectedSnapshotID.String())),
			HaveField("Status.State", Equal(storagev1alpha1.VolumeSnapshotStateReady)),
		))

		By("creating a user encryption secret")
		userEncryptionSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "user-encryption-",
			},
			Data: map[string][]byte{
				userEncryptionDataKey: []byte(userEncryptionData),
			},
		}
		Expect(k8sClient.Create(ctx, userEncryptionSecret)).To(Succeed())

		By("creating a volume from snapshot of encrypted volume with explicit encryption (should fail)")
		volume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{
				VolumeClassRef: &corev1.LocalObjectReference{Name: vc.Name},
				VolumePoolRef:  &corev1.LocalObjectReference{Name: vp.Name},
				Resources: corev1alpha1.ResourceList{
					corev1alpha1.ResourceStorage: size,
				},
				VolumeDataSource: storagev1alpha1.VolumeDataSource{
					VolumeSnapshotRef: &corev1.LocalObjectReference{Name: volumeSnapshot.Name},
				},
				Encryption: &storagev1alpha1.VolumeEncryption{
					SecretRef: corev1.LocalObjectReference{Name: userEncryptionSecret.Name},
				},
			},
		}
		Expect(k8sClient.Create(ctx, volume)).To(Succeed())
		DeferCleanup(expectVolumeDeleted, volume)

		By("verifying the volume creation fails due to encryption conflict")
		Consistently(srv).Should(SatisfyAll(
			HaveField("Volumes", HaveLen(1)),
		), "Only source volume exists, new volume not created due to encryption conflict")
	})

	It("should allow explicit encryption when creating volume from snapshot of unencrypted volume", func(ctx SpecContext) {
		size := resource.MustParse("10Mi")
		userEncryptionDataKey := "userEncryptionKey"
		userEncryptionData := "user-encryption-data"

		By("creating an unencrypted source volume")
		sourceVolume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "source-volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{
				VolumeClassRef: &corev1.LocalObjectReference{Name: vc.Name},
				VolumePoolRef:  &corev1.LocalObjectReference{Name: vp.Name},
				Resources: corev1alpha1.ResourceList{
					corev1alpha1.ResourceStorage: size,
				},
			},
		}
		Expect(k8sClient.Create(ctx, sourceVolume)).To(Succeed())
		DeferCleanup(expectVolumeDeleted, sourceVolume)

		By("waiting for the runtime to report the source volume")
		Eventually(srv).Should(SatisfyAll(
			HaveField("Volumes", HaveLen(1)),
		))

		By("setting the source volume status in the runtime")
		_, iriSourceVolume := GetSingleMapEntry(srv.Volumes)
		iriSourceVolume = &testingvolume.FakeVolume{Volume: proto.Clone(iriSourceVolume.Volume).(*iri.Volume)}
		iriSourceVolume.Status.State = iri.VolumeState_VOLUME_AVAILABLE
		iriSourceVolume.Status.Access = &iri.VolumeAccess{
			Driver: "test-driver",
			Handle: "test-volume-id",
		}
		srv.SetVolumes([]*testingvolume.FakeVolume{iriSourceVolume})

		By("waiting for the source volume status to be updated")
		Eventually(Object(sourceVolume)).Should(HaveField("Status.State", Equal(storagev1alpha1.VolumeStateAvailable)))

		By("creating a volume snapshot")
		volumeSnapshot := &storagev1alpha1.VolumeSnapshot{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "snapshot-",
			},
			Spec: storagev1alpha1.VolumeSnapshotSpec{
				VolumeRef: &corev1.LocalObjectReference{Name: sourceVolume.Name},
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

		By("setting the volume snapshot status in the runtime")
		_, iriVolumeSnapshot := GetSingleMapEntry(srv.VolumeSnapshots)
		iriVolumeSnapshot = &testingvolume.FakeVolumeSnapshot{VolumeSnapshot: proto.Clone(iriVolumeSnapshot.VolumeSnapshot).(*iri.VolumeSnapshot)}
		iriVolumeSnapshot.Status.State = iri.VolumeSnapshotState_VOLUME_SNAPSHOT_READY
		iriVolumeSnapshot.Status.Size = size.Value()
		srv.SetVolumeSnapshots([]*testingvolume.FakeVolumeSnapshot{iriVolumeSnapshot})

		By("waiting for the volume snapshot status to be updated")
		expectedSnapshotID := poolletutils.MakeID(testingvolume.FakeRuntimeName, iriVolumeSnapshot.Metadata.Id)
		Eventually(Object(volumeSnapshot)).Should(SatisfyAll(
			HaveField("Status.SnapshotID", Equal(expectedSnapshotID.String())),
			HaveField("Status.State", Equal(storagev1alpha1.VolumeSnapshotStateReady)),
		))

		By("creating a user encryption secret")
		userEncryptionSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "user-encryption-",
			},
			Data: map[string][]byte{
				userEncryptionDataKey: []byte(userEncryptionData),
			},
		}
		Expect(k8sClient.Create(ctx, userEncryptionSecret)).To(Succeed())

		By("creating a volume from unencrypted snapshot with explicit encryption (should work)")
		volume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{
				VolumeClassRef: &corev1.LocalObjectReference{Name: vc.Name},
				VolumePoolRef:  &corev1.LocalObjectReference{Name: vp.Name},
				Resources: corev1alpha1.ResourceList{
					corev1alpha1.ResourceStorage: size,
				},
				VolumeDataSource: storagev1alpha1.VolumeDataSource{
					VolumeSnapshotRef: &corev1.LocalObjectReference{Name: volumeSnapshot.Name},
				},
				Encryption: &storagev1alpha1.VolumeEncryption{
					SecretRef: corev1.LocalObjectReference{Name: userEncryptionSecret.Name},
				},
			},
		}
		Expect(k8sClient.Create(ctx, volume)).To(Succeed())
		DeferCleanup(expectVolumeDeleted, volume)

		By("waiting for the volume to be created in the runtime")
		Eventually(srv).Should(SatisfyAll(
			HaveField("Volumes", HaveLen(2)),
		))

		By("verifying the volume has the correct data source")
		var iriVolume *iri.Volume
		for _, iriVol := range srv.Volumes {
			if iriVol.Spec.VolumeDataSource != nil && iriVol.Spec.VolumeDataSource.SnapshotDataSource != nil {
				iriVolume = iriVol.Volume
				break
			}
		}
		Expect(iriVolume).NotTo(BeNil())
		Expect(iriVolume.Spec.VolumeDataSource).NotTo(BeNil())
		Expect(iriVolume.Spec.VolumeDataSource.SnapshotDataSource).NotTo(BeNil())
		Expect(iriVolume.Spec.VolumeDataSource.SnapshotDataSource.SnapshotId).To(Equal(volumeSnapshot.Name))

		By("verifying user's explicit encryption is used")
		Expect(iriVolume.Spec.Encryption).NotTo(BeNil())
		Expect(iriVolume.Spec.Encryption.SecretData).To(HaveKeyWithValue(userEncryptionDataKey, []byte(userEncryptionData)))
	})

	It("should not create volume from snapshot when source volume encryption secret is not found", func(ctx SpecContext) {
		size := resource.MustParse("10Mi")

		By("creating an encrypted source volume with non-existent secret")
		sourceVolume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "source-volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{
				VolumeClassRef: &corev1.LocalObjectReference{Name: vc.Name},
				VolumePoolRef:  &corev1.LocalObjectReference{Name: vp.Name},
				Resources: corev1alpha1.ResourceList{
					corev1alpha1.ResourceStorage: size,
				},
				Encryption: &storagev1alpha1.VolumeEncryption{
					SecretRef: corev1.LocalObjectReference{Name: "non-existent-secret"},
				},
			},
		}
		Expect(k8sClient.Create(ctx, sourceVolume)).To(Succeed())
		DeferCleanup(expectVolumeDeleted, sourceVolume)

		By("verifying the source volume is not created due to missing encryption secret")
		Consistently(srv).Should(HaveField("Volumes", HaveLen(0)), "No volumes created due to missing encryption secret")

	})

	It("should not create volume from snapshot when source volume is not found", func(ctx SpecContext) {
		size := resource.MustParse("10Mi")

		By("creating a volume snapshot with non-existent source volume")
		volumeSnapshot := &storagev1alpha1.VolumeSnapshot{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "snapshot-",
			},
			Spec: storagev1alpha1.VolumeSnapshotSpec{
				VolumeRef: &corev1.LocalObjectReference{Name: "non-existent-volume"},
			},
		}
		Expect(k8sClient.Create(ctx, volumeSnapshot)).To(Succeed())
		DeferCleanup(expectVolumeSnapshotDeleted, volumeSnapshot)

		By("creating a volume from the snapshot")
		volume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{
				VolumeClassRef: &corev1.LocalObjectReference{Name: vc.Name},
				VolumePoolRef:  &corev1.LocalObjectReference{Name: vp.Name},
				Resources: corev1alpha1.ResourceList{
					corev1alpha1.ResourceStorage: size,
				},
				VolumeDataSource: storagev1alpha1.VolumeDataSource{
					VolumeSnapshotRef: &corev1.LocalObjectReference{Name: volumeSnapshot.Name},
				},
			},
		}
		Expect(k8sClient.Create(ctx, volume)).To(Succeed())
		DeferCleanup(expectVolumeDeleted, volume)

		By("verifying the volume creation fails due to missing source volume")
		Eventually(Object(volume)).Should(HaveField("Status.State", Equal(storagev1alpha1.VolumeStatePending)))

		Consistently(srv).Should(HaveField("Volumes", HaveLen(0)), "No volumes created due to missing source volume")
	})

	It("should not create volume from snapshot when snapshot is not found", func(ctx SpecContext) {
		size := resource.MustParse("10Mi")

		By("creating a volume with non-existent snapshot reference")
		volume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{
				VolumeClassRef: &corev1.LocalObjectReference{Name: vc.Name},
				VolumePoolRef:  &corev1.LocalObjectReference{Name: vp.Name},
				Resources: corev1alpha1.ResourceList{
					corev1alpha1.ResourceStorage: size,
				},
				VolumeDataSource: storagev1alpha1.VolumeDataSource{
					VolumeSnapshotRef: &corev1.LocalObjectReference{Name: "non-existent-snapshot"},
				},
			},
		}
		Expect(k8sClient.Create(ctx, volume)).To(Succeed())
		DeferCleanup(expectVolumeDeleted, volume)

		By("verifying the volume creation fails due to missing snapshot")
		Eventually(Object(volume)).Should(HaveField("Status.State", Equal(storagev1alpha1.VolumeStatePending)))

		Consistently(srv).Should(HaveField("Volumes", HaveLen(0)), "No volumes created due to missing snapshot")
	})

})

func GetSingleMapEntry[K comparable, V any](m map[K]V) (K, V) {
	if n := len(m); n != 1 {
		Fail(fmt.Sprintf("Expected for map to have a single entry but got %d", n), 1)
	}
	for k, v := range m {
		return k, v
	}
	panic("unreachable")
}
