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
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	. "github.com/onmetal/onmetal-api/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("App", func() {
	ctx := SetupContext()
	ns := SetupTest(ctx)

	const fieldOwner = client.FieldOwner("fieldowner.test.api.onmetal.de/onmetal-apiserver")

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
					MachineClassRef: corev1.LocalObjectReference{Name: "my-class"},
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
					MachineClassRef: corev1.LocalObjectReference{Name: "my-class"},
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
					MachineClassRef: corev1.LocalObjectReference{Name: "my-class"},
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
					MachineClassRef: corev1.LocalObjectReference{Name: "my-class"},
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
					MachineClassRef: corev1.LocalObjectReference{Name: "my-class"},
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
					MachineClassRef: corev1.LocalObjectReference{Name: "my-class"},
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

		It("should allow listing volumes filtering by volume pool name", func() {
			const (
				volumePool1 = "volume-pool-1"
				volumePool2 = "volume-pool-2"
			)

			By("creating a volume on volume pool 1")
			volume1 := &storagev1alpha1.Volume{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    ns.Name,
					GenerateName: "volume-",
				},
				Spec: storagev1alpha1.VolumeSpec{
					VolumeClassRef: &corev1.LocalObjectReference{Name: "my-class"},
					VolumePoolRef:  &corev1.LocalObjectReference{Name: volumePool1},
					Resources: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("10Gi"),
					},
				},
			}
			Expect(k8sClient.Create(ctx, volume1)).To(Succeed())

			By("creating a volume on volume pool 2")
			volume2 := &storagev1alpha1.Volume{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    ns.Name,
					GenerateName: "volume-",
				},
				Spec: storagev1alpha1.VolumeSpec{
					VolumeClassRef: &corev1.LocalObjectReference{Name: "my-class"},
					VolumePoolRef:  &corev1.LocalObjectReference{Name: volumePool2},
					Resources: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("10Gi"),
					},
				},
			}
			Expect(k8sClient.Create(ctx, volume2)).To(Succeed())

			By("creating a volume on no volume pool")
			volume3 := &storagev1alpha1.Volume{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    ns.Name,
					GenerateName: "volume-",
				},
				Spec: storagev1alpha1.VolumeSpec{
					VolumeClassRef: &corev1.LocalObjectReference{Name: "my-class"},
					Resources: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("10Gi"),
					},
				},
			}
			Expect(k8sClient.Create(ctx, volume3)).To(Succeed())

			By("listing all volumes on volume pool 1")
			volumesOnVolumePool1List := &storagev1alpha1.VolumeList{}
			Expect(k8sClient.List(ctx, volumesOnVolumePool1List,
				client.InNamespace(ns.Name),
				client.MatchingFields{storagev1alpha1.VolumeVolumePoolRefNameField: volumePool1},
			))

			By("inspecting the items")
			Expect(volumesOnVolumePool1List.Items).To(ConsistOf(*volume1))

			By("listing all volumes on volume pool 2")
			volumesOnVolumePool2List := &storagev1alpha1.VolumeList{}
			Expect(k8sClient.List(ctx, volumesOnVolumePool2List,
				client.InNamespace(ns.Name),
				client.MatchingFields{storagev1alpha1.VolumeVolumePoolRefNameField: volumePool2},
			))

			By("inspecting the items")
			Expect(volumesOnVolumePool2List.Items).To(ConsistOf(*volume2))

			By("listing all volumes on no volume pool")
			volumesOnNoVolumePoolList := &storagev1alpha1.VolumeList{}
			Expect(k8sClient.List(ctx, volumesOnNoVolumePoolList,
				client.InNamespace(ns.Name),
				client.MatchingFields{storagev1alpha1.VolumeVolumePoolRefNameField: ""},
			))

			By("inspecting the items")
			Expect(volumesOnNoVolumePoolList.Items).To(ConsistOf(*volume3))
		})
	})
})
