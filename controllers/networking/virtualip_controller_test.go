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
	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/apis/networking/v1alpha1"
	"github.com/onmetal/onmetal-api/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("VirtualIPReconciler", func() {
	ctx := testutils.SetupContext()
	ns := SetupTest(ctx)

	It("should set a virtual ip to unbound if nothing binds it", func() {
		By("creating a virtual ip")
		virtualIP := &networkingv1alpha1.VirtualIP{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "vip-",
			},
			Spec: networkingv1alpha1.VirtualIPSpec{
				Type:     networkingv1alpha1.VirtualIPTypePublic,
				IPFamily: corev1.IPv4Protocol,
			},
		}
		Expect(k8sClient.Create(ctx, virtualIP)).To(Succeed())

		By("waiting for the status to be updated")
		virtualIPKey := client.ObjectKeyFromObject(virtualIP)
		Eventually(func() networkingv1alpha1.VirtualIPPhase {
			Expect(k8sClient.Get(ctx, virtualIPKey, virtualIP)).To(Succeed())
			return virtualIP.Status.Phase
		}).Should(Equal(networkingv1alpha1.VirtualIPPhaseUnbound))
	})

	It("should set the virtual ip status to bound if it gets bound and set it to unbound if it gets unbound", func() {
		By("creating a virtual ip")
		virtualIP := &networkingv1alpha1.VirtualIP{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "vip-",
			},
			Spec: networkingv1alpha1.VirtualIPSpec{
				Type:     networkingv1alpha1.VirtualIPTypePublic,
				IPFamily: corev1.IPv4Protocol,
			},
		}
		Expect(k8sClient.Create(ctx, virtualIP)).To(Succeed())

		By("updating the virtual ip to have an ip allocated")
		virtualIP.Status.IP = commonv1alpha1.MustParseNewIP("10.0.0.1")
		Expect(k8sClient.Status().Update(ctx, virtualIP)).To(Succeed())

		By("creating a virtual ip claim")
		virtualIPClaim := &networkingv1alpha1.VirtualIPClaim{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "vip-claim-",
			},
			Spec: networkingv1alpha1.VirtualIPClaimSpec{
				Type:     networkingv1alpha1.VirtualIPTypePublic,
				IPFamily: corev1.IPv4Protocol,
				VirtualIPRef: &corev1.LocalObjectReference{
					Name: virtualIP.Name,
				},
			},
		}
		Expect(k8sClient.Create(ctx, virtualIPClaim)).To(Succeed())

		By("waiting for the virtual ip to be bound")
		virtualIPKey := client.ObjectKeyFromObject(virtualIP)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, virtualIPKey, virtualIP)).To(Succeed())

			g.Expect(virtualIP.Status.Phase).To(Equal(networkingv1alpha1.VirtualIPPhaseBound))
			g.Expect(virtualIP.Spec.ClaimRef).To(Equal(&commonv1alpha1.LocalUIDReference{
				Name: virtualIPClaim.Name,
				UID:  virtualIPClaim.UID,
			}))
		}).Should(Succeed())

		By("deleting the virtual ip claim")
		Expect(k8sClient.Delete(ctx, virtualIPClaim)).To(Succeed())

		By("waiting for the virtual ip to be unbound again")
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, virtualIPKey, virtualIP)).To(Succeed())

			g.Expect(virtualIP.Status.Phase).To(Equal(networkingv1alpha1.VirtualIPPhaseUnbound))
			g.Expect(virtualIP.Spec.ClaimRef).To(BeNil())
		}).Should(Succeed())
	})
})
