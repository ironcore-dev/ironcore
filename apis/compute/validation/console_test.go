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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Console", func() {
	DescribeTable("ValidateConsole",
		func(console *compute.Console, match types.GomegaMatcher) {
			errList := ValidateConsole(console)
			Expect(errList).To(match)
		},
		Entry("missing name",
			&compute.Console{},
			ContainElement(RequiredField("metadata.name")),
		),
		Entry("missing namespace",
			&compute.Console{ObjectMeta: metav1.ObjectMeta{Name: "foo"}},
			ContainElement(RequiredField("metadata.namespace")),
		),
		Entry("bad name",
			&compute.Console{ObjectMeta: metav1.ObjectMeta{Name: "foo*"}},
			ContainElement(InvalidField("metadata.name")),
		),
		Entry("no machine ref",
			&compute.Console{},
			ContainElement(RequiredField("spec.machineRef")),
		),
		Entry("invalid machine ref name",
			&compute.Console{
				Spec: compute.ConsoleSpec{
					MachineRef: corev1.LocalObjectReference{Name: "foo*"},
				},
			},
			ContainElement(InvalidField("spec.machineRef.name")),
		),
	)

	DescribeTable("ValidateConsoleUpdate",
		func(newConsole, oldConsole *compute.Console, match types.GomegaMatcher) {
			errList := ValidateConsoleUpdate(newConsole, oldConsole)
			Expect(errList).To(match)
		},
		Entry("immutable machineRef",
			&compute.Console{
				Spec: compute.ConsoleSpec{
					MachineRef: corev1.LocalObjectReference{Name: "foo"},
				},
			},
			&compute.Console{
				Spec: compute.ConsoleSpec{
					MachineRef: corev1.LocalObjectReference{Name: "bar"},
				},
			},
			ContainElement(ImmutableField("spec.machineRef")),
		),
	)
})
