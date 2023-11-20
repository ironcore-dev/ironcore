// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package validation_test

import (
	"github.com/ironcore-dev/ironcore/internal/apis/core"
	"github.com/ironcore-dev/ironcore/internal/apis/storage"
	. "github.com/ironcore-dev/ironcore/internal/apis/storage/validation"
	. "github.com/ironcore-dev/ironcore/internal/testutils/validation"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("BucketClass", func() {
	DescribeTable("ValidateBucketClass",
		func(bucketClass *storage.BucketClass, match types.GomegaMatcher) {
			errList := ValidateBucketClass(bucketClass)
			Expect(errList).To(match)
		},
		Entry("missing name",
			&storage.BucketClass{},
			ContainElement(RequiredField("metadata.name")),
		),
		Entry("bad name",
			&storage.BucketClass{ObjectMeta: metav1.ObjectMeta{Name: "foo*"}},
			ContainElement(InvalidField("metadata.name")),
		),
		Entry("missing required capabilities",
			&storage.BucketClass{},
			ContainElements(
				InvalidField("capabilities[tps]"),
				InvalidField("capabilities[iops]"),
			),
		),
		Entry("invalid capabilities",
			&storage.BucketClass{
				Capabilities: core.ResourceList{
					core.ResourceTPS:  resource.MustParse("-1"),
					core.ResourceIOPS: resource.MustParse("-1"),
				},
			},
			ContainElements(
				InvalidField("capabilities[tps]"),
				InvalidField("capabilities[iops]"),
			),
		),
		Entry("valid capabilities",
			&storage.BucketClass{
				Capabilities: core.ResourceList{
					core.ResourceTPS:  resource.MustParse("500Mi"),
					core.ResourceIOPS: resource.MustParse("100"),
				},
			},
			Not(ContainElements(
				InvalidField("capabilities[tps]"),
				InvalidField("capabilities[iops]"),
			)),
		),
	)

	DescribeTable("ValidateBucketClassUpdate",
		func(newBucketClass, oldBucketClass *storage.BucketClass, match types.GomegaMatcher) {
			errList := ValidateBucketClassUpdate(newBucketClass, oldBucketClass)
			Expect(errList).To(match)
		},
		Entry("immutable capabilities",
			&storage.BucketClass{
				Capabilities: core.ResourceList{
					core.ResourceIOPS: resource.MustParse("1000"),
				},
			},
			&storage.BucketClass{
				Capabilities: core.ResourceList{},
			},
			ContainElement(ImmutableField("capabilities")),
		),
	)
})
