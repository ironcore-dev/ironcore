// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/bucket/v1alpha1"
	irimeta "github.com/ironcore-dev/ironcore/iri/apis/meta/v1alpha1"
	bucketpoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/bucketpoollet/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("DeleteBucket", func() {
	ns, _, srv := SetupTest()
	bucketClass := SetupBucketClass("250Mi", "1500")

	It("should correctly delete a bucket", func(ctx SpecContext) {
		By("creating a bucket")
		createRes, err := srv.CreateBucket(ctx, &iri.CreateBucketRequest{
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
		Expect(createRes).NotTo(BeNil())

		By("deleting the bucket")
		deleteRes, err := srv.DeleteBucket(ctx, &iri.DeleteBucketRequest{
			BucketId: createRes.Bucket.Metadata.Id,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(deleteRes).NotTo(BeNil())

		By("verifying the bucket is deleted")
		ironcoreBucket := &storagev1alpha1.Bucket{}
		ironcoreBucketKey := client.ObjectKey{Namespace: ns.Name, Name: createRes.Bucket.Metadata.Id}
		err = k8sClient.Get(ctx, ironcoreBucketKey, ironcoreBucket)
		Expect(apierrors.IsNotFound(err)).To(BeTrue())
	})
})
