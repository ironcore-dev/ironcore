// Copyright 2021 OnMetal authors
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

package storage

import (
	. "github.com/onmetal/onmetal-api/utils/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1alpha1 "github.com/onmetal/onmetal-api/api/core/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
)

var _ = Describe("BucketScheduler", func() {
	ctx := SetupContext()
	ns, _ := SetupTest(ctx)

	It("should schedule buckets on bucket pools", func() {
		By("creating a bucket pool")
		bucketPool := &storagev1alpha1.BucketPool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-pool-",
			},
		}
		Expect(k8sClient.Create(ctx, bucketPool)).To(Succeed(), "failed to create bucket pool")

		By("patching the bucket pool status to contain a bucket class")
		bucketPoolBase := bucketPool.DeepCopy()
		bucketPool.Status.AvailableBucketClasses = []corev1.LocalObjectReference{{Name: "my-bucketclass"}}
		Expect(k8sClient.Status().Patch(ctx, bucketPool, client.MergeFrom(bucketPoolBase))).
			To(Succeed(), "failed to patch bucket pool status")

		By("creating a bucket w/ the requested bucket class")
		bucket := &storagev1alpha1.Bucket{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-bucket-",
			},
			Spec: storagev1alpha1.BucketSpec{
				BucketClassRef: &corev1.LocalObjectReference{Name: "my-bucketclass"},
			},
		}
		Expect(k8sClient.Create(ctx, bucket)).To(Succeed(), "failed to create bucket")

		By("waiting for the bucket to be scheduled onto the bucket pool")
		bucketKey := client.ObjectKeyFromObject(bucket)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, bucketKey, bucket)).To(Succeed(), "failed to get bucket")
			g.Expect(bucket.Spec.BucketPoolRef).To(Equal(&corev1.LocalObjectReference{Name: bucketPool.Name}))
		}).Should(Succeed())
	})

	It("should schedule schedule buckets onto bucket pools if the pool becomes available later than the bucket", func() {
		By("creating a bucket w/ the requested bucket class")
		bucket := &storagev1alpha1.Bucket{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-bucket-",
			},
			Spec: storagev1alpha1.BucketSpec{
				BucketClassRef: &corev1.LocalObjectReference{Name: "my-bucketclass"},
			},
		}
		Expect(k8sClient.Create(ctx, bucket)).To(Succeed(), "failed to create bucket")

		bucketPools := &storagev1alpha1.BucketPoolList{}
		Expect(k8sClient.List(ctx, bucketPools)).To(Succeed(), "failed to create bucket pool")

		By("waiting for the bucket to indicate it is pending")
		bucketKey := client.ObjectKeyFromObject(bucket)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, bucketKey, bucket)).To(Succeed())
			g.Expect(bucket.Spec.BucketPoolRef).To(BeNil())
		}).Should(Succeed())

		By("creating a bucket pool")
		bucketPool := &storagev1alpha1.BucketPool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-pool-",
			},
		}
		Expect(k8sClient.Create(ctx, bucketPool)).To(Succeed(), "failed to create bucket pool")

		By("patching the bucket pool status to contain a bucket class")
		bucketPoolBase := bucketPool.DeepCopy()
		bucketPool.Status.AvailableBucketClasses = []corev1.LocalObjectReference{{Name: "my-bucketclass"}}
		Expect(k8sClient.Status().Patch(ctx, bucketPool, client.MergeFrom(bucketPoolBase))).
			To(Succeed(), "failed to patch bucket pool status")

		By("waiting for the bucket to be scheduled onto the bucket pool")
		Eventually(func() *corev1.LocalObjectReference {
			Expect(k8sClient.Get(ctx, bucketKey, bucket)).To(Succeed(), "failed to get bucket")
			return bucket.Spec.BucketPoolRef
		}).Should(Equal(&corev1.LocalObjectReference{Name: bucketPool.Name}))
	})

	It("should schedule onto bucket pools with matching labels", func() {
		By("creating a bucket pool w/o matching labels")
		bucketPoolNoMatchingLabels := &storagev1alpha1.BucketPool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-pool-",
			},
		}
		Expect(k8sClient.Create(ctx, bucketPoolNoMatchingLabels)).To(Succeed(), "failed to create bucket pool")

		By("patching the bucket pool status to contain a bucket class")
		bucketPoolNoMatchingLabelsBase := bucketPoolNoMatchingLabels.DeepCopy()
		bucketPoolNoMatchingLabels.Status.AvailableBucketClasses = []corev1.LocalObjectReference{{Name: "my-bucketclass"}}
		Expect(k8sClient.Status().Patch(ctx, bucketPoolNoMatchingLabels, client.MergeFrom(bucketPoolNoMatchingLabelsBase))).
			To(Succeed(), "failed to patch bucket pool status")

		By("creating a bucket pool w/ matching labels")
		bucketPoolMatchingLabels := &storagev1alpha1.BucketPool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-pool-",
				Labels: map[string]string{
					"foo": "bar",
				},
			},
		}
		Expect(k8sClient.Create(ctx, bucketPoolMatchingLabels)).To(Succeed(), "failed to create bucket pool")

		By("patching the bucket pool status to contain a bucket class")
		bucketPoolMatchingLabelsBase := bucketPoolMatchingLabels.DeepCopy()
		bucketPoolMatchingLabels.Status.AvailableBucketClasses = []corev1.LocalObjectReference{{Name: "my-bucketclass"}}
		Expect(k8sClient.Status().Patch(ctx, bucketPoolMatchingLabels, client.MergeFrom(bucketPoolMatchingLabelsBase))).
			To(Succeed(), "failed to patch bucket pool status")

		By("creating a bucket w/ the requested bucket class")
		bucket := &storagev1alpha1.Bucket{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-bucket-",
			},
			Spec: storagev1alpha1.BucketSpec{
				BucketPoolSelector: map[string]string{
					"foo": "bar",
				},
				BucketClassRef: &corev1.LocalObjectReference{Name: "my-bucketclass"},
			},
		}
		Expect(k8sClient.Create(ctx, bucket)).To(Succeed(), "failed to create bucket")

		By("waiting for the bucket to be scheduled onto the bucket pool")
		bucketKey := client.ObjectKeyFromObject(bucket)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, bucketKey, bucket)).To(Succeed(), "failed to get bucket")
			g.Expect(bucket.Spec.BucketPoolRef).To(Equal(&corev1.LocalObjectReference{Name: bucketPoolMatchingLabels.Name}))
		}).Should(Succeed())
	})

	It("should schedule a bucket with corresponding tolerations onto a bucket pool with taints", func() {
		By("creating a bucket pool w/ taints")
		taintedBucketPool := &storagev1alpha1.BucketPool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-pool-",
			},
			Spec: storagev1alpha1.BucketPoolSpec{
				Taints: []corev1alpha1.Taint{
					{
						Key:    "key",
						Value:  "value",
						Effect: corev1alpha1.TaintEffectNoSchedule,
					},
					{
						Key:    "key1",
						Effect: corev1alpha1.TaintEffectNoSchedule,
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, taintedBucketPool)).To(Succeed(), "failed to create the bucket pool")

		By("patching the bucket pool status to contain a bucket class")
		bucketPoolBase := taintedBucketPool.DeepCopy()
		taintedBucketPool.Status.AvailableBucketClasses = []corev1.LocalObjectReference{{Name: "my-bucketclass"}}
		Expect(k8sClient.Status().Patch(ctx, taintedBucketPool, client.MergeFrom(bucketPoolBase))).
			To(Succeed(), "failed to patch the bucket pool status")

		By("creating a bucket")
		bucket := &storagev1alpha1.Bucket{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-bucket-",
			},
			Spec: storagev1alpha1.BucketSpec{
				BucketClassRef: &corev1.LocalObjectReference{Name: "my-bucketclass"},
			},
		}
		Expect(k8sClient.Create(ctx, bucket)).To(Succeed(), "failed to create the bucket")

		By("observing the bucket isn't scheduled onto the bucket pool")
		bucketKey := client.ObjectKeyFromObject(bucket)
		Consistently(func() *corev1.LocalObjectReference {
			Expect(k8sClient.Get(ctx, bucketKey, bucket)).To(Succeed())
			return bucket.Spec.BucketPoolRef
		}).Should(BeNil())

		By("patching the bucket to contain only one of the corresponding tolerations")
		bucketBase := bucket.DeepCopy()
		bucket.Spec.Tolerations = append(bucket.Spec.Tolerations, corev1alpha1.Toleration{
			Key:      "key",
			Value:    "value",
			Effect:   corev1alpha1.TaintEffectNoSchedule,
			Operator: corev1alpha1.TolerationOpEqual,
		})
		Expect(k8sClient.Patch(ctx, bucket, client.MergeFrom(bucketBase))).To(Succeed(), "failed to patch the bucket's spec")

		By("observing the bucket isn't scheduled onto the bucket pool")
		Consistently(func() *corev1.LocalObjectReference {
			Expect(k8sClient.Get(ctx, bucketKey, bucket)).To(Succeed())
			return bucket.Spec.BucketPoolRef
		}).Should(BeNil())

		By("patching the bucket to contain all of the corresponding tolerations")
		bucketBase = bucket.DeepCopy()
		bucket.Spec.Tolerations = append(bucket.Spec.Tolerations, corev1alpha1.Toleration{
			Key:      "key1",
			Effect:   corev1alpha1.TaintEffectNoSchedule,
			Operator: corev1alpha1.TolerationOpExists,
		})
		Expect(k8sClient.Patch(ctx, bucket, client.MergeFrom(bucketBase))).To(Succeed(), "failed to patch the bucket's spec")

		By("observing the bucket is scheduled onto the bucket pool")
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, bucketKey, bucket)).To(Succeed(), "failed to get the bucket")
			g.Expect(bucket.Spec.BucketPoolRef).To(Equal(&corev1.LocalObjectReference{Name: taintedBucketPool.Name}))
		}).Should(Succeed())
	})
})
