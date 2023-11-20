// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	"github.com/ironcore-dev/ironcore/internal/apis/networking"
	. "github.com/ironcore-dev/ironcore/internal/testutils/validation"
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
		Entry("invalid target ref name",
			&networking.VirtualIP{
				Spec: networking.VirtualIPSpec{
					TargetRef: &commonv1alpha1.LocalUIDReference{
						Name: "foo*",
					},
				},
			},
			ContainElement(InvalidField("spec.targetRef.name")),
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
					TargetRef: &commonv1alpha1.LocalUIDReference{
						Name: "foo",
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
		Entry("mutable target reference",
			&networking.VirtualIP{
				Spec: networking.VirtualIPSpec{
					TargetRef: &commonv1alpha1.LocalUIDReference{
						Name: "bar",
					},
				},
			},
			&networking.VirtualIP{
				Spec: networking.VirtualIPSpec{
					TargetRef: &commonv1alpha1.LocalUIDReference{
						Name: "foo",
					},
				},
			},
			Not(ContainElement(ForbiddenField("spec"))),
		),
	)
})
