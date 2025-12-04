// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	bucketbrokerv1alpha1 "github.com/ironcore-dev/ironcore/broker/bucketbroker/api/v1alpha1"
	brokerutils "github.com/ironcore-dev/ironcore/broker/common/utils"
	iri "github.com/ironcore-dev/ironcore/iri/apis/bucket/v1alpha1"
	irimeta "github.com/ironcore-dev/ironcore/iri/apis/meta/v1alpha1"
	bucketpoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/bucketpoollet/api/v1alpha1"
	poolletutils "github.com/ironcore-dev/ironcore/poollet/common/utils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("CreateBucket", func() {
	ns, _, srv := SetupTest()
	bucketClass := SetupBucketClass("250Mi", "1500")

	It("should correctly create a bucket", func(ctx SpecContext) {
		By("creating a bucket")
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

		By("getting the ironcore bucket")
		ironcoreBucket := &storagev1alpha1.Bucket{}
		ironcoreBucketKey := client.ObjectKey{Namespace: ns.Name, Name: res.Bucket.Metadata.Id}
		Expect(k8sClient.Get(ctx, ironcoreBucketKey, ironcoreBucket)).To(Succeed())

		By("inspecting the ironcore bucket")
		Expect(ironcoreBucket.Labels).To(Equal(map[string]string{
			poolletutils.DownwardAPILabel(bucketpoolletv1alpha1.BucketDownwardAPIPrefix, "root-bucket-uid"): "foobar",
			bucketbrokerv1alpha1.CreatedLabel: "true",
			bucketbrokerv1alpha1.ManagerLabel: bucketbrokerv1alpha1.BucketBrokerManager,
		}))
		encodedIRIAnnotations, err := brokerutils.EncodeAnnotationsAnnotation(nil)
		Expect(err).NotTo(HaveOccurred())
		encodedIRILabels, err := brokerutils.EncodeLabelsAnnotation(map[string]string{
			bucketpoolletv1alpha1.BucketUIDLabel: "foobar",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(ironcoreBucket.Annotations).To(Equal(map[string]string{
			bucketbrokerv1alpha1.AnnotationsAnnotation: encodedIRIAnnotations,
			bucketbrokerv1alpha1.LabelsAnnotation:      encodedIRILabels,
		}))
		Expect(ironcoreBucket.Spec.BucketClassRef.Name).To(Equal(bucketClass.Name))

	})
})
