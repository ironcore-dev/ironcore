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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	networkv1alpha1 "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
)

var _ = Describe("ReservedIPReconciler", func() {
	ns := SetupTest()
	PIt("should reserve an IP", func() {
		By("creating a subnet")
		cidr := commonv1alpha1.MustParseIPPrefix("10.0.0.0/16")
		subnet := &networkv1alpha1.Subnet{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "reserved-ip-",
			},
			Spec: networkv1alpha1.SubnetSpec{
				Ranges: []networkv1alpha1.RangeType{
					{
						CIDR: &cidr,
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, subnet)).To(Succeed())

		By("creating a reserved ip")
		ip := commonv1alpha1.MustParseIP("10.0.0.1")
		reservedIP := &networkv1alpha1.ReservedIP{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-",
			},
			Spec: networkv1alpha1.ReservedIPSpec{
				Subnet: corev1.LocalObjectReference{Name: subnet.Name},
				IP:     commonv1alpha1.PtrToIP(ip),
			},
		}
		Expect(k8sClient.Create(ctx, reservedIP)).To(Succeed())

		By("waiting for the ipam range to be available")
		Eventually(func(g Gomega) {
			ipamRangeKey := client.ObjectKey{Namespace: ns.Name, Name: networkv1alpha1.ReservedIPIPAMName(reservedIP.Name)}
			ipamRange := &networkv1alpha1.IPAMRange{}
			err := k8sClient.Get(ctx, ipamRangeKey, ipamRange)
			Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(ipamRange.Spec.Requests).To(ConsistOf(networkv1alpha1.IPAMRangeRequest{
				IPs: commonv1alpha1.PtrToIPRange(commonv1alpha1.IPRangeFrom(ip, ip)),
			}))
		}, timeout, interval).Should(Succeed())

		By("waiting for the reserved ip to report the allocated ip")
		reservedIPKey := client.ObjectKeyFromObject(reservedIP)
		Eventually(func(g Gomega) {
			err := k8sClient.Get(ctx, reservedIPKey, reservedIP)
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(reservedIP.Status).To(MatchFields(IgnoreExtras|IgnoreMissing, Fields{
				"State": Equal(networkv1alpha1.ReservedIPStateReady),
				"IP":    PointTo(Equal(ip)),
			}))
		}, timeout, interval).Should(Succeed())
	})
})
