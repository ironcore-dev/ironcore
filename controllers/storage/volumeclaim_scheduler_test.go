/*
 * Copyright (c) 2022 by the OnMetal authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package storage

import (
	storagev1alpha1 "github.com/onmetal/onmetal-api/apis/storage/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("VolumeClaimScheduler", func() {
	ns := SetupTest(ctx)

	var volume, volume2 *storagev1alpha1.Volume
	var volumeClaim *storagev1alpha1.VolumeClaim

	BeforeEach(func() {
		// 100Gi volume
		volume = &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{
				Resources: map[corev1.ResourceName]resource.Quantity{
					"storage": resource.MustParse("100Gi"),
				},
				StoragePool: corev1.LocalObjectReference{
					Name: "my-storagepool",
				},
				StorageClassRef: corev1.LocalObjectReference{
					Name: "my-volumeclass",
				},
			},
		}
		// 10Gi volume
		volume2 = &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{
				Resources: map[corev1.ResourceName]resource.Quantity{
					"storage": resource.MustParse("10Gi"),
				},
				StoragePool: corev1.LocalObjectReference{
					Name: "my-storagepool",
				},
				StorageClassRef: corev1.LocalObjectReference{
					Name: "my-volumeclass",
				},
			},
		}
		volumeClaim = &storagev1alpha1.VolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-volume-claim-",
			},
			Spec: storagev1alpha1.VolumeClaimSpec{
				Resources: map[corev1.ResourceName]resource.Quantity{
					"storage": resource.MustParse("100Gi"),
				},
				Selector: &metav1.LabelSelector{},
				StorageClassRef: corev1.LocalObjectReference{
					Name: "my-volumeclass",
				},
			},
		}
	})

	It("Should claim a volume matching the volumeclaim resource requirements", func() {
		By("creating a volume w/ a set of resources")
		Expect(k8sClient.Create(ctx, volume)).To(Succeed(), "failed to create volume")

		By("patching the volume status to available")
		volumeBase := volume.DeepCopy()
		volume.Status.State = storagev1alpha1.VolumeStateAvailable
		Expect(k8sClient.Status().Patch(ctx, volume, client.MergeFrom(volumeBase))).
			To(Succeed(), "failed to patch volume status")

		By("creating a volumeclaim which should claim the matching volume")
		Expect(k8sClient.Create(ctx, volumeClaim)).To(Succeed(), "failed to create volumeclaim")

		By("waiting for the volume to reference the claim")
		volumeKey := client.ObjectKeyFromObject(volume)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, volumeKey, volume)).To(Succeed(), "failed to get volume")
			g.Expect(volume.Spec.ClaimRef.Name).To(Equal(volumeClaim.Name))
			g.Expect(volume.Spec.ClaimRef.UID).To(Equal(volumeClaim.UID))
		}, timeout, interval).Should(Succeed())

		By("waiting for the volumeclaim to reference the volume")
		claimKey := client.ObjectKeyFromObject(volumeClaim)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, claimKey, volumeClaim)).To(Succeed(), "failed to get volumeclaim")
			g.Expect(volumeClaim.Spec.VolumeRef.Name).To(Equal(volume.Name))
		}, timeout, interval).Should(Succeed())
	})

	It("Should not claim a volume if volumeclaim with matching resource requirements is found", func() {
		By("creating a volume w/ a set of resources")
		Expect(k8sClient.Create(ctx, volume2)).To(Succeed(), "failed to create volume")

		By("patching the volume status to available")
		volume2Base := volume2.DeepCopy()
		volume2.Status.State = storagev1alpha1.VolumeStateAvailable
		Expect(k8sClient.Status().Patch(ctx, volume2, client.MergeFrom(volume2Base))).
			To(Succeed(), "failed to patch volume status")

		By("creating a volumeclaim which should claim the matching volume")
		Expect(k8sClient.Create(ctx, volumeClaim)).To(Succeed(), "failed to create volumeclaim")

		By("waiting for the volume to reference the claim")
		volume2Key := client.ObjectKeyFromObject(volume2)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, volume2Key, volume2)).To(Succeed(), "failed to get volume")
			g.Expect(volume2.Spec.ClaimRef.Name).To(Equal(""))
			g.Expect(volume2.Spec.ClaimRef.UID).To(Equal(types.UID("")))
		}, timeout, interval).Should(Succeed())

		By("waiting for the volumeclaim to reference the volume")
		claimKey := client.ObjectKeyFromObject(volumeClaim)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, claimKey, volumeClaim)).To(Succeed(), "failed to get volumeclaim")
			g.Expect(volumeClaim.Spec.VolumeRef.Name).To(Equal(""))
		}, timeout, interval).Should(Succeed())
	})

	It("Should not claim a volume if the volume status is not set to available", func() {
		By("creating a volume w/ a set of resources")
		Expect(k8sClient.Create(ctx, volume)).To(Succeed(), "failed to create volume")

		By("patching the volume status to available")
		volumeBase := volume.DeepCopy()
		volume.Status.State = storagev1alpha1.VolumeStatePending
		Expect(k8sClient.Status().Patch(ctx, volume, client.MergeFrom(volumeBase))).
			To(Succeed(), "failed to patch volume status")

		By("creating a volumeclaim which should claim the matching volume")
		Expect(k8sClient.Create(ctx, volumeClaim)).To(Succeed(), "failed to create volumeclaim")

		By("waiting for the volume to reference the claim")
		volumeKey := client.ObjectKeyFromObject(volume)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, volumeKey, volume)).To(Succeed(), "failed to get volume")
			g.Expect(volume.Spec.ClaimRef.Name).To(Equal(""))
			g.Expect(volume.Spec.ClaimRef.UID).To(Equal(types.UID("")))
		}, timeout, interval).Should(Succeed())

		By("waiting for the volumeclaim to reference the volume")
		claimKey := client.ObjectKeyFromObject(volumeClaim)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, claimKey, volumeClaim)).To(Succeed(), "failed to get volumeclaim")
			g.Expect(volumeClaim.Spec.VolumeRef.Name).To(Equal(""))
		}, timeout, interval).Should(Succeed())
	})

	It("Should not claim a volume when the storageclasses are different", func() {
		By("creating a volume w/ a set of resources")
		Expect(k8sClient.Create(ctx, volume)).To(Succeed(), "failed to create volume")

		By("patching the volume status to available")
		volumeBase := volume.DeepCopy()
		volume.Status.State = storagev1alpha1.VolumeStateAvailable
		Expect(k8sClient.Status().Patch(ctx, volume, client.MergeFrom(volumeBase))).
			To(Succeed(), "failed to patch volume status")

		By("creating a volumeclaim which should claim the matching volume")
		volumeClaim.Spec.StorageClassRef = corev1.LocalObjectReference{
			Name: "my-volumeclass2",
		}
		Expect(k8sClient.Create(ctx, volumeClaim)).To(Succeed(), "failed to create volumeclaim")

		By("waiting for the volumeclaim to reference the volume")
		claimKey := client.ObjectKeyFromObject(volumeClaim)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, claimKey, volumeClaim)).To(Succeed(), "failed to get volumeclaim")
			g.Expect(volumeClaim.Spec.VolumeRef.Name).To(Equal(""))
		}, timeout, interval).Should(Succeed())

		By("waiting for the volume to reference the claim")
		volumeKey := client.ObjectKeyFromObject(volume)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, volumeKey, volume)).To(Succeed(), "failed to get volume")
			g.Expect(volume.Spec.ClaimRef.Name).To(Equal(""))
			g.Expect(volume.Spec.ClaimRef.UID).To(Equal(types.UID("")))
		}, timeout, interval).Should(Succeed())
	})

	It("Should claim one volume out of two where the resources match", func() {
		By("creating a 100Gi volume")
		Expect(k8sClient.Create(ctx, volume)).To(Succeed(), "failed to create volume")

		By("patching the volume status to available")
		volumeBase := volume.DeepCopy()
		volume.Status.State = storagev1alpha1.VolumeStateAvailable
		Expect(k8sClient.Status().Patch(ctx, volume, client.MergeFrom(volumeBase))).
			To(Succeed(), "failed to patch volume status")

		By("creating a 10Gi volume")
		Expect(k8sClient.Create(ctx, volume2)).To(Succeed(), "failed to create volume")

		By("patching the volume status to available")
		volumeBase2 := volume2.DeepCopy()
		volume2.Status.State = storagev1alpha1.VolumeStateAvailable
		Expect(k8sClient.Status().Patch(ctx, volume2, client.MergeFrom(volumeBase2))).
			To(Succeed(), "failed to patch volume status")

		By("creating a volumeclaim which should claim the matching volume")
		Expect(k8sClient.Create(ctx, volumeClaim)).To(Succeed(), "failed to create volumeclaim")

		By("waiting for the correct volume to reference the claim")
		volumeKey := client.ObjectKeyFromObject(volume)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, volumeKey, volume)).To(Succeed(), "failed to get volume")
			g.Expect(volume.Spec.ClaimRef.Name).To(Equal(volumeClaim.Name))
			g.Expect(volume.Spec.ClaimRef.UID).To(Equal(volumeClaim.UID))
		}, timeout, interval).Should(Succeed())

		By("waiting for the incorrect volume to not be claimed")
		volumeKey2 := client.ObjectKeyFromObject(volume2)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, volumeKey2, volume2)).To(Succeed(), "failed to get volume")
			g.Expect(volume2.Spec.ClaimRef.Name).To(Equal(""))
			g.Expect(volume2.Spec.ClaimRef.UID).To(Equal(types.UID("")))
		}, timeout, interval).Should(Succeed())

		By("waiting for the volumeclaim to reference the volume")
		claimKey := client.ObjectKeyFromObject(volumeClaim)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, claimKey, volumeClaim)).To(Succeed(), "failed to get volumeclaim")
			g.Expect(volumeClaim.Spec.VolumeRef.Name).To(Equal(volume.Name))
		}, timeout, interval).Should(Succeed())
	})

	It("Should not claim a volume when the volumeref is set", func() {
		By("creating a volume w/ a set of resources")
		volume.Spec.ClaimRef = storagev1alpha1.ClaimReference{
			Name: "my-volume",
			UID:  "12345",
		}
		Expect(k8sClient.Create(ctx, volume)).To(Succeed(), "failed to create volume")

		By("patching the volume status to available")
		volumeBase := volume.DeepCopy()
		volume.Status.State = storagev1alpha1.VolumeStateAvailable
		Expect(k8sClient.Status().Patch(ctx, volume, client.MergeFrom(volumeBase))).
			To(Succeed(), "failed to patch volume status")

		By("creating a volumeclaim w/ a volumeref")
		Expect(k8sClient.Create(ctx, volumeClaim)).To(Succeed(), "failed to create volumeclaim")

		By("waiting for the volumeclaim to reference the volume")
		claimKey := client.ObjectKeyFromObject(volumeClaim)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, claimKey, volumeClaim)).To(Succeed(), "failed to get volumeclaim")
			g.Expect(volumeClaim.Spec.VolumeRef.Name).To(Equal(""))
		}, timeout, interval).Should(Succeed())
	})
})
