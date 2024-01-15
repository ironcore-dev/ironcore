// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package app_test

import (
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	. "github.com/ironcore-dev/ironcore/utils/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Compute", func() {
	var (
		ctx          = SetupContext()
		ns           = SetupTest(ctx)
		machineClass = &computev1alpha1.MachineClass{}
	)

	const (
		fieldOwner = client.FieldOwner("fieldowner.test.ironcore.dev/ironcore-apiserver")
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
					Device: ptr.To[string]("oda"),
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
					Device: ptr.To[string]("oda"),
					VolumeSource: computev1alpha1.VolumeSource{
						EmptyDisk: &computev1alpha1.EmptyDiskVolumeSource{},
					},
				},
				{
					Name:   "bar",
					Device: ptr.To[string]("odb"),
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
					Device: ptr.To[string]("odb"),
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
			)).To(Succeed())

			By("inspecting the items")
			Expect(machinesOnMachinePool1List.Items).To(ConsistOf(*machine1))

			By("listing all machines on machine pool 2")
			machinesOnMachinePool2List := &computev1alpha1.MachineList{}
			Expect(k8sClient.List(ctx, machinesOnMachinePool2List,
				client.InNamespace(ns.Name),
				client.MatchingFields{computev1alpha1.MachineMachinePoolRefNameField: machinePool2},
			)).To(Succeed())

			By("inspecting the items")
			Expect(machinesOnMachinePool2List.Items).To(ConsistOf(*machine2))

			By("listing all machines on no machine pool")
			machinesOnNoMachinePoolList := &computev1alpha1.MachineList{}
			Expect(k8sClient.List(ctx, machinesOnNoMachinePoolList,
				client.InNamespace(ns.Name),
				client.MatchingFields{computev1alpha1.MachineMachinePoolRefNameField: ""},
			)).To(Succeed())

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
})
