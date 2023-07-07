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
	"time"

	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	corev1alpha1 "github.com/onmetal/onmetal-api/api/core/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	testingmachine "github.com/onmetal/onmetal-api/ori/testing/machine"
	"github.com/onmetal/onmetal-api/utils/quota"
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
	ns, _, mc, srv := SetupTest()

	It("should calculate pool capacity", func(ctx SpecContext) {
		var (
			sharedCpu    = resource.MustParse("10")
			staticCpu    = resource.MustParse("20")
			sharedMemory = resource.MustParse("10Gi")
			staticMemory = resource.MustParse("20Gi")
		)

		By("setting pool info")
		srv.SetPoolInfo(testingmachine.FakePoolInfo{
			SharedCpu:    sharedCpu.AsDec().UnscaledBig().Int64(),
			StaticCpu:    staticCpu.AsDec().UnscaledBig().Int64(),
			SharedMemory: sharedMemory.AsDec().UnscaledBig().Uint64(),
			StaticMemory: staticMemory.AsDec().UnscaledBig().Uint64(),
		})

		By("creating a machine pool")
		machinePool := &computev1alpha1.MachinePool{
			ObjectMeta: metav1.ObjectMeta{
				Name: TestMachinePool,
			},
		}
		Expect(k8sClient.Create(ctx, machinePool)).To(Succeed(), "failed to create machine pool")

		By("creating a second machine class")
		mc2 := &computev1alpha1.MachineClass{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-mc-",
			},
			Capabilities: corev1alpha1.ResourceList{
				corev1alpha1.ResourceCPU:    resource.MustParse("2"),
				corev1alpha1.ResourceMemory: resource.MustParse("2Gi"),
			},
			Mode: computev1alpha1.ModeShared,
		}
		Expect(k8sClient.Create(ctx, mc2)).To(Succeed(), "failed to create machine class")

		srv.SetMachineClasses([]*testingmachine.FakeMachineClass{
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

		By("checking if the capacity is correct")
		Eventually(Object(machinePool)).Within(time.Second).Should(SatisfyAll(
			HaveField("Status.Capacity", Satisfy(func(capacity corev1alpha1.ResourceList) bool {
				return quota.Equals(capacity, corev1alpha1.ResourceList{
					corev1alpha1.SharedResourceCPU:    sharedCpu,
					corev1alpha1.ResourceCPU:          staticCpu,
					corev1alpha1.SharedResourceMemory: sharedMemory,
					corev1alpha1.ResourceMemory:       staticMemory,
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
					Name: mc2.Name,
				},
				MachinePoolRef: &corev1.LocalObjectReference{
					Name: machinePool.Name,
				},
			},
		}
		Expect(k8sClient.Create(ctx, machine)).To(Succeed(), "failed to create machine")

		By("checking if the allocatable resources are correct")
		Eventually(Object(machinePool)).Within(time.Second).Should(SatisfyAll(
			HaveField("Status.Allocatable", Satisfy(func(allocatable corev1alpha1.ResourceList) bool {
				return quota.Equals(allocatable, corev1alpha1.ResourceList{
					corev1alpha1.SharedResourceCPU:    resource.MustParse("8"),
					corev1alpha1.ResourceCPU:          staticCpu,
					corev1alpha1.SharedResourceMemory: resource.MustParse("8Gi"),
					corev1alpha1.ResourceMemory:       staticMemory,
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
					Name: mc.Name,
				},
				MachinePoolRef: &corev1.LocalObjectReference{
					Name: machinePool.Name,
				},
			},
		}
		Expect(k8sClient.Create(ctx, machine2)).To(Succeed(), "failed to create test machine class")

		By("checking if the allocatable resources are correct")
		Eventually(Object(machinePool)).Within(time.Second).Should(SatisfyAll(
			HaveField("Status.Allocatable", Satisfy(func(allocatable corev1alpha1.ResourceList) bool {
				return quota.Equals(allocatable, corev1alpha1.ResourceList{
					corev1alpha1.SharedResourceCPU:    resource.MustParse("8"),
					corev1alpha1.ResourceCPU:          resource.MustParse("19"),
					corev1alpha1.SharedResourceMemory: resource.MustParse("8Gi"),
					corev1alpha1.ResourceMemory:       resource.MustParse("19Gi"),
				})
			})),
		))
	})
})
