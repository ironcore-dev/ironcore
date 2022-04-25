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
	"github.com/onmetal/onmetal-api/apis/networking"
	. "github.com/onmetal/onmetal-api/testutils/validation"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("VirtualIP", func() {
	DescribeTable("ValidateVirtualIP",
		func(virtualIP *networking.VirtualIP, match types.GomegaMatcher) {
			errList := ValidateVirtualIP(virtualIP)
			Expect(errList).To(match)
		},
		Entry("missing name",
			&networking.VirtualIP{},
			ContainElement(RequiredField("metadata.name")),
		),
		Entry("missing namespace",
			&networking.VirtualIP{ObjectMeta: metav1.ObjectMeta{Name: "foo"}},
			ContainElement(RequiredField("metadata.namespace")),
		),
		Entry("bad name",
			&networking.VirtualIP{ObjectMeta: metav1.ObjectMeta{Name: "foo*"}},
			ContainElement(InvalidField("metadata.name")),
		),
		Entry("no type",
			&networking.VirtualIP{},
			ContainElement(RequiredField("spec.type")),
		),
		Entry("invalid type",
			&networking.VirtualIP{
				Spec: networking.VirtualIPSpec{
					Type: "invalid",
				},
			},
			ContainElement(NotSupportedField("spec.type")),
		),
		Entry("no ip family",
			&networking.VirtualIP{},
			ContainElement(RequiredField("spec.ipFamily")),
		),
		Entry("invalid ip family",
			&networking.VirtualIP{
				Spec: networking.VirtualIPSpec{
					IPFamily: "invalid",
				},
			},
			ContainElement(NotSupportedField("spec.ipFamily")),
		),
		Entry("valid virtual ip",
			&networking.VirtualIP{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "bar",
				},
				Spec: networking.VirtualIPSpec{
					Type:     networking.VirtualIPTypePublic,
					IPFamily: corev1.IPv4Protocol,
					NetworkInterfaceSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"foo": "bar"},
					},
				},
			},
			BeEmpty(),
		),
	)

	DescribeTable("ValidateVirtualIPUpdate",
		func(newVirtualIP, oldVirtualIP *networking.VirtualIP, match types.GomegaMatcher) {
			errList := ValidateVirtualIPUpdate(newVirtualIP, oldVirtualIP)
			Expect(errList).To(match)
		},
		Entry("immutable type",
			&networking.VirtualIP{
				Spec: networking.VirtualIPSpec{
					Type: networking.VirtualIPTypePublic,
				},
			},
			&networking.VirtualIP{
				Spec: networking.VirtualIPSpec{
					Type: "other",
				},
			},
			ContainElement(ForbiddenField("spec")),
		),
		Entry("immutable ip family",
			&networking.VirtualIP{
				Spec: networking.VirtualIPSpec{
					IPFamily: corev1.IPv6Protocol,
				},
			},
			&networking.VirtualIP{
				Spec: networking.VirtualIPSpec{
					IPFamily: corev1.IPv4Protocol,
				},
			},
			ContainElement(ForbiddenField("spec")),
		),
		Entry("mutable network interface selector",
			&networking.VirtualIP{
				Spec: networking.VirtualIPSpec{
					NetworkInterfaceSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"foo": "bar"},
					},
				},
			},
			&networking.VirtualIP{
				Spec: networking.VirtualIPSpec{
					NetworkInterfaceSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"bar": "baz"},
					},
				},
			},
			Not(ContainElement(ForbiddenField("spec"))),
		),
	)
})
