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
	. "github.com/onmetal/onmetal-api/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
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
	})
})
