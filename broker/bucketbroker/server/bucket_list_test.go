// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	iri "github.com/ironcore-dev/ironcore/iri/apis/bucket/v1alpha1"
	irimeta "github.com/ironcore-dev/ironcore/iri/apis/meta/v1alpha1"
	bucketpoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/bucketpoollet/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ListBuckets", func() {
	_, _, srv := SetupTest()
	bucketClass := SetupBucketClass("250Mi", "1500")

	It("should correctly list buckets", func(ctx SpecContext) {
		By("creating multiple buckets")
		const noOfBuckets = 3

		buckets := make([]any, noOfBuckets)
		for i := 0; i < noOfBuckets; i++ {
			res, err := srv.CreateBucket(ctx, &iri.CreateBucketRequest{
				Bucket: &iri.Bucket{
					Metadata: &irimeta.ObjectMetadata{
						Labels: map[string]string{
							bucketpoolletv1alpha1.BucketUIDLabel: "foobar",
						},
					},
					Spec: &iri.BucketSpec{
						Class: bucketClass.Name,
					},
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
			buckets[i] = res.Bucket
		}

		By("listing the buckets")
		Expect(srv.ListBuckets(ctx, &iri.ListBucketsRequest{})).To(HaveField("Buckets", ConsistOf(buckets...)))
	})
})
