// Copyright 2021 OnMetal authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package compute

import (
	. "github.com/onmetal/onmetal-api/utils/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"

	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	corev1alpha1 "github.com/onmetal/onmetal-api/api/core/v1alpha1"
)

var _ = Describe("MachineScheduler", func() {
	ctx := SetupContext()
	ns, machineClass := SetupTest(ctx)

	It("should schedule machines on machine pools", func() {
		By("creating a machine pool")
		machinePool := &computev1alpha1.MachinePool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-pool-",
			},
		}
		Expect(k8sClient.Create(ctx, machinePool)).To(Succeed(), "failed to create machine pool")

		By("patching the machine pool status to contain a machine class")
		machinePoolBase := machinePool.DeepCopy()
		machinePool.Status.AvailableMachineClasses = []corev1.LocalObjectReference{{Name: machineClass.Name}}
		machinePool.Status.Allocatable = corev1alpha1.ResourceList{
			corev1alpha1.ResourceCPU:    resource.MustParse("10"),
			corev1alpha1.ResourceMemory: resource.MustParse("10Gi"),
		}
		Expect(k8sClient.Status().Patch(ctx, machinePool, client.MergeFrom(machinePoolBase))).
			To(Succeed(), "failed to patch machine pool status")

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
		machineKey := client.ObjectKeyFromObject(machine)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, machineKey, machine)).To(Succeed(), "failed to get machine")
			g.Expect(machine.Spec.MachinePoolRef).To(Equal(&corev1.LocalObjectReference{Name: machinePool.Name}))
			g.Expect(machine.Status.State).To(Equal(computev1alpha1.MachineStatePending))
		}).Should(Succeed())
	})

	It("should schedule schedule machines onto machine pools if the pool becomes available later than the machine", func() {
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
		machineKey := client.ObjectKeyFromObject(machine)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, machineKey, machine)).To(Succeed())
			g.Expect(machine.Spec.MachinePoolRef).To(BeNil())
			g.Expect(machine.Status.State).To(Equal(computev1alpha1.MachineStatePending))
		}).Should(Succeed())

		By("creating a machine pool")
		machinePool := &computev1alpha1.MachinePool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-pool-",
			},
		}
		Expect(k8sClient.Create(ctx, machinePool)).To(Succeed(), "failed to create machine pool")

		By("patching the machine pool status to contain a machine class")
		machinePoolBase := machinePool.DeepCopy()
		machinePool.Status.AvailableMachineClasses = []corev1.LocalObjectReference{{Name: machineClass.Name}}
		machinePool.Status.Allocatable = corev1alpha1.ResourceList{
			corev1alpha1.ResourceCPU:    resource.MustParse("10"),
			corev1alpha1.ResourceMemory: resource.MustParse("10Gi"),
		}
		Expect(k8sClient.Status().Patch(ctx, machinePool, client.MergeFrom(machinePoolBase))).
			To(Succeed(), "failed to patch machine pool status")

		By("waiting for the machine to be scheduled onto the machine pool")
		Eventually(func() *corev1.LocalObjectReference {
			Expect(k8sClient.Get(ctx, machineKey, machine)).To(Succeed(), "failed to get machine")
			return machine.Spec.MachinePoolRef
		}).Should(Equal(&corev1.LocalObjectReference{Name: machinePool.Name}))
	})

	It("should schedule onto machine pools with matching labels", func() {
		By("creating a machine pool w/o matching labels")
		machinePoolNoMatchingLabels := &computev1alpha1.MachinePool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-pool-",
			},
		}
		Expect(k8sClient.Create(ctx, machinePoolNoMatchingLabels)).To(Succeed(), "failed to create machine pool")

		By("patching the machine pool status to contain a machine class")
		machinePoolNoMatchingLabelsBase := machinePoolNoMatchingLabels.DeepCopy()
		machinePoolNoMatchingLabels.Status.AvailableMachineClasses = []corev1.LocalObjectReference{{Name: machineClass.Name}}
		machinePoolNoMatchingLabels.Status.Allocatable = corev1alpha1.ResourceList{
			corev1alpha1.ResourceCPU:    resource.MustParse("10"),
			corev1alpha1.ResourceMemory: resource.MustParse("10Gi"),
		}
		Expect(k8sClient.Status().Patch(ctx, machinePoolNoMatchingLabels, client.MergeFrom(machinePoolNoMatchingLabelsBase))).
			To(Succeed(), "failed to patch machine pool status")

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
		machinePoolMatchingLabelsBase := machinePoolMatchingLabels.DeepCopy()
		machinePoolMatchingLabels.Status.AvailableMachineClasses = []corev1.LocalObjectReference{{Name: machineClass.Name}}
		machinePoolMatchingLabels.Status.Allocatable = corev1alpha1.ResourceList{
			corev1alpha1.ResourceCPU:    resource.MustParse("10"),
			corev1alpha1.ResourceMemory: resource.MustParse("10Gi"),
		}
		Expect(k8sClient.Status().Patch(ctx, machinePoolMatchingLabels, client.MergeFrom(machinePoolMatchingLabelsBase))).
			To(Succeed(), "failed to patch machine pool status")

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
		machineKey := client.ObjectKeyFromObject(machine)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, machineKey, machine)).To(Succeed(), "failed to get machine")
			g.Expect(machine.Spec.MachinePoolRef).To(Equal(&corev1.LocalObjectReference{Name: machinePoolMatchingLabels.Name}))
			g.Expect(machine.Status.State).To(Equal(computev1alpha1.MachineStatePending))
		}).Should(Succeed())
	})

	It("should schedule a machine with corresponding tolerations onto a machine pool with taints", func() {
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
		machinePoolBase := taintedMachinePool.DeepCopy()
		taintedMachinePool.Status.AvailableMachineClasses = []corev1.LocalObjectReference{{Name: machineClass.Name}}
		taintedMachinePool.Status.Allocatable = corev1alpha1.ResourceList{
			corev1alpha1.ResourceCPU:    resource.MustParse("10"),
			corev1alpha1.ResourceMemory: resource.MustParse("10Gi"),
		}
		Expect(k8sClient.Status().Patch(ctx, taintedMachinePool, client.MergeFrom(machinePoolBase))).
			To(Succeed(), "failed to patch the machine pool status")

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
		machineKey := client.ObjectKeyFromObject(machine)
		Consistently(func() *corev1.LocalObjectReference {
			Expect(k8sClient.Get(ctx, machineKey, machine)).To(Succeed())
			return machine.Spec.MachinePoolRef
		}).Should(BeNil())

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
		Consistently(func() *corev1.LocalObjectReference {
			Expect(k8sClient.Get(ctx, machineKey, machine)).To(Succeed())
			return machine.Spec.MachinePoolRef
		}).Should(BeNil())

		By("patching the machine to contain all of the corresponding tolerations")
		machineBase = machine.DeepCopy()
		machine.Spec.Tolerations = append(machine.Spec.Tolerations, commonv1alpha1.Toleration{
			Key:      "key1",
			Effect:   commonv1alpha1.TaintEffectNoSchedule,
			Operator: commonv1alpha1.TolerationOpExists,
		})
		Expect(k8sClient.Patch(ctx, machine, client.MergeFrom(machineBase))).To(Succeed(), "failed to patch the machine's spec")

		By("observing the machine is scheduled onto the machine pool")
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, machineKey, machine)).To(Succeed(), "failed to get the machine")
			g.Expect(machine.Spec.MachinePoolRef).To(Equal(&corev1.LocalObjectReference{Name: taintedMachinePool.Name}))
		}).Should(Succeed())
	})

	It("should schedule a shared machine on shared pool and a static machine on a static pool", func() {
		By("creating a machine pool")
		machinePool := &computev1alpha1.MachinePool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-pool-",
			},
		}
		Expect(k8sClient.Create(ctx, machinePool)).To(Succeed(), "failed to create machine pool")

		By("patching the machine pool status to contain a machine class")
		machinePoolBase := machinePool.DeepCopy()
		machinePool.Status.AvailableMachineClasses = []corev1.LocalObjectReference{{Name: machineClass.Name}}
		machinePool.Status.Allocatable = corev1alpha1.ResourceList{
			corev1alpha1.ResourceCPU:    resource.MustParse("10"),
			corev1alpha1.ResourceMemory: resource.MustParse("10Gi"),
		}
		Expect(k8sClient.Status().Patch(ctx, machinePool, client.MergeFrom(machinePoolBase))).
			To(Succeed(), "failed to patch machine pool status")

		By("creating a shared machine pool")
		sharedMachinePool := &computev1alpha1.MachinePool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "shared-test-pool-",
			},
		}
		Expect(k8sClient.Create(ctx, sharedMachinePool)).To(Succeed(), "failed to create the shared machine pool")

		By("creating a shared machine class")
		sharedMachineClass := &computev1alpha1.MachineClass{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "shared-machine-class-",
			},
			Capabilities: corev1alpha1.ResourceList{
				corev1alpha1.ResourceCPU:    resource.MustParse("1"),
				corev1alpha1.ResourceMemory: resource.MustParse("1Gi"),
			},
			Mode: computev1alpha1.ModeShared,
		}
		Expect(k8sClient.Create(ctx, sharedMachineClass)).To(Succeed(), "failed to create test shared machine class")

		By("patching the shared machine pool status to contain a shared machine class")
		sharedMachinePoolBase := sharedMachinePool.DeepCopy()
		sharedMachinePool.Status.AvailableMachineClasses = []corev1.LocalObjectReference{{Name: sharedMachineClass.Name}}
		sharedMachinePool.Status.Allocatable = corev1alpha1.ResourceList{
			corev1alpha1.ResourceSharedCPU:    resource.MustParse("10"),
			corev1alpha1.ResourceSharedMemory: resource.MustParse("10Gi"),
		}
		Expect(k8sClient.Status().Patch(ctx, sharedMachinePool, client.MergeFrom(sharedMachinePoolBase))).
			To(Succeed(), "failed to patch the shared machine pool status")

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

		By("creating a shared machine")
		sharedMachine := &computev1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-shared-machine-",
			},
			Spec: computev1alpha1.MachineSpec{
				Image: "my-image",
				MachineClassRef: corev1.LocalObjectReference{
					Name: sharedMachineClass.Name,
				},
			},
		}
		Expect(k8sClient.Create(ctx, sharedMachine)).To(Succeed(), "failed to create the shared machine")

		By("checking that the shared machine is scheduled onto the shared machine pool")
		Eventually(Object(sharedMachine)).Should(SatisfyAll(
			HaveField("Spec.MachinePoolRef.Name", Equal(sharedMachinePool.Name)),
		))
	})

	It("should schedule machines on a mixed pool", func() {
		By("creating a shared machine class")
		sharedMachineClass := &computev1alpha1.MachineClass{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "shared-machine-class-",
			},
			Capabilities: corev1alpha1.ResourceList{
				corev1alpha1.ResourceCPU:    resource.MustParse("1"),
				corev1alpha1.ResourceMemory: resource.MustParse("1Gi"),
			},
			Mode: computev1alpha1.ModeShared,
		}
		Expect(k8sClient.Create(ctx, sharedMachineClass)).To(Succeed(), "failed to create test shared machine class")

		By("creating a machine pool")
		machinePool := &computev1alpha1.MachinePool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-pool-",
			},
		}
		Expect(k8sClient.Create(ctx, machinePool)).To(Succeed(), "failed to create machine pool")

		By("patching the machine pool status to contain a machine class")
		machinePoolBase := machinePool.DeepCopy()
		machinePool.Status.AvailableMachineClasses = []corev1.LocalObjectReference{{Name: machineClass.Name}, {Name: sharedMachineClass.Name}}
		machinePool.Status.Allocatable = corev1alpha1.ResourceList{
			corev1alpha1.ResourceCPU:          resource.MustParse("10"),
			corev1alpha1.ResourceMemory:       resource.MustParse("10Gi"),
			corev1alpha1.ResourceSharedCPU:    resource.MustParse("10"),
			corev1alpha1.ResourceSharedMemory: resource.MustParse("10Gi"),
		}
		Expect(k8sClient.Status().Patch(ctx, machinePool, client.MergeFrom(machinePoolBase))).
			To(Succeed(), "failed to patch machine pool status")

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

		By("creating a shared machine")
		sharedMachine := &computev1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-shared-machine-",
			},
			Spec: computev1alpha1.MachineSpec{
				Image: "my-image",
				MachineClassRef: corev1.LocalObjectReference{
					Name: sharedMachineClass.Name,
				},
			},
		}
		Expect(k8sClient.Create(ctx, sharedMachine)).To(Succeed(), "failed to create the shared machine")

		By("checking that the shared machine is scheduled onto the machine pool")
		Eventually(Object(sharedMachine)).Should(SatisfyAll(
			HaveField("Spec.MachinePoolRef.Name", Equal(machinePool.Name)),
		))
	})

	It("should schedule a machines once the capacity is sufficient", func() {
		By("creating a machine pool")
		machinePool := &computev1alpha1.MachinePool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-pool-",
			},
		}
		Expect(k8sClient.Create(ctx, machinePool)).To(Succeed(), "failed to create machine pool")

		By("patching the machine pool status to contain a machine class")
		machinePoolBase := machinePool.DeepCopy()
		machinePool.Status.AvailableMachineClasses = []corev1.LocalObjectReference{{Name: machineClass.Name}}
		Expect(k8sClient.Status().Patch(ctx, machinePool, client.MergeFrom(machinePoolBase))).
			To(Succeed(), "failed to patch machine pool status")

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
		Consistently(Object(machine)).Should(SatisfyAll(
			HaveField("Spec.MachinePoolRef", BeNil()),
		))

		By("patching the machine pool status to contain a machine class")
		machinePoolBase = machinePool.DeepCopy()
		machinePool.Status.Allocatable = corev1alpha1.ResourceList{
			corev1alpha1.ResourceCPU:    resource.MustParse("10"),
			corev1alpha1.ResourceMemory: resource.MustParse("10Gi"),
		}
		Expect(k8sClient.Status().Patch(ctx, machinePool, client.MergeFrom(machinePoolBase))).
			To(Succeed(), "failed to patch machine pool status")

		By("checking that the machine is scheduled onto the machine pool")
		Eventually(Object(machine)).Should(SatisfyAll(
			HaveField("Spec.MachinePoolRef.Name", Equal(machinePool.Name)),
		))
	})
})
