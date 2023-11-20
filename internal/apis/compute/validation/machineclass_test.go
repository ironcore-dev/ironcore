// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"github.com/ironcore-dev/ironcore/internal/apis/compute"
	"github.com/ironcore-dev/ironcore/internal/apis/core"
	. "github.com/ironcore-dev/ironcore/internal/testutils/validation"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
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
		Entry("missing required capabilities",
			&compute.MachineClass{},
			ContainElements(
				InvalidField("capabilities[cpu]"),
				InvalidField("capabilities[memory]"),
			),
		),
		Entry("invalid capabilities",
			&compute.MachineClass{
				Capabilities: core.ResourceList{
					core.ResourceCPU:    resource.MustParse("-1"),
					core.ResourceMemory: resource.MustParse("-1Gi"),
				},
			},
			ContainElements(
				InvalidField("capabilities[cpu]"),
				InvalidField("capabilities[memory]"),
			),
		),
		Entry("valid capabilities",
			&compute.MachineClass{
				Capabilities: core.ResourceList{
					core.ResourceCPU:    resource.MustParse("300m"),
					core.ResourceMemory: resource.MustParse("2Gi"),
				},
			},
			Not(ContainElements(
				InvalidField("capabilities[cpu]"),
				InvalidField("capabilities[memory]"),
			)),
		),
	)

	DescribeTable("ValidateMachineClassUpdate",
		func(newMachineClass, oldMachineClass *compute.MachineClass, match types.GomegaMatcher) {
			errList := ValidateMachineClassUpdate(newMachineClass, oldMachineClass)
			Expect(errList).To(match)
		},
		Entry("immutable capabilities",
			&compute.MachineClass{
				Capabilities: core.ResourceList{
					core.ResourceMemory: resource.MustParse("1Gi"),
				},
			},
			&compute.MachineClass{
				Capabilities: core.ResourceList{},
			},
			ContainElement(ImmutableField("capabilities")),
		),
	)
})
