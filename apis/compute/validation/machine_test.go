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
	"github.com/onmetal/onmetal-api/apis/compute"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Machine", func() {
	DescribeTable("ValidateMachine",
		func(machine *compute.Machine, match types.GomegaMatcher) {
			errList := ValidateMachine(machine)
			Expect(errList).To(match)
		},
		Entry("missing name",
			&compute.Machine{},
			ContainElement(RequiredField("metadata.name")),
		),
		Entry("missing namespace",
			&compute.Machine{ObjectMeta: metav1.ObjectMeta{Name: "foo"}},
			ContainElement(RequiredField("metadata.namespace")),
		),
		Entry("bad name",
			&compute.Machine{ObjectMeta: metav1.ObjectMeta{Name: "foo*"}},
			ContainElement(InvalidField("metadata.name")),
		),
		Entry("no machine class ref",
			&compute.Machine{},
			ContainElement(RequiredField("spec.machineClassRef")),
		),
		Entry("invalid machine class ref name",
			&compute.Machine{
				Spec: compute.MachineSpec{
					MachineClassRef: corev1.LocalObjectReference{Name: "foo*"},
				},
			},
			ContainElement(InvalidField("spec.machineClassRef.name")),
		),
		Entry("invalid volume name",
			&compute.Machine{
				Spec: compute.MachineSpec{
					Volumes: []compute.Volume{{Name: "foo*"}},
				},
			},
			ContainElement(InvalidField("spec.volume[0].name")),
		),
		Entry("invalid volumeClaimRef name",
			&compute.Machine{
				Spec: compute.MachineSpec{
					Volumes: []compute.Volume{
						{Name: "foo", VolumeSource: compute.VolumeSource{
							VolumeClaimRef: &corev1.LocalObjectReference{Name: "foo*"}},
						},
					},
				},
			},
			ContainElement(InvalidField("spec.volume[0].volumeClaimRef.name")),
		),
		Entry("invalid ignition ref name",
			&compute.Machine{
				Spec: compute.MachineSpec{
					IgnitionRef: &commonv1alpha1.ConfigMapKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "foo*",
						},
					},
				},
			},
			ContainElement(InvalidField("spec.ignitionRef.name")),
		),
	)

	DescribeTable("ValidateMachineUpdate",
		func(newMachine, oldMachine *compute.Machine, match types.GomegaMatcher) {
			errList := ValidateMachineUpdate(newMachine, oldMachine)
			Expect(errList).To(match)
		},
		Entry("immutable machineClassRef",
			&compute.Machine{
				Spec: compute.MachineSpec{
					MachineClassRef: corev1.LocalObjectReference{Name: "foo"},
				},
			},
			&compute.Machine{
				Spec: compute.MachineSpec{
					MachineClassRef: corev1.LocalObjectReference{Name: "bar"},
				},
			},
			ContainElement(ImmutableField("spec.machineClassRef")),
		),
		Entry("immutable machinePoolRef if set",
			&compute.Machine{
				Spec: compute.MachineSpec{
					MachinePoolRef: corev1.LocalObjectReference{Name: "foo"},
				},
			},
			&compute.Machine{
				Spec: compute.MachineSpec{
					MachinePoolRef: corev1.LocalObjectReference{Name: "bar"},
				},
			},
			ContainElement(ImmutableField("spec.machinePoolRef")),
		),
		Entry("mutable machinePoolRef if not set",
			&compute.Machine{
				Spec: compute.MachineSpec{
					MachinePoolRef: corev1.LocalObjectReference{Name: "foo"},
				},
			},
			&compute.Machine{},
			Not(ContainElement(ImmutableField("spec.machinePoolRef"))),
		),
	)
})
