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

var _ = Describe("NetworkInterfaceBinding", func() {
	DescribeTable("ValidateNetworkInterfaceBinding",
		func(networkInterfaceBinding *networking.NetworkInterfaceBinding, match types.GomegaMatcher) {
			errList := ValidateNetworkInterfaceBinding(networkInterfaceBinding)
			Expect(errList).To(match)
		},
		Entry("missing name",
			&networking.NetworkInterfaceBinding{},
			ContainElement(RequiredField("metadata.name")),
		),
		Entry("missing namespace",
			&networking.NetworkInterfaceBinding{ObjectMeta: metav1.ObjectMeta{Name: "foo"}},
			ContainElement(RequiredField("metadata.namespace")),
		),
		Entry("bad name",
			&networking.NetworkInterfaceBinding{ObjectMeta: metav1.ObjectMeta{Name: "foo*"}},
			ContainElement(InvalidField("metadata.name")),
		),
		Entry("invalid ip",
			&networking.NetworkInterfaceBinding{
				IPs: []commonv1alpha1.IP{{}},
			},
			ContainElement(InvalidField("ips[0]")),
		),
		Entry("duplicate ip family",
			&networking.NetworkInterfaceBinding{
				IPs: []commonv1alpha1.IP{
					commonv1alpha1.MustParseIP("10.0.0.1"),
					commonv1alpha1.MustParseIP("10.0.0.2"),
				},
			},
			ContainElement(ForbiddenField("ips[1]")),
		),
		Entry("valid network interface binding with two ips",
			&networking.NetworkInterfaceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "foo",
				},
				IPs: []commonv1alpha1.IP{
					commonv1alpha1.MustParseIP("10.0.0.1"),
					commonv1alpha1.MustParseIP("beef::"),
				},
			},
			BeEmpty(),
		),
	)

	DescribeTable("ValidateNetworkInterfaceBindingUpdate",
		func(newNetworkInterfaceBinding, oldNetworkInterfaceBinding *networking.NetworkInterfaceBinding, match types.GomegaMatcher) {
			errList := ValidateNetworkInterfaceBindingUpdate(newNetworkInterfaceBinding, oldNetworkInterfaceBinding)
			Expect(errList).To(match)
		},
	)
})
