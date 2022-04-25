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
	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	"github.com/onmetal/onmetal-api/apis/networking"
	. "github.com/onmetal/onmetal-api/testutils/validation"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("VirtualIPRouting", func() {
	DescribeTable("ValidateVirtualIPRouting",
		func(virtualIPRouting *networking.VirtualIPRouting, match types.GomegaMatcher) {
			errList := ValidateVirtualIPRouting(virtualIPRouting)
			Expect(errList).To(match)
		},
		Entry("missing name",
			&networking.VirtualIPRouting{},
			ContainElement(RequiredField("metadata.name")),
		),
		Entry("missing namespace",
			&networking.VirtualIPRouting{ObjectMeta: metav1.ObjectMeta{Name: "foo"}},
			ContainElement(RequiredField("metadata.namespace")),
		),
		Entry("bad name",
			&networking.VirtualIPRouting{ObjectMeta: metav1.ObjectMeta{Name: "foo*"}},
			ContainElement(InvalidField("metadata.name")),
		),
		Entry("invalid subset ip",
			&networking.VirtualIPRouting{
				Subsets: []networking.VirtualIPRoutingSubset{{}},
			},
			ContainElement(InvalidField("subsets[0].ip")),
		),
		Entry("missing subset ref name",
			&networking.VirtualIPRouting{
				Subsets: []networking.VirtualIPRoutingSubset{{}},
			},
			ContainElement(RequiredField("subsets[0].targetRef.name")),
		),
		Entry("invalid subset ref name",
			&networking.VirtualIPRouting{
				Subsets: []networking.VirtualIPRoutingSubset{{
					TargetRef: networking.LocalUIDReference{Name: "foo*"},
				}},
			},
			ContainElement(InvalidField("subsets[0].targetRef.name")),
		),
		Entry("duplicate subset entry",
			&networking.VirtualIPRouting{
				Subsets: []networking.VirtualIPRoutingSubset{
					{
						IP:        commonv1alpha1.MustParseIP("10.0.0.1"),
						TargetRef: networking.LocalUIDReference{Name: "foo*"},
					},
					{
						IP:        commonv1alpha1.MustParseIP("10.0.0.1"),
						TargetRef: networking.LocalUIDReference{Name: "foo*"},
					},
				},
			},
			ContainElement(DuplicateField("subsets[1]")),
		),
	)

	DescribeTable("ValidateVirtualIPRoutingUpdate",
		func(newVirtualIPRouting, oldVirtualIPRouting *networking.VirtualIPRouting, match types.GomegaMatcher) {
			errList := ValidateVirtualIPRoutingUpdate(newVirtualIPRouting, oldVirtualIPRouting)
			Expect(errList).To(match)
		},
	)
})
