// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package controllers_test

import (
	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/bucket/v1alpha1"
	"github.com/ironcore-dev/ironcore/iri/testing/bucket"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

var _ = Describe("BucketPoolController", func() {
	_, bucketPool, bucketClass, srv := SetupTest()

	It("should have the default bucket class in the pool", func(ctx SpecContext) {
		By("checking if the default bucket classes are present")
		Eventually(Object(bucketPool)).Should(SatisfyAll(
			HaveField("Status.AvailableBucketClasses", ContainElements([]corev1.LocalObjectReference{
				{
					Name: bucketClass.Name,
				},
			}))),
		)
	})

	It("should add bucket classes to the pool", func(ctx SpecContext) {
		By("creating a second bucket class")
		testBucketClass := &storagev1alpha1.BucketClass{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-bc-1-",
			},
			Capabilities: corev1alpha1.ResourceList{
				corev1alpha1.ResourceTPS:  resource.MustParse("250Mi"),
				corev1alpha1.ResourceIOPS: resource.MustParse("1000"),
			},
		}
		Expect(k8sClient.Create(ctx, testBucketClass)).To(Succeed(), "failed to create test bucket class")
		DeferCleanup(k8sClient.Delete, testBucketClass)

		srv.SetBucketClasses([]*bucket.FakeBucketClass{
			{
				BucketClass: iri.BucketClass{
					Name: bucketClass.Name,
					Capabilities: &iri.BucketClassCapabilities{
						Tps:  262144000,
						Iops: 15000,
					},
				},
			},
			{
				BucketClass: iri.BucketClass{
					Name: testBucketClass.Name,
					Capabilities: &iri.BucketClassCapabilities{
						Tps:  262144000,
						Iops: 1000,
					},
				},
			},
		})

		By("checking if the test bucket class is present")
		Eventually(Object(bucketPool)).Should(SatisfyAll(
			HaveField("Status.AvailableBucketClasses", ContainElements([]corev1.LocalObjectReference{
				{Name: bucketClass.Name},
				{Name: testBucketClass.Name},
			})),
		))
	})

	It("should enforce topology labels", func(ctx SpecContext) {
		By("patching the bucket pool with incorrect topology labels")
		Eventually(Update(bucketPool, func() {
			if bucketPool.Labels == nil {
				bucketPool.Labels = make(map[string]string)
			}
			bucketPool.Labels["topology.ironcore.dev/region"] = "wrong-region"
			bucketPool.Labels["topology.ironcore.dev/zone"] = "wrong-zone"
		})).Should(Succeed())

		By("checking if the reconciler resets the topology labels to its original values")
		Eventually(Object(bucketPool)).Should(SatisfyAll(
			HaveField("ObjectMeta.Labels", HaveKeyWithValue("topology.ironcore.dev/region", "test-region-1")),
			HaveField("ObjectMeta.Labels", HaveKeyWithValue("topology.ironcore.dev/zone", "test-zone-1")),
		))
	})
})
