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
	"github.com/onmetal/onmetal-api/internal/apis/storage"
	. "github.com/onmetal/onmetal-api/internal/apis/storage/validation"
	. "github.com/onmetal/onmetal-api/internal/testutils/validation"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("BucketPool", func() {
	DescribeTable("ValidateBucketPool",
		func(bucketPool *storage.BucketPool, match types.GomegaMatcher) {
			errList := ValidateBucketPool(bucketPool)
			Expect(errList).To(match)
		},
		Entry("missing name",
			&storage.BucketPool{},
			ContainElement(RequiredField("metadata.name")),
		),
		Entry("bad name",
			&storage.BucketPool{ObjectMeta: metav1.ObjectMeta{Name: "foo*"}},
			ContainElement(InvalidField("metadata.name")),
		),
		Entry("dns subdomain name",
			&storage.BucketPool{ObjectMeta: metav1.ObjectMeta{Name: "foo.bar.baz"}},
			Not(ContainElement(InvalidField("metadata.name"))),
		),
	)

	DescribeTable("ValidateBucketUpdate",
		func(newBucketPool, oldBucketPool *storage.BucketPool, match types.GomegaMatcher) {
			errList := ValidateBucketPoolUpdate(newBucketPool, oldBucketPool)
			Expect(errList).To(match)
		},
		Entry("immutable providerID if set",
			&storage.BucketPool{
				Spec: storage.BucketPoolSpec{
					ProviderID: "foo",
				},
			},
			&storage.BucketPool{
				Spec: storage.BucketPoolSpec{
					ProviderID: "bar",
				},
			},
			ContainElement(ImmutableField("spec.providerID")),
		),
		Entry("mutable providerID if not set",
			&storage.BucketPool{
				Spec: storage.BucketPoolSpec{
					ProviderID: "foo",
				},
			},
			&storage.BucketPool{},
			Not(ContainElement(ImmutableField("spec.providerID"))),
		),
	)
})
