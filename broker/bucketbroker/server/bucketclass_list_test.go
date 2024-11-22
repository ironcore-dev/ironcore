// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	iri "github.com/ironcore-dev/ironcore/iri/apis/bucket/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("ListBucketClasses", func() {
	_, bucketPool, srv := SetupTest()
	bucketClass1 := SetupBucketClass("100Mi", "100")
	bucketClass2 := SetupBucketClass("200Mi", "200")

	It("should correctly list available bucket classes", func(ctx SpecContext) {
		By("patching bucket classes in the bucket pool status")
		bucketPool.Status.AvailableBucketClasses = []corev1.LocalObjectReference{
			{Name: bucketClass1.Name},
			{Name: bucketClass2.Name},
		}
		Expect(k8sClient.Status().Update(ctx, bucketPool)).To(Succeed())

		res, err := srv.ListBucketClasses(ctx, &iri.ListBucketClassesRequest{})
		Expect(err).NotTo(HaveOccurred())
		Expect(res.BucketClasses).To(HaveLen(2))
		Expect(res.BucketClasses).To(ContainElements(
			&iri.BucketClass{Name: bucketClass1.Name, Capabilities: &iri.BucketClassCapabilities{Tps: 104857600, Iops: 100}},
			&iri.BucketClass{Name: bucketClass2.Name, Capabilities: &iri.BucketClassCapabilities{Tps: 209715200, Iops: 200}},
		))
	})

})
