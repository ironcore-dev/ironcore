// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	"github.com/ironcore-dev/ironcore/internal/apis/ipam"
	"github.com/ironcore-dev/ironcore/internal/apis/networking"
	. "github.com/ironcore-dev/ironcore/internal/testutils/validation"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("NetworkInterface", func() {
	DescribeTable("ValidateNetworkInterface",
		func(networkInterface *networking.NetworkInterface, match types.GomegaMatcher) {
			errList := ValidateNetworkInterface(networkInterface)
			Expect(errList).To(match)
		},
		Entry("missing name",
			&networking.NetworkInterface{},
			ContainElement(RequiredField("metadata.name")),
		),
		Entry("missing namespace",
			&networking.NetworkInterface{ObjectMeta: metav1.ObjectMeta{Name: "foo"}},
			ContainElement(RequiredField("metadata.namespace")),
		),
		Entry("bad name",
			&networking.NetworkInterface{ObjectMeta: metav1.ObjectMeta{Name: "foo*"}},
			ContainElement(InvalidField("metadata.name")),
		),
		Entry("no network ref",
			&networking.NetworkInterface{},
			ContainElement(RequiredField("spec.networkRef")),
		),
		Entry("invalid network ref name",
			&networking.NetworkInterface{
				Spec: networking.NetworkInterfaceSpec{
					NetworkRef: corev1.LocalObjectReference{Name: "foo*"},
				},
			},
			ContainElement(InvalidField("spec.networkRef.name")),
		),
		Entry("invalid machine ref name",
			&networking.NetworkInterface{
				Spec: networking.NetworkInterfaceSpec{
					MachineRef: &commonv1alpha1.LocalUIDReference{Name: "foo*"},
				},
			},
			ContainElement(InvalidField("spec.machineRef.name")),
		),
		Entry("missing ip families",
			&networking.NetworkInterface{},
			ContainElement(RequiredField("spec.ipFamilies")),
		),
		Entry("not supported ip family",
			&networking.NetworkInterface{
				Spec: networking.NetworkInterfaceSpec{
					IPFamilies: []corev1.IPFamily{"foo"},
				},
			},
			ContainElement(NotSupportedField("spec.ipFamilies[0]")),
		),
		Entry("duplicate ip family",
			&networking.NetworkInterface{
				Spec: networking.NetworkInterfaceSpec{
					IPFamilies: []corev1.IPFamily{"IPv4", "IPv4"},
				},
			},
			ContainElement(DuplicateField("spec.ipFamilies[1]")),
		),
		Entry("ephemeral prefix name present",
			&networking.NetworkInterface{
				Spec: networking.NetworkInterfaceSpec{
					IPs: []networking.IPSource{{Ephemeral: &networking.EphemeralPrefixSource{
						PrefixTemplate: &ipam.PrefixTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{Name: "foo"},
						}}},
					},
				},
			},
			ContainElement(ForbiddenField("spec.ips[0].ephemeral.prefixTemplate.metadata.name")),
		),
		Entry("ephemeral prefix namespace present",
			&networking.NetworkInterface{
				Spec: networking.NetworkInterfaceSpec{
					IPs: []networking.IPSource{{Ephemeral: &networking.EphemeralPrefixSource{
						PrefixTemplate: &ipam.PrefixTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{Namespace: "foo"},
						}}},
					},
				},
			},
			ContainElement(ForbiddenField("spec.ips[0].ephemeral.prefixTemplate.metadata.namespace")),
		),
		Entry("ephemeral prefix ip family mismatch",
			&networking.NetworkInterface{
				Spec: networking.NetworkInterfaceSpec{
					IPFamilies: []corev1.IPFamily{corev1.IPv4Protocol},
					IPs: []networking.IPSource{{Ephemeral: &networking.EphemeralPrefixSource{
						PrefixTemplate: &ipam.PrefixTemplateSpec{
							Spec: ipam.PrefixSpec{
								IPFamily: corev1.IPv6Protocol,
							},
						}}},
					},
				},
			},
			ContainElement(ForbiddenField("spec.ips[0].ephemeral.prefixTemplate.spec.ipFamily")),
		),
		Entry("ephemeral prefix does not create a single ip",
			&networking.NetworkInterface{
				Spec: networking.NetworkInterfaceSpec{
					IPFamilies: []corev1.IPFamily{corev1.IPv4Protocol},
					IPs: []networking.IPSource{{Ephemeral: &networking.EphemeralPrefixSource{
						PrefixTemplate: &ipam.PrefixTemplateSpec{
							Spec: ipam.PrefixSpec{
								PrefixLength: 24,
							},
						}}},
					},
				},
			},
			ContainElement(ForbiddenField("spec.ips[0].ephemeral.prefixTemplate.spec.prefixLength")),
		),
		Entry("invalid virtual ip reference",
			&networking.NetworkInterface{
				Spec: networking.NetworkInterfaceSpec{
					VirtualIP: &networking.VirtualIPSource{
						VirtualIPRef: &corev1.LocalObjectReference{Name: "foo*"},
					},
				},
			},
			ContainElement(InvalidField("spec.virtualIP.virtualIPRef.name")),
		),
		Entry("missing virtual ip ephemeral virtual ip template",
			&networking.NetworkInterface{
				Spec: networking.NetworkInterfaceSpec{
					VirtualIP: &networking.VirtualIPSource{
						Ephemeral: &networking.EphemeralVirtualIPSource{},
					},
				},
			},
			ContainElement(RequiredField("spec.virtualIP.ephemeral.virtualIPTemplate")),
		),
		Entry("multiple virtual ip sources",
			&networking.NetworkInterface{
				Spec: networking.NetworkInterfaceSpec{
					VirtualIP: &networking.VirtualIPSource{
						VirtualIPRef: &corev1.LocalObjectReference{Name: "foo"},
						Ephemeral:    &networking.EphemeralVirtualIPSource{},
					},
				},
			},
			ContainElement(InvalidField("spec.virtualIP.ephemeral")),
		),
	)

	DescribeTable("ValidateNetworkInterfaceUpdate",
		func(newNetworkInterface, oldNetworkInterface *networking.NetworkInterface, match types.GomegaMatcher) {
			errList := ValidateNetworkInterfaceUpdate(newNetworkInterface, oldNetworkInterface)
			Expect(errList).To(match)
		},
		Entry("immutable networkRef",
			&networking.NetworkInterface{
				Spec: networking.NetworkInterfaceSpec{
					NetworkRef: corev1.LocalObjectReference{Name: "foo"},
				},
			},
			&networking.NetworkInterface{
				Spec: networking.NetworkInterfaceSpec{
					NetworkRef: corev1.LocalObjectReference{Name: "bar"},
				},
			},
			ContainElement(ForbiddenField("spec")),
		),
		Entry("immutable ip families",
			&networking.NetworkInterface{
				Spec: networking.NetworkInterfaceSpec{
					IPFamilies: []corev1.IPFamily{corev1.IPv4Protocol},
				},
			},
			&networking.NetworkInterface{
				Spec: networking.NetworkInterfaceSpec{
					IPFamilies: []corev1.IPFamily{corev1.IPv6Protocol},
				},
			},
			ContainElement(ForbiddenField("spec")),
		),
		Entry("mutable machine ref",
			&networking.NetworkInterface{
				Spec: networking.NetworkInterfaceSpec{
					MachineRef: &commonv1alpha1.LocalUIDReference{Name: "bar"},
				},
			},
			&networking.NetworkInterface{
				Spec: networking.NetworkInterfaceSpec{
					MachineRef: &commonv1alpha1.LocalUIDReference{Name: "foo"},
				},
			},
			Not(ContainElement(ForbiddenField("spec"))),
		),
		Entry("mutable ips",
			&networking.NetworkInterface{
				Spec: networking.NetworkInterfaceSpec{
					IPs: []networking.IPSource{
						{
							Value: commonv1alpha1.MustParseNewIP("10.0.0.1"),
						},
					},
				},
			},
			&networking.NetworkInterface{
				Spec: networking.NetworkInterfaceSpec{
					IPs: []networking.IPSource{
						{
							Value: commonv1alpha1.MustParseNewIP("10.0.0.2"),
						},
					},
				},
			},
			Not(ContainElement(ForbiddenField("spec"))),
		),
		Entry("mutable virtual ip",
			&networking.NetworkInterface{
				Spec: networking.NetworkInterfaceSpec{
					VirtualIP: &networking.VirtualIPSource{
						VirtualIPRef: &corev1.LocalObjectReference{Name: "my-virtualip"},
					},
				},
			},
			&networking.NetworkInterface{},
			Not(ContainElement(ForbiddenField("spec"))),
		),
	)
})
