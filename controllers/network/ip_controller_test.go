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
	"fmt"
	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	networkv1alpha1 "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ConsistOfSlice is a helper to create a types.GomegaMatcher that does ConsistsOf with the elements of a slice.
func ConsistOfSlice(slice interface{}) types.GomegaMatcher {
	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Slice {
		Fail(fmt.Sprintf("value of type %T is not a slice: %#v", slice, slice))
	}
	n := v.Len()
	elems := make([]interface{}, 0, n)
	for i := 0; i < n; i++ {
		elems = append(elems, v.Index(i).Interface())
	}

	return ConsistOf(elems...)
}

var _ = Describe("IPReconciler", func() {
	ns := SetupTest()

	var tenPrefixWithFirstIPRemoved []commonv1alpha1.IPPrefix
	BeforeEach(func() {
		tenPrefixWithFirstIPRemoved = []commonv1alpha1.IPPrefix{
			commonv1alpha1.MustParseIPPrefix("10.0.0.1/32"),
			commonv1alpha1.MustParseIPPrefix("10.0.0.2/31"),
			commonv1alpha1.MustParseIPPrefix("10.0.0.4/30"),
			commonv1alpha1.MustParseIPPrefix("10.0.0.8/29"),
			commonv1alpha1.MustParseIPPrefix("10.0.0.16/28"),
			commonv1alpha1.MustParseIPPrefix("10.0.0.32/27"),
			commonv1alpha1.MustParseIPPrefix("10.0.0.64/26"),
			commonv1alpha1.MustParseIPPrefix("10.0.0.128/25"),
			commonv1alpha1.MustParseIPPrefix("10.0.1.0/24"),
			commonv1alpha1.MustParseIPPrefix("10.0.2.0/23"),
			commonv1alpha1.MustParseIPPrefix("10.0.4.0/22"),
			commonv1alpha1.MustParseIPPrefix("10.0.8.0/21"),
			commonv1alpha1.MustParseIPPrefix("10.0.16.0/20"),
			commonv1alpha1.MustParseIPPrefix("10.0.32.0/19"),
			commonv1alpha1.MustParseIPPrefix("10.0.64.0/18"),
			commonv1alpha1.MustParseIPPrefix("10.0.128.0/17"),
			commonv1alpha1.MustParseIPPrefix("10.1.0.0/16"),
			commonv1alpha1.MustParseIPPrefix("10.2.0.0/15"),
			commonv1alpha1.MustParseIPPrefix("10.4.0.0/14"),
			commonv1alpha1.MustParseIPPrefix("10.8.0.0/13"),
			commonv1alpha1.MustParseIPPrefix("10.16.0.0/12"),
			commonv1alpha1.MustParseIPPrefix("10.32.0.0/11"),
			commonv1alpha1.MustParseIPPrefix("10.64.0.0/10"),
			commonv1alpha1.MustParseIPPrefix("10.128.0.0/9"),
		}
	})

	It("should allocate ips", func() {
		By("creating a prefix")
		prefixValue := commonv1alpha1.MustParseIPPrefix("10.0.0.0/24")
		prefix := &networkv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-ip-",
			},
			Spec: networkv1alpha1.PrefixSpec{
				PrefixSpace: networkv1alpha1.PrefixSpace{
					Prefix: prefixValue,
				},
			},
		}
		Expect(k8sClient.Create(ctx, prefix)).To(Succeed())

		By("creating an ip")
		ip := &networkv1alpha1.IP{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-child-",
			},
			Spec: networkv1alpha1.IPSpec{
				PrefixRef: &networkv1alpha1.PrefixReference{
					Kind: networkv1alpha1.PrefixKind,
					Name: prefix.Name,
				},
			},
		}
		Expect(k8sClient.Create(ctx, ip)).To(Succeed())

		By("waiting for the child prefix to be marked as ready and report its available ranges")
		expectedIP := commonv1alpha1.MustParseIP("10.0.0.0")
		ipKey := client.ObjectKeyFromObject(ip)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, ipKey, ip)).To(Succeed())
			g.Expect(IsIPReady(ip)).To(BeTrue(), "ip is not ready: %v", ip)
			g.Expect(ip.Spec.IP).To(Equal(expectedIP))
		}, timeout, interval).Should(Succeed())

		By("asserting there is a single valid allocation")
		list := &networkv1alpha1.PrefixAllocationList{}
		Expect(ListOwned(ctx, k8sClient, k8sClient.Scheme(), ip, list, client.InNamespace(ns.Name))).To(Succeed())
		Expect(list.Items).To(HaveLen(1))
		allocation := &list.Items[0]
		Expect(allocation.Spec).To(Equal(networkv1alpha1.PrefixAllocationSpec{
			PrefixRef: &networkv1alpha1.PrefixReference{
				Kind: networkv1alpha1.PrefixKind,
				Name: prefix.Name,
			},
			PrefixAllocationRequest: networkv1alpha1.PrefixAllocationRequest{
				RangeLength: 1,
			},
		}))
		Expect(IsPrefixAllocationSucceeded(allocation)).To(BeTrue(), "allocation is not ready: %v", allocation)
		Expect(allocation.Status.Range).To(Equal(commonv1alpha1.PtrToIPRange(commonv1alpha1.IPRangeFrom(expectedIP, expectedIP))))
	})

	It("should leave ips in pending state when they can't be allocated", func() {
		By("creating a prefix")
		prefixValue := commonv1alpha1.MustParseIPPrefix("10.0.0.0/24")
		prefix := &networkv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-ip-",
			},
			Spec: networkv1alpha1.PrefixSpec{
				PrefixSpace: networkv1alpha1.PrefixSpace{
					Prefix: prefixValue,
				},
			},
		}
		Expect(k8sClient.Create(ctx, prefix)).To(Succeed())

		By("creating an ip")
		ip := &networkv1alpha1.IP{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "ip-",
			},
			Spec: networkv1alpha1.IPSpec{
				PrefixRef: &networkv1alpha1.PrefixReference{
					Kind: networkv1alpha1.PrefixKind,
					Name: prefix.Name,
				},
				IP: commonv1alpha1.MustParseIP("9.9.9.9"),
			},
		}
		Expect(k8sClient.Create(ctx, ip)).To(Succeed())

		By("asserting there are no allocations and the child does not become ready")
		ipKey := client.ObjectKeyFromObject(ip)
		Consistently(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, ipKey, ip)).Should(Succeed())
			g.Expect(IsIPReady(ip)).Should(BeFalse(), "ip is ready: %#v", ip)

			list := &networkv1alpha1.PrefixAllocationList{}
			g.Expect(list.Items).To(SatisfyAny(
				BeEmpty(),
				SatisfyAll(
					HaveLen(1),
					WithTransform(func(items []networkv1alpha1.PrefixAllocation) bool {
						return IsPrefixAllocationSucceeded(&items[0])
					}, BeFalse()),
				),
			))
		}, timeout, interval).Should(Succeed())
	})

	It("should assign an ip to matching parents", func() {
		By("creating an ip")
		ip := &networkv1alpha1.IP{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "ip-",
			},
			Spec: networkv1alpha1.IPSpec{
				PrefixSelector: &networkv1alpha1.PrefixSelector{
					Kind: networkv1alpha1.PrefixKind,
					LabelSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"foo": "bar",
						},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, ip)).To(Succeed())

		By("checking that the ip is not being assigned")
		ipKey := client.ObjectKeyFromObject(ip)
		Consistently(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, ipKey, ip)).Should(Succeed())
			g.Expect(IsIPReady(ip)).To(BeFalse(), "ip is ready: %v", ip)
			g.Expect(ip.Spec.PrefixRef).To(BeNil())
		}).Should(Succeed())

		By("creating a prefix that would fit but does not match")
		prefixValue := commonv1alpha1.MustParseIPPrefix("10.0.0.0/24")
		notMatchingPrefix := &networkv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-not-match-",
			},
			Spec: networkv1alpha1.PrefixSpec{
				PrefixSpace: networkv1alpha1.PrefixSpace{
					Prefix: prefixValue,
				},
			},
		}
		Expect(k8sClient.Create(ctx, notMatchingPrefix)).To(Succeed())

		By("checking that the ip is not being assigned")
		Consistently(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, ipKey, ip)).Should(Succeed())
			g.Expect(IsIPReady(ip)).To(BeFalse(), "ip is ready: %v", ip)
			g.Expect(ip.Spec.PrefixRef).To(BeNil())
		}).Should(Succeed())

		By("creating a prefix that fits and matches")
		prefix := &networkv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-ip-match-",
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
		Expect(k8sClient.Create(ctx, prefix)).To(Succeed())

		By("waiting for the ip to be marked as ready")
		expectedIP := commonv1alpha1.MustParseIP("10.0.0.0")
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, ipKey, ip)).To(Succeed())
			g.Expect(IsIPReady(ip)).To(BeTrue(), "ip is not ready: %v", ip)
			g.Expect(ip.Spec.PrefixRef).To(Equal(&networkv1alpha1.PrefixReference{
				Kind: networkv1alpha1.PrefixKind,
				Name: prefix.Name,
			}))
			g.Expect(ip.Spec.IP).To(Equal(expectedIP))
		}, timeout, interval).Should(Succeed())

		By("asserting there is a single valid allocation")
		list := &networkv1alpha1.PrefixAllocationList{}
		Expect(ListOwned(ctx, k8sClient, k8sClient.Scheme(), ip, list, client.InNamespace(ns.Name))).To(Succeed())
		Expect(list.Items).To(HaveLen(1))
		allocation := &list.Items[0]
		Expect(allocation.Spec).To(Equal(networkv1alpha1.PrefixAllocationSpec{
			PrefixRef: &networkv1alpha1.PrefixReference{
				Kind: networkv1alpha1.PrefixKind,
				Name: prefix.Name,
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
				RangeLength: 1,
			},
		}))
		Expect(IsPrefixAllocationSucceeded(allocation)).To(BeTrue(), "allocation is not ready: %v", allocation)
		Expect(allocation.Status.Range).To(Equal(commonv1alpha1.PtrToIPRange(commonv1alpha1.IPRangeFrom(expectedIP, expectedIP))))
	})

	It("should allocate ips from cluster prefixes by reference", func() {
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

		By("creating an ip referencing the cluster prefix")
		ip := &networkv1alpha1.IP{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-",
			},
			Spec: networkv1alpha1.IPSpec{
				PrefixRef: &networkv1alpha1.PrefixReference{
					Kind: networkv1alpha1.ClusterPrefixKind,
					Name: clusterPrefix.Name,
				},
				IP: commonv1alpha1.MustParseIP("10.0.0.0"),
			},
		}
		Expect(k8sClient.Create(ctx, ip)).To(Succeed())

		By("waiting for the ip to be allocated")
		clusterPrefixKey := client.ObjectKeyFromObject(clusterPrefix)
		ipKey := client.ObjectKeyFromObject(ip)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, ipKey, ip)).To(Succeed())
			g.Expect(IsIPReady(ip)).To(BeTrue(), "ip is not ready: %#v", ip)

			Expect(k8sClient.Get(ctx, clusterPrefixKey, clusterPrefix)).To(Succeed())
			g.Expect(clusterPrefix.Status.Available).To(ConsistOfSlice(tenPrefixWithFirstIPRemoved))
		}, timeout, interval).Should(Succeed())
	})

	It("should allocate ips from cluster prefixes by selector", func() {
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
		ip := &networkv1alpha1.IP{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-",
			},
			Spec: networkv1alpha1.IPSpec{
				PrefixSelector: &networkv1alpha1.PrefixSelector{
					Kind: networkv1alpha1.ClusterPrefixKind,
					LabelSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"foo": "bar",
						},
					},
				},
				IP: commonv1alpha1.MustParseIP("10.0.0.0"),
			},
		}
		Expect(k8sClient.Create(ctx, ip)).To(Succeed())

		By("waiting for the ip to be allocated")
		clusterPrefixKey := client.ObjectKeyFromObject(clusterPrefix)
		ipKey := client.ObjectKeyFromObject(ip)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, ipKey, ip)).To(Succeed())
			g.Expect(IsIPReady(ip)).To(BeTrue(), "ip is not ready: %#v", ip)

			Expect(k8sClient.Get(ctx, clusterPrefixKey, clusterPrefix)).To(Succeed())
			g.Expect(clusterPrefix.Status.Available).To(ConsistOfSlice(tenPrefixWithFirstIPRemoved))
		}, timeout, interval).Should(Succeed())
	})

	It("should dynamically allocate ips from cluster prefixes by reference", func() {
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

		By("creating an ip referencing the cluster prefix")
		ip := &networkv1alpha1.IP{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-",
			},
			Spec: networkv1alpha1.IPSpec{
				PrefixRef: &networkv1alpha1.PrefixReference{
					Kind: networkv1alpha1.ClusterPrefixKind,
					Name: clusterPrefix.Name,
				},
				IP: commonv1alpha1.MustParseIP("10.0.0.0"),
			},
		}
		Expect(k8sClient.Create(ctx, ip)).To(Succeed())

		By("waiting for the prefix to be allocated")
		clusterPrefixKey := client.ObjectKeyFromObject(clusterPrefix)
		ipKey := client.ObjectKeyFromObject(ip)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, ipKey, ip)).To(Succeed())
			g.Expect(IsIPReady(ip)).To(BeTrue(), "ip is not ready: %#v", ip)

			Expect(k8sClient.Get(ctx, clusterPrefixKey, clusterPrefix)).To(Succeed())
			g.Expect(clusterPrefix.Status.Available).To(ConsistOfSlice(tenPrefixWithFirstIPRemoved))
		}, timeout, interval).Should(Succeed())
	})
})
