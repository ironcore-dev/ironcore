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
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	. "github.com/onmetal/onmetal-api/utils/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

var _ = Describe("MachineEphemeralVolumeController", func() {
	ns := SetupNamespace(&k8sClient)
	machineClass := SetupMachineClass()

	It("should manage ephemeral volumes for a machine", func(ctx SpecContext) {
		By("creating a volume that will be referenced by the machine")
		refVolume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "ref-volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{},
		}
		Expect(k8sClient.Create(ctx, refVolume)).To(Succeed())

		By("creating a machine")
		machine := &computev1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "machine-",
			},
			Spec: computev1alpha1.MachineSpec{
				MachineClassRef: corev1.LocalObjectReference{Name: machineClass.Name},
				Volumes: []computev1alpha1.Volume{
					{
						Name: "ref-volume",
						VolumeSource: computev1alpha1.VolumeSource{
							VolumeRef: &corev1.LocalObjectReference{Name: refVolume.Name},
						},
					},
					{
						Name: "ephem-volume",
						VolumeSource: computev1alpha1.VolumeSource{
							Ephemeral: &computev1alpha1.EphemeralVolumeSource{
								VolumeTemplate: &storagev1alpha1.VolumeTemplateSpec{
									Spec: storagev1alpha1.VolumeSpec{},
								},
							},
						},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, machine)).To(Succeed())

		By("creating an undesired controlled volume")
		undesiredControlledVolume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "undesired-ctrl-volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{},
		}
		_ = ctrl.SetControllerReference(machine, undesiredControlledVolume, k8sClient.Scheme())
		Expect(k8sClient.Create(ctx, undesiredControlledVolume)).To(Succeed())

		By("waiting for the ephemeral volume to exist")
		ephemVolume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns.Name,
				Name:      computev1alpha1.MachineEphemeralVolumeName(machine.Name, "ephem-volume"),
			},
		}
		Eventually(Get(ephemVolume)).Should(Succeed())

		By("asserting the referenced volume still exists")
		Consistently(Get(refVolume)).Should(Succeed())

		By("waiting for the undesired controlled volume to be gone")
		Eventually(Get(undesiredControlledVolume)).Should(Satisfy(apierrors.IsNotFound))
	})
})
