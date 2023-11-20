// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package validation_test

import (
	"github.com/ironcore-dev/ironcore/internal/apis/storage"
	. "github.com/ironcore-dev/ironcore/internal/apis/storage/validation"
	. "github.com/ironcore-dev/ironcore/internal/testutils/validation"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Bucket", func() {
	DescribeTable("ValidateBucket",
		func(bucket *storage.Bucket, match types.GomegaMatcher) {
			errList := ValidateBucket(bucket)
			Expect(errList).To(match)
		},
		Entry("missing name",
			&storage.Bucket{},
			ContainElement(RequiredField("metadata.name")),
		),
		Entry("missing namespace",
			&storage.Bucket{ObjectMeta: metav1.ObjectMeta{Name: "foo"}},
			ContainElement(RequiredField("metadata.namespace")),
		),
		Entry("bad name",
			&storage.Bucket{ObjectMeta: metav1.ObjectMeta{Name: "foo*"}},
			ContainElement(InvalidField("metadata.name")),
		),
		Entry("no bucket class ref",
			&storage.Bucket{},
			Not(ContainElement(RequiredField("spec.bucketClassRef"))),
		),
		Entry("invalid bucket class ref name",
			&storage.Bucket{
				Spec: storage.BucketSpec{
					BucketClassRef: &corev1.LocalObjectReference{Name: "foo*"},
				},
			},
			ContainElement(InvalidField("spec.bucketClassRef.name")),
		),
		Entry("valid bucket pool ref name subdomain",
			&storage.Bucket{
				Spec: storage.BucketSpec{
					BucketPoolRef: &corev1.LocalObjectReference{Name: "foo.bar.baz"},
				},
			},
			Not(ContainElement(InvalidField("spec.bucketPoolRef.name"))),
		),
	)

	DescribeTable("ValidateBucketUpdate",
		func(newBucket, oldBucket *storage.Bucket, match types.GomegaMatcher) {
			errList := ValidateBucketUpdate(newBucket, oldBucket)
			Expect(errList).To(match)
		},
		Entry("immutable bucketClassRef",
			&storage.Bucket{
				Spec: storage.BucketSpec{
					BucketClassRef: &corev1.LocalObjectReference{Name: "foo"},
				},
			},
			&storage.Bucket{
				Spec: storage.BucketSpec{
					BucketClassRef: &corev1.LocalObjectReference{Name: "bar"},
				},
			},
			ContainElement(ImmutableField("spec.bucketClassRef")),
		),
		Entry("classful: immutable bucketPoolRef if set",
			&storage.Bucket{
				Spec: storage.BucketSpec{
					BucketClassRef: &corev1.LocalObjectReference{Name: "foo"},
					BucketPoolRef:  &corev1.LocalObjectReference{Name: "foo"},
				},
			},
			&storage.Bucket{
				Spec: storage.BucketSpec{
					BucketClassRef: &corev1.LocalObjectReference{Name: "foo"},
					BucketPoolRef:  &corev1.LocalObjectReference{Name: "bar"},
				},
			},
			ContainElement(ImmutableField("spec.bucketPoolRef")),
		),
		Entry("classful: mutable bucketPoolRef if not set",
			&storage.Bucket{
				Spec: storage.BucketSpec{
					BucketClassRef: &corev1.LocalObjectReference{Name: "foo"},
					BucketPoolRef:  &corev1.LocalObjectReference{Name: "foo"},
				},
			},
			&storage.Bucket{
				Spec: storage.BucketSpec{
					BucketClassRef: &corev1.LocalObjectReference{Name: "foo"},
				},
			},
			Not(ContainElement(ImmutableField("spec.bucketPoolRef"))),
		),
	)
})
