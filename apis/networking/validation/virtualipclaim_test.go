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

var _ = Describe("VirtualIPClaim", func() {
	DescribeTable("ValidateVirtualIPClaim",
		func(vipClaim *networking.VirtualIPClaim, match types.GomegaMatcher) {
			errList := ValidateVirtualIPClaim(vipClaim)
			Expect(errList).To(match)
		},
		Entry("missing name",
			&networking.VirtualIPClaim{},
			ContainElement(RequiredField("metadata.name")),
		),
		Entry("missing namespace",
			&networking.VirtualIPClaim{ObjectMeta: metav1.ObjectMeta{Name: "foo"}},
			ContainElement(RequiredField("metadata.namespace")),
		),
		Entry("bad name",
			&networking.VirtualIPClaim{ObjectMeta: metav1.ObjectMeta{Name: "foo*"}},
			ContainElement(InvalidField("metadata.name")),
		),
		Entry("no type",
			&networking.VirtualIPClaim{},
			ContainElement(RequiredField("spec.type")),
		),
		Entry("invalid type",
			&networking.VirtualIPClaim{
				Spec: networking.VirtualIPClaimSpec{
					Type: "invalid",
				},
			},
			ContainElement(NotSupportedField("spec.type")),
		),
		Entry("no ip family",
			&networking.VirtualIPClaim{},
			ContainElement(RequiredField("spec.ipFamily")),
		),
		Entry("invalid ip family",
			&networking.VirtualIPClaim{
				Spec: networking.VirtualIPClaimSpec{
					IPFamily: "invalid",
				},
			},
			ContainElement(NotSupportedField("spec.ipFamily")),
		),
		Entry("invalid claim ref name",
			&networking.VirtualIPClaim{
				Spec: networking.VirtualIPClaimSpec{
					VirtualIPRef: &corev1.LocalObjectReference{
						Name: "foo*",
					},
				},
			},
			ContainElement(InvalidField("spec.virtualIPRef.name")),
		),
		Entry("valid virtual ip claim",
			&networking.VirtualIPClaim{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "bar",
				},
				Spec: networking.VirtualIPClaimSpec{
					Type:     networking.VirtualIPTypePublic,
					IPFamily: corev1.IPv4Protocol,
					VirtualIPRef: &corev1.LocalObjectReference{
						Name: "foo",
					},
				},
			},
			BeEmpty(),
		),
	)

	DescribeTable("ValidateVirtualIPClaimUpdate",
		func(newVIPClaim, oldVIPClaim *networking.VirtualIPClaim, match types.GomegaMatcher) {
			errList := ValidateVirtualIPClaimUpdate(newVIPClaim, oldVIPClaim)
			Expect(errList).To(match)
		},
		Entry("immutable type",
			&networking.VirtualIPClaim{
				Spec: networking.VirtualIPClaimSpec{
					Type: networking.VirtualIPTypePublic,
				},
			},
			&networking.VirtualIPClaim{
				Spec: networking.VirtualIPClaimSpec{
					Type: "other",
				},
			},
			ContainElement(ForbiddenField("spec")),
		),
		Entry("immutable ip family",
			&networking.VirtualIPClaim{
				Spec: networking.VirtualIPClaimSpec{
					IPFamily: corev1.IPv6Protocol,
				},
			},
			&networking.VirtualIPClaim{
				Spec: networking.VirtualIPClaimSpec{
					IPFamily: corev1.IPv4Protocol,
				},
			},
			ContainElement(ForbiddenField("spec")),
		),
		Entry("immutable virtual ip reference",
			&networking.VirtualIPClaim{
				Spec: networking.VirtualIPClaimSpec{
					VirtualIPRef: &corev1.LocalObjectReference{
						Name: "bar",
					},
				},
			},
			&networking.VirtualIPClaim{
				Spec: networking.VirtualIPClaimSpec{
					VirtualIPRef: &corev1.LocalObjectReference{
						Name: "foo",
					},
				},
			},
			ContainElement(ForbiddenField("spec")),
		),
	)
})
