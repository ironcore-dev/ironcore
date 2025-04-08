// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

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
		Entry("invalid resource name",
			&core.ResourceQuota{
				Spec: core.ResourceQuotaSpec{
					Hard: core.ResourceList{
						"count/foo": resource.MustParse("2"),
					},
				},
			},
			ContainElement(InvalidField("spec.hard[count/foo]")),
		),
		Entry("invalid resource name",
			&core.ResourceQuota{
				Spec: core.ResourceQuotaSpec{
					Hard: core.ResourceList{
						"bar": resource.MustParse("2"),
					},
				},
			},
			ContainElement(InvalidField("spec.hard[bar]")),
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
