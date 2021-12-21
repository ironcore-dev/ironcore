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

package v1alpha1

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("volume validation webhook", func() {
	ns := SetupTest()
	Context("upon volume update", func() {
		It("signals error if storageclass is changed", func() {
			By("creating a storage pool")
			storagePool := &StoragePool{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "test-pool-",
				},
			}
			Expect(k8sClient.Create(ctx, storagePool)).To(Succeed(), "failed to create storage pool")

			By("patching the storage pool status to contain a storage class")
			storagePoolBase := storagePool.DeepCopy()
			storagePool.Status.AvailableStorageClasses = []corev1.LocalObjectReference{{Name: "my-volumeclass"}}
			Expect(k8sClient.Status().Patch(ctx, storagePool, client.MergeFrom(storagePoolBase))).
				To(Succeed(), "failed to patch storage pool status")

			By("creating a volume w/ the requested storage class")
			volume := &Volume{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    ns.Name,
					GenerateName: "test-volume-",
				},
				Spec: VolumeSpec{
					StorageClass: corev1.LocalObjectReference{
						Name: "my-volumeclass",
					},
				},
			}
			Expect(k8sClient.Create(ctx, volume)).To(Succeed(), "failed to create volume")
			newStorageClass := v1.LocalObjectReference{Name: "newclass"}
			volume.Spec.StorageClass = newStorageClass
			err := k8sClient.Update(ctx, volume)
			Expect(err).To(HaveOccurred())
			path := field.NewPath("spec")
			fieldErr := field.Invalid(path.Child("storageClass"), newStorageClass, "field is immutable")
			fieldErrList := field.ErrorList{fieldErr}
			Expect(err.Error()).To(ContainSubstring(fieldErrList.ToAggregate().Error()))
		})

		It("keeps silent when update storagepool once by scheduler", func() {
			volume := &Volume{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    ns.Name,
					GenerateName: "volume-",
				},
			}
			Expect(k8sClient.Create(ctx, volume)).To(Succeed())

			newStoragePoolName := "new_storagepool"
			volume.Spec.StoragePool.Name = newStoragePoolName
			err := k8sClient.Update(ctx, volume)
			Expect(err).NotTo(HaveOccurred())
		})

	})
})
