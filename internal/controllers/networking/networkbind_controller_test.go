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
	"github.com/google/go-cmp/cmp/cmpopts"
	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	. "github.com/onmetal/onmetal-api/utils/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var IgnoreNetworkPeeringStatusVolatileFields = cmpopts.IgnoreFields(
	networkingv1alpha1.NetworkPeeringStatus{},
	"LastPhaseTransitionTime",
)

var _ = Describe("NetworkBindReconciler", func() {
	ns1, _ := SetupTest()
	ns2 := SetupNamespace(&k8sClient)

	It("should bind two networks in the same namespace referencing each other", func(ctx SpecContext) {
		By("creating the first network")
		network1 := &networkingv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns1.Name,
				Name:      "network-1",
			},
			Spec: networkingv1alpha1.NetworkSpec{
				Handle: "network-1-handle",
				Peerings: []networkingv1alpha1.NetworkPeering{
					{
						Name: "peering",
						NetworkRef: commonv1alpha1.UIDReference{
							Name: "network-2",
						},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, network1)).To(Succeed())

		By("patching the first network as available")
		baseNetwork1 := network1.DeepCopy()
		network1.Status.State = networkingv1alpha1.NetworkStateAvailable
		Expect(k8sClient.Status().Patch(ctx, network1, client.MergeFrom(baseNetwork1))).To(Succeed())

		By("creating the second network")
		network2 := &networkingv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns1.Name,
				Name:      "network-2",
			},
			Spec: networkingv1alpha1.NetworkSpec{
				Handle: "network-2-handle",
				Peerings: []networkingv1alpha1.NetworkPeering{
					{
						Name: "peering",
						NetworkRef: commonv1alpha1.UIDReference{
							Name: "network-1",
						},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, network2)).To(Succeed())

		By("patching the second network to available")
		baseNetwork2 := network2.DeepCopy()
		network2.Status.State = networkingv1alpha1.NetworkStateAvailable
		Expect(k8sClient.Status().Patch(ctx, network2, client.MergeFrom(baseNetwork2))).To(Succeed())

		By("waiting for both networks to reference each other")
		network1Key := client.ObjectKeyFromObject(network1)
		network2Key := client.ObjectKeyFromObject(network2)
		Eventually(ctx, func(g Gomega) {
			Expect(k8sClient.Get(ctx, network1Key, network1)).To(Succeed())
			Expect(k8sClient.Get(ctx, network2Key, network2)).To(Succeed())

			g.Expect(network1.Spec.Peerings).To(ConsistOf(networkingv1alpha1.NetworkPeering{
				Name: "peering",
				NetworkRef: commonv1alpha1.UIDReference{
					Name: network2.Name,
					UID:  network2.UID,
				},
			}))
			g.Expect(network2.Spec.Peerings).To(ConsistOf(networkingv1alpha1.NetworkPeering{
				Name: "peering",
				NetworkRef: commonv1alpha1.UIDReference{
					Name: network1.Name,
					UID:  network1.UID,
				},
			}))

			g.Expect(network1.Status.Peerings).To(ContainElement(BeComparableTo(networkingv1alpha1.NetworkPeeringStatus{
				Name:          "peering",
				NetworkHandle: "network-2-handle",
				Phase:         networkingv1alpha1.NetworkPeeringPhaseBound,
			}, IgnoreNetworkPeeringStatusVolatileFields)))
			g.Expect(network2.Status.Peerings).To(ContainElement(BeComparableTo(networkingv1alpha1.NetworkPeeringStatus{
				Name:          "peering",
				NetworkHandle: "network-1-handle",
				Phase:         networkingv1alpha1.NetworkPeeringPhaseBound,
			}, IgnoreNetworkPeeringStatusVolatileFields)))
		}).Should(Succeed())
	})

	It("should leave two networks in pending if not both exactly reference each other", func(ctx SpecContext) {
		By("creating the first network")
		network1 := &networkingv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns1.Name,
				Name:      "network-1",
			},
			Spec: networkingv1alpha1.NetworkSpec{
				Handle: "network-1-handle",
				Peerings: []networkingv1alpha1.NetworkPeering{
					{
						Name: "peering",
						NetworkRef: commonv1alpha1.UIDReference{
							Name: "network-2",
						},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, network1)).To(Succeed())

		By("patching the first network as available")
		baseNetwork1 := network1.DeepCopy()
		network1.Status.State = networkingv1alpha1.NetworkStateAvailable
		Expect(k8sClient.Status().Patch(ctx, network1, client.MergeFrom(baseNetwork1))).To(Succeed())

		By("creating the second network")
		network2 := &networkingv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns1.Name,
				Name:      "network-2",
			},
			Spec: networkingv1alpha1.NetworkSpec{
				Handle: "network-2-handle",
				Peerings: []networkingv1alpha1.NetworkPeering{
					{
						Name: "peering",
						NetworkRef: commonv1alpha1.UIDReference{
							Name: "network-other",
						},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, network2)).To(Succeed())

		By("patching the second network to available")
		baseNetwork2 := network2.DeepCopy()
		network2.Status.State = networkingv1alpha1.NetworkStateAvailable
		Expect(k8sClient.Status().Patch(ctx, network2, client.MergeFrom(baseNetwork2))).To(Succeed())

		By("asserting both networks don't get peered")
		network1Key := client.ObjectKeyFromObject(network1)
		network2Key := client.ObjectKeyFromObject(network2)

		reportBothPeeringsAsPending := func(g Gomega) {
			Expect(k8sClient.Get(ctx, network1Key, network1)).To(Succeed())
			Expect(k8sClient.Get(ctx, network2Key, network2)).To(Succeed())

			g.Expect(network1.Spec.Peerings).To(ConsistOf(networkingv1alpha1.NetworkPeering{
				Name: "peering",
				NetworkRef: commonv1alpha1.UIDReference{
					Name: "network-2",
				},
			}))
			g.Expect(network2.Spec.Peerings).To(ConsistOf(networkingv1alpha1.NetworkPeering{
				Name: "peering",
				NetworkRef: commonv1alpha1.UIDReference{
					Name: "network-other",
				},
			}))

			g.Expect(network1.Status.Peerings).To(ContainElement(BeComparableTo(networkingv1alpha1.NetworkPeeringStatus{
				Name:  "peering",
				Phase: networkingv1alpha1.NetworkPeeringPhasePending,
			}, IgnoreNetworkPeeringStatusVolatileFields)))
			g.Expect(network2.Status.Peerings).To(ContainElement(BeComparableTo(networkingv1alpha1.NetworkPeeringStatus{
				Name:  "peering",
				Phase: networkingv1alpha1.NetworkPeeringPhasePending,
			}, IgnoreNetworkPeeringStatusVolatileFields)))
		}

		By("waiting for both peerings to be reported as pending")
		Eventually(reportBothPeeringsAsPending).WithContext(ctx).Should(Succeed())

		By("asserting it stays that way")
		Consistently(reportBothPeeringsAsPending).WithContext(ctx).Should(Succeed())
	})

	It("should peer two networks from different namespaces if they reference each other correctly", func(ctx SpecContext) {
		By("creating the first network in the first namespace")
		network1 := &networkingv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns1.Name,
				Name:      "network-1",
			},
			Spec: networkingv1alpha1.NetworkSpec{
				Handle: "network-1-handle",
				Peerings: []networkingv1alpha1.NetworkPeering{
					{
						Name: "peering",
						NetworkRef: commonv1alpha1.UIDReference{
							Namespace: ns2.Name,
							Name:      "network-2",
						},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, network1)).To(Succeed())

		By("patching the first network as available")
		baseNetwork1 := network1.DeepCopy()
		network1.Status.State = networkingv1alpha1.NetworkStateAvailable
		Expect(k8sClient.Status().Patch(ctx, network1, client.MergeFrom(baseNetwork1))).To(Succeed())

		By("creating the second network in the second namespace")
		network2 := &networkingv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns2.Name,
				Name:      "network-2",
			},
			Spec: networkingv1alpha1.NetworkSpec{
				Handle: "network-2-handle",
				Peerings: []networkingv1alpha1.NetworkPeering{
					{
						Name: "peering",
						NetworkRef: commonv1alpha1.UIDReference{
							Namespace: ns1.Name,
							Name:      "network-1",
						},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, network2)).To(Succeed())

		By("patching the second network to available")
		baseNetwork2 := network2.DeepCopy()
		network2.Status.State = networkingv1alpha1.NetworkStateAvailable
		Expect(k8sClient.Status().Patch(ctx, network2, client.MergeFrom(baseNetwork2))).To(Succeed())

		By("waiting for both networks to reference each other")
		network1Key := client.ObjectKeyFromObject(network1)
		network2Key := client.ObjectKeyFromObject(network2)
		Eventually(ctx, func(g Gomega) {
			Expect(k8sClient.Get(ctx, network1Key, network1)).To(Succeed())
			Expect(k8sClient.Get(ctx, network2Key, network2)).To(Succeed())

			g.Expect(network1.Spec.Peerings).To(ConsistOf(networkingv1alpha1.NetworkPeering{
				Name: "peering",
				NetworkRef: commonv1alpha1.UIDReference{
					Namespace: ns2.Name,
					Name:      network2.Name,
					UID:       network2.UID,
				},
			}))
			g.Expect(network2.Spec.Peerings).To(ConsistOf(networkingv1alpha1.NetworkPeering{
				Name: "peering",
				NetworkRef: commonv1alpha1.UIDReference{
					Namespace: ns1.Name,
					Name:      network1.Name,
					UID:       network1.UID,
				},
			}))

			g.Expect(network1.Status.Peerings).To(ContainElement(BeComparableTo(networkingv1alpha1.NetworkPeeringStatus{
				Name:          "peering",
				NetworkHandle: "network-2-handle",
				Phase:         networkingv1alpha1.NetworkPeeringPhaseBound,
			}, IgnoreNetworkPeeringStatusVolatileFields)))
			g.Expect(network2.Status.Peerings).To(ContainElement(BeComparableTo(networkingv1alpha1.NetworkPeeringStatus{
				Name:          "peering",
				NetworkHandle: "network-1-handle",
				Phase:         networkingv1alpha1.NetworkPeeringPhaseBound,
			}, IgnoreNetworkPeeringStatusVolatileFields)))
		}).Should(Succeed())
	})
})
