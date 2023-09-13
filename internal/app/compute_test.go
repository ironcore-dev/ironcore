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

package app_test

import (
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	corev1alpha1 "github.com/onmetal/onmetal-api/api/core/v1alpha1"
	. "github.com/onmetal/onmetal-api/utils/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

var _ = Describe("Compute", func() {
	var (
		ctx          = SetupContext()
		ns           = SetupTest(ctx)
		machineClass = &computev1alpha1.MachineClass{}
	)

	const (
		fieldOwner = client.FieldOwner("fieldowner.test.api.onmetal.de/onmetal-apiserver")
	)

	BeforeEach(func() {
		*machineClass = computev1alpha1.MachineClass{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "machine-class-",
			},
			Capabilities: corev1alpha1.ResourceList{
				corev1alpha1.ResourceCPU:    resource.MustParse("1"),
				corev1alpha1.ResourceMemory: resource.MustParse("1Gi"),
			},
		}
		Expect(k8sClient.Create(ctx, machineClass)).To(Succeed(), "failed to create test machine class")
		DeferCleanup(k8sClient.Delete, ctx, machineClass)
	})

	Context("Machine", func() {
		It("should correctly apply machines with volumes and default devices", func() {
			By("applying a machine with volumes")
			machine := &computev1alpha1.Machine{
				TypeMeta: metav1.TypeMeta{
					APIVersion: computev1alpha1.SchemeGroupVersion.String(),
					Kind:       "Machine",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: ns.Name,
					Name:      "my-machine",
				},
				Spec: computev1alpha1.MachineSpec{
					MachineClassRef: corev1.LocalObjectReference{Name: machineClass.Name},
					Volumes: []computev1alpha1.Volume{
						{
							Name: "foo",
							VolumeSource: computev1alpha1.VolumeSource{
								EmptyDisk: &computev1alpha1.EmptyDiskVolumeSource{},
							},
						},
					},
				},
			}
			Expect(k8sClient.Patch(ctx, machine, client.Apply, fieldOwner)).To(Succeed())

			By("inspecting the machine's volumes")
			Expect(machine.Spec.Volumes).To(Equal([]computev1alpha1.Volume{
				{
					Name:   "foo",
					Device: pointer.String("oda"),
					VolumeSource: computev1alpha1.VolumeSource{
						EmptyDisk: &computev1alpha1.EmptyDiskVolumeSource{},
					},
				},
			}))

			By("applying a changed machine with a second volume")
			machine = &computev1alpha1.Machine{
				TypeMeta: metav1.TypeMeta{
					APIVersion: computev1alpha1.SchemeGroupVersion.String(),
					Kind:       "Machine",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: ns.Name,
					Name:      "my-machine",
				},
				Spec: computev1alpha1.MachineSpec{
					MachineClassRef: corev1.LocalObjectReference{Name: machineClass.Name},
					Volumes: []computev1alpha1.Volume{
						{
							Name: "foo",
							VolumeSource: computev1alpha1.VolumeSource{
								EmptyDisk: &computev1alpha1.EmptyDiskVolumeSource{},
							},
						},
						{
							Name: "bar",
							VolumeSource: computev1alpha1.VolumeSource{
								EmptyDisk: &computev1alpha1.EmptyDiskVolumeSource{},
							},
						},
					},
				},
			}
			Expect(k8sClient.Patch(ctx, machine, client.Apply, fieldOwner)).To(Succeed())

			By("inspecting the machine's volumes")
			Expect(machine.Spec.Volumes).To(Equal([]computev1alpha1.Volume{
				{
					Name:   "foo",
					Device: pointer.String("oda"),
					VolumeSource: computev1alpha1.VolumeSource{
						EmptyDisk: &computev1alpha1.EmptyDiskVolumeSource{},
					},
				},
				{
					Name:   "bar",
					Device: pointer.String("odb"),
					VolumeSource: computev1alpha1.VolumeSource{
						EmptyDisk: &computev1alpha1.EmptyDiskVolumeSource{},
					},
				},
			}))

			By("applying a changed machine with the first volume removed")
			machine = &computev1alpha1.Machine{
				TypeMeta: metav1.TypeMeta{
					APIVersion: computev1alpha1.SchemeGroupVersion.String(),
					Kind:       "Machine",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: ns.Name,
					Name:      "my-machine",
				},
				Spec: computev1alpha1.MachineSpec{
					MachineClassRef: corev1.LocalObjectReference{Name: machineClass.Name},
					Volumes: []computev1alpha1.Volume{
						{
							Name: "bar",
							VolumeSource: computev1alpha1.VolumeSource{
								EmptyDisk: &computev1alpha1.EmptyDiskVolumeSource{},
							},
						},
					},
				},
			}
			Expect(k8sClient.Patch(ctx, machine, client.Apply, fieldOwner)).To(Succeed())

			By("inspecting the machine's volumes")
			Expect(machine.Spec.Volumes).To(Equal([]computev1alpha1.Volume{
				{
					Name:   "bar",
					Device: pointer.String("odb"),
					VolumeSource: computev1alpha1.VolumeSource{
						EmptyDisk: &computev1alpha1.EmptyDiskVolumeSource{},
					},
				},
			}))
		})

		It("should allow listing machines filtering by machine pool name", func() {
			const (
				machinePool1 = "machine-pool-1"
				machinePool2 = "machine-pool-2"
			)

			By("creating a machine on machine pool 1")
			machine1 := &computev1alpha1.Machine{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    ns.Name,
					GenerateName: "machine-",
				},
				Spec: computev1alpha1.MachineSpec{
					MachineClassRef: corev1.LocalObjectReference{Name: machineClass.Name},
					MachinePoolRef:  &corev1.LocalObjectReference{Name: machinePool1},
				},
			}
			Expect(k8sClient.Create(ctx, machine1)).To(Succeed())

			By("creating a machine on machine pool 2")
			machine2 := &computev1alpha1.Machine{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    ns.Name,
					GenerateName: "machine-",
				},
				Spec: computev1alpha1.MachineSpec{
					MachineClassRef: corev1.LocalObjectReference{Name: machineClass.Name},
					MachinePoolRef:  &corev1.LocalObjectReference{Name: machinePool2},
				},
			}
			Expect(k8sClient.Create(ctx, machine2)).To(Succeed())

			By("creating a machine on no machine pool")
			machine3 := &computev1alpha1.Machine{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    ns.Name,
					GenerateName: "machine-",
				},
				Spec: computev1alpha1.MachineSpec{
					MachineClassRef: corev1.LocalObjectReference{Name: machineClass.Name},
				},
			}
			Expect(k8sClient.Create(ctx, machine3)).To(Succeed())

			By("listing all machines on machine pool 1")
			machinesOnMachinePool1List := &computev1alpha1.MachineList{}
			Expect(k8sClient.List(ctx, machinesOnMachinePool1List,
				client.InNamespace(ns.Name),
				client.MatchingFields{computev1alpha1.MachineMachinePoolRefNameField: machinePool1},
			))

			By("inspecting the items")
			Expect(machinesOnMachinePool1List.Items).To(ConsistOf(*machine1))

			By("listing all machines on machine pool 2")
			machinesOnMachinePool2List := &computev1alpha1.MachineList{}
			Expect(k8sClient.List(ctx, machinesOnMachinePool2List,
				client.InNamespace(ns.Name),
				client.MatchingFields{computev1alpha1.MachineMachinePoolRefNameField: machinePool2},
			))

			By("inspecting the items")
			Expect(machinesOnMachinePool2List.Items).To(ConsistOf(*machine2))

			By("listing all machines on no machine pool")
			machinesOnNoMachinePoolList := &computev1alpha1.MachineList{}
			Expect(k8sClient.List(ctx, machinesOnNoMachinePoolList,
				client.InNamespace(ns.Name),
				client.MatchingFields{computev1alpha1.MachineMachinePoolRefNameField: ""},
			))

			By("inspecting the items")
			Expect(machinesOnNoMachinePoolList.Items).To(ConsistOf(*machine3))
		})

		It("should allow listing machines by machine class name", func() {
			By("creating another machine class")
			machineClass2 := &computev1alpha1.MachineClass{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "machine-class-",
				},
				Capabilities: corev1alpha1.ResourceList{
					corev1alpha1.ResourceCPU:    resource.MustParse("3"),
					corev1alpha1.ResourceMemory: resource.MustParse("10Gi"),
				},
			}
			Expect(k8sClient.Create(ctx, machineClass2)).To(Succeed())
			DeferCleanup(k8sClient.Delete, ctx, machineClass2)

			By("creating a machine")
			machine1 := &computev1alpha1.Machine{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    ns.Name,
					GenerateName: "machine-",
				},
				Spec: computev1alpha1.MachineSpec{
					MachineClassRef: corev1.LocalObjectReference{Name: machineClass.Name},
				},
			}
			Expect(k8sClient.Create(ctx, machine1)).To(Succeed())

			By("creating a machine with the other machine class")
			machine2 := &computev1alpha1.Machine{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    ns.Name,
					GenerateName: "machine-",
				},
				Spec: computev1alpha1.MachineSpec{
					MachineClassRef: corev1.LocalObjectReference{Name: machineClass2.Name},
				},
			}
			Expect(k8sClient.Create(ctx, machine2)).To(Succeed())

			By("listing machines with the first machine class name")
			machineList := &computev1alpha1.MachineList{}
			Expect(k8sClient.List(ctx, machineList, client.MatchingFields{
				computev1alpha1.MachineMachineClassRefNameField: machineClass.Name,
			})).To(Succeed())

			By("inspecting the retrieved list to only have the machine with the correct machine class")
			Expect(machineList.Items).To(ConsistOf(HaveField("UID", machine1.UID)))

			By("listing machines with the second machine class name")
			Expect(k8sClient.List(ctx, machineList, client.MatchingFields{
				computev1alpha1.MachineMachineClassRefNameField: machineClass2.Name,
			})).To(Succeed())

			By("inspecting the retrieved list to only have the machine with the correct machine class")
			Expect(machineList.Items).To(ConsistOf(HaveField("UID", machine2.UID)))
		})
	})

	Context("MachinePool resources", func() {
		It("should be masked", func() {
			By("creating a new machine pool")
			machinePool := &computev1alpha1.MachinePool{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "machine-pool-",
				},
				Spec: computev1alpha1.MachinePoolSpec{
					ProviderID: "test",
				},
			}
			Expect(k8sClient.Create(ctx, machinePool)).To(Succeed())

			By("patching the status")
			Eventually(UpdateStatus(machinePool, func() {
				machinePool.Status.Capacity = corev1alpha1.ResourceList{
					corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeMachineClass, "test-class"): resource.MustParse("10"),
				}
				machinePool.Status.Allocatable = corev1alpha1.ResourceList{
					corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeMachineClass, "test-class"): resource.MustParse("5"),
				}
			})).Should(Succeed())

			By("checking that the resources are hidden by using GET")
			Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(machinePool), machinePool)).To(Succeed())
			Expect(machinePool.Status.Allocatable).To(BeNil())
			Expect(machinePool.Status.Capacity).To(BeNil())

			By("checking that the resources are hidden by using LIST")
			machinePools := &computev1alpha1.MachinePoolList{}
			Expect(k8sClient.List(ctx, machinePools)).To(Succeed())
			Expect(machinePools.Items).To(ContainElement(SatisfyAll(
				HaveField("Status.Allocatable", BeNil()),
				HaveField("Status.Capacity", BeNil()),
			)))

			By("checking that the resources are shown using elevated user by using GET")
			Expect(elevatedK8sClient.Get(ctx, client.ObjectKeyFromObject(machinePool), machinePool)).To(Succeed())
			Expect(machinePool.Status.Capacity).To(Equal(corev1alpha1.ResourceList{
				corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeMachineClass, "test-class"): resource.MustParse("10"),
			}))
			Expect(machinePool.Status.Allocatable).To(Equal(corev1alpha1.ResourceList{
				corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeMachineClass, "test-class"): resource.MustParse("5"),
			}))

			By("checking that the resources are shown using elevated user by using LIST")
			Expect(elevatedK8sClient.List(ctx, machinePools)).To(Succeed())
			Expect(machinePools.Items).To(ContainElement(SatisfyAll(
				HaveField("Status.Capacity", Equal(corev1alpha1.ResourceList{
					corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeMachineClass, "test-class"): resource.MustParse("10"),
				})),
				HaveField("Status.Allocatable", Equal(corev1alpha1.ResourceList{
					corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeMachineClass, "test-class"): resource.MustParse("5"),
				})),
			)))
		})
	})
})
