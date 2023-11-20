// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package compute

import (
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	"github.com/ironcore-dev/ironcore/utils/annotations"
	. "github.com/ironcore-dev/ironcore/utils/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

var _ = Describe("MachineEphemeralNetworkInterfaceController", func() {
	ns := SetupNamespace(&k8sClient)
	machineClass := SetupMachineClass()

	It("should create ephemeral network interfaces for machines", func(ctx SpecContext) {
		By("creating a machine")
		machine := &computev1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "machine-",
			},
			Spec: computev1alpha1.MachineSpec{
				MachineClassRef: corev1.LocalObjectReference{Name: machineClass.Name},
				NetworkInterfaces: []computev1alpha1.NetworkInterface{
					{
						Name: "ephem-nic",
						NetworkInterfaceSource: computev1alpha1.NetworkInterfaceSource{
							Ephemeral: &computev1alpha1.EphemeralNetworkInterfaceSource{
								NetworkInterfaceTemplate: &networkingv1alpha1.NetworkInterfaceTemplateSpec{
									Spec: networkingv1alpha1.NetworkInterfaceSpec{
										NetworkRef: corev1.LocalObjectReference{Name: "my-network"},
										IPs: []networkingv1alpha1.IPSource{
											{Value: commonv1alpha1.MustParseNewIP("10.0.0.2")},
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

		By("waiting for the ephemeral network interface to exist")
		ephemNic := &networkingv1alpha1.NetworkInterface{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns.Name,
				Name:      computev1alpha1.MachineEphemeralNetworkInterfaceName(machine.Name, "ephem-nic"),
			},
		}
		Eventually(Object(ephemNic)).Should(SatisfyAll(
			BeControlledBy(machine),
			HaveField("Spec.NetworkRef", corev1.LocalObjectReference{Name: "my-network"}),
			HaveField("Spec.IPs", []networkingv1alpha1.IPSource{
				{Value: commonv1alpha1.MustParseNewIP("10.0.0.2")},
			}),
		))
	})

	It("should delete undesired ephemeral network interfaces", func(ctx SpecContext) {
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

		By("creating an undesired network interface")
		undesiredNic := &networkingv1alpha1.NetworkInterface{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "undesired-nic-",
			},
			Spec: networkingv1alpha1.NetworkInterfaceSpec{
				NetworkRef: corev1.LocalObjectReference{Name: "my-network"},
				IPs: []networkingv1alpha1.IPSource{
					{Value: commonv1alpha1.MustParseNewIP("10.0.0.1")},
				},
			},
		}
		annotations.SetDefaultEphemeralManagedBy(undesiredNic)
		Expect(ctrl.SetControllerReference(machine, undesiredNic, k8sClient.Scheme())).To(Succeed())
		Expect(k8sClient.Create(ctx, undesiredNic)).To(Succeed())

		By("waiting for the undesired network interface to be gone")
		Eventually(Get(undesiredNic)).Should(Satisfy(apierrors.IsNotFound))
	})

	It("should not delete an externally managed network interface", func(ctx SpecContext) {
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

		By("creating an externally managed network interface")
		externalNic := &networkingv1alpha1.NetworkInterface{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "external-nic-",
			},
			Spec: networkingv1alpha1.NetworkInterfaceSpec{
				NetworkRef: corev1.LocalObjectReference{Name: "my-network"},
				IPs: []networkingv1alpha1.IPSource{
					{Value: commonv1alpha1.MustParseNewIP("10.0.0.1")},
				},
			},
		}
		Expect(ctrl.SetControllerReference(machine, externalNic, k8sClient.Scheme())).To(Succeed())
		Expect(k8sClient.Create(ctx, externalNic)).To(Succeed())

		By("asserting the network interface is not being deleted")
		Consistently(Object(externalNic)).Should(HaveField("DeletionTimestamp", BeNil()))
	})
})
