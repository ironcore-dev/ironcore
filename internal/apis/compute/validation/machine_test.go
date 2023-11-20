// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	"github.com/ironcore-dev/ironcore/internal/apis/compute"
	. "github.com/ironcore-dev/ironcore/internal/testutils/validation"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func mustParseNewQuantity(s string) *resource.Quantity {
	q := resource.MustParse(s)
	return &q
}

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
		Entry("invalid machine power",
			&compute.Machine{
				Spec: compute.MachineSpec{
					Power: "invalid",
				},
			},
			ContainElement(NotSupportedField("spec.power")),
		),
		Entry("valid machine power",
			&compute.Machine{
				Spec: compute.MachineSpec{
					Power: compute.PowerOn,
				},
			},
			Not(ContainElement(NotSupportedField("spec.power"))),
		),
		Entry("no image",
			&compute.Machine{},
			Not(ContainElement(RequiredField("spec.image"))),
		),
		Entry("invalid machine class ref name",
			&compute.Machine{
				Spec: compute.MachineSpec{
					MachineClassRef: corev1.LocalObjectReference{Name: "foo*"},
				},
			},
			ContainElement(InvalidField("spec.machineClassRef.name")),
		),
		Entry("valid machine pool ref subdomain name",
			&compute.Machine{
				Spec: compute.MachineSpec{
					MachinePoolRef: &corev1.LocalObjectReference{Name: "foo.bar.baz"},
				},
			},
			Not(ContainElement(InvalidField("spec.machinePoolRef.name"))),
		),
		Entry("invalid volume name",
			&compute.Machine{
				Spec: compute.MachineSpec{
					Volumes: []compute.Volume{{Name: "foo*"}},
				},
			},
			ContainElement(InvalidField("spec.volume[0].name")),
		),
		Entry("duplicate volume name",
			&compute.Machine{
				Spec: compute.MachineSpec{
					Volumes: []compute.Volume{
						{Name: "foo"},
						{Name: "foo"},
					},
				},
			},
			ContainElement(DuplicateField("spec.volume[1].name")),
		),
		Entry("invalid volumeRef name",
			&compute.Machine{
				Spec: compute.MachineSpec{
					Volumes: []compute.Volume{
						{
							Name:   "foo",
							Device: "oda",
							VolumeSource: compute.VolumeSource{
								VolumeRef: &corev1.LocalObjectReference{Name: "foo*"},
							},
						},
					},
				},
			},
			ContainElement(InvalidField("spec.volume[0].volumeRef.name")),
		),
		Entry("invalid empty disk size limit quantity",
			&compute.Machine{
				Spec: compute.MachineSpec{
					Volumes: []compute.Volume{
						{
							Name:   "foo",
							Device: "oda",
							VolumeSource: compute.VolumeSource{
								EmptyDisk: &compute.EmptyDiskVolumeSource{SizeLimit: mustParseNewQuantity("-1Gi")},
							},
						},
					},
				},
			},
			ContainElement(InvalidField("spec.volume[0].emptyDisk.sizeLimit")),
		),
		Entry("duplicate machine volume device",
			&compute.Machine{
				Spec: compute.MachineSpec{
					Volumes: []compute.Volume{
						{
							Device: "vda",
						},
						{
							Device: "vda",
						},
					},
				},
			},
			ContainElement(DuplicateField("spec.volume[1].device")),
		),
		Entry("empty machine volume device",
			&compute.Machine{
				Spec: compute.MachineSpec{
					Volumes: []compute.Volume{
						{},
					},
				},
			},
			ContainElement(InvalidField("spec.volume[0].device")),
		),
		Entry("invalid volume device",
			&compute.Machine{
				Spec: compute.MachineSpec{
					Volumes: []compute.Volume{
						{
							Device: "foobar",
						},
					},
				},
			},
			ContainElement(InvalidField("spec.volume[0].device")),
		),
		Entry("invalid ignition ref name",
			&compute.Machine{
				Spec: compute.MachineSpec{
					IgnitionRef: &commonv1alpha1.SecretKeySelector{
						Name: "foo*",
					},
				},
			},
			ContainElement(InvalidField("spec.ignitionRef.name")),
		),
		Entry("invalid ignition ref name",
			&compute.Machine{
				Spec: compute.MachineSpec{
					ImagePullSecretRef: &corev1.LocalObjectReference{
						Name: "foo*",
					},
				},
			},
			ContainElement(InvalidField("spec.imagePullSecretRef.name")),
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
					MachinePoolRef: &corev1.LocalObjectReference{Name: "foo"},
				},
			},
			&compute.Machine{
				Spec: compute.MachineSpec{
					MachinePoolRef: &corev1.LocalObjectReference{Name: "bar"},
				},
			},
			ContainElement(ImmutableField("spec.machinePoolRef")),
		),
		Entry("mutable machinePoolRef if not set",
			&compute.Machine{
				Spec: compute.MachineSpec{
					MachinePoolRef: &corev1.LocalObjectReference{Name: "foo"},
				},
			},
			&compute.Machine{},
			Not(ContainElement(ImmutableField("spec.machinePoolRef"))),
		),
	)
})
