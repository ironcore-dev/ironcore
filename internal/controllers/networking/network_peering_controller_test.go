// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package networking

import (
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	. "github.com/ironcore-dev/ironcore/utils/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

var _ = Describe("NetworkPeeringController", func() {
	ns := SetupNamespace(&k8sClient)
	ns1 := SetupNamespace(&k8sClient)

	It("should peer networks in the same namespace referencing a single parent network", func(ctx SpecContext) {
		By("creating a network network-1")
		network1 := &networkingv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns.Name,
				Name:      "network-1",
			},
			Spec: networkingv1alpha1.NetworkSpec{
				Peerings: []networkingv1alpha1.NetworkPeering{
					{
						Name: "peering-1",
						NetworkRef: networkingv1alpha1.NetworkPeeringNetworkRef{
							Name:      "network-2",
							Namespace: ns.Name,
						},
					},
					{
						Name: "peering-2",
						NetworkRef: networkingv1alpha1.NetworkPeeringNetworkRef{
							Name: "network-3",
						},
					},
					{
						Name: "peering-3",
						NetworkRef: networkingv1alpha1.NetworkPeeringNetworkRef{
							Name:      "network-4",
							Namespace: ns.Name,
						},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, network1)).To(Succeed())

		By("creating a network network-2")
		network2 := &networkingv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns.Name,
				Name:      "network-2",
			},
			Spec: networkingv1alpha1.NetworkSpec{
				Peerings: []networkingv1alpha1.NetworkPeering{
					{
						Name: "peering-2",
						NetworkRef: networkingv1alpha1.NetworkPeeringNetworkRef{
							Name:      "network-1",
							Namespace: ns.Name,
						},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, network2)).To(Succeed())

		By("creating a network network-3")
		network3 := &networkingv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns.Name,
				Name:      "network-3",
			},
			Spec: networkingv1alpha1.NetworkSpec{
				Peerings: []networkingv1alpha1.NetworkPeering{
					{
						Name: "peering-3",
						NetworkRef: networkingv1alpha1.NetworkPeeringNetworkRef{
							Name:      "network-1",
							Namespace: ns.Name,
						},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, network3)).To(Succeed())

		By("creating a network network-4")
		network4 := &networkingv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns.Name,
				Name:      "network-4",
			},
			Spec: networkingv1alpha1.NetworkSpec{
				Peerings: []networkingv1alpha1.NetworkPeering{
					{
						Name: "peering-4",
						NetworkRef: networkingv1alpha1.NetworkPeeringNetworkRef{
							Name:      "network-1",
							Namespace: ns.Name,
						},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, network4)).To(Succeed())

		By("patching networks as available")
		baseNetwork1 := network1.DeepCopy()
		network1.Status.State = networkingv1alpha1.NetworkStateAvailable
		Expect(k8sClient.Status().Patch(ctx, network1, client.MergeFrom(baseNetwork1))).To(Succeed())

		baseNetwork2 := network2.DeepCopy()
		network2.Status.State = networkingv1alpha1.NetworkStateAvailable
		Expect(k8sClient.Status().Patch(ctx, network2, client.MergeFrom(baseNetwork2))).To(Succeed())

		baseNetwork3 := network3.DeepCopy()
		network3.Status.State = networkingv1alpha1.NetworkStateAvailable
		Expect(k8sClient.Status().Patch(ctx, network3, client.MergeFrom(baseNetwork3))).To(Succeed())

		baseNetwork4 := network4.DeepCopy()
		network4.Status.State = networkingv1alpha1.NetworkStateAvailable
		Expect(k8sClient.Status().Patch(ctx, network4, client.MergeFrom(baseNetwork4))).To(Succeed())

		By("waiting for networks to reference each other")
		Eventually(Object(network1)).
			Should(SatisfyAll(
				HaveField("Spec.PeeringClaimRefs", ConsistOf(networkingv1alpha1.NetworkPeeringClaimRef{
					Namespace: network2.Namespace,
					Name:      network2.Name,
					UID:       network2.UID,
				}, networkingv1alpha1.NetworkPeeringClaimRef{
					Namespace: network3.Namespace,
					Name:      network3.Name,
					UID:       network3.UID,
				}, networkingv1alpha1.NetworkPeeringClaimRef{
					Namespace: network4.Namespace,
					Name:      network4.Name,
					UID:       network4.UID,
				})),
				HaveField("Status.State", Equal(networkingv1alpha1.NetworkStateAvailable)),
				HaveField("Status.Peerings", ConsistOf(networkingv1alpha1.NetworkPeeringStatus{
					Name:  network1.Spec.Peerings[0].Name,
					State: networkingv1alpha1.NetworkPeeringStatePending,
				}, networkingv1alpha1.NetworkPeeringStatus{
					Name:  network1.Spec.Peerings[1].Name,
					State: networkingv1alpha1.NetworkPeeringStatePending,
				}, networkingv1alpha1.NetworkPeeringStatus{
					Name:  network1.Spec.Peerings[2].Name,
					State: networkingv1alpha1.NetworkPeeringStatePending,
				})),
			))

		Eventually(Object(network2)).
			Should(SatisfyAll(
				HaveField("Spec.PeeringClaimRefs", ConsistOf(networkingv1alpha1.NetworkPeeringClaimRef{
					Namespace: network1.Namespace,
					Name:      network1.Name,
					UID:       network1.UID,
				})),
				HaveField("Status.State", Equal(networkingv1alpha1.NetworkStateAvailable)),
				HaveField("Status.Peerings", ConsistOf(networkingv1alpha1.NetworkPeeringStatus{
					Name:  network2.Spec.Peerings[0].Name,
					State: networkingv1alpha1.NetworkPeeringStatePending,
				})),
			))

		Eventually(Object(network3)).
			Should(SatisfyAll(
				HaveField("Spec.PeeringClaimRefs", ConsistOf(networkingv1alpha1.NetworkPeeringClaimRef{
					Namespace: network1.Namespace,
					Name:      network1.Name,
					UID:       network1.UID,
				})),
				HaveField("Status.State", Equal(networkingv1alpha1.NetworkStateAvailable)),
				HaveField("Status.Peerings", ConsistOf(networkingv1alpha1.NetworkPeeringStatus{
					Name:  network3.Spec.Peerings[0].Name,
					State: networkingv1alpha1.NetworkPeeringStatePending,
				})),
			))

		Eventually(Object(network4)).
			Should(SatisfyAll(
				HaveField("Spec.PeeringClaimRefs", ConsistOf(networkingv1alpha1.NetworkPeeringClaimRef{
					Namespace: network1.Namespace,
					Name:      network1.Name,
					UID:       network1.UID,
				})),
				HaveField("Status.State", Equal(networkingv1alpha1.NetworkStateAvailable)),
				HaveField("Status.Peerings", ConsistOf(networkingv1alpha1.NetworkPeeringStatus{
					Name:  network4.Spec.Peerings[0].Name,
					State: networkingv1alpha1.NetworkPeeringStatePending,
				})),
			))

		By("deleting the networks")
		Expect(k8sClient.Delete(ctx, network1)).To(Succeed())
		Expect(k8sClient.Delete(ctx, network2)).To(Succeed())
		Expect(k8sClient.Delete(ctx, network3)).To(Succeed())
		Expect(k8sClient.Delete(ctx, network4)).To(Succeed())

		By("waiting for networks to be gone")
		Eventually(Get(network1)).Should(Satisfy(apierrors.IsNotFound))
		Eventually(Get(network2)).Should(Satisfy(apierrors.IsNotFound))
		Eventually(Get(network3)).Should(Satisfy(apierrors.IsNotFound))
		Eventually(Get(network4)).Should(Satisfy(apierrors.IsNotFound))
	})

	It("should peer two networks from different namespaces if they reference each other correctly", func(ctx SpecContext) {
		By("creating a network network-1")
		network1 := &networkingv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns.Name,
				Name:      "network-1",
			},
			Spec: networkingv1alpha1.NetworkSpec{
				Peerings: []networkingv1alpha1.NetworkPeering{
					{
						Name: "peering-1",
						NetworkRef: networkingv1alpha1.NetworkPeeringNetworkRef{
							Name:      "network-2",
							Namespace: ns1.Name,
						},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, network1)).To(Succeed())

		By("creating a network network-2")
		network2 := &networkingv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns1.Name,
				Name:      "network-2",
			},
			Spec: networkingv1alpha1.NetworkSpec{
				Peerings: []networkingv1alpha1.NetworkPeering{
					{
						Name: "peering-2",
						NetworkRef: networkingv1alpha1.NetworkPeeringNetworkRef{
							Name:      "network-1",
							Namespace: ns.Name,
						},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, network2)).To(Succeed())

		By("patching networks as available")
		baseNetwork1 := network1.DeepCopy()
		network1.Status.State = networkingv1alpha1.NetworkStateAvailable
		Expect(k8sClient.Status().Patch(ctx, network1, client.MergeFrom(baseNetwork1))).To(Succeed())

		baseNetwork2 := network2.DeepCopy()
		network2.Status.State = networkingv1alpha1.NetworkStateAvailable
		Expect(k8sClient.Status().Patch(ctx, network2, client.MergeFrom(baseNetwork2))).To(Succeed())

		By("waiting for networks to reference each other")
		Eventually(Object(network1)).
			Should(SatisfyAll(
				HaveField("Spec.PeeringClaimRefs", ConsistOf(networkingv1alpha1.NetworkPeeringClaimRef{
					Namespace: network2.Namespace,
					Name:      network2.Name,
					UID:       network2.UID,
				})),
				HaveField("Status.State", Equal(networkingv1alpha1.NetworkStateAvailable)),
				HaveField("Status.Peerings", ConsistOf(networkingv1alpha1.NetworkPeeringStatus{
					Name:  network1.Spec.Peerings[0].Name,
					State: networkingv1alpha1.NetworkPeeringStatePending,
				})),
			))

		Eventually(Object(network2)).
			Should(SatisfyAll(
				HaveField("Spec.PeeringClaimRefs", ConsistOf(networkingv1alpha1.NetworkPeeringClaimRef{
					Namespace: network1.Namespace,
					Name:      network1.Name,
					UID:       network1.UID,
				})),
				HaveField("Status.State", Equal(networkingv1alpha1.NetworkStateAvailable)),
				HaveField("Status.Peerings", ConsistOf(networkingv1alpha1.NetworkPeeringStatus{
					Name:  network2.Spec.Peerings[0].Name,
					State: networkingv1alpha1.NetworkPeeringStatePending,
				})),
			))

		By("deleting the networks")
		Expect(k8sClient.Delete(ctx, network1)).To(Succeed())
		Expect(k8sClient.Delete(ctx, network2)).To(Succeed())

		By("waiting for networks to be gone")
		Eventually(Get(network1)).Should(Satisfy(apierrors.IsNotFound))
		Eventually(Get(network2)).Should(Satisfy(apierrors.IsNotFound))
	})

	It("should not peer two networks if they dont exactly reference each other", func(ctx SpecContext) {
		By("creating a network network-1")
		network1 := &networkingv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns.Name,
				Name:      "network-1",
			},
			Spec: networkingv1alpha1.NetworkSpec{
				Peerings: []networkingv1alpha1.NetworkPeering{
					{
						Name: "peering-1",
						NetworkRef: networkingv1alpha1.NetworkPeeringNetworkRef{
							Name:      "network-2",
							Namespace: ns1.Name,
						},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, network1)).To(Succeed())

		By("creating a network network-2")
		network2 := &networkingv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns.Name,
				Name:      "network-2",
			},
			Spec: networkingv1alpha1.NetworkSpec{
				Peerings: []networkingv1alpha1.NetworkPeering{
					{
						Name: "peering-2",
						NetworkRef: networkingv1alpha1.NetworkPeeringNetworkRef{
							Name:      "network-other",
							Namespace: ns.Name,
						},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, network2)).To(Succeed())

		By("patching networks as available")
		baseNetwork1 := network1.DeepCopy()
		network1.Status.State = networkingv1alpha1.NetworkStateAvailable
		Expect(k8sClient.Status().Patch(ctx, network1, client.MergeFrom(baseNetwork1))).To(Succeed())

		baseNetwork2 := network2.DeepCopy()
		network2.Status.State = networkingv1alpha1.NetworkStateAvailable
		Expect(k8sClient.Status().Patch(ctx, network2, client.MergeFrom(baseNetwork2))).To(Succeed())

		By("ensuring both networks do not get peered")
		Eventually(Object(network1)).
			Should(SatisfyAll(
				HaveField("Spec.PeeringClaimRefs", BeEmpty()),
				HaveField("Status.State", Equal(networkingv1alpha1.NetworkStateAvailable)),
				HaveField("Status.Peerings", ConsistOf(networkingv1alpha1.NetworkPeeringStatus{
					Name:  network1.Spec.Peerings[0].Name,
					State: networkingv1alpha1.NetworkPeeringStatePending,
				})),
			))

		Eventually(Object(network2)).
			Should(SatisfyAll(
				HaveField("Spec.PeeringClaimRefs", BeEmpty()),
				HaveField("Status.State", Equal(networkingv1alpha1.NetworkStateAvailable)),
				HaveField("Status.Peerings", ConsistOf(networkingv1alpha1.NetworkPeeringStatus{
					Name:  network2.Spec.Peerings[0].Name,
					State: networkingv1alpha1.NetworkPeeringStatePending,
				})),
			))

		By("deleting the networks")
		Expect(k8sClient.Delete(ctx, network1)).To(Succeed())
		Expect(k8sClient.Delete(ctx, network2)).To(Succeed())

		By("waiting for networks to be gone")
		Eventually(Get(network1)).Should(Satisfy(apierrors.IsNotFound))
		Eventually(Get(network2)).Should(Satisfy(apierrors.IsNotFound))
	})

	It("should peer networks in the same namespace referencing each other", func(ctx SpecContext) {
		By("creating a network network-1")
		network1 := &networkingv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns.Name,
				Name:      "network-1",
			},
			Spec: networkingv1alpha1.NetworkSpec{
				Peerings: []networkingv1alpha1.NetworkPeering{
					{
						Name: "peering-1",
						NetworkRef: networkingv1alpha1.NetworkPeeringNetworkRef{
							Name:      "network-2",
							Namespace: ns.Name,
						},
					},
					{
						Name: "peering-2",
						NetworkRef: networkingv1alpha1.NetworkPeeringNetworkRef{
							Name:      "network-3",
							Namespace: ns.Name,
						},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, network1)).To(Succeed())

		By("creating a network network-2")
		network2 := &networkingv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns.Name,
				Name:      "network-2",
			},
			Spec: networkingv1alpha1.NetworkSpec{
				Peerings: []networkingv1alpha1.NetworkPeering{
					{
						Name: "peering-1",
						NetworkRef: networkingv1alpha1.NetworkPeeringNetworkRef{
							Name:      "network-1",
							Namespace: ns.Name,
						},
					},
					{
						Name: "peering-2",
						NetworkRef: networkingv1alpha1.NetworkPeeringNetworkRef{
							Name:      "network-3",
							Namespace: ns.Name,
						},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, network2)).To(Succeed())

		By("creating a network network-3")
		network3 := &networkingv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns.Name,
				Name:      "network-3",
			},
			Spec: networkingv1alpha1.NetworkSpec{
				Peerings: []networkingv1alpha1.NetworkPeering{
					{
						Name: "peering-1",
						NetworkRef: networkingv1alpha1.NetworkPeeringNetworkRef{
							Name:      "network-1",
							Namespace: ns.Name,
						},
					},
					{
						Name: "peering-2",
						NetworkRef: networkingv1alpha1.NetworkPeeringNetworkRef{
							Name:      "network-2",
							Namespace: ns.Name,
						},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, network3)).To(Succeed())

		By("patching networks as available")
		baseNetwork1 := network1.DeepCopy()
		network1.Status.State = networkingv1alpha1.NetworkStateAvailable
		Expect(k8sClient.Status().Patch(ctx, network1, client.MergeFrom(baseNetwork1))).To(Succeed())

		baseNetwork2 := network2.DeepCopy()
		network2.Status.State = networkingv1alpha1.NetworkStateAvailable
		Expect(k8sClient.Status().Patch(ctx, network2, client.MergeFrom(baseNetwork2))).To(Succeed())

		baseNetwork3 := network3.DeepCopy()
		network3.Status.State = networkingv1alpha1.NetworkStateAvailable
		Expect(k8sClient.Status().Patch(ctx, network3, client.MergeFrom(baseNetwork3))).To(Succeed())

		By("waiting for networks to reference each other")
		Eventually(Object(network1)).
			Should(SatisfyAll(
				HaveField("Spec.PeeringClaimRefs", ConsistOf(networkingv1alpha1.NetworkPeeringClaimRef{
					Namespace: network2.Namespace,
					Name:      network2.Name,
					UID:       network2.UID,
				}, networkingv1alpha1.NetworkPeeringClaimRef{
					Namespace: network3.Namespace,
					Name:      network3.Name,
					UID:       network3.UID,
				})),
				HaveField("Status.State", Equal(networkingv1alpha1.NetworkStateAvailable)),
				HaveField("Status.Peerings", ConsistOf(networkingv1alpha1.NetworkPeeringStatus{
					Name:  network1.Spec.Peerings[0].Name,
					State: networkingv1alpha1.NetworkPeeringStatePending,
				}, networkingv1alpha1.NetworkPeeringStatus{
					Name:  network1.Spec.Peerings[1].Name,
					State: networkingv1alpha1.NetworkPeeringStatePending,
				})),
			))

		Eventually(Object(network2)).
			Should(SatisfyAll(
				HaveField("Spec.PeeringClaimRefs", ConsistOf(networkingv1alpha1.NetworkPeeringClaimRef{
					Namespace: network1.Namespace,
					Name:      network1.Name,
					UID:       network1.UID,
				}, networkingv1alpha1.NetworkPeeringClaimRef{
					Namespace: network3.Namespace,
					Name:      network3.Name,
					UID:       network3.UID,
				})),
				HaveField("Status.State", Equal(networkingv1alpha1.NetworkStateAvailable)),
				HaveField("Status.Peerings", ConsistOf(networkingv1alpha1.NetworkPeeringStatus{
					Name:  network2.Spec.Peerings[0].Name,
					State: networkingv1alpha1.NetworkPeeringStatePending,
				}, networkingv1alpha1.NetworkPeeringStatus{
					Name:  network2.Spec.Peerings[1].Name,
					State: networkingv1alpha1.NetworkPeeringStatePending,
				})),
			))

		Eventually(Object(network3)).
			Should(SatisfyAll(
				HaveField("Spec.PeeringClaimRefs", ConsistOf(networkingv1alpha1.NetworkPeeringClaimRef{
					Namespace: network1.Namespace,
					Name:      network1.Name,
					UID:       network1.UID,
				}, networkingv1alpha1.NetworkPeeringClaimRef{
					Namespace: network2.Namespace,
					Name:      network2.Name,
					UID:       network2.UID,
				})),
				HaveField("Status.State", Equal(networkingv1alpha1.NetworkStateAvailable)),
				HaveField("Status.Peerings", ConsistOf(networkingv1alpha1.NetworkPeeringStatus{
					Name:  network3.Spec.Peerings[0].Name,
					State: networkingv1alpha1.NetworkPeeringStatePending,
				}, networkingv1alpha1.NetworkPeeringStatus{
					Name:  network3.Spec.Peerings[1].Name,
					State: networkingv1alpha1.NetworkPeeringStatePending,
				})),
			))

		By("deleting the networks")
		Expect(k8sClient.Delete(ctx, network1)).To(Succeed())
		Expect(k8sClient.Delete(ctx, network2)).To(Succeed())
		Expect(k8sClient.Delete(ctx, network3)).To(Succeed())

		By("waiting for networks to be gone")
		Eventually(Get(network1)).Should(Satisfy(apierrors.IsNotFound))
		Eventually(Get(network2)).Should(Satisfy(apierrors.IsNotFound))
		Eventually(Get(network3)).Should(Satisfy(apierrors.IsNotFound))
	})
})
