// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package controllers_test

import (
	"fmt"

	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
	testingvolume "github.com/ironcore-dev/ironcore/iri/testing/volume"
	ironcoreclient "github.com/ironcore-dev/ironcore/utils/client"
	poolletproviderid "github.com/ironcore-dev/ironcore/utils/poollet"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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

		Expect(ironcoreclient.PatchAddReconcileAnnotation(ctx, k8sClient, volume)).Should(Succeed())

		By("Waiting for the ironcore volume Status to be up-to-date")
		expectedVolumeID := poolletproviderid.MakeID(testingvolume.FakeRuntimeName, iriVolume.Metadata.Id)
		Eventually(Object(volume)).Should(SatisfyAll(
			HaveField("Status.State", storagev1alpha1.VolumeStateAvailable),
			HaveField("Status.VolumeID", expectedVolumeID.String()),
			HaveField("Status.Access.Driver", volumeDriver),
			HaveField("Status.Access.SecretRef", Not(BeNil())),
			HaveField("Status.Access.VolumeAttributes", HaveKeyWithValue(MonitorsKey, volumeMonitors)),
			HaveField("Status.Access.VolumeAttributes", HaveKeyWithValue(ImageKey, volumeImage)),
		))

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
