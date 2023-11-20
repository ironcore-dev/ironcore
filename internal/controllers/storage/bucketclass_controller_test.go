// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	. "github.com/ironcore-dev/controller-utils/testutils"
	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	. "github.com/ironcore-dev/ironcore/utils/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
)

var _ = Describe("BucketClass controller", func() {
	ns := SetupNamespace(&k8sClient)

	It("should finalize the bucket class if no bucket is using it", func(ctx SpecContext) {
		By("creating the bucket class consumed by the bucket")
		bucketClass := &storagev1alpha1.BucketClass{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "bucketclass-",
			},
			Capabilities: corev1alpha1.ResourceList{
				corev1alpha1.ResourceTPS:  resource.MustParse("100Mi"),
				corev1alpha1.ResourceIOPS: resource.MustParse("100"),
			},
		}
		Expect(k8sClient.Create(ctx, bucketClass)).Should(Succeed())

		By("creating the bucket")
		bucket := &storagev1alpha1.Bucket{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "bucket-",
			},
			Spec: storagev1alpha1.BucketSpec{
				BucketClassRef: &corev1.LocalObjectReference{Name: bucketClass.Name},
			},
		}
		Expect(k8sClient.Create(ctx, bucket)).Should(Succeed())

		By("checking the finalizer is present")
		bucketClassKey := client.ObjectKeyFromObject(bucketClass)
		Eventually(func(g Gomega) []string {
			err := k8sClient.Get(ctx, bucketClassKey, bucketClass)
			Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
			g.Expect(err).NotTo(HaveOccurred())
			return bucketClass.Finalizers
		}).Should(ContainElement(storagev1alpha1.BucketClassFinalizer))

		By("issuing a delete request for the bucket class")
		Expect(k8sClient.Delete(ctx, bucketClass)).Should(Succeed())

		By("asserting the bucket class is still present as the bucket is referencing it")
		Consistently(func(g Gomega) []string {
			err := k8sClient.Get(ctx, bucketClassKey, bucketClass)
			g.Expect(err).NotTo(HaveOccurred())
			return bucketClass.Finalizers
		}).Should(ContainElement(storagev1alpha1.BucketClassFinalizer))

		By("deleting the referencing bucket")
		Expect(k8sClient.Delete(ctx, bucket)).Should(Succeed())

		By("waiting for the bucket class to be gone")
		Eventually(func() error {
			return k8sClient.Get(ctx, bucketClassKey, bucketClass)
		}).Should(MatchErrorFunc(apierrors.IsNotFound))
	})
})
