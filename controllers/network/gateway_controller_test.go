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
	"inet.af/netaddr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"

	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	networkv1alpha1 "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
)

var _ = Describe("gateway controller", func() {
	ns := SetupTest()
	Context("reconciling creation", func() {
		It("sets the ControllerReference of the corresponding IPAMRange to the Gateway", func() {
			parsedPrefix, err := netaddr.ParseIPPrefix("192.168.0.0/24")
			Expect(err).ToNot(HaveOccurred())

			subnet := &networkv1alpha1.Subnet{
				TypeMeta: metav1.TypeMeta{
					APIVersion: networkv1alpha1.GroupVersion.String(),
					Kind:       networkv1alpha1.SubnetGK.Kind,
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    ns.Name,
					GenerateName: "subnet-",
				},
				Spec: networkv1alpha1.SubnetSpec{
					Ranges: []networkv1alpha1.RangeType{
						{CIDR: &commonv1alpha1.CIDR{IPPrefix: parsedPrefix}},
					},
				},
			}
			Expect(k8sClient.Create(ctx, subnet)).To(Succeed())

			gw := &networkv1alpha1.Gateway{
				TypeMeta: metav1.TypeMeta{
					APIVersion: networkv1alpha1.GroupVersion.String(),
					Kind:       "Gateway",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    ns.Name,
					GenerateName: "gateway-",
				},
				Spec: networkv1alpha1.GatewaySpec{
					Subnet: v1.LocalObjectReference{Name: subnet.Name},
				},
			}
			Expect(k8sClient.Create(ctx, gw)).To(Succeed())

			ipamRange := newIPAMRangeFromGateway(gw)
			ipamRangeKey := client.ObjectKeyFromObject(ipamRange)
			Eventually(func(g Gomega) {
				err := k8sClient.Get(ctx, ipamRangeKey, ipamRange)
				Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
				g.Expect(err).NotTo(HaveOccurred())

				g.Expect(ipamRange.OwnerReferences).To(ContainElement(metav1.OwnerReference{
					APIVersion:         networkv1alpha1.GroupVersion.String(),
					Kind:               "Gateway",
					Name:               gw.Name,
					UID:                gw.UID,
					BlockOwnerDeletion: pointer.BoolPtr(true),
					Controller:         pointer.BoolPtr(true),
				}))
			}, timeout, interval).Should(Succeed())
		})

		It("creates the corresponding IPAMRange", func() {
			parsedPrefix, err := netaddr.ParseIPPrefix("192.168.0.0/24")
			Expect(err).ToNot(HaveOccurred())

			subnet := &networkv1alpha1.Subnet{
				TypeMeta: metav1.TypeMeta{
					APIVersion: networkv1alpha1.GroupVersion.String(),
					Kind:       networkv1alpha1.SubnetGK.Kind,
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    ns.Name,
					GenerateName: "subnet-",
				},
				Spec: networkv1alpha1.SubnetSpec{
					Ranges: []networkv1alpha1.RangeType{
						{CIDR: &commonv1alpha1.CIDR{IPPrefix: parsedPrefix}},
					},
				},
			}
			Expect(k8sClient.Create(ctx, subnet)).To(Succeed())

			gw := &networkv1alpha1.Gateway{
				TypeMeta: metav1.TypeMeta{
					APIVersion: networkv1alpha1.GroupVersion.String(),
					Kind:       "Gateway",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    ns.Name,
					GenerateName: "gateway-",
				},
				Spec: networkv1alpha1.GatewaySpec{
					Subnet: v1.LocalObjectReference{Name: subnet.Name},
				},
			}
			Expect(k8sClient.Create(ctx, gw)).To(Succeed())

			ipamRange := newIPAMRangeFromGateway(gw)
			ipamRangeKey := client.ObjectKeyFromObject(ipamRange)
			Eventually(func() error {
				err := k8sClient.Get(ctx, ipamRangeKey, ipamRange)
				Expect(client.IgnoreNotFound(err)).To(Succeed())
				return err
			}, timeout, interval).Should(Succeed())
		})

		It("updates the status of the Gateway", func() {
			parsedPrefix, err := netaddr.ParseIPPrefix("192.168.0.0/24")
			Expect(err).ToNot(HaveOccurred())

			subnet := &networkv1alpha1.Subnet{
				TypeMeta: metav1.TypeMeta{
					APIVersion: networkv1alpha1.GroupVersion.String(),
					Kind:       networkv1alpha1.SubnetGK.Kind,
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    ns.Name,
					GenerateName: "subnet-",
				},
				Spec: networkv1alpha1.SubnetSpec{
					Ranges: []networkv1alpha1.RangeType{
						{CIDR: &commonv1alpha1.CIDR{IPPrefix: parsedPrefix}},
					},
				},
			}
			Expect(k8sClient.Create(ctx, subnet)).To(Succeed())

			gw := &networkv1alpha1.Gateway{
				TypeMeta: metav1.TypeMeta{
					APIVersion: networkv1alpha1.GroupVersion.String(),
					Kind:       "Gateway",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    ns.Name,
					GenerateName: "gateway-",
				},
				Spec: networkv1alpha1.GatewaySpec{
					Subnet: v1.LocalObjectReference{Name: subnet.Name},
				},
			}
			Expect(k8sClient.Create(ctx, gw)).To(Succeed())

			gwKey := client.ObjectKeyFromObject(gw)
			Eventually(func() []commonv1alpha1.IPAddr {
				err := k8sClient.Get(ctx, gwKey, gw)
				Expect(client.IgnoreNotFound(err)).To(Succeed())
				return gw.Status.IPs
			}, timeout, interval).ShouldNot(BeEmpty())

			subnetIPRange := subnet.Spec.Ranges[0].CIDR.IPPrefix.Range()
			gatewayIP := gw.Status.IPs[0].IP
			Expect(subnetIPRange.Contains(gatewayIP)).To(BeTrue(), "The Subnet IP range contains the IP.")
		})
	})
})
