/*
 * Copyright (c) 2022 by the OnMetal authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package storage

import (
	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/apis/storage/v1alpha1"
	"github.com/onmetal/onmetal-api/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

var _ = Describe("VolumeReconciler", func() {
	ctx := testutils.SetupContext()
	ns := SetupTest(ctx)

	It("should schedule, bind and unbind a volume to a machine", func() {
		By("creating a volume")
		volume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{
				VolumeClassRef: &corev1.LocalObjectReference{Name: "my-class"},
				VolumePoolRef:  &corev1.LocalObjectReference{Name: "my-pool"},
				Resources: corev1.ResourceList{
					"storage": resource.MustParse("1Gi"),
				},
			},
		}
		Expect(k8sClient.Create(ctx, volume)).To(Succeed())

		By("creating a machine referencing the volume")
		machine := &computev1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "machine-",
			},
			Spec: computev1alpha1.MachineSpec{
				MachineClassRef: corev1.LocalObjectReference{Name: "my-class"},
				MachinePoolRef:  &corev1.LocalObjectReference{Name: "my-pool"},
				Image:           "my-image",
				Volumes: []computev1alpha1.Volume{
					{
						Name: "my-volume",
						VolumeSource: computev1alpha1.VolumeSource{
							VolumeRef: &corev1.LocalObjectReference{Name: volume.Name},
						},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, machine)).To(Succeed())

		By("waiting for the volume to be bound to the machine")
		Eventually(Object(volume)).Should(SatisfyAll(
			HaveField("Status.Phase", Equal(storagev1alpha1.VolumePhaseBound)),
			HaveField("Spec.ClaimRef", Equal(&commonv1alpha1.LocalUIDReference{
				Name: machine.Name,
				UID:  machine.UID,
			})),
		))

		By("deleting the machine")
		Expect(k8sClient.Delete(ctx, machine)).To(Succeed())

		By("waiting for the volume to be unbound again")
		Eventually(Object(volume)).Should(SatisfyAll(
			HaveField("Status.Phase", Equal(storagev1alpha1.VolumePhaseUnbound)),
			HaveField("Spec.ClaimRef", BeNil()),
		))
	})

	It("should not claim a volume if it is marked as unclaimable", func() {
		By("creating an unclaimable volume")
		volume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{
				VolumeClassRef: &corev1.LocalObjectReference{Name: "my-class"},
				VolumePoolRef:  &corev1.LocalObjectReference{Name: "my-pool"},
				Unclaimable:    true,
				Resources: corev1.ResourceList{
					"storage": resource.MustParse("1Gi"),
				},
			},
		}
		Expect(k8sClient.Create(ctx, volume)).To(Succeed())

		By("waiting for the volume to report as unbound")
		Eventually(Object(volume)).Should(SatisfyAll(
			HaveField("Status.Phase", Equal(storagev1alpha1.VolumePhaseUnbound)),
			HaveField("Spec.ClaimRef", BeNil()),
		))

		By("creating a machine referencing the volume")
		machine := &computev1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "machine-",
			},
			Spec: computev1alpha1.MachineSpec{
				MachineClassRef: corev1.LocalObjectReference{Name: "my-class"},
				MachinePoolRef:  &corev1.LocalObjectReference{Name: "my-pool"},
				Image:           "my-image",
				Volumes: []computev1alpha1.Volume{
					{
						Name: "my-volume",
						VolumeSource: computev1alpha1.VolumeSource{
							VolumeRef: &corev1.LocalObjectReference{Name: volume.Name},
						},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, machine)).To(Succeed())

		By("asserting the volume does not get claimed")
		Consistently(Object(volume)).Should(SatisfyAll(
			HaveField("Status.Phase", Equal(storagev1alpha1.VolumePhaseUnbound)),
			HaveField("Spec.ClaimRef", BeNil()),
		))
	})
})
