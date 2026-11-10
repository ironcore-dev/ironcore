// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0
package networking

import (
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	. "github.com/ironcore-dev/ironcore/utils/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

var _ = Describe("NetworkInterfaceReleaseReconciler", func() {
	ns := SetupNamespace(&k8sClient)

	It("should release network interfaces whose owner is gone", func(ctx SpecContext) {
		By("creating a network interface")
		nic := &networkingv1alpha1.NetworkInterface{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "nic-",
			},
			Spec: networkingv1alpha1.NetworkInterfaceSpec{
				NetworkRef: corev1.LocalObjectReference{Name: "my-network"},
				IPs:        []networkingv1alpha1.IPSource{{Value: commonv1alpha1.MustParseNewIP("10.0.0.1")}},
			},
		}
		Expect(k8sClient.Create(ctx, nic)).To(Succeed())

		By("setting an available status on the network interface")
		Eventually(UpdateStatus(nic, func() {
			nic.Status = networkingv1alpha1.NetworkInterfaceStatus{
				State: networkingv1alpha1.NetworkInterfaceStateAvailable,
				IPs:   []commonv1alpha1.IP{*commonv1alpha1.MustParseNewIP("10.0.0.1")},
			}
		})).Should(Succeed())

		By("adding a machine reference claiming the network interface")
		Eventually(Update(nic, func() {
			nic.Spec.MachineRef = &commonv1alpha1.LocalUIDReference{
				Name: "should-not-exist",
				UID:  uuid.NewUUID(),
			}
		})).Should(Succeed())

		By("waiting for the network interface to be released")
		Eventually(Object(nic)).Should(SatisfyAll(
			HaveField("Spec.MachineRef", BeNil()),
			HaveField("Status.State", Equal(networkingv1alpha1.NetworkInterfaceStatePending)),
			HaveField("Status.IPs", BeEmpty()),
		))
	})
})
