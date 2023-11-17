// Copyright 2023 IronCore authors
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
	"github.com/ironcore-dev/ironcore/internal/apis/core"
	. "github.com/ironcore-dev/ironcore/internal/apis/core/validation"
	. "github.com/ironcore-dev/ironcore/internal/testutils/validation"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Resourcequota", func() {
	DescribeTable("ValidateResourceQuota",
		func(resourceQuota *core.ResourceQuota, match types.GomegaMatcher) {
			errList := ValidateResourceQuota(resourceQuota)
			Expect(errList).To(match)
		},
		Entry("missing name",
			&core.ResourceQuota{},
			ContainElement(RequiredField("metadata.name")),
		),
		Entry("missing namespace",
			&core.ResourceQuota{ObjectMeta: metav1.ObjectMeta{Name: "foo"}},
			ContainElement(RequiredField("metadata.namespace")),
		),
		Entry("bad name",
			&core.ResourceQuota{ObjectMeta: metav1.ObjectMeta{Name: "foo*"}},
			ContainElement(InvalidField("metadata.name")),
		),
		Entry("negative quantity",
			&core.ResourceQuota{
				Spec: core.ResourceQuotaSpec{
					Hard: core.ResourceList{
						core.ResourceRequestsMemory: resource.MustParse("-1Gi"),
					},
				},
			},
			ContainElement(InvalidField("spec.hard[requests.memory]")),
		),
	)

	DescribeTable("ValidateResourceQuotaUpdate",
		func(newResourceQuota, oldResourceQuota *core.ResourceQuota, match types.GomegaMatcher) {
			errList := ValidateResourceQuotaUpdate(newResourceQuota, oldResourceQuota)
			Expect(errList).To(match)
		},
		Entry("resource version missing",
			&core.ResourceQuota{},
			&core.ResourceQuota{},
			ContainElement(InvalidField("metadata.resourceVersion")),
		),
	)
})
