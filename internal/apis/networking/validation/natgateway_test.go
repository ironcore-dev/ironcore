// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"github.com/ironcore-dev/ironcore/internal/apis/networking"
	. "github.com/ironcore-dev/ironcore/internal/testutils/validation"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
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
					PortsPerNetworkInterface: ptr.To[int32](3),
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
