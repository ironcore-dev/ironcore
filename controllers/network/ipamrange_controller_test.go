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
	"reflect"
	"time"

	"k8s.io/apimachinery/pkg/types"

	corev1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"inet.af/netaddr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

		It("should reconcile parent without children", func() {
			parent := createParentIPAMRange(ctx, ns)
			validateAllocations(
				ctx,
				parent,
				map[string]networkv1alpha1.IPAMRangeAllocationState{
					parentCIDR: networkv1alpha1.IPAMRangeAllocationFree,
				},
				nil,
				nil,
			)
		})
		It("should update parent allocations with CIDR request", func() {
			parent := createParentIPAMRange(ctx, ns)

			meta := metav1.ObjectMeta{
				Name:      "child",
				Namespace: ns.Name,
			}
			child := createIPAMRange(ctx, meta, "192.168.1.0/25", parent.Name, nil, 0, 0)
			validateAllocations(
				ctx,
				parent,
				map[string]networkv1alpha1.IPAMRangeAllocationState{
					"192.168.1.0/25":   networkv1alpha1.IPAMRangeAllocationUsed,
					"192.168.1.128/25": networkv1alpha1.IPAMRangeAllocationFree,
				},
				nil,
				nil,
			)
			validateAllocations(
				ctx,
				child,
				map[string]networkv1alpha1.IPAMRangeAllocationState{
					"192.168.1.0/25": networkv1alpha1.IPAMRangeAllocationFree,
				},
				nil,
				nil,
			)
		})
		It("should update parent allocations with size request", func() {
			parent := createParentIPAMRange(ctx, ns)

			meta := metav1.ObjectMeta{
				Name:      "child",
				Namespace: ns.Name,
			}
			child := createIPAMRange(ctx, meta, "", parent.Name, nil, 25, 0)
			validateAllocations(
				ctx,
				parent,
				map[string]networkv1alpha1.IPAMRangeAllocationState{
					"192.168.1.0/25":   networkv1alpha1.IPAMRangeAllocationUsed,
					"192.168.1.128/25": networkv1alpha1.IPAMRangeAllocationFree,
				},
				nil,
				nil,
			)
			validateAllocations(
				ctx,
				child,
				map[string]networkv1alpha1.IPAMRangeAllocationState{
					"192.168.1.0/25": networkv1alpha1.IPAMRangeAllocationFree,
				},
				nil,
				nil,
			)
		})
		It("should update parent allocations with ip request", func() {
			parent := createParentIPAMRange(ctx, ns)

			meta := metav1.ObjectMeta{
				Name:      "child",
				Namespace: ns.Name,
			}
			fromIP, _ := netaddr.ParseIP("192.168.1.0")
			toIP, _ := netaddr.ParseIP("192.168.1.127")
			ipRange := &commonv1alpha1.IPRange{
				From: commonv1alpha1.NewIPAddr(fromIP),
				To:   commonv1alpha1.NewIPAddr(toIP),
			}
			child := createIPAMRange(ctx, meta, "", parent.Name, ipRange, 0, 0)
			validateAllocations(
				ctx,
				parent,
				map[string]networkv1alpha1.IPAMRangeAllocationState{
					"192.168.1.128/25": networkv1alpha1.IPAMRangeAllocationFree,
				},
				nil,
				ipRange,
			)
			validateAllocations(
				ctx,
				child,
				nil,
				nil,
				ipRange,
			)
		})

		// TODO: currently it's possible to request only single IP address with IPCount
		// TODO: so 2 tests below are commented
		//It("should update parent allocations with ip count request", func() {
		//	parent := createParentIPAMRange(ctx, ns)
		//
		//	meta := metav1.ObjectMeta{
		//		Name:      "child",
		//		Namespace: ns.Name,
		//	}
		//	child := createIPAMRange(ctx, meta, "", parent.Name, nil, 0, 128)
		//	validateAllocations(
		//		ctx,
		//		parent,
		//		map[string]networkv1alpha1.IPAMRangeAllocationState{
		//			"192.168.1.0/25":   networkv1alpha1.IPAMRangeAllocationUsed,
		//			"192.168.1.128/25": networkv1alpha1.IPAMRangeAllocationFree,
		//		},
		//		nil,
		//	)
		//	validateAllocations(
		//		ctx,
		//		child,
		//		map[string]networkv1alpha1.IPAMRangeAllocationState{
		//			"192.168.1.0/25": networkv1alpha1.IPAMRangeAllocationFree,
		//		},
		//		nil,
		//	)
		//})
		//It("allocations should fail if ip count is too big", func() {})

		It("allocation should fail if CIDR is out of range", func() {
			parent := createParentIPAMRange(ctx, ns)

			meta := metav1.ObjectMeta{
				Name:      "child",
				Namespace: ns.Name,
			}
			child := createIPAMRange(ctx, meta, "192.168.2.0/25", parent.Name, nil, 0, 0)
			validateAllocations(
				ctx,
				parent,
				map[string]networkv1alpha1.IPAMRangeAllocationState{
					parentCIDR: networkv1alpha1.IPAMRangeAllocationFree,
				},
				child.Spec.Requests,
				nil,
			)
			validateAllocations(
				ctx,
				child,
				nil,
				child.Spec.Requests,
				nil,
			)
		})
		It("allocation should fail if size is too big", func() {
			parent := createParentIPAMRange(ctx, ns)

			meta := metav1.ObjectMeta{
				Name:      "child",
				Namespace: ns.Name,
			}
			child := createIPAMRange(ctx, meta, "", parent.Name, nil, 23, 0)
			validateAllocations(
				ctx,
				parent,
				map[string]networkv1alpha1.IPAMRangeAllocationState{
					parentCIDR: networkv1alpha1.IPAMRangeAllocationFree,
				},
				child.Spec.Requests,
				nil,
			)
			validateAllocations(
				ctx,
				child,
				nil,
				child.Spec.Requests,
				nil,
			)
		})
		It("allocation should fail if ip is out of range", func() {
			parent := createParentIPAMRange(ctx, ns)

			meta := metav1.ObjectMeta{
				Name:      "child",
				Namespace: ns.Name,
			}
			fromIP, _ := netaddr.ParseIP("192.168.2.0")
			toIP, _ := netaddr.ParseIP("192.168.2.127")
			ipRange := &commonv1alpha1.IPRange{
				From: commonv1alpha1.NewIPAddr(fromIP),
				To:   commonv1alpha1.NewIPAddr(toIP),
			}
			child := createIPAMRange(ctx, meta, "", parent.Name, ipRange, 0, 0)
			validateAllocations(
				ctx,
				parent,
				map[string]networkv1alpha1.IPAMRangeAllocationState{
					parentCIDR: networkv1alpha1.IPAMRangeAllocationFree,
				},
				child.Spec.Requests,
				nil,
			)
			validateAllocations(
				ctx,
				child,
				nil,
				child.Spec.Requests,
				nil,
			)
		})

		//It("should update allocations when CIDR is changed", func(){})
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

func validateAllocations(
	ctx context.Context,
	obj *networkv1alpha1.IPAMRange,
	cidrToState map[string]networkv1alpha1.IPAMRangeAllocationState,
	failedRequests []networkv1alpha1.IPAMRangeRequest,
	ipRange *commonv1alpha1.IPRange,
) {
	Eventually(func() bool {
		key := types.NamespacedName{Name: obj.Name, Namespace: obj.Namespace}
		freshObj := &networkv1alpha1.IPAMRange{}
		if err := k8sClient.Get(ctx, key, freshObj); err != nil {
			return false
		}

		type check struct {
			state networkv1alpha1.IPAMRangeAllocationState
			valid bool
		}
		cidrCL := make(map[string]check)
		for cidr, state := range cidrToState {
			cidrCL[cidr] = check{
				state: state,
				valid: false,
			}
		}
		failedRequestsCL := make(map[string]check)
		for _, r := range failedRequests {
			failedRequestsCL[getRequestKey(r)] = check{
				state: networkv1alpha1.IPAMRangeAllocationFailed,
				valid: false,
			}
		}

		for _, alloc := range freshObj.Status.Allocations {
			if alloc.IPs != nil {
				if !reflect.DeepEqual(alloc.IPs, ipRange) {
					return false
				}
			}
			if alloc.CIDR != nil {
				if s, ok := cidrCL[alloc.CIDR.String()]; ok {
					if s.state == alloc.State {
						s.valid = true
						cidrCL[alloc.CIDR.String()] = s
					}
				} else {
					return false
				}
			}
			if alloc.State == networkv1alpha1.IPAMRangeAllocationFailed {
				k := getRequestKey(*alloc.Request)
				if s, ok := failedRequestsCL[k]; ok {
					if s.state == alloc.State {
						s.valid = true
						failedRequestsCL[k] = s
					}
				} else {
					return false
				}
			}
		}

		for _, v := range cidrCL {
			if !v.valid {
				return false
			}
		}
		for _, v := range failedRequestsCL {
			if !v.valid {
				return false
			}
		}

		return true
	}, timeout, interval).Should(BeTrue())
}

func getRequestKey(req networkv1alpha1.IPAMRangeRequest) (key string) {
	if req.IPs != nil {
		key = req.IPs.From.String() + "-" + req.IPs.To.String()
	} else {
		key = fmt.Sprint(req)
	}

	return
}
