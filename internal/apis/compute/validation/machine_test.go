// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	"github.com/ironcore-dev/ironcore/internal/apis/compute"
	"github.com/ironcore-dev/ironcore/internal/apis/ipam"
	"github.com/ironcore-dev/ironcore/internal/apis/networking"
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
		Entry("invalid empty disk size limit quantity",
			&compute.Machine{
				Spec: compute.MachineSpec{
					Volumes: []compute.Volume{
						{
							Name:   "foo",
							Device: "oda",
							VolumeSource: compute.VolumeSource{
								LocalDisk: &compute.LocalDiskVolumeSource{SizeLimit: mustParseNewQuantity("-1Gi")},
							},
						},
					},
				},
			},
			ContainElement(InvalidField("spec.volume[0].localDisk.sizeLimit")),
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

	DescribeTable("ValidateMachineNetworkInterface",
		func(machine *compute.Machine, match types.GomegaMatcher) {
			errList := ValidateMachine(machine)
			Expect(errList).To(match)
		},
		Entry("invalid networkInterface name",
			&compute.Machine{
				Spec: compute.MachineSpec{
					NetworkInterfaces: []compute.NetworkInterface{{Name: "bar*"}},
				},
			},
			ContainElement(InvalidField("spec.networkInterface[0].name")),
		),
		Entry("duplicate networkInterface name",
			&compute.Machine{
				Spec: compute.MachineSpec{
					NetworkInterfaces: []compute.NetworkInterface{
						{Name: "foo"},
						{Name: "foo"},
					},
				},
			},
			ContainElement(DuplicateField("spec.networkInterface[1].name")),
		),
		Entry("invalid networkInterfaceRef name",
			&compute.Machine{
				Spec: compute.MachineSpec{
					NetworkInterfaces: []compute.NetworkInterface{
						{
							Name: "bar",
							NetworkInterfaceSource: compute.NetworkInterfaceSource{
								NetworkInterfaceRef: &corev1.LocalObjectReference{Name: "foo*"},
							},
						},
					},
				},
			},
			ContainElement(InvalidField("spec.networkInterface[0].networkInterfaceRef.name")),
		),
		Entry("invalid networkInterface prefix length for IPv4 IPSource",
			&compute.Machine{
				Spec: compute.MachineSpec{
					NetworkInterfaces: []compute.NetworkInterface{
						{
							Name: "foo",
							NetworkInterfaceSource: compute.NetworkInterfaceSource{
								Ephemeral: &compute.EphemeralNetworkInterfaceSource{
									NetworkInterfaceTemplate: &networking.NetworkInterfaceTemplateSpec{
										Spec: networking.NetworkInterfaceSpec{
											NetworkRef: corev1.LocalObjectReference{Name: "bar"},
											IPs: []networking.IPSource{{
												Ephemeral: &networking.EphemeralPrefixSource{
													PrefixTemplate: &ipam.PrefixTemplateSpec{
														Spec: ipam.PrefixSpec{
															IPFamily: corev1.IPv4Protocol,
															Prefix:   commonv1alpha1.MustParseNewIPPrefix("10.0.0.0/24"),
														},
													}},
											}},
										},
									},
								},
							},
						},
					},
				},
			},
			ContainElement(ForbiddenField("spec.networkInterface[0].ephemeral.networkInterfaceTemplate.ips[0].ephemeral.prefixTemplate.spec.prefix")),
		),
		Entry("invalid ephemral networkInterface prefix length for IPv6 IPSource",
			&compute.Machine{
				Spec: compute.MachineSpec{
					NetworkInterfaces: []compute.NetworkInterface{
						{
							Name: "foo",
							NetworkInterfaceSource: compute.NetworkInterfaceSource{
								Ephemeral: &compute.EphemeralNetworkInterfaceSource{
									NetworkInterfaceTemplate: &networking.NetworkInterfaceTemplateSpec{
										Spec: networking.NetworkInterfaceSpec{
											NetworkRef: corev1.LocalObjectReference{Name: "bar"},
											IPs: []networking.IPSource{{
												Ephemeral: &networking.EphemeralPrefixSource{
													PrefixTemplate: &ipam.PrefixTemplateSpec{
														Spec: ipam.PrefixSpec{
															IPFamily: corev1.IPv6Protocol,
															Prefix:   commonv1alpha1.MustParseNewIPPrefix("2001:db8::/64"),
														},
													}},
											}},
										},
									},
								},
							},
						},
					},
				},
			},
			ContainElement(ForbiddenField("spec.networkInterface[0].ephemeral.networkInterfaceTemplate.ips[0].ephemeral.prefixTemplate.spec.prefix")),
		),
		Entry("invalid networkInterface prefix length derived from parent prefix for IPv4 IPSource",
			&compute.Machine{
				Spec: compute.MachineSpec{
					NetworkInterfaces: []compute.NetworkInterface{
						{
							Name: "foo",
							NetworkInterfaceSource: compute.NetworkInterfaceSource{
								Ephemeral: &compute.EphemeralNetworkInterfaceSource{
									NetworkInterfaceTemplate: &networking.NetworkInterfaceTemplateSpec{
										Spec: networking.NetworkInterfaceSpec{
											NetworkRef: corev1.LocalObjectReference{Name: "bar"},
											IPFamilies: []corev1.IPFamily{"IPv4"},
											IPs: []networking.IPSource{{
												Ephemeral: &networking.EphemeralPrefixSource{
													PrefixTemplate: &ipam.PrefixTemplateSpec{
														Spec: ipam.PrefixSpec{
															IPFamily: corev1.IPv4Protocol,
															ParentSelector: &metav1.LabelSelector{
																MatchLabels: map[string]string{"foo": "bar"},
															},
															PrefixLength: 40,
														},
													}},
											}},
										},
									},
								},
							},
						},
					},
				},
			},
			ContainElement(ForbiddenField("spec.networkInterface[0].ephemeral.networkInterfaceTemplate.ips[0].ephemeral.prefixTemplate.spec.prefixLength")),
		),
		Entry("invalid ephemral networkInterface prefix length for IPv6 IPSource",
			&compute.Machine{
				Spec: compute.MachineSpec{
					NetworkInterfaces: []compute.NetworkInterface{
						{
							Name: "foo",
							NetworkInterfaceSource: compute.NetworkInterfaceSource{
								Ephemeral: &compute.EphemeralNetworkInterfaceSource{
									NetworkInterfaceTemplate: &networking.NetworkInterfaceTemplateSpec{
										Spec: networking.NetworkInterfaceSpec{
											NetworkRef: corev1.LocalObjectReference{Name: "bar"},
											IPFamilies: []corev1.IPFamily{"IPv6"},
											IPs: []networking.IPSource{{
												Ephemeral: &networking.EphemeralPrefixSource{
													PrefixTemplate: &ipam.PrefixTemplateSpec{
														Spec: ipam.PrefixSpec{
															IPFamily: corev1.IPv6Protocol,
															ParentSelector: &metav1.LabelSelector{
																MatchLabels: map[string]string{"foo": "bar"},
															},
															PrefixLength: 132,
														},
													}},
											}},
										},
									},
								},
							},
						},
					},
				},
			},
			ContainElement(ForbiddenField("spec.networkInterface[0].ephemeral.networkInterfaceTemplate.ips[0].ephemeral.prefixTemplate.spec.prefixLength")),
		),
	)

	DescribeTable("ValidateMachineNetworkInterface prefix length derived from parent prefix",
		func(machine *compute.Machine, match types.GomegaMatcher) {
			errList := ValidateMachine(machine)
			Expect(errList).To(match)
		},
		Entry("valid networkInterface prefix length derived from parent prefix for IPv4 IPSource",
			&compute.Machine{
				Spec: compute.MachineSpec{
					NetworkInterfaces: []compute.NetworkInterface{
						{
							Name: "foo",
							NetworkInterfaceSource: compute.NetworkInterfaceSource{
								Ephemeral: &compute.EphemeralNetworkInterfaceSource{
									NetworkInterfaceTemplate: &networking.NetworkInterfaceTemplateSpec{
										Spec: networking.NetworkInterfaceSpec{
											NetworkRef: corev1.LocalObjectReference{Name: "bar"},
											IPFamilies: []corev1.IPFamily{"IPv4"},
											IPs: []networking.IPSource{{
												Ephemeral: &networking.EphemeralPrefixSource{
													PrefixTemplate: &ipam.PrefixTemplateSpec{
														Spec: ipam.PrefixSpec{
															IPFamily: corev1.IPv4Protocol,
															ParentRef: &corev1.LocalObjectReference{
																Name: "root",
															},
															PrefixLength: 32,
														},
													}},
											}},
										},
									},
								},
							},
						},
					},
				},
			},
			Not(ContainElement(ForbiddenField("spec.networkInterface[0].ephemeral.networkInterfaceTemplate.ips[0].ephemeral.prefixTemplate.spec.prefixLength"))),
		),
		Entry("valid ephemral networkInterface prefix length derived from parent prefix for IPv6 IPSource",
			&compute.Machine{
				Spec: compute.MachineSpec{
					NetworkInterfaces: []compute.NetworkInterface{
						{
							Name: "foo",
							NetworkInterfaceSource: compute.NetworkInterfaceSource{
								Ephemeral: &compute.EphemeralNetworkInterfaceSource{
									NetworkInterfaceTemplate: &networking.NetworkInterfaceTemplateSpec{
										Spec: networking.NetworkInterfaceSpec{
											NetworkRef: corev1.LocalObjectReference{Name: "bar"},
											IPFamilies: []corev1.IPFamily{"IPv6"},
											IPs: []networking.IPSource{{
												Ephemeral: &networking.EphemeralPrefixSource{
													PrefixTemplate: &ipam.PrefixTemplateSpec{
														Spec: ipam.PrefixSpec{
															IPFamily: corev1.IPv6Protocol,
															ParentRef: &corev1.LocalObjectReference{
																Name: "root",
															},
															PrefixLength: 128,
														},
													}},
											}},
										},
									},
								},
							},
						},
					},
				},
			},
			Not(ContainElement(ForbiddenField("spec.networkInterface[0].ephemeral.networkInterfaceTemplate.ips[0].ephemeral.prefixTemplate.spec.prefixLength"))),
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
		Entry("mutate local boot disk image",
			&compute.Machine{
				Spec: compute.MachineSpec{
					Volumes: []compute.Volume{
						{
							Name:   "rootdisk",
							Device: "vda",
							VolumeSource: compute.VolumeSource{
								LocalDisk: &compute.LocalDiskVolumeSource{
									Image: "newImage",
								},
							},
						},
					},
				},
			},
			&compute.Machine{
				Spec: compute.MachineSpec{
					Volumes: []compute.Volume{
						{
							Name:   "rootdisk",
							Device: "vda",
							VolumeSource: compute.VolumeSource{
								LocalDisk: &compute.LocalDiskVolumeSource{
									Image: "oldImage",
								},
							},
						},
					},
				},
			},
			ContainElement(ImmutableField("spec.volumes[rootdisk].localDisk.image")),
		),
		Entry("remove local boot disk",
			&compute.Machine{
				Spec: compute.MachineSpec{},
			},
			&compute.Machine{
				Spec: compute.MachineSpec{
					Volumes: []compute.Volume{
						{
							Name:   "rootdisk",
							Device: "vda",
							VolumeSource: compute.VolumeSource{
								LocalDisk: &compute.LocalDiskVolumeSource{
									Image: "image",
								},
							},
						},
					},
				},
			},
			ContainElement(InvalidField("spec.volumes[rootdisk].localDisk.image")),
		),
		Entry("duplicate volume name",
			&compute.Machine{
				Spec: compute.MachineSpec{
					Volumes: []compute.Volume{
						{Name: "foo"},
						{Name: "bar"},
					},
				},
				Status: compute.MachineStatus{
					Volumes: []compute.VolumeStatus{
						{Name: "foo"},
						{Name: "bar"},
					},
				},
			},
			&compute.Machine{
				Spec: compute.MachineSpec{
					Volumes: []compute.Volume{
						{Name: "foo"},
					},
				},
				Status: compute.MachineStatus{
					Volumes: []compute.VolumeStatus{
						{Name: "foo"},
						{Name: "bar"},
					},
				},
			},
			ContainElement(DuplicateField("spec.volume[1].name")),
		),
	)
})
