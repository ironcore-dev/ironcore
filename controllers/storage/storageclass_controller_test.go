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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	storagev1alpha1 "github.com/onmetal/onmetal-api/apis/storage/v1alpha1"
)

var _ = Describe("storageclass controller", func() {
	ns := SetupTest(ctx)
	It("should finalize the storageclass if no volume is using it", func() {
		By("creating the storageclass consumed by the volume")
		storageClass := &storagev1alpha1.StorageClass{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "storageclass-",
			},
			Spec: storagev1alpha1.StorageClassSpec{},
		}
		Expect(k8sClient.Create(ctx, storageClass)).Should(Succeed())

		By("creating the volume")
		volume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{
				StorageClassRef: corev1.LocalObjectReference{
					Name: storageClass.Name,
				},
			},
		}
		Expect(k8sClient.Create(ctx, volume)).Should(Succeed())

		By("checking the finalizer is present")
		storageClassKey := client.ObjectKeyFromObject(storageClass)
		Eventually(func(g Gomega) []string {
			err := k8sClient.Get(ctx, storageClassKey, storageClass)
			Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
			g.Expect(err).NotTo(HaveOccurred())
			return storageClass.Finalizers
		}, timeout, interval).Should(ContainElement(storagev1alpha1.StorageClassFinalizer))

		By("issuing a delete request for the storage class")
		Expect(k8sClient.Delete(ctx, storageClass)).Should(Succeed())

		By("asserting the storage class is still present as the volume is referencing it")
		Consistently(func(g Gomega) []string {
			err := k8sClient.Get(ctx, storageClassKey, storageClass)
			g.Expect(err).NotTo(HaveOccurred())
			return storageClass.Finalizers
		}, timeout).Should(ContainElement(storagev1alpha1.StorageClassFinalizer))

		By("deleting the referencing volume")
		Expect(k8sClient.Delete(ctx, volume)).Should(Succeed())

		By("waiting for the storage class to be gone")
		Eventually(func() error {
			return k8sClient.Get(ctx, storageClassKey, storageClass)
		}, timeout, interval).Should(MatchErrorFunc(apierrors.IsNotFound))
	})
})
