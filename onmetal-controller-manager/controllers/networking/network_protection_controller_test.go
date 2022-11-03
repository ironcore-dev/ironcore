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

package networking

import (
	"github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/apis/networking/v1alpha1"
	"github.com/onmetal/onmetal-api/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("NetworkProtectionReconciler", func() {
	ctx := testutils.SetupContext()
	ns := SetupTest(ctx)

	var (
		network *networkingv1alpha1.Network
	)

	BeforeEach(func() {
		By("creating a network")
		network = &networkingv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "my-network-",
			},
		}
		Expect(k8sClient.Create(ctx, network)).To(Succeed())
	})

	It("should add and remove a finalizer for a network in use/not used by a network interface", func() {
		By("creating a network interface referencing this network")
		networkInterface := &networkingv1alpha1.NetworkInterface{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "my-networkinterface-",
			},
			Spec: networkingv1alpha1.NetworkInterfaceSpec{
				NetworkRef: corev1.LocalObjectReference{
					Name: network.Name,
				},
				IPFamilies: []corev1.IPFamily{
					corev1.IPv4Protocol,
				},
				IPs: []networkingv1alpha1.IPSource{{
					Value: v1alpha1.MustParseNewIP("10.0.0.1"),
				}},
			},
		}
		Expect(k8sClient.Create(ctx, networkInterface)).To(Succeed())

		By("ensuring that the network finalizer has been set")
		networkKey := types.NamespacedName{
			Namespace: ns.Name,
			Name:      network.Name,
		}
		Eventually(func(g Gomega) {
			err := k8sClient.Get(ctx, networkKey, network)
			Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(network.GetFinalizers()).To(ContainElement(networkFinalizer))
		}).Should(Succeed())

		By("deleting the network interface")
		Expect(k8sClient.Delete(ctx, networkInterface)).To(Succeed())

		By("deleting the network")
		Expect(k8sClient.Delete(ctx, network)).To(Succeed())

		By("ensuring that the network has been deleted")
		Eventually(func(g Gomega) {
			err := k8sClient.Get(ctx, networkKey, network)
			Expect(client.IgnoreNotFound(err)).To(BeNil())
		}).Should(Succeed())
	})

	It("should remove a finalizer for a network in deletion state once the reference network interface is deleted", func() {
		By("creating a network interface referencing this network")
		networkInterface := &networkingv1alpha1.NetworkInterface{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "my-networkinterface-",
			},
			Spec: networkingv1alpha1.NetworkInterfaceSpec{
				NetworkRef: corev1.LocalObjectReference{
					Name: network.Name,
				},
				IPFamilies: []corev1.IPFamily{
					corev1.IPv4Protocol,
				},
				IPs: []networkingv1alpha1.IPSource{{
					Value: v1alpha1.MustParseNewIP("10.0.0.1"),
				}},
			},
		}
		Expect(k8sClient.Create(ctx, networkInterface)).To(Succeed())

		By("ensuring that the network finalizer has been set")
		networkKey := types.NamespacedName{
			Namespace: ns.Name,
			Name:      network.Name,
		}
		Eventually(func(g Gomega) {
			err := k8sClient.Get(ctx, networkKey, network)
			Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(network.GetFinalizers()).To(ContainElement(networkFinalizer))
		}).Should(Succeed())

		By("deleting the network")
		Expect(k8sClient.Delete(ctx, network)).To(Succeed())

		By("ensuring that the network has a deletion timestamp set and the finalizer still present")
		Eventually(func(g Gomega) {
			err := k8sClient.Get(ctx, networkKey, network)
			Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(network.DeletionTimestamp.IsZero()).To(BeFalse())
			g.Expect(network.GetFinalizers()).To(ContainElement(networkFinalizer))
		}).Should(Succeed())

		By("deleting the network interface")
		Expect(k8sClient.Delete(ctx, networkInterface)).To(Succeed())

		By("ensuring that the network has been deleted")
		Eventually(func(g Gomega) {
			err := k8sClient.Get(ctx, networkKey, network)
			Expect(client.IgnoreNotFound(err)).To(BeNil())
		}).Should(Succeed())
	})

	It("should keep a finalizer if one of two network interfaces is removed", func() {
		By("creating the first network interface referencing this network")
		networkInterface := &networkingv1alpha1.NetworkInterface{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "my-networkinterface-",
			},
			Spec: networkingv1alpha1.NetworkInterfaceSpec{
				NetworkRef: corev1.LocalObjectReference{
					Name: network.Name,
				},
				IPFamilies: []corev1.IPFamily{
					corev1.IPv4Protocol,
				},
				IPs: []networkingv1alpha1.IPSource{{
					Value: v1alpha1.MustParseNewIP("10.0.0.1"),
				}},
			},
		}
		Expect(k8sClient.Create(ctx, networkInterface)).To(Succeed())

		By("creating a second network interface referencing this network")
		networkInterface2 := &networkingv1alpha1.NetworkInterface{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "my-networkinterface-",
			},
			Spec: networkingv1alpha1.NetworkInterfaceSpec{
				NetworkRef: corev1.LocalObjectReference{
					Name: network.Name,
				},
				IPFamilies: []corev1.IPFamily{
					corev1.IPv4Protocol,
				},
				IPs: []networkingv1alpha1.IPSource{{
					Value: v1alpha1.MustParseNewIP("10.0.0.2"),
				}},
			},
		}
		Expect(k8sClient.Create(ctx, networkInterface2)).To(Succeed())

		By("ensuring that the network finalizer has been set")
		networkKey := types.NamespacedName{
			Namespace: ns.Name,
			Name:      network.Name,
		}
		Eventually(func(g Gomega) {
			err := k8sClient.Get(ctx, networkKey, network)
			Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(network.GetFinalizers()).To(ContainElement(networkFinalizer))
		}).Should(Succeed())

		By("deleting the first network interface")
		Expect(k8sClient.Delete(ctx, networkInterface)).To(Succeed())

		By("deleting the network")
		Expect(k8sClient.Delete(ctx, network)).To(Succeed())

		By("ensuring that the network has a deletion timestamp set and finalizer still present")
		Eventually(func(g Gomega) {
			err := k8sClient.Get(ctx, networkKey, network)
			Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(network.DeletionTimestamp.IsZero()).To(BeFalse())
			g.Expect(network.GetFinalizers()).To(ContainElement(networkFinalizer))
		}).Should(Succeed())
	})

	It("should add/remove a finalizer for a network in use/not used by an alias prefix", func() {
		By("creating an alias prefix referencing this network")
		aliasPrefix := &networkingv1alpha1.AliasPrefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "my-aliasprefix-",
			},
			Spec: networkingv1alpha1.AliasPrefixSpec{
				NetworkRef: corev1.LocalObjectReference{
					Name: network.Name,
				},
			},
		}
		Expect(k8sClient.Create(ctx, aliasPrefix)).To(Succeed())

		By("ensuring that the network finalizer has been set")
		networkKey := types.NamespacedName{
			Namespace: ns.Name,
			Name:      network.Name,
		}
		Eventually(func(g Gomega) {
			err := k8sClient.Get(ctx, networkKey, network)
			Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(network.GetFinalizers()).To(ContainElement(networkFinalizer))
		}).Should(Succeed())

		By("deleting the alias prefix")
		Expect(k8sClient.Delete(ctx, aliasPrefix)).To(Succeed())

		By("deleting the network")
		Expect(k8sClient.Delete(ctx, network)).To(Succeed())

		By("ensuring that the network has been deleted")
		Eventually(func(g Gomega) {
			err := k8sClient.Get(ctx, networkKey, network)
			Expect(client.IgnoreNotFound(err)).To(BeNil())
		}).Should(Succeed())
	})

	It("should keep a finalizer if one of two alias prefix is removed", func() {
		By("creating the first alias prefix referencing this network")
		aliasPrefix := &networkingv1alpha1.AliasPrefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "my-aliasprefix-",
			},
			Spec: networkingv1alpha1.AliasPrefixSpec{
				NetworkRef: corev1.LocalObjectReference{
					Name: network.Name,
				},
			},
		}
		Expect(k8sClient.Create(ctx, aliasPrefix)).To(Succeed())

		By("creating a second alias prefix referencing this network")
		aliasPrefix2 := &networkingv1alpha1.AliasPrefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "my-aliasprefix-",
			},
			Spec: networkingv1alpha1.AliasPrefixSpec{
				NetworkRef: corev1.LocalObjectReference{
					Name: network.Name,
				},
			},
		}
		Expect(k8sClient.Create(ctx, aliasPrefix2)).To(Succeed())

		By("ensuring that the network finalizer has been set")
		networkKey := types.NamespacedName{
			Namespace: ns.Name,
			Name:      network.Name,
		}
		Eventually(func(g Gomega) {
			err := k8sClient.Get(ctx, networkKey, network)
			Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(network.GetFinalizers()).To(ContainElement(networkFinalizer))
		}).Should(Succeed())

		By("deleting the first alias prefix")
		Expect(k8sClient.Delete(ctx, aliasPrefix)).To(Succeed())

		By("deleting the network")
		Expect(k8sClient.Delete(ctx, network)).To(Succeed())

		By("ensuring that the network has a deletion timestamp set and finalizer still present")
		Eventually(func(g Gomega) {
			err := k8sClient.Get(ctx, networkKey, network)
			Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(network.DeletionTimestamp.IsZero()).To(BeFalse())
			g.Expect(network.GetFinalizers()).To(ContainElement(networkFinalizer))
		}).Should(Succeed())
	})

	It("should allow deletion of an unused network", func() {
		By("deleting the network")
		Expect(k8sClient.Delete(ctx, network)).To(Succeed())

		By("ensuring that the network is not found")
		networkKey := types.NamespacedName{
			Namespace: ns.Name,
			Name:      network.Name,
		}
		Eventually(func(g Gomega) {
			err := k8sClient.Get(ctx, networkKey, network)
			Expect(client.IgnoreNotFound(err)).To(BeNil())
		}).Should(Succeed())
	})
})
