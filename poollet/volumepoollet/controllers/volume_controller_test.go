// Copyright 2022 OnMetal authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controllers_test

import (
	"fmt"

	corev1alpha1 "github.com/onmetal/onmetal-api/api/core/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/volume/v1alpha1"
	onmetalapiclient "github.com/onmetal/onmetal-api/utils/client"
	. "github.com/onmetal/onmetal-api/utils/testing"
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
	ctx := SetupContext()
	ns, vp, vc, expandableVc, srv := SetupTest(ctx)

	It("should create a basic volume", func() {
		size := resource.MustParse("10Mi")
		volumeMonitors := "test-monotors"
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

		By("waiting for the runtime to report the volume")
		Eventually(srv).Should(SatisfyAll(
			HaveField("Volumes", HaveLen(1)),
		))

		_, oriVolume := GetSingleMapEntry(srv.Volumes)

		Expect(oriVolume.Spec.Image).To(Equal(""))
		Expect(oriVolume.Spec.Class).To(Equal(vc.Name))
		Expect(oriVolume.Spec.Encryption).To(BeNil())
		Expect(oriVolume.Spec.Resources.StorageBytes).To(Equal(uint64(size.Value())))

		oriVolume.Status.Access = &ori.VolumeAccess{
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
		oriVolume.Status.State = ori.VolumeState_VOLUME_AVAILABLE

		Expect(onmetalapiclient.PatchAddReconcileAnnotation(ctx, k8sClient, volume)).Should(Succeed())

		Eventually(Object(volume)).Should(SatisfyAll(
			HaveField("Status.State", storagev1alpha1.VolumeStateAvailable),
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

	It("should create a volume with encryption secret", func() {
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

		By("waiting for the runtime to report the volume")
		Eventually(srv).Should(SatisfyAll(
			HaveField("Volumes", HaveLen(1)),
		))

		_, oriVolume := GetSingleMapEntry(srv.Volumes)

		Expect(oriVolume.Spec.Image).To(Equal(""))
		Expect(oriVolume.Spec.Class).To(Equal(vc.Name))
		Expect(oriVolume.Spec.Resources.StorageBytes).To(Equal(uint64(size.Value())))
		Expect(oriVolume.Spec.Encryption.SecretData).NotTo(HaveKeyWithValue(encryptionDataKey, encryptionData))

	})

	It("should expand a volume", func() {
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

		By("waiting for the runtime to report the volume")
		Eventually(srv).Should(SatisfyAll(
			HaveField("Volumes", HaveLen(1)),
		))

		_, oriVolume := GetSingleMapEntry(srv.Volumes)

		Expect(oriVolume.Spec.Image).To(Equal(""))
		Expect(oriVolume.Spec.Class).To(Equal(vc.Name))
		Expect(oriVolume.Spec.Resources.StorageBytes).To(Equal(uint64(size.Value())))

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

		Eventually(func() uint64 {
			_, oriVolume = GetSingleMapEntry(srv.Volumes)
			return oriVolume.Spec.Resources.StorageBytes
		}).Should(Equal(uint64(newSize.Value())))
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
