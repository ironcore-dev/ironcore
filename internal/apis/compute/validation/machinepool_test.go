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
	"github.com/onmetal/onmetal-api/internal/apis/compute"
	. "github.com/onmetal/onmetal-api/internal/testutils/validation"
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
