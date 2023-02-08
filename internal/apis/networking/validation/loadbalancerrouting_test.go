/*
 * Copyright (c) 2022 by the OnMetal authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package validation

import (
	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	"github.com/onmetal/onmetal-api/internal/apis/networking"
	. "github.com/onmetal/onmetal-api/internal/testutils/validation"
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
		Entry("duplicate destination",
			&networking.LoadBalancerRouting{
				Destinations: []commonv1alpha1.LocalUIDReference{
					{Name: "foo"},
					{Name: "foo"},
				},
			},
			ContainElement(DuplicateField("destinations[1]")),
		),
	)
})
