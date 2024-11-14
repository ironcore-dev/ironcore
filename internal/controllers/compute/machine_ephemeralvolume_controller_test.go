// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package compute

import (
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	"github.com/ironcore-dev/ironcore/utils/annotations"
	. "github.com/ironcore-dev/ironcore/utils/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
									Spec: storagev1alpha1.EphemeralVolumeSpec{},
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
		By("Verifying OwnerRef is updated for ephemeral volume")
		Expect(ephemVolume).Should(HaveField("ObjectMeta.OwnerReferences", ConsistOf(MatchFields(IgnoreExtras, Fields{
			"APIVersion": Equal(computev1alpha1.SchemeGroupVersion.String()),
			"Kind":       Equal("Machine"),
			"Name":       Equal(machine.Name),
		})),
		))

		By("asserting the referenced volume still exists")
		Consistently(Get(refVolume)).Should(Succeed())

		By("waiting for the undesired controlled volume to be gone")
		Eventually(Get(undesiredControlledVolume)).Should(Satisfy(apierrors.IsNotFound))
	})

	It("should not delete externally managed volumes for a machine", func(ctx SpecContext) {
		By("creating an undesired controlled volume")
		externalVolume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "external-volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{},
		}
		Expect(k8sClient.Create(ctx, externalVolume)).To(Succeed())

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
						Name: "primary",
						VolumeSource: computev1alpha1.VolumeSource{
							VolumeRef: &corev1.LocalObjectReference{Name: externalVolume.Name},
						},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, machine)).To(Succeed())

		By("asserting that the external volume is not being deleted")
		Consistently(Object(externalVolume)).Should(HaveField("DeletionTimestamp", BeNil()))
	})

	It("verify ownerRef is set for ephemeral volumes based on reclaim policy", func(ctx SpecContext) {
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
									Spec: storagev1alpha1.EphemeralVolumeSpec{
										ReclaimPolicy: storagev1alpha1.Retain,
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
		By("Verifying OwnerRef is not set for ephemeral volume when reclaim policy is retain")
		Eventually(Object(ephemVolume)).Should(HaveField("ObjectMeta.OwnerReferences", BeEmpty()))

		By("Updating reclaim policy to delete")
		baseMachine := machine.DeepCopy()
		machineVolumes := machine.Spec.Volumes
		for _, machineVolume := range machineVolumes {
			if machineVolume.Ephemeral == nil {
				continue
			}
			machineVolume.Ephemeral.VolumeTemplate.Spec.ReclaimPolicy = storagev1alpha1.Delete
		}
		machine.Spec.Volumes = machineVolumes
		Expect(k8sClient.Patch(ctx, machine, client.MergeFrom(baseMachine))).To(Succeed())
		By("Verifying OwnerRef is updated for ephemeral volume")
		Eventually(Object(ephemVolume)).Should(HaveField("ObjectMeta.OwnerReferences", ConsistOf(MatchFields(IgnoreExtras, Fields{
			"APIVersion": Equal(computev1alpha1.SchemeGroupVersion.String()),
			"Kind":       Equal("Machine"),
			"Name":       Equal(machine.Name),
		})),
		))

	})
})
