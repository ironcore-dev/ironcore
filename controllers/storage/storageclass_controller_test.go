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
	It("removes the finalizer from the storageclass only if there's no volume still using the storageclass", func() {
		By("creating the storageclass consumed by the volume")
		sc := &storagev1alpha1.StorageClass{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "storageclass-",
			},
			Spec: storagev1alpha1.StorageClassSpec{},
		}
		Expect(k8sClient.Create(ctx, sc)).Should(Succeed())

		By("creating the volume")
		vol := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{
				StorageClass: corev1.LocalObjectReference{
					Name: sc.Name,
				},
			},
		}
		Expect(k8sClient.Create(ctx, vol)).Should(Succeed())

		By("checking the finalizer was added")
		scKey := client.ObjectKeyFromObject(sc)
		Eventually(func(g Gomega) []string {
			err := k8sClient.Get(ctx, scKey, sc)
			Expect(client.IgnoreNotFound(err)).To(Succeed(), "errors other than `not found` are not expected")
			g.Expect(err).NotTo(HaveOccurred())
			return sc.Finalizers
		}, interval).Should(ContainElement(storagev1alpha1.StorageClassFinalizer))

		By("checking the storageclass and its finalizer consistently exist upon deletion ")
		Expect(k8sClient.Delete(ctx, sc)).Should(Succeed())

		Consistently(func(g Gomega) []string {
			err := k8sClient.Get(ctx, scKey, sc)
			Expect(client.IgnoreNotFound(err)).To(Succeed(), "errors other than `not found` are not expected")
			g.Expect(err).NotTo(HaveOccurred())
			return sc.Finalizers
		}, interval).Should(ContainElement(storagev1alpha1.StorageClassFinalizer))

		By("checking the storageclass is eventually gone after the deletion of the volume")
		Expect(k8sClient.Delete(ctx, vol)).Should(Succeed())
		Eventually(func() bool {
			err := k8sClient.Get(ctx, scKey, sc)
			Expect(client.IgnoreNotFound(err)).To(Succeed(), "error other than `not found` not expected")
			return apierrors.IsNotFound(err)
		}, timeout, interval).Should(BeTrue(), "`not found` expected")
	})
})
