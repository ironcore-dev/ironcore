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
	"github.com/onmetal/onmetal-api/apis/compute"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("NetworkInterface", func() {
	DescribeTable("ValidateNetworkInterface",
		func(networkInterface *compute.NetworkInterface, match types.GomegaMatcher) {
			errList := ValidateNetworkInterface(networkInterface)
			Expect(errList).To(match)
		},
		Entry("missing name",
			&compute.NetworkInterface{},
			ContainElement(RequiredField("metadata.name")),
		),
		Entry("missing namespace",
			&compute.NetworkInterface{ObjectMeta: metav1.ObjectMeta{Name: "foo"}},
			ContainElement(RequiredField("metadata.namespace")),
		),
		Entry("bad name",
			&compute.NetworkInterface{ObjectMeta: metav1.ObjectMeta{Name: "foo*"}},
			ContainElement(InvalidField("metadata.name")),
		),
		Entry("no network ref",
			&compute.NetworkInterface{},
			ContainElement(RequiredField("spec.networkRef")),
		),
		Entry("invalid network ref name",
			&compute.NetworkInterface{
				Spec: compute.NetworkInterfaceSpec{
					NetworkRef: corev1.LocalObjectReference{Name: "foo*"},
				},
			},
			ContainElement(InvalidField("spec.networkRef.name")),
		),
		Entry("invalid machine ref name",
			&compute.NetworkInterface{
				Spec: compute.NetworkInterfaceSpec{
					MachineRef: corev1.LocalObjectReference{Name: "foo*"},
				},
			},
			ContainElement(InvalidField("spec.machineRef.name")),
		),
		Entry("missing ip families",
			&compute.NetworkInterface{},
			ContainElement(RequiredField("spec.ipFamilies")),
		),
		Entry("not supported ip family",
			&compute.NetworkInterface{
				Spec: compute.NetworkInterfaceSpec{
					IPFamilies: []corev1.IPFamily{"foo"},
				},
			},
			ContainElement(NotSupportedField("spec.ipFamilies[0]")),
		),
		Entry("duplicate ip family",
			&compute.NetworkInterface{
				Spec: compute.NetworkInterfaceSpec{
					IPFamilies: []corev1.IPFamily{"IPv4", "IPv4"},
				},
			},
			ContainElement(DuplicateField("spec.ipFamilies[1]")),
		),
		Entry("missing ephemeral prefix name",
			&compute.NetworkInterface{
				Spec: compute.NetworkInterfaceSpec{
					IPs: []compute.IPSource{{EphemeralPrefix: &compute.EphemeralPrefixSource{
						PrefixTemplate: &compute.PrefixTemplate{}}},
					},
				},
			},
			ContainElement(RequiredField("spec.ips[0].ephemeralPrefix.metadata.name")),
		),
		Entry("invalid ephemeral prefix name",
			&compute.NetworkInterface{
				Spec: compute.NetworkInterfaceSpec{
					IPs: []compute.IPSource{{EphemeralPrefix: &compute.EphemeralPrefixSource{
						PrefixTemplate: &compute.PrefixTemplate{
							ObjectMeta: metav1.ObjectMeta{Name: "foo*"},
						}}},
					},
				},
			},
			ContainElement(InvalidField("spec.ips[0].ephemeralPrefix.metadata.name")),
		),
		Entry("missing ephemeral prefix namespace",
			&compute.NetworkInterface{
				Spec: compute.NetworkInterfaceSpec{
					IPs: []compute.IPSource{{EphemeralPrefix: &compute.EphemeralPrefixSource{
						PrefixTemplate: &compute.PrefixTemplate{
							ObjectMeta: metav1.ObjectMeta{Name: "foo"},
						}}},
					},
				},
			},
			ContainElement(RequiredField("spec.ips[0].ephemeralPrefix.metadata.namespace")),
		),
		// TODO: add validation tests for Prefix spec once new Prefix is merged
	)

	DescribeTable("ValidateNetworkInterfaceUpdate",
		func(newNetworkInterface, oldNetworkInterface *compute.NetworkInterface, match types.GomegaMatcher) {
			errList := ValidateNetworkInterfaceUpdate(newNetworkInterface, oldNetworkInterface)
			Expect(errList).To(match)
		},
		Entry("immutable networkRef if set",
			&compute.NetworkInterface{
				Spec: compute.NetworkInterfaceSpec{
					NetworkRef: corev1.LocalObjectReference{Name: "foo"},
				},
			},
			&compute.NetworkInterface{
				Spec: compute.NetworkInterfaceSpec{
					NetworkRef: corev1.LocalObjectReference{Name: "bar"},
				},
			},
			ContainElement(ImmutableField("spec.networkRef")),
		),
	)
})
