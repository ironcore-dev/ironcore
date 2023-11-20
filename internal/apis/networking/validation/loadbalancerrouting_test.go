// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"github.com/ironcore-dev/ironcore/internal/apis/networking"
	. "github.com/ironcore-dev/ironcore/internal/testutils/validation"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("LoadBalancerRouting", func() {
	DescribeTable("ValidateLoadBalancerRouting",
		func(loadBalancerRouting *networking.LoadBalancerRouting, match types.GomegaMatcher) {
			errList := ValidateLoadBalancerRouting(loadBalancerRouting)
			Expect(errList).To(match)
		},
		Entry("missing name",
			&networking.LoadBalancerRouting{},
			ContainElement(RequiredField("metadata.name")),
		),
		Entry("missing namespace",
			&networking.LoadBalancerRouting{ObjectMeta: metav1.ObjectMeta{Name: "foo"}},
			ContainElement(RequiredField("metadata.namespace")),
		),
		Entry("bad name",
			&networking.LoadBalancerRouting{ObjectMeta: metav1.ObjectMeta{Name: "foo*"}},
			ContainElement(InvalidField("metadata.name")),
		),
		Entry("invalid destination ip",
			&networking.LoadBalancerRouting{
				Destinations: []networking.LoadBalancerDestination{{}},
			},
			ContainElement(InvalidField("destinations[0].ip")),
		),
		Entry("invalid destination targetRef name",
			&networking.LoadBalancerRouting{
				Destinations: []networking.LoadBalancerDestination{
					{TargetRef: &networking.LoadBalancerTargetRef{Name: "foo*"}},
				},
			},
			ContainElement(InvalidField("destinations[0].targetRef.name")),
		),
	)
})
