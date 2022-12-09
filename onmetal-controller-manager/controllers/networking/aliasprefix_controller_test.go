/*
 * Copyright (c) 2021 by the OnMetal authors.
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
	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	"github.com/onmetal/onmetal-api/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("AliasPrefixReconciler", func() {
	ctx := testutils.SetupContext()
	ns := SetupTest(ctx)

	It("should reconcile the prefix and routing destinations", func() {
		By("creating a network")
		network := &networkingv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "network-",
			},
			Spec: networkingv1alpha1.NetworkSpec{
				Handle: "foo",
			},
		}
		Expect(k8sClient.Create(ctx, network)).To(Succeed())

		By("patching the network to be available")
		networkBase := network.DeepCopy()
		network.Status.State = networkingv1alpha1.NetworkStateAvailable
		Expect(k8sClient.Status().Patch(ctx, network, client.MergeFrom(networkBase))).To(Succeed())

		By("creating an alias prefix")
		aliasPrefix := &networkingv1alpha1.AliasPrefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "alias-prefix-",
			},
			Spec: networkingv1alpha1.AliasPrefixSpec{
				NetworkRef: corev1.LocalObjectReference{Name: network.Name},
				Prefix: networkingv1alpha1.PrefixSource{
					Value: commonv1alpha1.MustParseNewIPPrefix("10.0.0.0/28"),
				},
				NetworkInterfaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"foo": "bar"},
				},
			},
		}
		Expect(k8sClient.Create(ctx, aliasPrefix)).To(Succeed())

		By("waiting for the alias prefix to report the prefix")
		aliasPrefixKey := client.ObjectKeyFromObject(aliasPrefix)
		Eventually(func() *commonv1alpha1.IPPrefix {
			Expect(k8sClient.Get(ctx, aliasPrefixKey, aliasPrefix)).To(Succeed())
			return aliasPrefix.Status.Prefix
		}).Should(Equal(commonv1alpha1.MustParseNewIPPrefix("10.0.0.0/28")))

		By("waiting for the alias prefix routing to exist with no destinations")
		aliasPrefixRouting := &networkingv1alpha1.AliasPrefixRouting{}
		Eventually(func(g Gomega) {
			err := k8sClient.Get(ctx, aliasPrefixKey, aliasPrefixRouting)
			Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(metav1.IsControlledBy(aliasPrefixRouting, aliasPrefix)).To(BeTrue(), "alias prefix routing is not controlled by alias prefix: %#v", aliasPrefixRouting.OwnerReferences)
			g.Expect(aliasPrefixRouting.NetworkRef).To(Equal(commonv1alpha1.LocalUIDReference{
				Name: network.Name,
				UID:  network.UID,
			}))
			g.Expect(aliasPrefixRouting.Destinations).To(BeEmpty())
		}).Should(Succeed())

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

		By("waiting for the alias prefix routing to be updated")
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, aliasPrefixKey, aliasPrefixRouting)).To(Succeed())

			g.Expect(aliasPrefixRouting.NetworkRef).To(Equal(commonv1alpha1.LocalUIDReference{
				Name: network.Name,
				UID:  network.UID,
			}))
			g.Expect(aliasPrefixRouting.Destinations).To(Equal([]commonv1alpha1.LocalUIDReference{
				{Name: nic.Name, UID: nic.UID},
			}))
		}).Should(Succeed())
	})
})
