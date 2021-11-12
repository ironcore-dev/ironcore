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
	parentsubCIDR = "10.0.0.0/16"
)

var _ = Describe("ReservedIPReconciler", func() {
	Context("Reconcile an ReservedIP", func() {
		timeinterval := time.Millisecond * 500
		ns := SetupTest()
		It("should create reserved ip instance", func() {
			subnet := createRootSubnet(context.Background(), ns)
			Eventually(func(g Gomega) {
				key := types.NamespacedName{Name: subnet.Name, Namespace: subnet.Namespace}
				obj := &networkv1alpha1.Subnet{}
				g.Expect(k8sClient.Get(ctx, key, obj)).Should(Succeed())
			}, timeout, interval).Should(Succeed())
			ip, _ := netaddr.ParseIP("10.0.0.1")
			resIP := commonv1alpha1.IPAddr{
				IP: ip,
			}
			reservedip := &networkv1alpha1.ReservedIP{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: ns.Name,
					Name:      "reservediptest",
				},
				Spec: networkv1alpha1.ReservedIPSpec{
					Subnet: corev1.LocalObjectReference{Name: subnet.Name},
					IP:     commonv1alpha1.IPAddr{IP: ip},
				},
			}
			Expect(k8sClient.Create(ctx, reservedip)).To(Succeed())
			Eventually(func(g Gomega) {
				key := types.NamespacedName{Name: fmt.Sprintf("reservedip-subnet-%s-%s", reservedip.Name, subnet.Name), Namespace: ns.Name}
				obj := &networkv1alpha1.IPAMRange{}
				g.Expect(k8sClient.Get(ctx, key, obj)).Should(Succeed())
				g.Expect(obj.Status).NotTo(BeNil())
				g.Expect(validateAllocatedIP(obj, resIP, networkv1alpha1.IPAMRangeAllocationFree)).To(BeTrue())
				key2 := types.NamespacedName{Name: "reservediptest", Namespace: ns.Name}
				obj2 := &networkv1alpha1.ReservedIP{}
				g.Expect(k8sClient.Get(ctx, key2, obj2)).Should(Succeed())
				g.Expect(obj2.Status.IP).NotTo(BeNil())
				g.Expect(obj2.Status.IP.IP).NotTo(BeNil())
				g.Expect(obj2.Status.IP.IP.String()).To(Equal("10.0.0.1"))
				g.Expect(obj2.Status.State).To(Equal(networkv1alpha1.ReservedIPStateReady))
			}, timeout, timeinterval).Should(Succeed())
		})
	})
})

func createRootSubnet(ctx context.Context, ns *corev1.Namespace) *networkv1alpha1.Subnet {
	meta := metav1.ObjectMeta{
		Name:      "rootsubnet",
		Namespace: ns.Name,
	}
	return createSubnet(ctx, parentsubCIDR, meta)
}

func createSubnet(ctx context.Context, cidrStr string, meta metav1.ObjectMeta) *networkv1alpha1.Subnet {
	spec := networkv1alpha1.SubnetSpec{}
	var cidr commonv1alpha1.CIDR
	if cidrStr != "" {
		prefix, err := netaddr.ParseIPPrefix(cidrStr)
		Expect(err).ToNot(HaveOccurred())
		cidr = commonv1alpha1.NewCIDR(prefix)
	}
	spec.RoutingDomain.Name = "routingdomain-sample"
	spec.Ranges = []networkv1alpha1.RangeType{
		{
			Size: 1,
			CIDR: cidr,
		},
	}
	instance := &networkv1alpha1.Subnet{
		Spec:       spec,
		ObjectMeta: meta,
	}
	Expect(k8sClient.Create(ctx, instance)).To(Succeed())
	return instance
}

func validateAllocatedIP(obj *networkv1alpha1.IPAMRange, ip commonv1alpha1.IPAddr, expState networkv1alpha1.IPAMRangeAllocationState) bool {
	for _, alloc := range obj.Status.Allocations {
		if alloc.IPs.From == ip && alloc.State == expState {
			return true
		}
	}
	return false
}
