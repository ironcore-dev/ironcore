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
	"github.com/onmetal/onmetal-api/internal/apis/networking"
	. "github.com/onmetal/onmetal-api/internal/testutils/validation"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

var _ = Describe("NATGateway", func() {
	DescribeTable("ValidateNATGateway",
		func(natGateway *networking.NATGateway, match types.GomegaMatcher) {
			errList := ValidateNATGateway(natGateway)
			Expect(errList).To(match)
		},
		Entry("missing name",
			&networking.NATGateway{},
			ContainElement(RequiredField("metadata.name")),
		),
		Entry("missing namespace",
			&networking.NATGateway{ObjectMeta: metav1.ObjectMeta{Name: "foo"}},
			ContainElement(RequiredField("metadata.namespace")),
		),
		Entry("bad name",
			&networking.NATGateway{ObjectMeta: metav1.ObjectMeta{Name: "foo*"}},
			ContainElement(InvalidField("metadata.name")),
		),
		Entry("no network ref",
			&networking.NATGateway{},
			ContainElement(InvalidField("spec.networkRef.name")),
		),
		Entry("invalid network ref name",
			&networking.NATGateway{
				Spec: networking.NATGatewaySpec{
					NetworkRef: corev1.LocalObjectReference{Name: "foo*"},
				},
			},
			ContainElement(InvalidField("spec.networkRef.name")),
		),
		Entry("missing type",
			&networking.NATGateway{},
			ContainElement(RequiredField("spec.type")),
		),
		Entry("ports per nic not power of 2",
			&networking.NATGateway{
				Spec: networking.NATGatewaySpec{
					PortsPerNetworkInterface: pointer.Int32(3),
				},
			},
			ContainElement(InvalidField("spec.portsPerNetworkInterface")),
		),
	)

	DescribeTable("ValidateNATGatewayUpdate",
		func(newNATGateway, oldNATGateway *networking.NATGateway, match types.GomegaMatcher) {
			errList := ValidateNATGatewayUpdate(newNATGateway, oldNATGateway)
			Expect(errList).To(match)
		},
		Entry("immutable networkRef",
			&networking.NATGateway{
				Spec: networking.NATGatewaySpec{
					NetworkRef: corev1.LocalObjectReference{Name: "foo"},
				},
			},
			&networking.NATGateway{
				Spec: networking.NATGatewaySpec{
					NetworkRef: corev1.LocalObjectReference{Name: "foo*"},
				},
			},
			ContainElement(ForbiddenField("spec.networkRef")),
		),
	)
})
