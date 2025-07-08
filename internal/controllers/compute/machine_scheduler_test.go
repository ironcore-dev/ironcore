// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package compute

import (
	"fmt"
	"math"

	. "github.com/ironcore-dev/ironcore/utils/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"

	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
)

var _ = Describe("MachineScheduler", func() {
	ns := SetupNamespace(&k8sClient)
	machineClass := SetupMachineClass()

	It("should schedule machines on machine pools", func(ctx SpecContext) {
		By("creating a machine pool")
		machinePool := &computev1alpha1.MachinePool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-pool-",
			},
		}
		Expect(k8sClient.Create(ctx, machinePool)).To(Succeed(), "failed to create machine pool")

		By("patching the machine pool status to contain a machine class")
		Eventually(UpdateStatus(machinePool, func() {
			machinePool.Status.AvailableMachineClasses = []corev1.LocalObjectReference{{Name: machineClass.Name}}
			machinePool.Status.Allocatable = corev1alpha1.ResourceList{
				corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeMachineClass, machineClass.Name): resource.MustParse("10"),
			}
		})).Should(Succeed())

		By("creating a machine w/ the requested machine class")
		machine := &computev1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-machine-",
			},
			Spec: computev1alpha1.MachineSpec{
				Image:           "my-image",
				MachineClassRef: corev1.LocalObjectReference{Name: machineClass.Name},
			},
		}
		Expect(k8sClient.Create(ctx, machine)).To(Succeed(), "failed to create machine")

		By("waiting for the machine to be scheduled onto the machine pool")
		Eventually(Object(machine)).Should(SatisfyAll(
			HaveField("Spec.MachinePoolRef", Equal(&corev1.LocalObjectReference{Name: machinePool.Name})),
			HaveField("Status.State", Equal(computev1alpha1.MachineStatePending)),
		))
	})

	It("should schedule schedule machines onto machine pools if the pool becomes available later than the machine", func(ctx SpecContext) {
		By("creating a machine w/ the requested machine class")
		machine := &computev1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-machine-",
			},
			Spec: computev1alpha1.MachineSpec{
				Image:           "my-image",
				MachineClassRef: corev1.LocalObjectReference{Name: machineClass.Name},
			},
		}
		Expect(k8sClient.Create(ctx, machine)).To(Succeed(), "failed to create machine")

		By("waiting for the machine to indicate it is pending")
		Eventually(Object(machine)).Should(SatisfyAll(
			HaveField("Spec.MachinePoolRef", BeNil()),
			HaveField("Status.State", Equal(computev1alpha1.MachineStatePending)),
		))

		By("creating a machine pool")
		machinePool := &computev1alpha1.MachinePool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-pool-",
			},
		}
		Expect(k8sClient.Create(ctx, machinePool)).To(Succeed(), "failed to create machine pool")

		By("patching the machine pool status to contain a machine class")
		Eventually(UpdateStatus(machinePool, func() {
			machinePool.Status.AvailableMachineClasses = []corev1.LocalObjectReference{{Name: machineClass.Name}}
			machinePool.Status.Allocatable = corev1alpha1.ResourceList{
				corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeMachineClass, machineClass.Name): resource.MustParse("10"),
			}
		})).Should(Succeed())

		By("waiting for the machine to be scheduled onto the machine pool")
		Eventually(Object(machine)).Should(SatisfyAll(
			HaveField("Spec.MachinePoolRef", Equal(&corev1.LocalObjectReference{Name: machinePool.Name})),
			HaveField("Status.State", Equal(computev1alpha1.MachineStatePending)),
		))
	})

	It("should schedule onto machine pools with matching labels", func(ctx SpecContext) {
		By("creating a machine pool w/o matching labels")
		machinePoolNoMatchingLabels := &computev1alpha1.MachinePool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-pool-",
			},
		}
		Expect(k8sClient.Create(ctx, machinePoolNoMatchingLabels)).To(Succeed(), "failed to create machine pool")

		By("patching the machine pool status to contain a machine class")
		Eventually(UpdateStatus(machinePoolNoMatchingLabels, func() {
			machinePoolNoMatchingLabels.Status.AvailableMachineClasses = []corev1.LocalObjectReference{{Name: machineClass.Name}}
			machinePoolNoMatchingLabels.Status.Allocatable = corev1alpha1.ResourceList{
				corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeMachineClass, machineClass.Name): resource.MustParse("10"),
			}
		})).Should(Succeed())

		By("creating a machine pool w/ matching labels")
		machinePoolMatchingLabels := &computev1alpha1.MachinePool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-pool-",
				Labels: map[string]string{
					"foo": "bar",
				},
			},
		}
		Expect(k8sClient.Create(ctx, machinePoolMatchingLabels)).To(Succeed(), "failed to create machine pool")

		By("patching the machine pool status to contain a machine class")
		Eventually(UpdateStatus(machinePoolMatchingLabels, func() {
			machinePoolMatchingLabels.Status.AvailableMachineClasses = []corev1.LocalObjectReference{{Name: machineClass.Name}}
			machinePoolMatchingLabels.Status.Allocatable = corev1alpha1.ResourceList{
				corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeMachineClass, machineClass.Name): resource.MustParse("10"),
			}
		})).Should(Succeed())

		By("creating a machine w/ the requested machine class")
		machine := &computev1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-machine-",
			},
			Spec: computev1alpha1.MachineSpec{
				Image: "my-image",
				MachinePoolSelector: map[string]string{
					"foo": "bar",
				},
				MachineClassRef: corev1.LocalObjectReference{
					Name: machineClass.Name,
				},
			},
		}
		Expect(k8sClient.Create(ctx, machine)).To(Succeed(), "failed to create machine")

		By("waiting for the machine to be scheduled onto the machine pool")
		Eventually(Object(machine)).Should(SatisfyAll(
			HaveField("Spec.MachinePoolRef", Equal(&corev1.LocalObjectReference{Name: machinePoolMatchingLabels.Name})),
			HaveField("Status.State", Equal(computev1alpha1.MachineStatePending)),
		))
	})

	It("should schedule a machine with corresponding tolerations onto a machine pool with taints", func(ctx SpecContext) {
		By("creating a machine pool w/ taints")
		taintedMachinePool := &computev1alpha1.MachinePool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-pool-",
			},
			Spec: computev1alpha1.MachinePoolSpec{
				Taints: []commonv1alpha1.Taint{
					{
						Key:    "key",
						Value:  "value",
						Effect: commonv1alpha1.TaintEffectNoSchedule,
					},
					{
						Key:    "key1",
						Effect: commonv1alpha1.TaintEffectNoSchedule,
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, taintedMachinePool)).To(Succeed(), "failed to create the machine pool")

		By("patching the machine pool status to contain a machine class")
		Eventually(UpdateStatus(taintedMachinePool, func() {
			taintedMachinePool.Status.AvailableMachineClasses = []corev1.LocalObjectReference{{Name: machineClass.Name}}
			taintedMachinePool.Status.Allocatable = corev1alpha1.ResourceList{
				corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeMachineClass, machineClass.Name): resource.MustParse("10"),
			}
		})).Should(Succeed())

		By("creating a machine")
		machine := &computev1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-machine-",
			},
			Spec: computev1alpha1.MachineSpec{
				Image: "my-image",
				MachineClassRef: corev1.LocalObjectReference{
					Name: machineClass.Name,
				},
			},
		}
		Expect(k8sClient.Create(ctx, machine)).To(Succeed(), "failed to create the machine")

		By("observing the machine isn't scheduled onto the machine pool")
		Consistently(Object(machine)).Should(SatisfyAll(
			HaveField("Spec.MachinePoolRef", BeNil()),
		))

		By("patching the machine to contain only one of the corresponding tolerations")
		machineBase := machine.DeepCopy()
		machine.Spec.Tolerations = append(machine.Spec.Tolerations, commonv1alpha1.Toleration{
			Key:      "key",
			Value:    "value",
			Effect:   commonv1alpha1.TaintEffectNoSchedule,
			Operator: commonv1alpha1.TolerationOpEqual,
		})
		Expect(k8sClient.Patch(ctx, machine, client.MergeFrom(machineBase))).To(Succeed(), "failed to patch the machine's spec")

		By("observing the machine isn't scheduled onto the machine pool")
		Consistently(Object(machine)).Should(SatisfyAll(
			HaveField("Spec.MachinePoolRef", BeNil()),
		))

		By("patching the machine to contain all of the corresponding tolerations")
		machineBase = machine.DeepCopy()
		machine.Spec.Tolerations = append(machine.Spec.Tolerations, commonv1alpha1.Toleration{
			Key:      "key1",
			Effect:   commonv1alpha1.TaintEffectNoSchedule,
			Operator: commonv1alpha1.TolerationOpExists,
		})
		Expect(k8sClient.Patch(ctx, machine, client.MergeFrom(machineBase))).To(Succeed(), "failed to patch the machine's spec")

		By("observing the machine is scheduled onto the machine pool")
		Eventually(Object(machine)).Should(SatisfyAll(
			HaveField("Spec.MachinePoolRef", Equal(&corev1.LocalObjectReference{Name: taintedMachinePool.Name})),
		))
	})

	It("should schedule machine on pool with most allocatable resources", func(ctx SpecContext) {
		By("creating a machine pool")
		machinePool := &computev1alpha1.MachinePool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-pool-",
			},
		}
		Expect(k8sClient.Create(ctx, machinePool)).To(Succeed(), "failed to create machine pool")

		By("patching the machine pool status to contain a machine class")
		Eventually(UpdateStatus(machinePool, func() {
			machinePool.Status.AvailableMachineClasses = []corev1.LocalObjectReference{{Name: machineClass.Name}}
			machinePool.Status.Allocatable = corev1alpha1.ResourceList{
				corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeMachineClass, machineClass.Name): resource.MustParse("10"),
			}
		})).Should(Succeed())

		By("creating a second machine pool")
		secondMachinePool := &computev1alpha1.MachinePool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "second-test-pool-",
			},
		}
		Expect(k8sClient.Create(ctx, secondMachinePool)).To(Succeed(), "failed to create the second machine pool")

		By("creating a second machine class")
		secondMachineClass := &computev1alpha1.MachineClass{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "second-machine-class-",
			},
			Capabilities: corev1alpha1.ResourceList{
				corev1alpha1.ResourceCPU:    resource.MustParse("1"),
				corev1alpha1.ResourceMemory: resource.MustParse("1Gi"),
			},
		}
		Expect(k8sClient.Create(ctx, secondMachineClass)).To(Succeed(), "failed to create second machine class")

		By("patching the second machine pool status to contain a both machine classes")
		Eventually(UpdateStatus(secondMachinePool, func() {
			secondMachinePool.Status.AvailableMachineClasses = []corev1.LocalObjectReference{
				{Name: machineClass.Name},
				{Name: secondMachineClass.Name},
			}
			secondMachinePool.Status.Allocatable = corev1alpha1.ResourceList{
				corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeMachineClass, machineClass.Name):       resource.MustParse("5"),
				corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeMachineClass, secondMachineClass.Name): resource.MustParse("100"),
			}
		})).Should(Succeed())

		By("creating a machine")
		machine := &computev1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-machine-",
			},
			Spec: computev1alpha1.MachineSpec{
				Image: "my-image",
				MachineClassRef: corev1.LocalObjectReference{
					Name: machineClass.Name,
				},
			},
		}
		Expect(k8sClient.Create(ctx, machine)).To(Succeed(), "failed to create the machine")

		By("checking that the machine is scheduled onto the machine pool")
		Eventually(Object(machine)).Should(SatisfyAll(
			HaveField("Spec.MachinePoolRef.Name", Equal(machinePool.Name)),
		))
	})

	It("should schedule machines evenly on pools", func(ctx SpecContext) {
		By("creating a machine pool")
		machinePool := &computev1alpha1.MachinePool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-pool-",
			},
		}
		Expect(k8sClient.Create(ctx, machinePool)).To(Succeed(), "failed to create machine pool")

		By("patching the machine pool status to contain a machine class")
		Eventually(UpdateStatus(machinePool, func() {
			machinePool.Status.AvailableMachineClasses = []corev1.LocalObjectReference{{Name: machineClass.Name}}
			machinePool.Status.Allocatable = corev1alpha1.ResourceList{
				corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeMachineClass, machineClass.Name): resource.MustParse("50"),
			}
		})).Should(Succeed())

		By("creating a second machine pool")
		secondMachinePool := &computev1alpha1.MachinePool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "second-test-pool-",
			},
		}
		Expect(k8sClient.Create(ctx, secondMachinePool)).To(Succeed(), "failed to create the second machine pool")

		By("patching the second machine pool status to contain a both machine classes")
		Eventually(UpdateStatus(secondMachinePool, func() {
			secondMachinePool.Status.AvailableMachineClasses = []corev1.LocalObjectReference{
				{Name: machineClass.Name},
			}
			secondMachinePool.Status.Allocatable = corev1alpha1.ResourceList{
				corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeMachineClass, machineClass.Name): resource.MustParse("50"),
			}
		})).Should(Succeed())

		By("creating machines")
		var machines []*computev1alpha1.Machine
		for i := 0; i < 50; i++ {
			machine := &computev1alpha1.Machine{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    ns.Name,
					GenerateName: fmt.Sprintf("test-machine-%d-", i),
				},
				Spec: computev1alpha1.MachineSpec{
					Image: "my-image",
					MachineClassRef: corev1.LocalObjectReference{
						Name: machineClass.Name,
					},
				},
			}
			Expect(k8sClient.Create(ctx, machine)).To(Succeed(), "failed to create the machine")
			machines = append(machines, machine)
		}

		By("checking that every machine is scheduled onto a machine pool")
		var numInstancesPool1, numInstancesPool2 int64
		for i := 0; i < 50; i++ {
			Eventually(Object(machines[i])).Should(SatisfyAll(
				HaveField("Spec.MachinePoolRef", Not(BeNil())),
			))

			switch machines[i].Spec.MachinePoolRef.Name {
			case machinePool.Name:
				numInstancesPool1++
			case secondMachinePool.Name:
				numInstancesPool2++
			}
		}

		By("checking that machine are roughly distributed")
		Expect(math.Abs(float64(numInstancesPool1 - numInstancesPool2))).To(BeNumerically("<", 5))
	})

	It("should schedule a machines once the capacity is sufficient", func(ctx SpecContext) {
		By("creating a machine pool")
		machinePool := &computev1alpha1.MachinePool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-pool-",
			},
		}
		Expect(k8sClient.Create(ctx, machinePool)).To(Succeed(), "failed to create machine pool")
		By("patching the machine pool status to contain a machine class")
		Eventually(UpdateStatus(machinePool, func() {
			machinePool.Status.AvailableMachineClasses = []corev1.LocalObjectReference{{Name: machineClass.Name}}
		})).Should(Succeed())

		By("creating a machine")
		machine := &computev1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-machine-",
			},
			Spec: computev1alpha1.MachineSpec{
				Image: "my-image",
				MachineClassRef: corev1.LocalObjectReference{
					Name: machineClass.Name,
				},
			},
		}
		Expect(k8sClient.Create(ctx, machine)).To(Succeed(), "failed to create the machine")

		By("checking that the machine isn't scheduled onto the machine pool")
		Consistently(Object(machine)).Should(SatisfyAll(
			HaveField("Spec.MachinePoolRef", BeNil()),
		))

		By("patching the machine pool status to contain a machine class")
		Eventually(UpdateStatus(machinePool, func() {
			machinePool.Status.Allocatable = corev1alpha1.ResourceList{
				corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeMachineClass, machineClass.Name): resource.MustParse("10"),
			}
		})).Should(Succeed())

		By("checking that the machine is scheduled onto the machine pool")
		Eventually(Object(machine)).Should(SatisfyAll(
			HaveField("Spec.MachinePoolRef", Equal(&corev1.LocalObjectReference{Name: machinePool.Name})),
		))
	})

	It("should schedule machine on pool with correctly allocatable resources", func(ctx SpecContext) {
		By("creating a machine pool")
		machinePool := &computev1alpha1.MachinePool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-pool-",
			},
		}
		Expect(k8sClient.Create(ctx, machinePool)).To(Succeed(), "failed to create machine pool")

		By("patching the machine pool status to contain a machine class")
		Eventually(UpdateStatus(machinePool, func() {
			machinePool.Status.AvailableMachineClasses = []corev1.LocalObjectReference{{Name: machineClass.Name}}
			machinePool.Status.Allocatable = corev1alpha1.ResourceList{
				corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeMachineClass, machineClass.Name): resource.MustParse("5"),
			}
		})).Should(Succeed())

		By("creating a second machine pool")
		secondMachinePool := &computev1alpha1.MachinePool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "second-test-pool-",
			},
		}
		Expect(k8sClient.Create(ctx, secondMachinePool)).To(Succeed(), "failed to create the second machine pool")

		By("creating a second machine class")
		secondMachineClass := &computev1alpha1.MachineClass{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "second-machine-class-",
			},
			Capabilities: corev1alpha1.ResourceList{
				corev1alpha1.ResourceCPU:    resource.MustParse("1"),
				corev1alpha1.ResourceMemory: resource.MustParse("1Gi"),
				"type-a.vendor.com/gpu":     resource.MustParse("1"),
			},
		}
		Expect(k8sClient.Create(ctx, secondMachineClass)).To(Succeed(), "failed to create second machine class")

		By("patching the second machine pool status to contain a both machine classes")
		Eventually(UpdateStatus(secondMachinePool, func() {
			secondMachinePool.Status.AvailableMachineClasses = []corev1.LocalObjectReference{
				{Name: machineClass.Name},
				{Name: secondMachineClass.Name},
			}
			secondMachinePool.Status.Allocatable = corev1alpha1.ResourceList{
				corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeMachineClass, machineClass.Name):       resource.MustParse("5"),
				corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeMachineClass, secondMachineClass.Name): resource.MustParse("5"),
			}
		})).Should(Succeed())

		By("creating a machine")
		machine := &computev1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-machine-",
			},
			Spec: computev1alpha1.MachineSpec{
				Image: "my-image",
				MachineClassRef: corev1.LocalObjectReference{
					Name: secondMachineClass.Name,
				},
			},
		}
		Expect(k8sClient.Create(ctx, machine)).To(Succeed(), "failed to create the machine")

		By("checking that the machine is scheduled onto the machine pool")
		Eventually(Object(machine)).Should(SatisfyAll(
			HaveField("Spec.MachinePoolRef.Name", Equal(secondMachinePool.Name)),
		))
	})
})
