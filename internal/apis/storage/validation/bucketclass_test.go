// Copyright 2022 OnMetal authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package validation_test

import (
	"github.com/onmetal/onmetal-api/internal/apis/core"
	"github.com/onmetal/onmetal-api/internal/apis/storage"
	. "github.com/onmetal/onmetal-api/internal/apis/storage/validation"
	. "github.com/onmetal/onmetal-api/internal/testutils/validation"
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
