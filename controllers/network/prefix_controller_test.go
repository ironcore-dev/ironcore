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
package network

import (
	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	networkv1alpha1 "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("PrefixReconciler", func() {
	ns := SetupTest()

	It("should mark root prefixes as ready", func() {
		By("creating a root prefix")
		prefixValue := commonv1alpha1.MustParseIPPrefix("10.0.0.0/24")
		prefix := &networkv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-root-",
			},
			Spec: networkv1alpha1.PrefixSpec{
				PrefixSpace: networkv1alpha1.PrefixSpace{
					Prefix: prefixValue,
				},
			},
		}
		Expect(k8sClient.Create(ctx, prefix)).To(Succeed())

		By("asserting there is no allocation for root prefixes")
		Consistently(func(g Gomega) {
			list := &networkv1alpha1.PrefixAllocationList{}
			g.Expect(ListOwned(ctx, k8sClient, k8sClient.Scheme(), prefix, list, client.InNamespace(ns.Name))).To(Succeed())
			g.Expect(list.Items).To(BeEmpty())
		}, timeout, interval).Should(Succeed())

		By("waiting for the prefix to be marked as ready and report its available ranges")
		prefixKey := client.ObjectKeyFromObject(prefix)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, prefixKey, prefix)).To(Succeed())
			g.Expect(IsPrefixReady(prefix)).To(BeTrue(), "prefix is not ready: %#v", prefix)
			g.Expect(prefix.Status.Available).To(Equal([]commonv1alpha1.IPPrefix{prefixValue}))
		}, timeout, interval).Should(Succeed())
	})

	It("should allocate child prefixes", func() {
		By("creating a root prefix")
		prefixValue := commonv1alpha1.MustParseIPPrefix("10.0.0.0/24")
		rootPrefix := &networkv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-root-",
			},
			Spec: networkv1alpha1.PrefixSpec{
				PrefixSpace: networkv1alpha1.PrefixSpace{
					Prefix: prefixValue,
				},
			},
		}
		Expect(k8sClient.Create(ctx, rootPrefix)).To(Succeed())

		By("creating a child prefix")
		childPrefix := &networkv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-child-",
			},
			Spec: networkv1alpha1.PrefixSpec{
				ParentRef: &networkv1alpha1.PrefixReference{
					Kind: networkv1alpha1.PrefixKind,
					Name: rootPrefix.Name,
				},
				PrefixSpace: networkv1alpha1.PrefixSpace{
					PrefixLength: 28,
				},
			},
		}
		Expect(k8sClient.Create(ctx, childPrefix)).To(Succeed())

		By("waiting for the child prefix to be marked as ready and report its available ranges")
		expectedChildPrefix := commonv1alpha1.MustParseIPPrefix("10.0.0.0/28")
		childPrefixKey := client.ObjectKeyFromObject(childPrefix)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, childPrefixKey, childPrefix)).To(Succeed())
			g.Expect(IsPrefixReady(childPrefix)).To(BeTrue(), "child prefix is not ready: %#v", childPrefix)
			g.Expect(childPrefix.Spec.ParentRef).To(Equal(&networkv1alpha1.PrefixReference{
				Kind: networkv1alpha1.PrefixKind,
				Name: rootPrefix.Name,
			}))
			g.Expect(childPrefix.Status.Available).To(Equal([]commonv1alpha1.IPPrefix{expectedChildPrefix}))
		}, timeout, interval).Should(Succeed())

		By("asserting the parent's available ranges have been updated")
		rootPrefixKey := client.ObjectKeyFromObject(rootPrefix)
		Expect(k8sClient.Get(ctx, rootPrefixKey, rootPrefix)).To(Succeed())
		Expect(rootPrefix.Status.Available).To(ConsistOf(
			commonv1alpha1.MustParseIPPrefix("10.0.0.16/28"),
			commonv1alpha1.MustParseIPPrefix("10.0.0.32/27"),
			commonv1alpha1.MustParseIPPrefix("10.0.0.64/26"),
			commonv1alpha1.MustParseIPPrefix("10.0.0.128/25"),
		))

		By("asserting there is a single valid allocation")
		list := &networkv1alpha1.PrefixAllocationList{}
		Expect(ListOwned(ctx, k8sClient, k8sClient.Scheme(), childPrefix, list, client.InNamespace(ns.Name))).To(Succeed())
		Expect(list.Items).To(HaveLen(1))
		allocation := &list.Items[0]
		Expect(allocation.Spec).To(Equal(networkv1alpha1.PrefixAllocationSpec{
			PrefixRef: &networkv1alpha1.PrefixReference{
				Kind: networkv1alpha1.PrefixKind,
				Name: rootPrefix.Name,
			},
			PrefixAllocationRequest: networkv1alpha1.PrefixAllocationRequest{
				PrefixLength: 28,
			},
		}))
		Expect(IsPrefixAllocationSucceeded(allocation)).To(BeTrue(), "allocation is not ready: %#v", allocation)
		Expect(allocation.Status.Prefix).To(Equal(expectedChildPrefix))
	})

	It("should leave prefixes in pending state when they can't be allocated", func() {
		By("creating a root prefix")
		prefixValue := commonv1alpha1.MustParseIPPrefix("10.0.0.0/24")
		rootPrefix := &networkv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-root-",
			},
			Spec: networkv1alpha1.PrefixSpec{
				PrefixSpace: networkv1alpha1.PrefixSpace{
					Prefix: prefixValue,
				},
			},
		}
		Expect(k8sClient.Create(ctx, rootPrefix)).To(Succeed())

		By("creating a child prefix")
		childPrefix := &networkv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-child-",
			},
			Spec: networkv1alpha1.PrefixSpec{
				ParentRef: &networkv1alpha1.PrefixReference{
					Kind: networkv1alpha1.PrefixKind,
					Name: rootPrefix.Name,
				},
				PrefixSpace: networkv1alpha1.PrefixSpace{
					PrefixLength: 8,
				},
			},
		}
		Expect(k8sClient.Create(ctx, childPrefix)).To(Succeed())

		By("asserting there is a single, non-succeeded allocation and the child does not become ready")
		childPrefixKey := client.ObjectKeyFromObject(childPrefix)
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, childPrefixKey, childPrefix)).Should(Succeed())
			g.Expect(IsPrefixReady(childPrefix)).Should(BeFalse(), "child prefix became ready: %#v", childPrefix)

			list := &networkv1alpha1.PrefixAllocationList{}
			g.Expect(ListOwned(ctx, k8sClient, k8sClient.Scheme(), childPrefix, list, client.InNamespace(ns.Name))).To(Succeed())
			g.Expect(list.Items).To(HaveLen(1))
			allocation := list.Items[0]
			g.Expect(IsPrefixAllocationSucceeded(&allocation)).To(BeFalse(), "prefix allocation is succeeded: %#v", allocation)
		}, timeout, interval).Should(Succeed())
	})

	It("should assign a prefix on matching parents", func() {
		By("creating a child prefix")
		childPrefix := &networkv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-child-",
			},
			Spec: networkv1alpha1.PrefixSpec{
				PrefixSpace: networkv1alpha1.PrefixSpace{
					PrefixLength: 28,
				},
				ParentSelector: &networkv1alpha1.PrefixSelector{
					Kind: networkv1alpha1.PrefixKind,
					LabelSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"foo": "bar",
						},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, childPrefix)).To(Succeed())

		By("checking that the child prefix is not being assigned")
		childPrefixKey := client.ObjectKeyFromObject(childPrefix)
		Consistently(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, childPrefixKey, childPrefix)).Should(Succeed())
			g.Expect(IsPrefixReady(childPrefix)).To(BeFalse(), "child prefix is ready: %#v", childPrefix)
			g.Expect(childPrefix.Spec.ParentRef).To(BeNil())
		}, timeout, interval).Should(Succeed())

		By("creating a root prefix that would fit but does not match")
		prefixValue := commonv1alpha1.MustParseIPPrefix("10.0.0.0/24")
		notMatchingRootPrefix := &networkv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-root-",
			},
			Spec: networkv1alpha1.PrefixSpec{
				PrefixSpace: networkv1alpha1.PrefixSpace{
					Prefix: prefixValue,
				},
			},
		}
		Expect(k8sClient.Create(ctx, notMatchingRootPrefix)).To(Succeed())

		By("checking that the child prefix is not being assigned")
		Consistently(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, childPrefixKey, childPrefix)).Should(Succeed())
			g.Expect(IsPrefixReady(childPrefix)).To(BeFalse(), "child prefix is ready: %#v", childPrefix)
			g.Expect(childPrefix.Spec.ParentRef).To(BeNil())
		}, timeout, interval).Should(Succeed())

		By("creating a root prefix that fits and matches")
		rootPrefix := &networkv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-root-",
				Labels: map[string]string{
					"foo": "bar",
				},
			},
			Spec: networkv1alpha1.PrefixSpec{
				PrefixSpace: networkv1alpha1.PrefixSpace{
					Prefix: prefixValue,
				},
			},
		}
		Expect(k8sClient.Create(ctx, rootPrefix)).To(Succeed())

		By("waiting for the child prefix to be marked as ready and report its available ranges")
		expectedChildPrefix := commonv1alpha1.MustParseIPPrefix("10.0.0.0/28")
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, childPrefixKey, childPrefix)).To(Succeed())
			g.Expect(IsPrefixReady(childPrefix)).To(BeTrue(), "child prefix is not ready: %#v", childPrefix)
			g.Expect(childPrefix.Spec.ParentRef).To(Equal(&networkv1alpha1.PrefixReference{
				Kind: networkv1alpha1.PrefixKind,
				Name: rootPrefix.Name,
			}))
			g.Expect(childPrefix.Status.Available).To(Equal([]commonv1alpha1.IPPrefix{expectedChildPrefix}))
		}, timeout, interval).Should(Succeed())

		By("asserting the parent's available ranges have been updated")
		rootPrefixKey := client.ObjectKeyFromObject(rootPrefix)
		Expect(k8sClient.Get(ctx, rootPrefixKey, rootPrefix)).To(Succeed())
		Expect(rootPrefix.Status.Available).To(ConsistOf(
			commonv1alpha1.MustParseIPPrefix("10.0.0.16/28"),
			commonv1alpha1.MustParseIPPrefix("10.0.0.32/27"),
			commonv1alpha1.MustParseIPPrefix("10.0.0.64/26"),
			commonv1alpha1.MustParseIPPrefix("10.0.0.128/25"),
		))

		By("asserting there is a single valid allocation")
		list := &networkv1alpha1.PrefixAllocationList{}
		Expect(ListOwned(ctx, k8sClient, k8sClient.Scheme(), childPrefix, list, client.InNamespace(ns.Name))).To(Succeed())
		Expect(list.Items).To(HaveLen(1))
		allocation := &list.Items[0]
		Expect(allocation.Spec).To(Equal(networkv1alpha1.PrefixAllocationSpec{
			PrefixRef: &networkv1alpha1.PrefixReference{
				Kind: networkv1alpha1.PrefixKind,
				Name: rootPrefix.Name,
			},
			PrefixSelector: &networkv1alpha1.PrefixSelector{
				Kind: networkv1alpha1.PrefixKind,
				LabelSelector: metav1.LabelSelector{
					MatchLabels: map[string]string{
						"foo": "bar",
					},
				},
			},
			PrefixAllocationRequest: networkv1alpha1.PrefixAllocationRequest{
				PrefixLength: 28,
			},
		}))
		Expect(IsPrefixAllocationSucceeded(allocation)).To(BeTrue(), "allocation is not ready: %#v", allocation)
		Expect(allocation.Status.Prefix).To(Equal(expectedChildPrefix))
	})

	It("should not distribute reserved prefixes", func() {
		By("creating a root prefix with reservations")
		prefixValue := commonv1alpha1.MustParseIPPrefix("10.0.0.0/8")
		rootPrefix := &networkv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-root-",
			},
			Spec: networkv1alpha1.PrefixSpec{
				PrefixSpace: networkv1alpha1.PrefixSpace{
					Prefix: prefixValue,
					Reservations: []commonv1alpha1.IPPrefix{
						commonv1alpha1.MustParseIPPrefix("10.0.0.0/9"),
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, rootPrefix)).To(Succeed())

		By("waiting for the prefix to be marked as ready and report its available ranges")
		rootPrefixKey := client.ObjectKeyFromObject(rootPrefix)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, rootPrefixKey, rootPrefix)).To(Succeed())
			g.Expect(IsPrefixReady(rootPrefix)).To(BeTrue(), "prefix is not ready: %#v", rootPrefix)
			g.Expect(rootPrefix.Status.Reserved).To(Equal([]commonv1alpha1.IPPrefix{
				commonv1alpha1.MustParseIPPrefix("10.0.0.0/9"),
			}))
			g.Expect(rootPrefix.Status.Available).To(Equal([]commonv1alpha1.IPPrefix{
				commonv1alpha1.MustParseIPPrefix("10.128.0.0/9"),
			}))
		}, timeout, interval).Should(Succeed())
	})

	It("should create dynamic reservations for sub-prefixes upon assignment", func() {
		By("creating a root prefix with reservations")
		prefixValue := commonv1alpha1.MustParseIPPrefix("10.0.0.0/8")
		rootPrefix := &networkv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-root-",
			},
			Spec: networkv1alpha1.PrefixSpec{
				PrefixSpace: networkv1alpha1.PrefixSpace{
					Prefix: prefixValue,
				},
			},
		}
		Expect(k8sClient.Create(ctx, rootPrefix)).To(Succeed())

		By("creating a child prefix with dynamic reservations")
		childPrefix := &networkv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-child-",
			},
			Spec: networkv1alpha1.PrefixSpec{
				ParentRef: &networkv1alpha1.PrefixReference{
					Kind: networkv1alpha1.PrefixKind,
					Name: rootPrefix.Name,
				},
				PrefixSpace: networkv1alpha1.PrefixSpace{
					PrefixLength:       9,
					ReservationLengths: []int8{10},
				},
			},
		}
		Expect(k8sClient.Create(ctx, childPrefix)).To(Succeed())

		By("waiting for the child prefix to be allocated and report its available and reserved prefixes")
		childPrefixKey := client.ObjectKeyFromObject(childPrefix)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, childPrefixKey, childPrefix)).To(Succeed())
			g.Expect(IsPrefixReady(childPrefix)).To(BeTrue(), "child prefix is not ready: %#v", childPrefix)
			g.Expect(childPrefix.Spec.Prefix).To(Equal(commonv1alpha1.MustParseIPPrefix("10.0.0.0/9")))
			g.Expect(childPrefix.Spec.Reservations).To(ConsistOf(commonv1alpha1.MustParseIPPrefix("10.0.0.0/10")))
			g.Expect(childPrefix.Status.Reserved).To(Equal([]commonv1alpha1.IPPrefix{
				commonv1alpha1.MustParseIPPrefix("10.0.0.0/10"),
			}))
			g.Expect(childPrefix.Status.Available).To(Equal([]commonv1alpha1.IPPrefix{
				commonv1alpha1.MustParseIPPrefix("10.64.0.0/10"),
			}))
		}, timeout, interval).Should(Succeed())
	})

	It("should allocate prefixes from cluster prefixes by reference", func() {
		By("creating a root cluster prefix")
		clusterPrefix := &networkv1alpha1.ClusterPrefix{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "prefix-test-",
			},
			Spec: networkv1alpha1.ClusterPrefixSpec{
				PrefixSpace: networkv1alpha1.PrefixSpace{
					Prefix: commonv1alpha1.MustParseIPPrefix("10.0.0.0/8"),
				},
			},
		}
		Expect(k8sClient.Create(ctx, clusterPrefix)).To(Succeed())

		By("creating a prefix referencing the cluster prefix")
		prefix := &networkv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-",
			},
			Spec: networkv1alpha1.PrefixSpec{
				ParentRef: &networkv1alpha1.PrefixReference{
					Kind: networkv1alpha1.ClusterPrefixKind,
					Name: clusterPrefix.Name,
				},
				PrefixSpace: networkv1alpha1.PrefixSpace{
					Prefix: commonv1alpha1.MustParseIPPrefix("10.0.0.0/9"),
				},
			},
		}
		Expect(k8sClient.Create(ctx, prefix)).To(Succeed())

		By("waiting for the prefix to be allocated")
		clusterPrefixKey := client.ObjectKeyFromObject(clusterPrefix)
		prefixKey := client.ObjectKeyFromObject(prefix)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, prefixKey, prefix)).To(Succeed())
			g.Expect(IsPrefixReady(prefix)).To(BeTrue(), "prefix is not ready: %#v", prefix)
			g.Expect(prefix.Status.Available).To(ConsistOf(commonv1alpha1.MustParseIPPrefix("10.0.0.0/9")))

			Expect(k8sClient.Get(ctx, clusterPrefixKey, clusterPrefix)).To(Succeed())
			g.Expect(clusterPrefix.Status.Available).To(ConsistOf(commonv1alpha1.MustParseIPPrefix("10.128.0.0/9")))
		}, timeout, interval).Should(Succeed())
	})

	It("should allocate prefixes from cluster prefixes by selector", func() {
		By("creating a root cluster prefix")
		clusterPrefix := &networkv1alpha1.ClusterPrefix{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "prefix-test-",
				Labels: map[string]string{
					"foo": "bar",
				},
			},
			Spec: networkv1alpha1.ClusterPrefixSpec{
				PrefixSpace: networkv1alpha1.PrefixSpace{
					Prefix: commonv1alpha1.MustParseIPPrefix("10.0.0.0/8"),
				},
			},
		}
		Expect(k8sClient.Create(ctx, clusterPrefix)).To(Succeed())

		By("creating a prefix referencing the cluster prefix")
		prefix := &networkv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-",
			},
			Spec: networkv1alpha1.PrefixSpec{
				ParentSelector: &networkv1alpha1.PrefixSelector{
					Kind: networkv1alpha1.ClusterPrefixKind,
					LabelSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"foo": "bar",
						},
					},
				},
				PrefixSpace: networkv1alpha1.PrefixSpace{
					Prefix: commonv1alpha1.MustParseIPPrefix("10.0.0.0/9"),
				},
			},
		}
		Expect(k8sClient.Create(ctx, prefix)).To(Succeed())

		By("waiting for the prefix to be allocated")
		clusterPrefixKey := client.ObjectKeyFromObject(clusterPrefix)
		prefixKey := client.ObjectKeyFromObject(prefix)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, prefixKey, prefix)).To(Succeed())
			g.Expect(IsPrefixReady(prefix)).To(BeTrue(), "prefix is not ready: %#v", prefix)
			g.Expect(prefix.Status.Available).To(ConsistOf(commonv1alpha1.MustParseIPPrefix("10.0.0.0/9")))

			Expect(k8sClient.Get(ctx, clusterPrefixKey, clusterPrefix)).To(Succeed())
			g.Expect(clusterPrefix.Status.Available).To(ConsistOf(commonv1alpha1.MustParseIPPrefix("10.128.0.0/9")))
		}, timeout, interval).Should(Succeed())
	})

	It("should dynamically allocate prefixes from cluster prefixes by reference", func() {
		By("creating a root cluster prefix")
		clusterPrefix := &networkv1alpha1.ClusterPrefix{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "prefix-test-",
			},
			Spec: networkv1alpha1.ClusterPrefixSpec{
				PrefixSpace: networkv1alpha1.PrefixSpace{
					Prefix: commonv1alpha1.MustParseIPPrefix("10.0.0.0/8"),
				},
			},
		}
		Expect(k8sClient.Create(ctx, clusterPrefix)).To(Succeed())

		By("creating a prefix referencing the cluster prefix")
		prefix := &networkv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-",
			},
			Spec: networkv1alpha1.PrefixSpec{
				ParentRef: &networkv1alpha1.PrefixReference{
					Kind: networkv1alpha1.ClusterPrefixKind,
					Name: clusterPrefix.Name,
				},
				PrefixSpace: networkv1alpha1.PrefixSpace{
					PrefixLength: 9,
				},
			},
		}
		Expect(k8sClient.Create(ctx, prefix)).To(Succeed())

		By("waiting for the prefix to be allocated")
		clusterPrefixKey := client.ObjectKeyFromObject(clusterPrefix)
		prefixKey := client.ObjectKeyFromObject(prefix)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, prefixKey, prefix)).To(Succeed())
			g.Expect(IsPrefixReady(prefix)).To(BeTrue(), "prefix is not ready: %#v", prefix)
			g.Expect(prefix.Status.Available).To(ConsistOf(commonv1alpha1.MustParseIPPrefix("10.0.0.0/9")))

			Expect(k8sClient.Get(ctx, clusterPrefixKey, clusterPrefix)).To(Succeed())
			g.Expect(clusterPrefix.Status.Available).To(ConsistOf(commonv1alpha1.MustParseIPPrefix("10.128.0.0/9")))
		}, timeout, interval).Should(Succeed())
	})
})
