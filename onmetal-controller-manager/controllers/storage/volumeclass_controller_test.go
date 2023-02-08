/*
 * Copyright (c) 2021 by the OnMetal authors.
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
	. "github.com/onmetal/controller-utils/testutils"
	. "github.com/onmetal/onmetal-api/utils/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
)

var _ = Describe("VolumeClass controller", func() {
	ctx := SetupContext()
	ns, _ := SetupTest(ctx)

	It("should finalize the volume class if no volume is using it", func() {
		By("creating the volume class consumed by the volume")
		volumeClass := &storagev1alpha1.VolumeClass{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "volumeclass-",
			},
			Capabilities: corev1.ResourceList{
				storagev1alpha1.ResourceTPS:  resource.MustParse("100Mi"),
				storagev1alpha1.ResourceIOPS: resource.MustParse("100"),
			},
		}
		Expect(k8sClient.Create(ctx, volumeClass)).Should(Succeed())

		By("creating the volume")
		volume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{
				VolumeClassRef: &corev1.LocalObjectReference{Name: volumeClass.Name},
				Resources: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("1Gi"),
				},
			},
		}
		Expect(k8sClient.Create(ctx, volume)).Should(Succeed())

		By("checking the finalizer is present")
		volumeClassKey := client.ObjectKeyFromObject(volumeClass)
		Eventually(func(g Gomega) []string {
			err := k8sClient.Get(ctx, volumeClassKey, volumeClass)
			Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
			g.Expect(err).NotTo(HaveOccurred())
			return volumeClass.Finalizers
		}).Should(ContainElement(storagev1alpha1.VolumeClassFinalizer))

		By("issuing a delete request for the volume class")
		Expect(k8sClient.Delete(ctx, volumeClass)).Should(Succeed())

		By("asserting the volume class is still present as the volume is referencing it")
		Consistently(func(g Gomega) []string {
			err := k8sClient.Get(ctx, volumeClassKey, volumeClass)
			g.Expect(err).NotTo(HaveOccurred())
			return volumeClass.Finalizers
		}).Should(ContainElement(storagev1alpha1.VolumeClassFinalizer))

		By("deleting the referencing volume")
		Expect(k8sClient.Delete(ctx, volume)).Should(Succeed())

		By("waiting for the volume class to be gone")
		Eventually(func() error {
			return k8sClient.Get(ctx, volumeClassKey, volumeClass)
		}).Should(MatchErrorFunc(apierrors.IsNotFound))
	})
})
