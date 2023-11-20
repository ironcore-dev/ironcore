// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	. "github.com/ironcore-dev/ironcore/utils/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"

	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
)

var _ = Describe("BucketScheduler", func() {
	ns := SetupNamespace(&k8sClient)

	BeforeEach(func(ctx SpecContext) {
		By("waiting for the cached client to report no bucket pools")
		Eventually(New(cacheK8sClient).ObjectList(&storagev1alpha1.BucketPoolList{})).Should(HaveField("Items", BeEmpty()))
	})

	AfterEach(func(ctx SpecContext) {
		By("deleting all bucket pools")
		Expect(k8sClient.DeleteAllOf(ctx, &storagev1alpha1.BucketPool{})).To(Succeed())
	})

	It("should schedule buckets on bucket pools", func(ctx SpecContext) {
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

	It("should schedule schedule buckets onto bucket pools if the pool becomes available later than the bucket", func(ctx SpecContext) {
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

	It("should schedule onto bucket pools with matching labels", func(ctx SpecContext) {
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

	It("should schedule a bucket with corresponding tolerations onto a bucket pool with taints", func(ctx SpecContext) {
		By("creating a bucket pool w/ taints")
		taintedBucketPool := &storagev1alpha1.BucketPool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-pool-",
			},
			Spec: storagev1alpha1.BucketPoolSpec{
				Taints: []commonv1alpha1.Taint{
					{
						Key:    "key",
						Value:  "value",
						Effect: commonv1alpha1.TaintEffectNoSchedule,
					},
					{
						Key:    "key1",
						Effect: commonv1alpha1.TaintEffectNoSchedule,
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
		bucket.Spec.Tolerations = append(bucket.Spec.Tolerations, commonv1alpha1.Toleration{
			Key:      "key",
			Value:    "value",
			Effect:   commonv1alpha1.TaintEffectNoSchedule,
			Operator: commonv1alpha1.TolerationOpEqual,
		})
		Expect(k8sClient.Patch(ctx, bucket, client.MergeFrom(bucketBase))).To(Succeed(), "failed to patch the bucket's spec")

		By("observing the bucket isn't scheduled onto the bucket pool")
		Consistently(func() *corev1.LocalObjectReference {
			Expect(k8sClient.Get(ctx, bucketKey, bucket)).To(Succeed())
			return bucket.Spec.BucketPoolRef
		}).Should(BeNil())

		By("patching the bucket to contain all of the corresponding tolerations")
		bucketBase = bucket.DeepCopy()
		bucket.Spec.Tolerations = append(bucket.Spec.Tolerations, commonv1alpha1.Toleration{
			Key:      "key1",
			Effect:   commonv1alpha1.TaintEffectNoSchedule,
			Operator: commonv1alpha1.TolerationOpExists,
		})
		Expect(k8sClient.Patch(ctx, bucket, client.MergeFrom(bucketBase))).To(Succeed(), "failed to patch the bucket's spec")

		By("observing the bucket is scheduled onto the bucket pool")
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, bucketKey, bucket)).To(Succeed(), "failed to get the bucket")
			g.Expect(bucket.Spec.BucketPoolRef).To(Equal(&corev1.LocalObjectReference{Name: taintedBucketPool.Name}))
		}).Should(Succeed())
	})
})
