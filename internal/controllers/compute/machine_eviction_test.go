// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package compute

import (
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	. "github.com/ironcore-dev/ironcore/utils/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"

	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
)

var _ = Describe("MachineEviction", func() {
	ns := SetupNamespace(&k8sClient)
	machineClass := SetupMachineClass()

	It("should evict a machine that does not tolerate a NoExecute taint", func(ctx SpecContext) {
		By("creating a machine pool")
		machinePool := &computev1alpha1.MachinePool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-pool-",
			},
		}
		Expect(k8sClient.Create(ctx, machinePool)).To(Succeed(), "failed to create machine pool")

		By("marking the machine pool as available")
		Eventually(UpdateStatus(machinePool, func() {
			machinePool.Status.AvailableMachineClasses = []corev1.LocalObjectReference{{Name: machineClass.Name}}
			machinePool.Status.Allocatable = corev1alpha1.ResourceList{
				corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeMachineClass, machineClass.Name): resource.MustParse("10"),
			}
		})).Should(Succeed())

		By("creating a machine bound to the pool without any tolerations")
		machine := &computev1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-machine-",
			},
			Spec: computev1alpha1.MachineSpec{
				MachineClassRef: corev1.LocalObjectReference{Name: machineClass.Name},
				MachinePoolRef:  &corev1.LocalObjectReference{Name: machinePool.Name},
			},
		}
		Expect(k8sClient.Create(ctx, machine)).To(Succeed(), "failed to create machine")

		By("waiting for the machine to be bound and pending")
		Eventually(Object(machine)).Should(SatisfyAll(
			HaveField("Spec.MachinePoolRef", Equal(&corev1.LocalObjectReference{Name: machinePool.Name})),
			HaveField("Status.State", Equal(computev1alpha1.MachineStatePending)),
		))

		By("adding a NoExecute taint to the machine pool")
		Eventually(Update(machinePool, func() {
			machinePool.Spec.Taints = []commonv1alpha1.Taint{
				{
					Key:    "maintenance",
					Value:  "some-update",
					Effect: commonv1alpha1.TaintEffectNoExecute,
				},
			}
		})).Should(Succeed())

		By("asserting the machine is evicted")
		Eventually(Get(machine)).Should(Satisfy(apierrors.IsNotFound))
	})

	It("should evict a machine whose toleration does not match the NoExecute taint but retain one that tolerates it", func(ctx SpecContext) {
		By("creating a machine pool")
		machinePool := &computev1alpha1.MachinePool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-pool-",
			},
		}
		Expect(k8sClient.Create(ctx, machinePool)).To(Succeed(), "failed to create machine pool")

		By("marking the machine pool as available")
		Eventually(UpdateStatus(machinePool, func() {
			machinePool.Status.AvailableMachineClasses = []corev1.LocalObjectReference{{Name: machineClass.Name}}
			machinePool.Status.Allocatable = corev1alpha1.ResourceList{
				corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeMachineClass, machineClass.Name): resource.MustParse("10"),
			}
		})).Should(Succeed())

		By("creating a machine that only tolerates the NoSchedule effect")
		machine := &computev1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-machine-",
			},
			Spec: computev1alpha1.MachineSpec{
				MachineClassRef: corev1.LocalObjectReference{Name: machineClass.Name},
				MachinePoolRef:  &corev1.LocalObjectReference{Name: machinePool.Name},
				Tolerations: []commonv1alpha1.Toleration{
					{
						Key:    "maintenance",
						Effect: commonv1alpha1.TaintEffectNoSchedule,
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, machine)).To(Succeed(), "failed to create machine")

		By("creating a machine that tolerates the NoExecute taint via an Exists toleration")
		machineWithTolerations := &computev1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-machine-",
			},
			Spec: computev1alpha1.MachineSpec{
				MachineClassRef: corev1.LocalObjectReference{Name: machineClass.Name},
				MachinePoolRef:  &corev1.LocalObjectReference{Name: machinePool.Name},
				Tolerations: []commonv1alpha1.Toleration{
					{
						Key:      "maintenance",
						Operator: commonv1alpha1.TolerationOpExists,
						Effect:   commonv1alpha1.TaintEffectNoExecute,
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, machineWithTolerations)).To(Succeed(), "failed to create machine")

		By("waiting for the non-tolerating machine to be bound and pending")
		Eventually(Object(machine)).Should(SatisfyAll(
			HaveField("Spec.MachinePoolRef", Equal(&corev1.LocalObjectReference{Name: machinePool.Name})),
			HaveField("Status.State", Equal(computev1alpha1.MachineStatePending)),
		))

		By("adding a NoExecute taint to the machine pool")
		Eventually(Update(machinePool, func() {
			machinePool.Spec.Taints = []commonv1alpha1.Taint{
				{
					Key:    "maintenance",
					Value:  "some-other-update",
					Effect: commonv1alpha1.TaintEffectNoExecute,
				},
			}
		})).Should(Succeed())

		By("asserting the non-tolerating machine is evicted")
		Eventually(Get(machine)).Should(Satisfy(apierrors.IsNotFound))

		By("asserting the tolerating machine is retained")
		Consistently(Object(machineWithTolerations)).Should(
			HaveField("ObjectMeta.DeletionTimestamp", BeNil()),
		)
	})
})
