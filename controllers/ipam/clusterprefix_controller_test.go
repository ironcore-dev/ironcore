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
package ipam

import (
	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	ipamv1alpha1 "github.com/onmetal/onmetal-api/apis/ipam/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"inet.af/netaddr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("ClusterPrefixReconciler", func() {
	It("should mark root prefixes as ready", func() {
		By("creating a root prefix")
		prefixValue := commonv1alpha1.MustParseIPPrefix("10.0.0.0/24")
		prefix := &ipamv1alpha1.ClusterPrefix{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-root-",
			},
			Spec: ipamv1alpha1.ClusterPrefixSpec{
				PrefixSpace: ipamv1alpha1.PrefixSpace{
					Prefix: prefixValue,
				},
			},
		}
		Expect(k8sClient.Create(ctx, prefix)).To(Succeed())

		By("asserting there is no allocation for root prefixes")
		Consistently(func(g Gomega) {
			list := &ipamv1alpha1.ClusterPrefixAllocationList{}
			g.Expect(ListOwned(ctx, k8sClient, k8sClient.Scheme(), prefix, list)).To(Succeed())
			g.Expect(list.Items).To(BeEmpty())
		}, timeout, interval).Should(Succeed())

		By("waiting for the prefix to be marked as ready and report its available ranges")
		prefixKey := client.ObjectKeyFromObject(prefix)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, prefixKey, prefix)).To(Succeed())
			g.Expect(IsClusterPrefixReady(prefix)).To(BeTrue(), "prefix is not ready: %#v", prefix)
			g.Expect(prefix.Status.Available).To(Equal([]commonv1alpha1.IPPrefix{prefixValue}))
		}, timeout, interval).Should(Succeed())
	})

	PIt("should allocate child prefixes", func() {
		By("creating a root prefix")
		prefixValue := commonv1alpha1.MustParseIPPrefix("10.0.0.0/24")
		rootPrefix := &ipamv1alpha1.ClusterPrefix{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-root-",
			},
			Spec: ipamv1alpha1.ClusterPrefixSpec{
				PrefixSpace: ipamv1alpha1.PrefixSpace{
					Prefix: prefixValue,
				},
			},
		}
		Expect(k8sClient.Create(ctx, rootPrefix)).To(Succeed())

		By("creating a child prefix")
		childPrefix := &ipamv1alpha1.ClusterPrefix{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-child-",
			},
			Spec: ipamv1alpha1.ClusterPrefixSpec{
				ParentRef: &corev1.LocalObjectReference{
					Name: rootPrefix.Name,
				},
				PrefixSpace: ipamv1alpha1.PrefixSpace{
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
			g.Expect(IsClusterPrefixReady(childPrefix)).To(BeTrue(), "child prefix is not ready: %#v", childPrefix)
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
		list := &ipamv1alpha1.ClusterPrefixAllocationList{}
		Expect(ListOwned(ctx, k8sClient, k8sClient.Scheme(), childPrefix, list)).To(Succeed())
		Expect(list.Items).To(HaveLen(1))
		allocation := &list.Items[0]
		Expect(allocation.Spec.PrefixRef).To(Equal(&corev1.LocalObjectReference{Name: rootPrefix.Name}))
		Expect(allocation.Spec.ClusterPrefixAllocationRequest.PrefixLength).To(Equal(int32(28))) // TODO: Also assert on prefix being empty
		Expect(IsClusterPrefixAllocationSucceeded(allocation)).To(BeTrue(), "allocation is not ready: %#v", allocation)
		Expect(allocation.Status.Prefix).To(Equal(expectedChildPrefix))
	})

	It("should leave prefixes in pending state when they can't be allocated", func() {
		By("creating a root prefix")
		prefixValue := commonv1alpha1.MustParseIPPrefix("10.0.0.0/24")
		rootPrefix := &ipamv1alpha1.ClusterPrefix{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-root-",
			},
			Spec: ipamv1alpha1.ClusterPrefixSpec{
				PrefixSpace: ipamv1alpha1.PrefixSpace{
					Prefix: prefixValue,
				},
			},
		}
		Expect(k8sClient.Create(ctx, rootPrefix)).To(Succeed())

		By("creating a child prefix")
		childPrefix := &ipamv1alpha1.ClusterPrefix{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-child-",
			},
			Spec: ipamv1alpha1.ClusterPrefixSpec{
				ParentRef: &corev1.LocalObjectReference{
					Name: rootPrefix.Name,
				},
				PrefixSpace: ipamv1alpha1.PrefixSpace{
					PrefixLength: 8,
				},
			},
		}
		Expect(k8sClient.Create(ctx, childPrefix)).To(Succeed())

		By("asserting there is a single, non-succeeded allocation and the child does not become ready")
		childPrefixKey := client.ObjectKeyFromObject(childPrefix)
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, childPrefixKey, childPrefix)).Should(Succeed())
			g.Expect(IsClusterPrefixReady(childPrefix)).Should(BeFalse(), "child prefix is ready: %#v", childPrefix)

			list := &ipamv1alpha1.ClusterPrefixAllocationList{}
			g.Expect(ListOwned(ctx, k8sClient, k8sClient.Scheme(), childPrefix, list)).To(Succeed())
			g.Expect(list.Items).To(HaveLen(1))
			allocation := list.Items[0]
			g.Expect(IsClusterPrefixAllocationSucceeded(&allocation)).To(BeFalse(), "prefix allocation is succeeded: %#v", allocation)
		}, timeout, interval).Should(Succeed())
	})

	It("should assign a prefix on matching parents", func() {
		By("creating a child prefix")
		childPrefix := &ipamv1alpha1.ClusterPrefix{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-child-",
			},
			Spec: ipamv1alpha1.ClusterPrefixSpec{
				PrefixSpace: ipamv1alpha1.PrefixSpace{
					PrefixLength: 28,
				},
				ParentSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"foo": "bar",
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, childPrefix)).To(Succeed())

		By("checking that the child prefix is not being assigned")
		childPrefixKey := client.ObjectKeyFromObject(childPrefix)
		Consistently(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, childPrefixKey, childPrefix)).Should(Succeed())
			g.Expect(IsClusterPrefixReady(childPrefix)).To(BeFalse(), "child prefix is ready: %#v", childPrefix)
			g.Expect(childPrefix.Spec.ParentRef).To(BeNil())
		}, timeout, interval).Should(Succeed())

		By("creating a root prefix that would fit but does not match")
		prefixValue := commonv1alpha1.MustParseIPPrefix("10.0.0.0/24")
		notMatchingRootPrefix := &ipamv1alpha1.ClusterPrefix{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-root-",
			},
			Spec: ipamv1alpha1.ClusterPrefixSpec{
				PrefixSpace: ipamv1alpha1.PrefixSpace{
					Prefix: prefixValue,
				},
			},
		}
		Expect(k8sClient.Create(ctx, notMatchingRootPrefix)).To(Succeed())

		By("checking that the child prefix is not being assigned")
		Consistently(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, childPrefixKey, childPrefix)).Should(Succeed())
			g.Expect(IsClusterPrefixReady(childPrefix)).To(BeFalse(), "child prefix is ready: %#v", childPrefix)
			g.Expect(childPrefix.Spec.ParentRef).To(BeNil())
		}, timeout, interval).Should(Succeed())

		By("creating a root prefix that fits and matches")
		rootPrefix := &ipamv1alpha1.ClusterPrefix{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-root-",
				Labels: map[string]string{
					"foo": "bar",
				},
			},
			Spec: ipamv1alpha1.ClusterPrefixSpec{
				PrefixSpace: ipamv1alpha1.PrefixSpace{
					Prefix: prefixValue,
				},
			},
		}
		Expect(k8sClient.Create(ctx, rootPrefix)).To(Succeed())

		By("waiting for the child prefix to be marked as ready and report its available ranges")
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, childPrefixKey, childPrefix)).To(Succeed())
			g.Expect(IsClusterPrefixReady(childPrefix)).To(BeTrue(), "child prefix is not ready: %#v", childPrefix)
			g.Expect(childPrefix.Spec.ParentRef).To(Equal(&corev1.LocalObjectReference{
				Name: rootPrefix.Name,
			}))
			g.Expect(childPrefix.Status.Available).To(ConsistOf(commonv1alpha1.MustParseIPPrefix("10.0.0.0/28")))
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
		list := &ipamv1alpha1.ClusterPrefixAllocationList{}
		Expect(ListOwned(ctx, k8sClient, k8sClient.Scheme(), childPrefix, list)).To(Succeed())
		Expect(list.Items).To(HaveLen(1))
		allocation := &list.Items[0]
		Expect(allocation.Spec).To(Equal(ipamv1alpha1.ClusterPrefixAllocationSpec{
			PrefixRef: &corev1.LocalObjectReference{
				Name: rootPrefix.Name,
			},
			PrefixSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"foo": "bar",
				},
			},
			ClusterPrefixAllocationRequest: ipamv1alpha1.ClusterPrefixAllocationRequest{
				PrefixLength: 28,
			},
		}))
		Expect(IsClusterPrefixAllocationSucceeded(allocation)).To(BeTrue(), "allocation is not ready: %#v", allocation)
		Expect(allocation.Status.Prefix).To(Equal(commonv1alpha1.MustParseIPPrefix("10.0.0.0/28")))
	})

	It("should not distribute reserved prefixes", func() {
		By("creating a root prefix with reservations")
		prefixValue := commonv1alpha1.MustParseIPPrefix("10.0.0.0/8")
		rootPrefix := &ipamv1alpha1.ClusterPrefix{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-root-",
			},
			Spec: ipamv1alpha1.ClusterPrefixSpec{
				PrefixSpace: ipamv1alpha1.PrefixSpace{
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
			g.Expect(IsClusterPrefixReady(rootPrefix)).To(BeTrue(), "prefix is not ready: %#v", rootPrefix)
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
		rootPrefix := &ipamv1alpha1.ClusterPrefix{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-root-",
			},
			Spec: ipamv1alpha1.ClusterPrefixSpec{
				PrefixSpace: ipamv1alpha1.PrefixSpace{
					Prefix: prefixValue,
				},
			},
		}
		Expect(k8sClient.Create(ctx, rootPrefix)).To(Succeed())

		By("creating a child prefix with dynamic reservations")
		childPrefix := &ipamv1alpha1.ClusterPrefix{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-child-",
			},
			Spec: ipamv1alpha1.ClusterPrefixSpec{
				ParentRef: &corev1.LocalObjectReference{
					Name: rootPrefix.Name,
				},
				PrefixSpace: ipamv1alpha1.PrefixSpace{
					PrefixLength:       9,
					ReservationLengths: []int32{10},
				},
			},
		}
		Expect(k8sClient.Create(ctx, childPrefix)).To(Succeed())

		By("waiting for the child prefix to be allocated and report its available and reserved prefixes")
		childPrefixKey := client.ObjectKeyFromObject(childPrefix)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, childPrefixKey, childPrefix)).To(Succeed())
			g.Expect(IsClusterPrefixReady(childPrefix)).To(BeTrue(), "child prefix is not ready: %#v", childPrefix)

			childPrefixValue := childPrefix.Spec.Prefix.IPPrefix

			// The prefix should be of the requested size and within the parent prefix.
			g.Expect(childPrefixValue).To(SatisfyAll(
				WithTransform(func(ipPrefix netaddr.IPPrefix) uint8 {
					return ipPrefix.Bits()
				}, BeEquivalentTo(9)),
				WithTransform(ipSetFromPrefix(prefixValue.IPPrefix).ContainsPrefix, BeTrue()),
			))

			g.Expect(childPrefix.Status.Reserved).To(HaveLen(1))
			reservation := childPrefix.Status.Reserved[0].IPPrefix
			g.Expect(reservation).To(SatisfyAll(
				WithTransform(func(ipPrefix netaddr.IPPrefix) uint8 {
					return ipPrefix.Bits()
				}, BeEquivalentTo(10)),
				WithTransform(ipSetFromPrefix(childPrefixValue).ContainsPrefix, BeTrue()),
			))

			g.Expect(reservation.String()).NotTo(Equal(childPrefixValue.String()))

			var remainingBldr netaddr.IPSetBuilder
			remainingBldr.AddPrefix(childPrefixValue)
			remainingBldr.RemovePrefix(reservation)
			remainingPrefixes := ExpectIPSet(remainingBldr.IPSet()).Prefixes()

			g.Expect(childPrefix.Status.Available).To(HaveLen(len(remainingPrefixes)))
		}, timeout, interval).Should(Succeed())
	})
})
