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
	"time"

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
	timeout  = time.Second * 10
	interval = time.Millisecond * 250

	parentCIDR = "192.168.1.0/24"
)

var _ = Describe("IPAMRangeReconciler", func() {
	Context("Reconcile an IPAMRange", func() {
		ctx := context.Background()
		ns := SetupTest(ctx)

		testFunc := func(
			obj *networkv1alpha1.IPAMRange,
			expectedAllocations map[string]networkv1alpha1.IPAMRangeAllocationState,
			ipRange *commonv1alpha1.IPRange,
			failedRequests []networkv1alpha1.IPAMRangeRequest,
			callback func(),
		) {
			Eventually(func(g Gomega) {
				key := types.NamespacedName{Name: obj.Name, Namespace: obj.Namespace}
				newObj := &networkv1alpha1.IPAMRange{}
				g.Expect(k8sClient.Get(ctx, key, newObj)).Should(Succeed())

				if len(expectedAllocations) > 0 {
					g.Expect(getAllocationStates(newObj)).To(Equal(expectedAllocations))
				}
				if ipRange != nil {
					g.Expect(getIPRanges(newObj)).To(ContainElement(ipRange))
				}
				g.Expect(getFailedRequests(newObj)).To(ContainElements(getRequestKeys(failedRequests)...))

				if callback != nil {
					go callback()
				}
			}, timeout, interval).Should(Succeed())
		}
		It("should reconcile parent without children", func() {
			parent := createParentIPAMRange(ctx, ns)

			By("Check parent allocations")
			expectedAllocations := map[string]networkv1alpha1.IPAMRangeAllocationState{
				parentCIDR: networkv1alpha1.IPAMRangeAllocationFree,
			}
			testFunc(parent, expectedAllocations, nil, nil, nil)
		})
		It("should update parent allocations with CIDR request", func() {
			parent := createParentIPAMRange(ctx, ns)
			child := createChildIPAMRange(ctx, parent, "192.168.1.0/25", nil, 0, 0)

			By("Check parent allocations")
			expectedParentAllocations := map[string]networkv1alpha1.IPAMRangeAllocationState{
				"192.168.1.0/25":   networkv1alpha1.IPAMRangeAllocationUsed,
				"192.168.1.128/25": networkv1alpha1.IPAMRangeAllocationFree,
			}
			testFunc(parent, expectedParentAllocations, nil, nil, nil)

			By("Check child allocations")
			expectedChildAllocations := map[string]networkv1alpha1.IPAMRangeAllocationState{
				"192.168.1.0/25": networkv1alpha1.IPAMRangeAllocationFree,
			}
			testFunc(child, expectedChildAllocations, nil, nil, nil)
		})
		It("should update parent allocations with size request", func() {
			parent := createParentIPAMRange(ctx, ns)
			child := createChildIPAMRange(ctx, parent, "", nil, 25, 0)

			By("Check parent allocations")
			expectedParentAllocations := map[string]networkv1alpha1.IPAMRangeAllocationState{
				"192.168.1.0/25":   networkv1alpha1.IPAMRangeAllocationUsed,
				"192.168.1.128/25": networkv1alpha1.IPAMRangeAllocationFree,
			}
			testFunc(parent, expectedParentAllocations, nil, nil, nil)

			By("Check child allocations")
			expectedChildAllocations := map[string]networkv1alpha1.IPAMRangeAllocationState{
				"192.168.1.0/25": networkv1alpha1.IPAMRangeAllocationFree,
			}
			testFunc(child, expectedChildAllocations, nil, nil, nil)
		})
		It("should update parent allocations with ip request", func() {
			parent := createParentIPAMRange(ctx, ns)

			fromIP, _ := netaddr.ParseIP("192.168.1.0")
			toIP, _ := netaddr.ParseIP("192.168.1.127")
			ipRange := &commonv1alpha1.IPRange{
				From: commonv1alpha1.NewIPAddr(fromIP),
				To:   commonv1alpha1.NewIPAddr(toIP),
			}
			child := createChildIPAMRange(ctx, parent, "", ipRange, 0, 0)

			By("Check parent allocations")
			expectedParentAllocations := map[string]networkv1alpha1.IPAMRangeAllocationState{
				"192.168.1.128/25": networkv1alpha1.IPAMRangeAllocationFree,
			}
			testFunc(parent, expectedParentAllocations, ipRange, nil, nil)

			By("Check child allocations")
			testFunc(child, nil, ipRange, nil, nil)
		})
		It("should update parent allocations with ip count request", func() {
			parent := createParentIPAMRange(ctx, ns)
			child := createChildIPAMRange(ctx, parent, "", nil, 0, 1)

			By("Check parent allocations")
			fromIP, _ := netaddr.ParseIP("192.168.1.1")
			toIP, _ := netaddr.ParseIP("192.168.1.1")
			ipRange := &commonv1alpha1.IPRange{
				From: commonv1alpha1.NewIPAddr(fromIP),
				To:   commonv1alpha1.NewIPAddr(toIP),
			}
			testFunc(parent, nil, ipRange, nil, nil)

			By("Check child allocations")
			testFunc(child, nil, ipRange, nil, nil)
		})
		It("allocation should fail if CIDR is out of range", func() {
			parent := createParentIPAMRange(ctx, ns)
			child := createChildIPAMRange(ctx, parent, "192.168.2.0/25", nil, 0, 0)

			By("Check parent allocations")
			expectedParentAllocations := map[string]networkv1alpha1.IPAMRangeAllocationState{
				parentCIDR: networkv1alpha1.IPAMRangeAllocationFree,
			}
			testFunc(parent, expectedParentAllocations, nil, child.Spec.Requests, nil)

			By("Check child allocations")
			testFunc(child, nil, nil, child.Spec.Requests, nil)
		})
		It("allocation should fail if size is too big", func() {
			parent := createParentIPAMRange(ctx, ns)
			child := createChildIPAMRange(ctx, parent, "", nil, 23, 0)

			By("Check parent allocations")
			expectedParentAllocations := map[string]networkv1alpha1.IPAMRangeAllocationState{
				parentCIDR: networkv1alpha1.IPAMRangeAllocationFree,
			}
			testFunc(parent, expectedParentAllocations, nil, child.Spec.Requests, nil)

			By("Check child allocations")
			testFunc(child, nil, nil, child.Spec.Requests, nil)
		})
		It("allocation should fail if ip is out of range", func() {
			parent := createParentIPAMRange(ctx, ns)

			fromIP, _ := netaddr.ParseIP("192.168.2.0")
			toIP, _ := netaddr.ParseIP("192.168.2.127")
			ipRange := &commonv1alpha1.IPRange{
				From: commonv1alpha1.NewIPAddr(fromIP),
				To:   commonv1alpha1.NewIPAddr(toIP),
			}
			child := createChildIPAMRange(ctx, parent, "", ipRange, 0, 0)

			By("Check parent allocations")
			expectedParentAllocations := map[string]networkv1alpha1.IPAMRangeAllocationState{
				parentCIDR: networkv1alpha1.IPAMRangeAllocationFree,
			}
			testFunc(parent, expectedParentAllocations, nil, child.Spec.Requests, nil)

			By("Check child allocations")
			testFunc(child, nil, nil, child.Spec.Requests, nil)
		})
		It("should update allocations when CIDR is changed", func() {
			parent := createParentIPAMRange(ctx, ns)
			child := createChildIPAMRange(ctx, parent, "192.168.2.0/25", nil, 0, 0)

			// Allocation should fail at first, because request CIDR is out of range
			// callback changes parent CIDR and wait for allocation to succeed
			callback := func() {
				prefix, err := netaddr.ParseIPPrefix("192.168.2.0/24")
				Expect(err).ToNot(HaveOccurred())

				updParent := &networkv1alpha1.IPAMRange{}
				Expect(k8sClient.Get(ctx, types.NamespacedName{
					Name:      parent.Name,
					Namespace: parent.Namespace,
				}, updParent)).ToNot(HaveOccurred())
				updParent.Spec.CIDRs = []commonv1alpha1.CIDR{commonv1alpha1.NewCIDR(prefix)}
				Expect(k8sClient.Update(ctx, updParent)).To(Succeed())

				By("Check parent allocations after CIDR change")
				expectedParentAllocations := map[string]networkv1alpha1.IPAMRangeAllocationState{
					"192.168.2.0/25":   networkv1alpha1.IPAMRangeAllocationUsed,
					"192.168.2.128/25": networkv1alpha1.IPAMRangeAllocationFree,
				}
				testFunc(updParent, expectedParentAllocations, nil, nil, nil)

				By("Check child allocations after CIDR change")
				expectedChildAllocations := map[string]networkv1alpha1.IPAMRangeAllocationState{
					"192.168.2.0/25": networkv1alpha1.IPAMRangeAllocationFree,
				}
				testFunc(child, expectedChildAllocations, nil, nil, nil)
			}

			By("Check parent allocations")
			expectedParentAllocations := map[string]networkv1alpha1.IPAMRangeAllocationState{
				parentCIDR: networkv1alpha1.IPAMRangeAllocationFree,
			}
			testFunc(parent, expectedParentAllocations, nil, child.Spec.Requests, callback)

			By("Check child allocations")
			testFunc(child, nil, nil, child.Spec.Requests, nil)
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

	var cidr *commonv1alpha1.CIDR
	if cidrStr != "" {
		prefix, err := netaddr.ParseIPPrefix(cidrStr)
		Expect(err).ToNot(HaveOccurred())
		c := commonv1alpha1.NewCIDR(prefix)
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
		spec.CIDRs = []commonv1alpha1.CIDR{*cidr}
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
