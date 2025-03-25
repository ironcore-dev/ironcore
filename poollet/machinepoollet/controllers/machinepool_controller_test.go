// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package controllers_test

import (
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	testingmachine "github.com/ironcore-dev/ironcore/iri/testing/machine"
	"github.com/ironcore-dev/ironcore/utils/quota"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

var _ = Describe("MachinePoolController", func() {
	ns, machinePool, machineClass, srv := SetupTest()

	It("should calculate pool capacity", func(ctx SpecContext) {
		var (
			machineClassCapacity, machineClass2Capacity int64 = 4, 10
		)

		By("creating a second machine class")
		machineClass2 := &computev1alpha1.MachineClass{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-mc2-",
			},
			Capabilities: corev1alpha1.ResourceList{
				corev1alpha1.ResourceCPU:    resource.MustParse("2"),
				corev1alpha1.ResourceMemory: resource.MustParse("2Gi"),
			},
		}
		Expect(k8sClient.Create(ctx, machineClass2)).To(Succeed(), "failed to create machine class")

		srv.SetMachineClasses([]*testingmachine.FakeMachineClassStatus{
			{
				MachineClassStatus: iri.MachineClassStatus{
					MachineClass: &iri.MachineClass{
						Name: machineClass.Name,
						Capabilities: &iri.MachineClassCapabilities{
							CpuMillis:   machineClass.Capabilities.CPU().MilliValue(),
							MemoryBytes: machineClass.Capabilities.Memory().Value(),
						},
					},
					Quantity: machineClassCapacity,
				},
			},
			{
				MachineClassStatus: iri.MachineClassStatus{
					MachineClass: &iri.MachineClass{
						Name: machineClass2.Name,
						Capabilities: &iri.MachineClassCapabilities{
							CpuMillis:   machineClass2.Capabilities.CPU().MilliValue(),
							MemoryBytes: machineClass2.Capabilities.Memory().Value(),
						},
					},
					Quantity: machineClass2Capacity,
				},
			},
		})

		By("checking if the capacity is correct")
		Eventually(Object(machinePool)).Should(SatisfyAll(
			HaveField("Status.Capacity", Satisfy(func(capacity corev1alpha1.ResourceList) bool {
				return quota.Equals(capacity, corev1alpha1.ResourceList{
					corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeMachineClass, machineClass.Name):  *resource.NewQuantity(machineClassCapacity, resource.DecimalSI),
					corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeMachineClass, machineClass2.Name): *resource.NewQuantity(machineClass2Capacity, resource.DecimalSI),
				})
			})),
		))

		By("creating a machine")
		machine := &computev1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-machine",
				Namespace:    ns.Name,
			},
			Spec: computev1alpha1.MachineSpec{
				MachineClassRef: corev1.LocalObjectReference{
					Name: machineClass2.Name,
				},
				MachinePoolRef: &corev1.LocalObjectReference{
					Name: machinePool.Name,
				},
			},
		}
		Expect(k8sClient.Create(ctx, machine)).To(Succeed(), "failed to create machine")

		By("checking if the allocatable resources are correct")
		Eventually(Object(machinePool)).Should(SatisfyAll(
			HaveField("Status.Allocatable", Satisfy(func(allocatable corev1alpha1.ResourceList) bool {
				return quota.Equals(allocatable, corev1alpha1.ResourceList{
					corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeMachineClass, machineClass.Name):  *resource.NewQuantity(machineClassCapacity, resource.DecimalSI),
					corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeMachineClass, machineClass2.Name): *resource.NewQuantity(machineClass2Capacity-1, resource.DecimalSI),
				})
			})),
		))

		By("creating a second machine")
		machine2 := &computev1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-machine",
				Namespace:    ns.Name,
			},
			Spec: computev1alpha1.MachineSpec{
				MachineClassRef: corev1.LocalObjectReference{
					Name: machineClass.Name,
				},
				MachinePoolRef: &corev1.LocalObjectReference{
					Name: machinePool.Name,
				},
			},
		}
		Expect(k8sClient.Create(ctx, machine2)).To(Succeed(), "failed to create test machine class")

		By("checking if the allocatable resources are correct")
		Eventually(Object(machinePool)).Should(SatisfyAll(
			HaveField("Status.Allocatable", Satisfy(func(allocatable corev1alpha1.ResourceList) bool {
				return quota.Equals(allocatable, corev1alpha1.ResourceList{
					corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeMachineClass, machineClass.Name):  *resource.NewQuantity(machineClassCapacity-1, resource.DecimalSI),
					corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeMachineClass, machineClass2.Name): *resource.NewQuantity(machineClass2Capacity-1, resource.DecimalSI),
				})
			})),
		))
	})

	It("should add machine classes to pool", func(ctx SpecContext) {
		srv.SetMachineClasses([]*testingmachine.FakeMachineClassStatus{})

		By("creating a machine class")
		machineClass := &computev1alpha1.MachineClass{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-mc-1-",
			},
			Capabilities: corev1alpha1.ResourceList{
				corev1alpha1.ResourceCPU:    resource.MustParse("2"),
				corev1alpha1.ResourceMemory: resource.MustParse("2Gi"),
			},
		}
		Expect(k8sClient.Create(ctx, machineClass)).To(Succeed(), "failed to create test machine class")

		srv.SetMachineClasses([]*testingmachine.FakeMachineClassStatus{
			{
				MachineClassStatus: iri.MachineClassStatus{
					MachineClass: &iri.MachineClass{
						Name: machineClass.Name,
						Capabilities: &iri.MachineClassCapabilities{
							CpuMillis:   machineClass.Capabilities.CPU().MilliValue(),
							MemoryBytes: machineClass.Capabilities.Memory().Value(),
						},
					},
				},
			},
		})

		By("checking if the default machine class is present")
		Eventually(Object(machinePool)).Should(SatisfyAll(
			HaveField("Status.AvailableMachineClasses", Equal([]corev1.LocalObjectReference{
				{
					Name: machineClass.Name,
				},
			}))),
		)

		By("creating a second machine class")
		machineClass2 := &computev1alpha1.MachineClass{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-mc-2-",
			},
			Capabilities: corev1alpha1.ResourceList{
				corev1alpha1.ResourceCPU:    resource.MustParse("3"),
				corev1alpha1.ResourceMemory: resource.MustParse("4Gi"),
			},
		}
		Expect(k8sClient.Create(ctx, machineClass2)).To(Succeed(), "failed to create test machine class")

		Eventually(Object(machinePool)).Should(SatisfyAll(
			HaveField("Status.AvailableMachineClasses", HaveLen(1))),
		)

		srv.SetMachineClasses([]*testingmachine.FakeMachineClassStatus{
			{
				MachineClassStatus: iri.MachineClassStatus{
					MachineClass: &iri.MachineClass{
						Name: machineClass.Name,
						Capabilities: &iri.MachineClassCapabilities{
							CpuMillis:   machineClass.Capabilities.CPU().MilliValue(),
							MemoryBytes: machineClass.Capabilities.Memory().Value(),
						},
					},
				},
			},
			{
				MachineClassStatus: iri.MachineClassStatus{
					MachineClass: &iri.MachineClass{
						Name: machineClass2.Name,
						Capabilities: &iri.MachineClassCapabilities{
							CpuMillis:   machineClass2.Capabilities.CPU().MilliValue(),
							MemoryBytes: machineClass2.Capabilities.Memory().Value(),
						},
					},
				},
			},
		})

		By("checking if the second machine class is present")
		Eventually(Object(machinePool)).Should(SatisfyAll(
			HaveField("Status.AvailableMachineClasses", ConsistOf(
				corev1.LocalObjectReference{
					Name: machineClass.Name,
				},
				corev1.LocalObjectReference{
					Name: machineClass2.Name,
				},
			))),
		)
	})

})
