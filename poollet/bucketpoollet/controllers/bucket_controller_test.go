// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package controllers_test

import (
	"fmt"

	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/bucket/v1alpha1"
	ironcoreclient "github.com/ironcore-dev/ironcore/utils/client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

const (
	AccessKeyID     = "AccessKeyID"
	SecretAccessKey = "SecretAccessKey"
)

var _ = Describe("BucketController", func() {
	ns, bp, bc, srv := SetupTest()

	It("should create a bucket", func(ctx SpecContext) {
		bucketEndpoint := "foo.com"
		accessKey := "foo"
		secretAccess := "bar"

		By("creating a bucket")
		bucket := &storagev1alpha1.Bucket{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "bucket-",
			},
			Spec: storagev1alpha1.BucketSpec{
				BucketClassRef: &corev1.LocalObjectReference{Name: bc.Name},
				BucketPoolRef:  &corev1.LocalObjectReference{Name: bp.Name},
			},
		}
		Expect(k8sClient.Create(ctx, bucket)).To(Succeed())
		DeferCleanup(expectBucketDeleted, bucket)

		By("waiting for the runtime to report the bucket")
		Eventually(srv).Should(SatisfyAll(
			HaveField("Buckets", HaveLen(1)),
		))

		_, iriBucket := GetSingleMapEntry(srv.Buckets)
		Expect(iriBucket.Spec.Class).To(Equal(bc.Name))

		iriBucket.Status.Access = &iri.BucketAccess{
			Endpoint: bucketEndpoint,
			SecretData: map[string][]byte{
				AccessKeyID:     []byte(accessKey),
				SecretAccessKey: []byte(secretAccess),
			},
		}
		iriBucket.Status.State = iri.BucketState_BUCKET_AVAILABLE

		Expect(ironcoreclient.PatchAddReconcileAnnotation(ctx, k8sClient, bucket)).Should(Succeed())

		Eventually(Object(bucket)).Should(SatisfyAll(
			HaveField("Status.State", storagev1alpha1.BucketStateAvailable),
			HaveField("Status.Access.SecretRef", Not(BeNil())),
			HaveField("Status.Access.Endpoint", Equal(bucketEndpoint)),
		))

		accessSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns.Name,
				Name:      bucket.Status.Access.SecretRef.Name,
			}}
		Eventually(Object(accessSecret)).Should(SatisfyAll(
			HaveField("Data", HaveKeyWithValue(AccessKeyID, []byte(accessKey))),
			HaveField("Data", HaveKeyWithValue(SecretAccessKey, []byte(secretAccess))),
		))

	})

})

func GetSingleMapEntry[K comparable, V any](m map[K]V) (K, V) {
	if n := len(m); n != 1 {
		Fail(fmt.Sprintf("Expected for map to have a single entry but got %d", n), 1)
	}
	for k, v := range m {
		return k, v
	}
	panic("unreachable")
}
