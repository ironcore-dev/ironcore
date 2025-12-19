// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package compute

import (
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	"github.com/ironcore-dev/ironcore/utils/annotations"
	. "github.com/ironcore-dev/ironcore/utils/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

var _ = Describe("MachineEphemeralVolumeController", func() {
	ns := SetupNamespace(&k8sClient)
	machineClass := SetupMachineClass()
	volumeClass := SetupVolumeClass()

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
		annotations.SetDefaultEphemeralManagedBy(undesiredControlledVolume)
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

	It("should manage ephemeral volumes for a machine and setting architecture hint", func(ctx SpecContext) {
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
						Name: "ephem-volume",
						VolumeSource: computev1alpha1.VolumeSource{
							Ephemeral: &computev1alpha1.EphemeralVolumeSource{
								VolumeTemplate: &storagev1alpha1.VolumeTemplateSpec{
									Spec: storagev1alpha1.VolumeSpec{
										VolumeClassRef: &corev1.LocalObjectReference{Name: volumeClass.Name},
										Resources: corev1alpha1.ResourceList{
											corev1alpha1.ResourceStorage: resource.MustParse("500Mi"),
										},
										DataSource: storagev1alpha1.VolumeDataSource{
											OSImage: &storagev1alpha1.OSDataSource{
												Image: "test-os-image",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, machine)).To(Succeed())

		By("waiting for the ephemeral volume to exist")
		ephemVolume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns.Name,
				Name:      computev1alpha1.MachineEphemeralVolumeName(machine.Name, "ephem-volume"),
			},
		}
		Eventually(Get(ephemVolume)).Should(Succeed())

		Eventually(Object(ephemVolume)).Should(SatisfyAll(
			HaveField("Spec.DataSource.OSImage", &storagev1alpha1.OSDataSource{
				Image:        "test-os-image",
				Architecture: ptr.To("amd64"),
			}),
		))
	})

	It("should not delete externally managed volumes for a machine", func(ctx SpecContext) {
		By("creating a machine")
		machine := &computev1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "machine-",
			},
			Spec: computev1alpha1.MachineSpec{
				MachineClassRef: corev1.LocalObjectReference{Name: machineClass.Name},
			},
		}
		Expect(k8sClient.Create(ctx, machine)).To(Succeed())

		By("creating an undesired controlled volume")
		externalVolume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "external-volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{},
		}
		_ = ctrl.SetControllerReference(machine, externalVolume, k8sClient.Scheme())
		Expect(k8sClient.Create(ctx, externalVolume)).To(Succeed())

		By("asserting that the external volume is not being deleted")
		Consistently(Object(externalVolume)).Should(HaveField("DeletionTimestamp", BeNil()))
	})
})
