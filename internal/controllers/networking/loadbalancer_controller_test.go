// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0
package networking

import (
	. "github.com/ironcore-dev/ironcore/utils/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"

	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
)

var _ = Describe("LoadBalancerReconciler", func() {
	ns := SetupNamespace(&k8sClient)

	It("should reconcile the prefix and routing destinations", func(ctx SpecContext) {
		By("creating a network")
		network := &networkingv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "network-",
			},
		}
		Expect(k8sClient.Create(ctx, network)).To(Succeed())

		By("setting the network to be available")
		Eventually(UpdateStatus(network, func() {
			network.Status.State = networkingv1alpha1.NetworkStateAvailable
		})).Should(Succeed())

		By("creating a load balancer")
		loadBalancer := &networkingv1alpha1.LoadBalancer{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "load-balancer-",
			},
			Spec: networkingv1alpha1.LoadBalancerSpec{
				Type: networkingv1alpha1.LoadBalancerTypePublic,
				IPFamilies: []corev1.IPFamily{
					corev1.IPv4Protocol,
				},
				NetworkRef: corev1.LocalObjectReference{Name: network.Name},
				NetworkInterfaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"foo": "bar"},
				},
			},
		}
		Expect(k8sClient.Create(ctx, loadBalancer)).To(Succeed())

		By("waiting for the load balancer routing to exist with no destinations")
		loadBalancerRouting := &networkingv1alpha1.LoadBalancerRouting{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: loadBalancer.Namespace,
				Name:      loadBalancer.Name,
			},
		}
		Eventually(Object(loadBalancerRouting)).Should(SatisfyAll(
			BeControlledBy(loadBalancer),
			HaveField("Destinations", BeEmpty()),
		))

		By("creating a network interface")
		nic := &networkingv1alpha1.NetworkInterface{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "nic-",
				Labels:       map[string]string{"foo": "bar"},
			},
			Spec: networkingv1alpha1.NetworkInterfaceSpec{
				NetworkRef: corev1.LocalObjectReference{Name: network.Name},
				IPFamilies: []corev1.IPFamily{
					corev1.IPv4Protocol,
				},
				IPs: []networkingv1alpha1.IPSource{
					{
						Value: commonv1alpha1.MustParseNewIP("10.0.0.1"),
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, nic)).To(Succeed())

		By("setting the network interface to be available with IP and provider ID")
		Eventually(Update(nic, func() {
			nic.Spec.ProviderID = "my://provider-id"
		})).Should(Succeed())
		Eventually(UpdateStatus(nic, func() {
			nic.Status.State = networkingv1alpha1.NetworkInterfaceStateAvailable
			nic.Status.IPs = commonv1alpha1.MustParseIPs("10.0.0.1")
		})).Should(Succeed())

		By("waiting for the load balancer routing to be updated")
		Eventually(Object(loadBalancerRouting)).Should(SatisfyAll(
			HaveField("NetworkRef", commonv1alpha1.LocalUIDReference{
				Name: network.Name,
				UID:  network.UID,
			}),
			HaveField("Destinations", []networkingv1alpha1.LoadBalancerDestination{
				{
					IP: commonv1alpha1.MustParseIP("10.0.0.1"),
					TargetRef: &networkingv1alpha1.LoadBalancerTargetRef{
						Name:       nic.Name,
						UID:        nic.UID,
						ProviderID: "my://provider-id",
					},
				},
			}),
		))

		By("removing the labels from the network interface")
		Eventually(Update(nic, func() {
			nic.Labels = nil
		})).Should(Succeed())

		By("waiting for the load balancer routing to be updated")
		Eventually(Object(loadBalancerRouting)).Should(SatisfyAll(
			HaveField("NetworkRef", commonv1alpha1.LocalUIDReference{
				Name: network.Name,
				UID:  network.UID,
			}),
			HaveField("Destinations", BeEmpty()),
		))
	})
})
