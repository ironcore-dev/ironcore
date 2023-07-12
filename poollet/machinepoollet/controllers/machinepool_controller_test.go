// Copyright 2022 OnMetal authors
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

package controllers_test

import (
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	corev1alpha1 "github.com/onmetal/onmetal-api/api/core/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"github.com/onmetal/onmetal-api/ori/testing/machine"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

const (
	TestMachinePool = "test-machinepool"
)

var _ = Describe("MachinePoolController", func() {
	_, _, _, srv := SetupTest()

	It("should add machine classes to pool", func(ctx SpecContext) {
		srv.SetMachineClasses([]*machine.FakeMachineClass{})

		By("creating a machine class")
		mc := &computev1alpha1.MachineClass{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-mc-1-",
			},
			Capabilities: corev1alpha1.ResourceList{
				corev1alpha1.ResourceCPU:    resource.MustParse("2"),
				corev1alpha1.ResourceMemory: resource.MustParse("2Gi"),
			},
		}
		Expect(k8sClient.Create(ctx, mc)).To(Succeed(), "failed to create test machine class")

		srv.SetMachineClasses([]*machine.FakeMachineClass{
			{
				MachineClass: ori.MachineClass{
					Name: mc.Name,
					Capabilities: &ori.MachineClassCapabilities{
						CpuMillis:   mc.Capabilities.CPU().MilliValue(),
						MemoryBytes: mc.Capabilities.Memory().AsDec().UnscaledBig().Uint64(),
					},
				},
			},
		})

		By("creating a machine pool")
		machinePool := &computev1alpha1.MachinePool{
			ObjectMeta: metav1.ObjectMeta{
				Name: TestMachinePool,
			},
		}
		Expect(k8sClient.Create(ctx, machinePool)).To(Succeed(), "failed to create test machine pool")

		By("checking if the default machine class is present")
		Eventually(Object(machinePool)).Should(SatisfyAll(
			HaveField("Status.AvailableMachineClasses", Equal([]corev1.LocalObjectReference{
				{
					Name: mc.Name,
				},
			}))),
		)

		By("creating a second machine class")
		mc2 := &computev1alpha1.MachineClass{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-mc-2-",
			},
			Capabilities: corev1alpha1.ResourceList{
				corev1alpha1.ResourceCPU:    resource.MustParse("3"),
				corev1alpha1.ResourceMemory: resource.MustParse("4Gi"),
			},
		}
		Expect(k8sClient.Create(ctx, mc2)).To(Succeed(), "failed to create test machine class")

		Eventually(Object(machinePool)).Should(SatisfyAll(
			HaveField("Status.AvailableMachineClasses", HaveLen(1))),
		)

		srv.SetMachineClasses([]*machine.FakeMachineClass{
			{
				MachineClass: ori.MachineClass{
					Name: mc.Name,
					Capabilities: &ori.MachineClassCapabilities{
						CpuMillis:   mc.Capabilities.CPU().MilliValue(),
						MemoryBytes: mc.Capabilities.Memory().AsDec().UnscaledBig().Uint64(),
					},
				},
			},
			{
				MachineClass: ori.MachineClass{
					Name: mc2.Name,
					Capabilities: &ori.MachineClassCapabilities{
						CpuMillis:   mc2.Capabilities.CPU().MilliValue(),
						MemoryBytes: mc2.Capabilities.Memory().AsDec().UnscaledBig().Uint64(),
					},
				},
			},
		})

		By("checking if the second machine class is present")
		Eventually(Object(machinePool)).Should(SatisfyAll(
			HaveField("Status.AvailableMachineClasses", ConsistOf(
				corev1.LocalObjectReference{
					Name: mc.Name,
				},
				corev1.LocalObjectReference{
					Name: mc2.Name,
				},
			))),
		)
	})
})
