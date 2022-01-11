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
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"inet.af/netaddr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	networkv1alpha1 "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
)

const (
	parentCIDR = "192.168.1.0/24"
)

var _ = Describe("IPAMRangeReconciler", func() {
	Context("Reconcile an IPAMRange", func() {
		ns := SetupTest()

		It("should reconcile parent without children", func() {
			parent := createParentIPAMRange(ctx, ns)

			By("Check parent allocations")
			expectedAllocations := map[string]networkv1alpha1.IPAMRangeAllocationState{
				parentCIDR: networkv1alpha1.IPAMRangeAllocationFree,
			}
			Eventually(func(g Gomega) {
				key := types.NamespacedName{Name: parent.Name, Namespace: parent.Namespace}
				obj := &networkv1alpha1.IPAMRange{}
				g.Expect(k8sClient.Get(ctx, key, obj)).Should(Succeed())
				g.Expect(getAllocationStates(obj)).To(Equal(expectedAllocations))
			}, timeout, interval).Should(Succeed())
		})
		It("should update parent allocations with CIDR request", func() {
			parent := createParentIPAMRange(ctx, ns)
			child := createChildIPAMRange(ctx, parent, "192.168.1.0/25", nil, 0, 0)

			By("Check parent allocations")
			expectedParentAllocations := map[string]networkv1alpha1.IPAMRangeAllocationState{
				"192.168.1.0/25":   networkv1alpha1.IPAMRangeAllocationUsed,
				"192.168.1.128/25": networkv1alpha1.IPAMRangeAllocationFree,
			}
			Eventually(func(g Gomega) {
				key := types.NamespacedName{Name: parent.Name, Namespace: parent.Namespace}
				obj := &networkv1alpha1.IPAMRange{}
				g.Expect(k8sClient.Get(ctx, key, obj)).Should(Succeed())
				g.Expect(getAllocationStates(obj)).To(Equal(expectedParentAllocations))
			}, timeout, interval).Should(Succeed())

			By("Check child allocations")
			expectedChildAllocations := map[string]networkv1alpha1.IPAMRangeAllocationState{
				"192.168.1.0/25": networkv1alpha1.IPAMRangeAllocationFree,
			}
			Eventually(func(g Gomega) {
				key := types.NamespacedName{Name: child.Name, Namespace: child.Namespace}
				obj := &networkv1alpha1.IPAMRange{}
				g.Expect(k8sClient.Get(ctx, key, obj)).Should(Succeed())
				g.Expect(getAllocationStates(obj)).To(Equal(expectedChildAllocations))
			}, timeout, interval).Should(Succeed())
		})
		It("should update parent allocations with size request", func() {
			parent := createParentIPAMRange(ctx, ns)
			child := createChildIPAMRange(ctx, parent, "", nil, 25, 0)

			By("Check parent allocations")
			expectedParentAllocations := map[string]networkv1alpha1.IPAMRangeAllocationState{
				"192.168.1.0/25":   networkv1alpha1.IPAMRangeAllocationUsed,
				"192.168.1.128/25": networkv1alpha1.IPAMRangeAllocationFree,
			}
			Eventually(func(g Gomega) {
				key := types.NamespacedName{Name: parent.Name, Namespace: parent.Namespace}
				obj := &networkv1alpha1.IPAMRange{}
				g.Expect(k8sClient.Get(ctx, key, obj)).Should(Succeed())
				g.Expect(getAllocationStates(obj)).To(Equal(expectedParentAllocations))
			}, timeout, interval).Should(Succeed())

			By("Check child allocations")
			expectedChildAllocations := map[string]networkv1alpha1.IPAMRangeAllocationState{
				"192.168.1.0/25": networkv1alpha1.IPAMRangeAllocationFree,
			}
			parsedCIDR := commonv1alpha1.MustParseIPPrefix("192.168.1.0/25")
			Eventually(func(g Gomega) {
				key := types.NamespacedName{Name: child.Name, Namespace: child.Namespace}
				obj := &networkv1alpha1.IPAMRange{}
				g.Expect(k8sClient.Get(ctx, key, obj)).Should(Succeed())
				g.Expect(getAllocationStates(obj)).To(Equal(expectedChildAllocations))
				g.Expect(obj.Status.Allocations).To(ContainElement(networkv1alpha1.IPAMRangeAllocationStatus{
					CIDR:  &parsedCIDR,
					State: networkv1alpha1.IPAMRangeAllocationFree,
					Request: &networkv1alpha1.IPAMRangeRequest{
						Size: int32(25),
					},
				}))
			}, timeout, interval).Should(Succeed())
		})
		It("should update parent allocations with ip request", func() {
			parent := createParentIPAMRange(ctx, ns)

			fromIP, _ := netaddr.ParseIP("192.168.1.0")
			toIP, _ := netaddr.ParseIP("192.168.1.127")
			ipRange := &commonv1alpha1.IPRange{
				From: commonv1alpha1.NewIP(fromIP),
				To:   commonv1alpha1.NewIP(toIP),
			}
			child := createChildIPAMRange(ctx, parent, "", ipRange, 0, 0)

			By("Check parent allocations")
			expectedParentAllocations := map[string]networkv1alpha1.IPAMRangeAllocationState{
				"192.168.1.128/25": networkv1alpha1.IPAMRangeAllocationFree,
			}
			Eventually(func(g Gomega) {
				key := types.NamespacedName{Name: parent.Name, Namespace: parent.Namespace}
				obj := &networkv1alpha1.IPAMRange{}
				g.Expect(k8sClient.Get(ctx, key, obj)).Should(Succeed())
				g.Expect(getAllocationStates(obj)).To(Equal(expectedParentAllocations))
				g.Expect(getIPRanges(obj)).To(ContainElement(ipRange))
			}, timeout, interval).Should(Succeed())

			By("Check child allocations")
			Eventually(func(g Gomega) {
				key := types.NamespacedName{Name: child.Name, Namespace: child.Namespace}
				obj := &networkv1alpha1.IPAMRange{}
				g.Expect(k8sClient.Get(ctx, key, obj)).Should(Succeed())
				g.Expect(getIPRanges(obj)).To(ContainElement(ipRange))
			}, timeout, interval).Should(Succeed())
		})
		It("should update parent allocations with ip count request", func() {
			parent := createParentIPAMRange(ctx, ns)
			child := createChildIPAMRange(ctx, parent, "", nil, 0, 1)

			By("Check parent allocations")
			fromIP, _ := netaddr.ParseIP("192.168.1.1")
			toIP, _ := netaddr.ParseIP("192.168.1.1")
			ipRange := &commonv1alpha1.IPRange{
				From: commonv1alpha1.NewIP(fromIP),
				To:   commonv1alpha1.NewIP(toIP),
			}
			Eventually(func(g Gomega) {
				key := types.NamespacedName{Name: parent.Name, Namespace: parent.Namespace}
				obj := &networkv1alpha1.IPAMRange{}
				g.Expect(k8sClient.Get(ctx, key, obj)).Should(Succeed())
				g.Expect(getIPRanges(obj)).To(ContainElement(ipRange))
			}, timeout, interval).Should(Succeed())

			By("Check child allocations")
			Eventually(func(g Gomega) {
				key := types.NamespacedName{Name: child.Name, Namespace: child.Namespace}
				obj := &networkv1alpha1.IPAMRange{}
				g.Expect(k8sClient.Get(ctx, key, obj)).Should(Succeed())
				g.Expect(getIPRanges(obj)).To(ContainElement(ipRange))
			}, timeout, interval).Should(Succeed())
		})
		It("allocation should fail if CIDR is out of range", func() {
			parent := createParentIPAMRange(ctx, ns)
			child := createChildIPAMRange(ctx, parent, "192.168.2.0/25", nil, 0, 0)

			By("Check parent allocations")
			expectedParentAllocations := map[string]networkv1alpha1.IPAMRangeAllocationState{
				parentCIDR: networkv1alpha1.IPAMRangeAllocationFree,
			}
			Eventually(func(g Gomega) {
				key := types.NamespacedName{Name: parent.Name, Namespace: parent.Namespace}
				obj := &networkv1alpha1.IPAMRange{}
				g.Expect(k8sClient.Get(ctx, key, obj)).Should(Succeed())
				g.Expect(getAllocationStates(obj)).To(Equal(expectedParentAllocations))
				g.Expect(getFailedRequests(obj)).To(ContainElements(getRequestKeys(child.Spec.Requests)...))
			}, timeout, interval).Should(Succeed())

			By("Check child allocations")
			Eventually(func(g Gomega) {
				key := types.NamespacedName{Name: child.Name, Namespace: child.Namespace}
				obj := &networkv1alpha1.IPAMRange{}
				g.Expect(k8sClient.Get(ctx, key, obj)).Should(Succeed())
				g.Expect(getFailedRequests(obj)).To(ContainElements(getRequestKeys(child.Spec.Requests)...))
			}, timeout, interval).Should(Succeed())
		})
		It("allocation should fail if size is too big", func() {
			parent := createParentIPAMRange(ctx, ns)
			child := createChildIPAMRange(ctx, parent, "", nil, 23, 0)

			By("Check parent allocations")
			expectedParentAllocations := map[string]networkv1alpha1.IPAMRangeAllocationState{
				parentCIDR: networkv1alpha1.IPAMRangeAllocationFree,
			}
			Eventually(func(g Gomega) {
				key := types.NamespacedName{Name: parent.Name, Namespace: parent.Namespace}
				obj := &networkv1alpha1.IPAMRange{}
				g.Expect(k8sClient.Get(ctx, key, obj)).Should(Succeed())
				g.Expect(getAllocationStates(obj)).To(Equal(expectedParentAllocations))
				g.Expect(getFailedRequests(obj)).To(ContainElements(getRequestKeys(child.Spec.Requests)...))
			}, timeout, interval).Should(Succeed())

			By("Check child allocations")
			Eventually(func(g Gomega) {
				key := types.NamespacedName{Name: child.Name, Namespace: child.Namespace}
				obj := &networkv1alpha1.IPAMRange{}
				g.Expect(k8sClient.Get(ctx, key, obj)).Should(Succeed())
				g.Expect(getFailedRequests(obj)).To(ContainElements(getRequestKeys(child.Spec.Requests)...))
			}, timeout, interval).Should(Succeed())
		})
		It("allocation should fail if ip is out of range", func() {
			parent := createParentIPAMRange(ctx, ns)

			fromIP, _ := netaddr.ParseIP("192.168.2.0")
			toIP, _ := netaddr.ParseIP("192.168.2.127")
			ipRange := &commonv1alpha1.IPRange{
				From: commonv1alpha1.NewIP(fromIP),
				To:   commonv1alpha1.NewIP(toIP),
			}
			child := createChildIPAMRange(ctx, parent, "", ipRange, 0, 0)

			By("Check parent allocations")
			expectedParentAllocations := map[string]networkv1alpha1.IPAMRangeAllocationState{
				parentCIDR: networkv1alpha1.IPAMRangeAllocationFree,
			}
			Eventually(func(g Gomega) {
				key := types.NamespacedName{Name: parent.Name, Namespace: parent.Namespace}
				obj := &networkv1alpha1.IPAMRange{}
				g.Expect(k8sClient.Get(ctx, key, obj)).Should(Succeed())
				g.Expect(getAllocationStates(obj)).To(Equal(expectedParentAllocations))
				g.Expect(getFailedRequests(obj)).To(ContainElements(getRequestKeys(child.Spec.Requests)...))
			}, timeout, interval).Should(Succeed())

			By("Check child allocations")
			Eventually(func(g Gomega) {
				key := types.NamespacedName{Name: child.Name, Namespace: child.Namespace}
				obj := &networkv1alpha1.IPAMRange{}
				g.Expect(k8sClient.Get(ctx, key, obj)).Should(Succeed())
				g.Expect(getFailedRequests(obj)).To(ContainElements(getRequestKeys(child.Spec.Requests)...))
			}, timeout, interval).Should(Succeed())
		})

		// TODO: Fix sporadic issue
		PIt("should update allocations when CIDR is changed", func() {
			parent := createParentIPAMRange(ctx, ns)
			child := createChildIPAMRange(ctx, parent, "192.168.2.0/25", nil, 0, 0)

			// Allocation should fail at first, because request CIDR is out of range
			By("Check parent allocations")
			expectedParentAllocations := map[string]networkv1alpha1.IPAMRangeAllocationState{
				parentCIDR: networkv1alpha1.IPAMRangeAllocationFree,
			}
			Eventually(func(g Gomega) {
				key := types.NamespacedName{Name: parent.Name, Namespace: parent.Namespace}
				obj := &networkv1alpha1.IPAMRange{}
				g.Expect(k8sClient.Get(ctx, key, obj)).Should(Succeed())
				g.Expect(getAllocationStates(obj)).To(Equal(expectedParentAllocations))
				g.Expect(getFailedRequests(obj)).To(ContainElements(getRequestKeys(child.Spec.Requests)...))

				// Change parent CIDR and wait for allocation to succeed
				prefix, err := netaddr.ParseIPPrefix("192.168.2.0/24")
				Expect(err).ToNot(HaveOccurred())

				updParent := &networkv1alpha1.IPAMRange{}
				Expect(k8sClient.Get(ctx, types.NamespacedName{
					Name:      parent.Name,
					Namespace: parent.Namespace,
				}, updParent)).ToNot(HaveOccurred())
				updParent.Spec.CIDRs = []commonv1alpha1.IPPrefix{commonv1alpha1.NewIPPrefix(prefix)}
				Expect(k8sClient.Update(ctx, updParent)).To(Succeed())
			}, timeout, interval).Should(Succeed())

			By("Check child allocations")
			Eventually(func(g Gomega) {
				key := types.NamespacedName{Name: child.Name, Namespace: child.Namespace}
				obj := &networkv1alpha1.IPAMRange{}
				g.Expect(k8sClient.Get(ctx, key, obj)).Should(Succeed())
				g.Expect(getFailedRequests(obj)).To(ContainElements(getRequestKeys(child.Spec.Requests)...))
			}, timeout, interval).Should(Succeed())

			By("Check parent allocations after CIDR change")
			expectedChangedParentAllocations := map[string]networkv1alpha1.IPAMRangeAllocationState{
				"192.168.2.0/25":   networkv1alpha1.IPAMRangeAllocationUsed,
				"192.168.2.128/25": networkv1alpha1.IPAMRangeAllocationFree,
			}
			Eventually(func(g Gomega) {
				key := types.NamespacedName{Name: parent.Name, Namespace: parent.Namespace}
				obj := &networkv1alpha1.IPAMRange{}
				g.Expect(k8sClient.Get(ctx, key, obj)).Should(Succeed())
				g.Expect(getAllocationStates(obj)).To(Equal(expectedChangedParentAllocations))
			}, timeout, interval).Should(Succeed())

			By("Check child allocations after CIDR change")
			expectedChangedChildAllocations := map[string]networkv1alpha1.IPAMRangeAllocationState{
				"192.168.2.0/25": networkv1alpha1.IPAMRangeAllocationFree,
			}
			Eventually(func(g Gomega) {
				key := types.NamespacedName{Name: child.Name, Namespace: child.Namespace}
				obj := &networkv1alpha1.IPAMRange{}
				g.Expect(k8sClient.Get(ctx, key, obj)).Should(Succeed())
				g.Expect(getAllocationStates(obj)).To(Equal(expectedChangedChildAllocations))
			}, timeout, interval).Should(Succeed())
		})
	})
})

func createIPAMRange(
	ctx context.Context,
	meta metav1.ObjectMeta,
	cidrStr, parentName string,
	ipRange *commonv1alpha1.IPRange,
	size, count int32,
) *networkv1alpha1.IPAMRange {
	spec := networkv1alpha1.IPAMRangeSpec{}

	var cidr *commonv1alpha1.IPPrefix
	if cidrStr != "" {
		prefix, err := netaddr.ParseIPPrefix(cidrStr)
		Expect(err).ToNot(HaveOccurred())
		c := commonv1alpha1.NewIPPrefix(prefix)
		cidr = &c
	}
	if parentName != "" {
		spec.Parent = &corev1.LocalObjectReference{
			Name: parentName,
		}
		spec.Requests = []networkv1alpha1.IPAMRangeRequest{
			{
				CIDR:    cidr,
				IPs:     ipRange,
				Size:    size,
				IPCount: count,
			},
		}
	} else if cidr != nil {
		spec.CIDRs = []commonv1alpha1.IPPrefix{*cidr}
	}

	instance := &networkv1alpha1.IPAMRange{
		Spec:       spec,
		ObjectMeta: meta,
	}
	Expect(k8sClient.Create(ctx, instance)).To(Succeed())

	return instance
}

func createParentIPAMRange(ctx context.Context, ns *corev1.Namespace) *networkv1alpha1.IPAMRange {
	meta := metav1.ObjectMeta{
		Name:      "parent",
		Namespace: ns.Name,
	}
	return createIPAMRange(ctx, meta, parentCIDR, "", nil, 0, 0)
}

func createChildIPAMRange(
	ctx context.Context,
	parent *networkv1alpha1.IPAMRange,
	cidrStr string,
	ipRange *commonv1alpha1.IPRange,
	size, count int32,
) *networkv1alpha1.IPAMRange {
	meta := metav1.ObjectMeta{
		Name:      "child",
		Namespace: parent.Namespace,
	}
	return createIPAMRange(ctx, meta, cidrStr, parent.Name, ipRange, size, count)
}

func getAllocationStates(obj *networkv1alpha1.IPAMRange) map[string]networkv1alpha1.IPAMRangeAllocationState {
	result := make(map[string]networkv1alpha1.IPAMRangeAllocationState)
	for _, alloc := range obj.Status.Allocations {
		if alloc.CIDR != nil {
			result[alloc.CIDR.String()] = alloc.State
		}
	}

	return result
}

func getIPRanges(obj *networkv1alpha1.IPAMRange) (result []*commonv1alpha1.IPRange) {
	for _, alloc := range obj.Status.Allocations {
		if alloc.IPs != nil {
			result = append(result, alloc.IPs)
		}
	}

	return
}

func getFailedRequests(obj *networkv1alpha1.IPAMRange) (result []string) {
	for _, alloc := range obj.Status.Allocations {
		if alloc.State == networkv1alpha1.IPAMRangeAllocationFailed {
			result = append(result, getRequestKey(*alloc.Request))
		}
	}

	return
}

func getRequestKey(req networkv1alpha1.IPAMRangeRequest) (key string) {
	if req.IPs != nil {
		key = req.IPs.From.String() + "-" + req.IPs.To.String()
	} else {
		key = fmt.Sprint(req)
	}

	return
}

func getRequestKeys(reqs []networkv1alpha1.IPAMRangeRequest) (result []interface{}) {
	for _, r := range reqs {
		result = append(result, getRequestKey(r))
	}

	return
}
