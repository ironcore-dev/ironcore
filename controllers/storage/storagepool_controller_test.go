///*
// * Copyright (c) 2021 by the OnMetal authors.
// *
// * Licensed under the Apache License, Version 2.0 (the "License");
// * you may not use this file except in compliance with the License.
// * You may obtain a copy of the License at
// *
// *     http://www.apache.org/licenses/LICENSE-2.0
// *
// * Unless required by applicable law or agreed to in writing, software
// * distributed under the License is distributed on an "AS IS" BASIS,
// * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// * See the License for the specific language governing permissions and
// * limitations under the License.
// */
//
package storage

import (
	"time"

	corev1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	storagev1alpha1 "github.com/onmetal/onmetal-api/apis/storage/v1alpha1"
)

var _ = Describe("StoragePoolReconciler", func() {
	Context("Reconcile a StoragePool", func() {
		ns := SetupTest(ctx)

		It("should set state as NotAvailable when Ready condition is not present", func() {
			storagePool := &storagev1alpha1.StoragePool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "no-ready-condition",
					Namespace: ns.Name,
				},
			}
			Expect(k8sClient.Create(ctx, storagePool)).To(Succeed())

			By("checking that StoragePool is in pending state")
			Eventually(func(g Gomega) {
				key := types.NamespacedName{
					Name:      storagePool.Name,
					Namespace: ns.Name,
				}
				obj := &storagev1alpha1.StoragePool{}
				err := k8sClient.Get(ctx, key, obj)
				Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
				g.Expect(err).NotTo(HaveOccurred())

				g.Expect(obj.Status.State).To(Equal(storagev1alpha1.StoragePoolStateNotAvailable))
			}, timeout, interval).Should(Succeed())
		})
		It("should set state as Pending when Ready condition is outdated", func() {
			storagePool := &storagev1alpha1.StoragePool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ready-out-of-date",
					Namespace: ns.Name,
				},
			}
			Expect(k8sClient.Create(ctx, storagePool)).To(Succeed())

			storagePool.Status.Conditions = []storagev1alpha1.StoragePoolCondition{
				{
					Type:               storagev1alpha1.StoragePoolConditionTypeReady,
					Status:             corev1.ConditionTrue,
					LastUpdateTime:     metav1.Time{Time: time.Now().Add(time.Duration(-1) * storagePoolGracePeriod)},
					LastTransitionTime: metav1.Now(),
				},
			}
			Expect(k8sClient.Status().Update(ctx, storagePool)).To(Succeed())

			By("checking that StoragePool is in pending state")
			Eventually(func(g Gomega) {
				key := types.NamespacedName{
					Name:      storagePool.Name,
					Namespace: ns.Name,
				}
				obj := &storagev1alpha1.StoragePool{}
				err := k8sClient.Get(ctx, key, obj)
				Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
				g.Expect(err).NotTo(HaveOccurred())

				g.Expect(obj.Status.State).To(Equal(storagev1alpha1.StoragePoolStatePending))
			}, timeout, interval).Should(Succeed())
		})
		It("should set state to Available when Ready condition is up-to-date", func() {
			storagePool := &storagev1alpha1.StoragePool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ready",
					Namespace: ns.Name,
				},
			}
			Expect(k8sClient.Create(ctx, storagePool)).To(Succeed())

			storagePool.Status.Conditions = []storagev1alpha1.StoragePoolCondition{
				{
					Type:               storagev1alpha1.StoragePoolConditionTypeReady,
					Status:             corev1.ConditionTrue,
					LastUpdateTime:     metav1.Now(),
					LastTransitionTime: metav1.Now(),
				},
			}
			Expect(k8sClient.Status().Update(ctx, storagePool)).To(Succeed())

			By("checking that StoragePool is in available state")
			Eventually(func(g Gomega) {
				key := types.NamespacedName{
					Name:      storagePool.Name,
					Namespace: ns.Name,
				}
				obj := &storagev1alpha1.StoragePool{}
				err := k8sClient.Get(ctx, key, obj)
				Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
				g.Expect(err).NotTo(HaveOccurred())

				g.Expect(obj.Status.State).To(Equal(storagev1alpha1.StoragePoolStateAvailable))
			}, timeout, interval).Should(Succeed())
		})
	})
})
