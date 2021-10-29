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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
)

// var ctx = ctrl.SetupSignalHandler()

var _ = Describe("MachineScheduler", func() {
	// ns := SetupTest(ctx)

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
		machinePool.Status.AvailableMachineClasses = []corev1.LocalObjectReference{{Name: "my-machineclass"}}
		Expect(k8sClient.Status().Patch(ctx, machinePool, client.MergeFrom(machinePoolBase))).
			To(Succeed(), "failed to patch machine pool status")

		By("creating a machine w/ the requested machine class")
		machine := &computev1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-machine-",
			},
			Spec: computev1alpha1.MachineSpec{
				MachineClass: corev1.LocalObjectReference{
					Name: "my-machineclass",
				},
			},
		}
		Expect(k8sClient.Create(ctx, machine)).To(Succeed(), "failed to create machine")

		By("waiting for the machine to be scheduled onto the machine pool")
		machineKey := client.ObjectKeyFromObject(machine)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, machineKey, machine)).To(Succeed(), "failed to get machine")
			g.Expect(machine.Spec.MachinePool.Name).To(Equal(machinePool.Name))
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
				MachineClass: corev1.LocalObjectReference{
					Name: "my-machineclass",
				},
			},
		}
		Expect(k8sClient.Create(ctx, machine)).To(Succeed(), "failed to create machine")

		By("waiting for the machine to indicate it is pending")
		machineKey := client.ObjectKeyFromObject(machine)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, machineKey, machine)).To(Succeed())
			g.Expect(machine.Spec.MachinePool.Name).To(BeEmpty())
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
		machinePool.Status.AvailableMachineClasses = []corev1.LocalObjectReference{{Name: "my-machineclass"}}
		Expect(k8sClient.Status().Patch(ctx, machinePool, client.MergeFrom(machinePoolBase))).
			To(Succeed(), "failed to patch machine pool status")

		By("waiting for the machine to be scheduled onto the machine pool")
		Eventually(func() string {
			Expect(k8sClient.Get(ctx, machineKey, machine)).To(Succeed(), "failed to get machine")
			return machine.Spec.MachinePool.Name
		}).Should(Equal(machinePool.Name))
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
		machinePoolNoMatchingLabels.Status.AvailableMachineClasses = []corev1.LocalObjectReference{{Name: "my-machineclass"}}
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
		machinePoolMatchingLabels.Status.AvailableMachineClasses = []corev1.LocalObjectReference{{Name: "my-machineclass"}}
		Expect(k8sClient.Status().Patch(ctx, machinePoolMatchingLabels, client.MergeFrom(machinePoolMatchingLabelsBase))).
			To(Succeed(), "failed to patch machine pool status")

		By("creating a machine w/ the requested machine class")
		machine := &computev1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-machine-",
			},
			Spec: computev1alpha1.MachineSpec{
				MachinePoolSelector: map[string]string{
					"foo": "bar",
				},
				MachineClass: corev1.LocalObjectReference{
					Name: "my-machineclass",
				},
			},
		}
		Expect(k8sClient.Create(ctx, machine)).To(Succeed(), "failed to create machine")

		By("waiting for the machine to be scheduled onto the machine pool")
		machineKey := client.ObjectKeyFromObject(machine)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, machineKey, machine)).To(Succeed(), "failed to get machine")
			g.Expect(machine.Spec.MachinePool.Name).To(Equal(machinePoolMatchingLabels.Name))
			g.Expect(machine.Status.State).To(Equal(computev1alpha1.MachineStatePending))
		}).Should(Succeed())
	})
})
