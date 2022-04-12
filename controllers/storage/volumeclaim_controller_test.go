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
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("VolumeClaimReconciler", func() {
	ns := SetupTest(ctx)

	var volume *storagev1alpha1.Volume
	var volumeClaim *storagev1alpha1.VolumeClaim

	BeforeEach(func() {
		volume = &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{
				VolumePoolRef: corev1.LocalObjectReference{
					Name: "my-volumepool",
				},
				Resources: map[corev1.ResourceName]resource.Quantity{
					"storage": resource.MustParse("100Gi"),
				},
				VolumeClassRef: corev1.LocalObjectReference{
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
				VolumeClassRef: corev1.LocalObjectReference{
					Name: "my-volumeclass",
				},
			},
		}
	})

	It("Should bound a volumeclaim if the volume has the correct claim ref", func() {
		By("creating a volume w/ a set of resources")
		Expect(k8sClient.Create(ctx, volume)).To(Succeed(), "failed to create volume")

		By("patching the volume status to available")
		volumeBase := volume.DeepCopy()
		volume.Status.State = storagev1alpha1.VolumeStateAvailable
		Expect(k8sClient.Status().Patch(ctx, volume, client.MergeFrom(volumeBase))).
			To(Succeed(), "failed to patch volume status")

		By("creating a volumeclaim which should claim the matching volume")
		Expect(k8sClient.Create(ctx, volumeClaim)).To(Succeed(), "failed to create volumeclaim")

		By("waiting for the volumeclaim phase to become bound")
		claimKey := client.ObjectKeyFromObject(volumeClaim)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, claimKey, volumeClaim)).To(Succeed(), "failed to get volumeclaim")
			g.Expect(volumeClaim.Status.Phase).To(Equal(storagev1alpha1.VolumeClaimBound))
		}, timeout, interval).Should(Succeed())
	})

	It("Should un-bind a volumeclaim if the underlying volume is deleted", func() {
		By("creating a volume w/ a set of resources")
		Expect(k8sClient.Create(ctx, volume)).To(Succeed(), "failed to create volume")

		By("patching the volume status to available")
		volumeBase := volume.DeepCopy()
		volume.Status.State = storagev1alpha1.VolumeStateAvailable
		Expect(k8sClient.Status().Patch(ctx, volume, client.MergeFrom(volumeBase))).
			To(Succeed(), "failed to patch volume status")

		By("creating a volumeclaim which should claim the matching volume")
		Expect(k8sClient.Create(ctx, volumeClaim)).To(Succeed(), "failed to create volumeclaim")

		By("waiting for the volume phase to become bound")
		volumeClaimKey := client.ObjectKeyFromObject(volumeClaim)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, volumeClaimKey, volumeClaim)).To(Succeed(), "failed to get volumeclaim")
			g.Expect(volumeClaim.Spec.VolumeRef.Name).To(Equal(volume.Name))
			g.Expect(volumeClaim.Status.Phase).To(Equal(storagev1alpha1.VolumeClaimBound))
		}, timeout, interval).Should(Succeed())

		By("deleting the volume")
		Expect(k8sClient.Delete(ctx, volume)).To(Succeed(), "failed to delete volume")

		By("waiting for the volumeclaim phase to become lost")
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, volumeClaimKey, volumeClaim)).To(Succeed(), "failed to get volumeclaim")
			g.Expect(volumeClaim.Status.Phase).To(Equal(storagev1alpha1.VolumeClaimLost))
		}, timeout, interval).Should(Succeed())
	})
})
