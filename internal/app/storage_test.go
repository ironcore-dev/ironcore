// Copyright 2022 IronCore authors
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

package app_test

import (
	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	. "github.com/ironcore-dev/ironcore/utils/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Storage", func() {
	var (
		ctx         = SetupContext()
		ns          = SetupTest(ctx)
		volumeClass = &storagev1alpha1.VolumeClass{}
		bucketClass = &storagev1alpha1.BucketClass{}
	)

	BeforeEach(func() {
		*volumeClass = storagev1alpha1.VolumeClass{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "volume-class-",
			},
			Capabilities: corev1alpha1.ResourceList{
				corev1alpha1.ResourceTPS:  resource.MustParse("50"),
				corev1alpha1.ResourceIOPS: resource.MustParse("3000"),
			},
		}
		Expect(k8sClient.Create(ctx, volumeClass)).To(Succeed(), "failed to create test volume class")
		DeferCleanup(k8sClient.Delete, ctx, volumeClass)

		*bucketClass = storagev1alpha1.BucketClass{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "bucket-class-",
			},
			Capabilities: corev1alpha1.ResourceList{
				corev1alpha1.ResourceTPS:  resource.MustParse("50"),
				corev1alpha1.ResourceIOPS: resource.MustParse("3000"),
			},
		}
		Expect(k8sClient.Create(ctx, bucketClass)).To(Succeed(), "failed to create test bucket class")
		DeferCleanup(k8sClient.Delete, ctx, bucketClass)
	})

	Context("Volume", func() {
		It("should allow listing volumes filtering by volume pool name", func() {
			const (
				volumePool1 = "volume-pool-1"
				volumePool2 = "volume-pool-2"
			)

			By("creating a volume on volume pool 1")
			volume1 := &storagev1alpha1.Volume{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    ns.Name,
					GenerateName: "volume-",
				},
				Spec: storagev1alpha1.VolumeSpec{
					VolumeClassRef: &corev1.LocalObjectReference{Name: "my-class"},
					VolumePoolRef:  &corev1.LocalObjectReference{Name: volumePool1},
					Resources: corev1alpha1.ResourceList{
						corev1alpha1.ResourceStorage: resource.MustParse("10Gi"),
					},
				},
			}
			Expect(k8sClient.Create(ctx, volume1)).To(Succeed())

			By("creating a volume on volume pool 2")
			volume2 := &storagev1alpha1.Volume{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    ns.Name,
					GenerateName: "volume-",
				},
				Spec: storagev1alpha1.VolumeSpec{
					VolumeClassRef: &corev1.LocalObjectReference{Name: "my-class"},
					VolumePoolRef:  &corev1.LocalObjectReference{Name: volumePool2},
					Resources: corev1alpha1.ResourceList{
						corev1alpha1.ResourceStorage: resource.MustParse("10Gi"),
					},
				},
			}
			Expect(k8sClient.Create(ctx, volume2)).To(Succeed())

			By("creating a volume on no volume pool")
			volume3 := &storagev1alpha1.Volume{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    ns.Name,
					GenerateName: "volume-",
				},
				Spec: storagev1alpha1.VolumeSpec{
					VolumeClassRef: &corev1.LocalObjectReference{Name: "my-class"},
					Resources: corev1alpha1.ResourceList{
						corev1alpha1.ResourceStorage: resource.MustParse("10Gi"),
					},
				},
			}
			Expect(k8sClient.Create(ctx, volume3)).To(Succeed())

			By("listing all volumes on volume pool 1")
			volumesOnVolumePool1List := &storagev1alpha1.VolumeList{}
			Expect(k8sClient.List(ctx, volumesOnVolumePool1List,
				client.InNamespace(ns.Name),
				client.MatchingFields{storagev1alpha1.VolumeVolumePoolRefNameField: volumePool1},
			)).To(Succeed())

			By("inspecting the items")
			Expect(volumesOnVolumePool1List.Items).To(ConsistOf(*volume1))

			By("listing all volumes on volume pool 2")
			volumesOnVolumePool2List := &storagev1alpha1.VolumeList{}
			Expect(k8sClient.List(ctx, volumesOnVolumePool2List,
				client.InNamespace(ns.Name),
				client.MatchingFields{storagev1alpha1.VolumeVolumePoolRefNameField: volumePool2},
			)).To(Succeed())

			By("inspecting the items")
			Expect(volumesOnVolumePool2List.Items).To(ConsistOf(*volume2))

			By("listing all volumes on no volume pool")
			volumesOnNoVolumePoolList := &storagev1alpha1.VolumeList{}
			Expect(k8sClient.List(ctx, volumesOnNoVolumePoolList,
				client.InNamespace(ns.Name),
				client.MatchingFields{storagev1alpha1.VolumeVolumePoolRefNameField: ""},
			)).To(Succeed())

			By("inspecting the items")
			Expect(volumesOnNoVolumePoolList.Items).To(ConsistOf(*volume3))
		})

		It("should allow listing volumes by volume class name", func() {
			By("creating another volume class")
			volumeClass2 := &storagev1alpha1.VolumeClass{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "volume-class-",
				},
				Capabilities: corev1alpha1.ResourceList{
					corev1alpha1.ResourceTPS:  resource.MustParse("30"),
					corev1alpha1.ResourceIOPS: resource.MustParse("1000"),
				},
			}
			Expect(k8sClient.Create(ctx, volumeClass2)).To(Succeed())
			DeferCleanup(k8sClient.Delete, ctx, volumeClass2)

			By("creating a volume")
			volume1 := &storagev1alpha1.Volume{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    ns.Name,
					GenerateName: "volume-",
				},
				Spec: storagev1alpha1.VolumeSpec{
					VolumeClassRef: &corev1.LocalObjectReference{Name: volumeClass.Name},
					Resources: corev1alpha1.ResourceList{
						corev1alpha1.ResourceStorage: resource.MustParse("10Gi"),
					},
				},
			}
			Expect(k8sClient.Create(ctx, volume1)).To(Succeed())

			By("creating a volume with the other volume class")
			volume2 := &storagev1alpha1.Volume{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    ns.Name,
					GenerateName: "volume-",
				},
				Spec: storagev1alpha1.VolumeSpec{
					VolumeClassRef: &corev1.LocalObjectReference{Name: volumeClass2.Name},
					Resources: corev1alpha1.ResourceList{
						corev1alpha1.ResourceStorage: resource.MustParse("10Gi"),
					},
				},
			}
			Expect(k8sClient.Create(ctx, volume2)).To(Succeed())

			By("listing volumes with the first volume class name")
			volumeList := &storagev1alpha1.VolumeList{}
			Expect(k8sClient.List(ctx, volumeList, client.MatchingFields{
				storagev1alpha1.VolumeVolumeClassRefNameField: volumeClass.Name,
			})).To(Succeed())

			By("inspecting the retrieved list to only have the volume with the correct volume class")
			Expect(volumeList.Items).To(ConsistOf(HaveField("UID", volume1.UID)))

			By("listing volumes with the second volume class name")
			Expect(k8sClient.List(ctx, volumeList, client.MatchingFields{
				storagev1alpha1.VolumeVolumeClassRefNameField: volumeClass2.Name,
			})).To(Succeed())

			By("inspecting the retrieved list to only have the volume with the correct volume class")
			Expect(volumeList.Items).To(ConsistOf(HaveField("UID", volume2.UID)))
		})
	})

	Context("Bucket", func() {
		It("should allow listing buckets filtering by bucket pool name", func() {
			const (
				bucketPool1 = "bucket-pool-1"
				bucketPool2 = "bucket-pool-2"
			)

			By("creating a bucket on bucket pool 1")
			bucket1 := &storagev1alpha1.Bucket{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    ns.Name,
					GenerateName: "bucket-",
				},
				Spec: storagev1alpha1.BucketSpec{
					BucketClassRef: &corev1.LocalObjectReference{Name: "my-class"},
					BucketPoolRef:  &corev1.LocalObjectReference{Name: bucketPool1},
				},
			}
			Expect(k8sClient.Create(ctx, bucket1)).To(Succeed())

			By("creating a bucket on bucket pool 2")
			bucket2 := &storagev1alpha1.Bucket{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    ns.Name,
					GenerateName: "bucket-",
				},
				Spec: storagev1alpha1.BucketSpec{
					BucketClassRef: &corev1.LocalObjectReference{Name: "my-class"},
					BucketPoolRef:  &corev1.LocalObjectReference{Name: bucketPool2},
				},
			}
			Expect(k8sClient.Create(ctx, bucket2)).To(Succeed())

			By("creating a bucket on no bucket pool")
			bucket3 := &storagev1alpha1.Bucket{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    ns.Name,
					GenerateName: "bucket-",
				},
				Spec: storagev1alpha1.BucketSpec{
					BucketClassRef: &corev1.LocalObjectReference{Name: "my-class"},
				},
			}
			Expect(k8sClient.Create(ctx, bucket3)).To(Succeed())

			By("listing all buckets on bucket pool 1")
			bucketsOnBucketPool1List := &storagev1alpha1.BucketList{}
			Expect(k8sClient.List(ctx, bucketsOnBucketPool1List,
				client.InNamespace(ns.Name),
				client.MatchingFields{storagev1alpha1.BucketBucketPoolRefNameField: bucketPool1},
			)).To(Succeed())

			By("inspecting the items")
			Expect(bucketsOnBucketPool1List.Items).To(ConsistOf(*bucket1))

			By("listing all buckets on bucket pool 2")
			bucketsOnBucketPool2List := &storagev1alpha1.BucketList{}
			Expect(k8sClient.List(ctx, bucketsOnBucketPool2List,
				client.InNamespace(ns.Name),
				client.MatchingFields{storagev1alpha1.BucketBucketPoolRefNameField: bucketPool2},
			)).To(Succeed())

			By("inspecting the items")
			Expect(bucketsOnBucketPool2List.Items).To(ConsistOf(*bucket2))

			By("listing all buckets on no bucket pool")
			bucketsOnNoBucketPoolList := &storagev1alpha1.BucketList{}
			Expect(k8sClient.List(ctx, bucketsOnNoBucketPoolList,
				client.InNamespace(ns.Name),
				client.MatchingFields{storagev1alpha1.BucketBucketPoolRefNameField: ""},
			)).To(Succeed())

			By("inspecting the items")
			Expect(bucketsOnNoBucketPoolList.Items).To(ConsistOf(*bucket3))
		})

		It("should allow listing buckets by bucket class name", func() {
			By("creating another bucket class")
			bucketClass2 := &storagev1alpha1.BucketClass{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "bucket-class-",
				},
				Capabilities: corev1alpha1.ResourceList{
					corev1alpha1.ResourceTPS:  resource.MustParse("30"),
					corev1alpha1.ResourceIOPS: resource.MustParse("1000"),
				},
			}
			Expect(k8sClient.Create(ctx, bucketClass2)).To(Succeed())
			DeferCleanup(k8sClient.Delete, ctx, bucketClass2)

			By("creating a bucket")
			bucket1 := &storagev1alpha1.Bucket{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    ns.Name,
					GenerateName: "bucket-",
				},
				Spec: storagev1alpha1.BucketSpec{
					BucketClassRef: &corev1.LocalObjectReference{Name: bucketClass.Name},
				},
			}
			Expect(k8sClient.Create(ctx, bucket1)).To(Succeed())

			By("creating a bucket with the other bucket class")
			bucket2 := &storagev1alpha1.Bucket{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    ns.Name,
					GenerateName: "bucket-",
				},
				Spec: storagev1alpha1.BucketSpec{
					BucketClassRef: &corev1.LocalObjectReference{Name: bucketClass2.Name},
				},
			}
			Expect(k8sClient.Create(ctx, bucket2)).To(Succeed())

			By("listing buckets with the first bucket class name")
			bucketList := &storagev1alpha1.BucketList{}
			Expect(k8sClient.List(ctx, bucketList, client.MatchingFields{
				storagev1alpha1.BucketBucketClassRefNameField: bucketClass.Name,
			})).To(Succeed())

			By("inspecting the retrieved list to only have the bucket with the correct bucket class")
			Expect(bucketList.Items).To(ConsistOf(HaveField("UID", bucket1.UID)))

			By("listing buckets with the second bucket class name")
			Expect(k8sClient.List(ctx, bucketList, client.MatchingFields{
				storagev1alpha1.BucketBucketClassRefNameField: bucketClass2.Name,
			})).To(Succeed())

			By("inspecting the retrieved list to only have the bucket with the correct bucket class")
			Expect(bucketList.Items).To(ConsistOf(HaveField("UID", bucket2.UID)))
		})
	})
})
