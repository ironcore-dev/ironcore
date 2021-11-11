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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"inet.af/netaddr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	networkv1alpha1 "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
)

var _ = Describe("ReservedIPReconciler", func() {

	Context("Reconcile an ReservedIP", func() {

		ns := "default"

		It("should create reserved ip instance", func() {
			ip, _ := netaddr.ParseIP("192.168.1.22")
			reservedip := &networkv1alpha1.ReservedIP{

				ObjectMeta: metav1.ObjectMeta{
					Namespace: ns,
					Name:      "reservediptest",
				},
				Spec: networkv1alpha1.ReservedIPSpec{
					Subnet: corev1.LocalObjectReference{Name: "testsubnet"},
					IP:     commonv1alpha1.IPAddr{IP: ip},
				},
			}
			Expect(k8sClient.Create(ctx, reservedip)).To(Succeed())
			Eventually(func(g Gomega) {
				key := types.NamespacedName{Name: fmt.Sprintf("reservedip-subnet-%s-%s", reservedip.Name, reservedip.Spec.Subnet.Name), Namespace: ns}
				obj := &networkv1alpha1.IPAMRange{}
				g.Expect(k8sClient.Get(ctx, key, obj)).Should(Succeed())
			}, timeout, interval).Should(Succeed())
		})

		It("should update reserved ip instance", func() {

			key := types.NamespacedName{Name: "reservediptest", Namespace: ns}
			reservedip := &networkv1alpha1.ReservedIP{}
			k8sClient.Get(ctx, key, reservedip)
			reservedip.Spec.Subnet = corev1.LocalObjectReference{Name: "testsubnet2"}
			Expect(k8sClient.Update(ctx, reservedip)).To(Succeed())

			Eventually(func(g Gomega) {
				key := types.NamespacedName{Name: fmt.Sprintf("reservedip-subnet-%s-%s", reservedip.Name, reservedip.Spec.Subnet.Name), Namespace: ns}
				obj := &networkv1alpha1.IPAMRange{}
				g.Expect(k8sClient.Get(ctx, key, obj)).Should(Succeed())
				g.Expect(obj.Spec.Parent.Name).To(Equal("subnet-testsubnet2"))

			}, timeout, interval).Should(Succeed())
		})
	})
})

func getAllocationState(obj *networkv1alpha1.IPAMRange) map[string]networkv1alpha1.IPAMRangeAllocationState {
	result := make(map[string]networkv1alpha1.IPAMRangeAllocationState)
	for _, alloc := range obj.Status.Allocations {
		if alloc.CIDR != nil {
			result[alloc.CIDR.String()] = alloc.State
		}
	}

	return result
}
