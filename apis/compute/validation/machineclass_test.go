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
	. "github.com/onmetal/onmetal-api/testutils/validation"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("MachineClass", func() {
	DescribeTable("ValidateMachineClass",
		func(machineClass *compute.MachineClass, match types.GomegaMatcher) {
			errList := ValidateMachineClass(machineClass)
			Expect(errList).To(match)
		},
		Entry("missing name",
			&compute.MachineClass{},
			ContainElement(RequiredField("metadata.name")),
		),
		Entry("bad name",
			&compute.MachineClass{ObjectMeta: metav1.ObjectMeta{Name: "foo*"}},
			ContainElement(InvalidField("metadata.name")),
		),
	)

	DescribeTable("ValidateMachineClassUpdate",
		func(newMachineClass, oldMachineClass *compute.MachineClass, match types.GomegaMatcher) {
			errList := ValidateMachineClassUpdate(newMachineClass, oldMachineClass)
			Expect(errList).To(match)
		},
		Entry("immutable capabilities",
			&compute.MachineClass{
				Capabilities: map[corev1.ResourceName]resource.Quantity{
					"ram": resource.MustParse("1Gi"),
				},
			},
			&compute.MachineClass{
				Capabilities: map[corev1.ResourceName]resource.Quantity{},
			},
			ContainElement(ImmutableField("capabilities")),
		),
	)
})
