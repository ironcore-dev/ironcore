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
	"github.com/onmetal/onmetal-api/testutils/validation"
	. "github.com/onmetal/onmetal-api/testutils/validation"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("NATGatewayRouting", func() {
	DescribeTable("ValidateNATGatewayRouting",
		func(natGatewayRouting *networking.NATGatewayRouting, match types.GomegaMatcher) {
			errList := ValidateNATGatewayRouting(natGatewayRouting)
			Expect(errList).To(match)
		},
		Entry("missing name",
			&networking.NATGatewayRouting{},
			ContainElement(validation.RequiredField("metadata.name")),
		),
		Entry("missing namespace",
			&networking.NATGatewayRouting{ObjectMeta: metav1.ObjectMeta{Name: "foo"}},
			ContainElement(validation.RequiredField("metadata.namespace")),
		),
		Entry("bad name",
			&networking.NATGatewayRouting{ObjectMeta: metav1.ObjectMeta{Name: "foo*"}},
			ContainElement(validation.InvalidField("metadata.name")),
		),
		Entry("duplicate destination",
			&networking.NATGatewayRouting{
				Destinations: []networking.NATGatewayDestination{
					{LocalUIDReference: commonv1alpha1.LocalUIDReference{Name: "foo"}},
					{LocalUIDReference: commonv1alpha1.LocalUIDReference{Name: "foo"}},
				},
			},
			ContainElement(DuplicateField("destinations[1]")),
		),
	)

	DescribeTable("ValidateNATGatewayRoutingUpdate",
		func(newNATGatewayRouting, oldNATGatewayRouting *networking.NATGatewayRouting, match types.GomegaMatcher) {
			errList := ValidateNATGatewayRoutingUpdate(newNATGatewayRouting, oldNATGatewayRouting)
			Expect(errList).To(match)
		},
	)
})
