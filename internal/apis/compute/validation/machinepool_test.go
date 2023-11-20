// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"github.com/ironcore-dev/ironcore/internal/apis/compute"
	. "github.com/ironcore-dev/ironcore/internal/testutils/validation"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("MachinePool", func() {
	DescribeTable("ValidateMachinePool",
		func(machinePool *compute.MachinePool, match types.GomegaMatcher) {
			errList := ValidateMachinePool(machinePool)
			Expect(errList).To(match)
		},
		Entry("missing name",
			&compute.MachinePool{},
			ContainElement(RequiredField("metadata.name")),
		),
		Entry("bad name",
			&compute.MachinePool{ObjectMeta: metav1.ObjectMeta{Name: "foo*"}},
			ContainElement(InvalidField("metadata.name")),
		),
		Entry("dns subdomain name",
			&compute.MachinePool{ObjectMeta: metav1.ObjectMeta{Name: "foo.bar.baz"}},
			Not(ContainElement(InvalidField("metadata.name"))),
		),
	)

	DescribeTable("ValidateMachinePoolUpdate",
		func(newMachinePool, oldMachinePool *compute.MachinePool, match types.GomegaMatcher) {
			errList := ValidateMachinePoolUpdate(newMachinePool, oldMachinePool)
			Expect(errList).To(match)
		},
		Entry("immutable providerID if set",
			&compute.MachinePool{
				Spec: compute.MachinePoolSpec{
					ProviderID: "foo",
				},
			},
			&compute.MachinePool{
				Spec: compute.MachinePoolSpec{
					ProviderID: "bar",
				},
			},
			ContainElement(ImmutableField("spec.providerID")),
		),
		Entry("mutable providerID if not set",
			&compute.MachinePool{
				Spec: compute.MachinePoolSpec{
					ProviderID: "foo",
				},
			},
			&compute.MachinePool{},
			Not(ContainElement(ImmutableField("spec.machinePoolRef"))),
		),
	)
})
