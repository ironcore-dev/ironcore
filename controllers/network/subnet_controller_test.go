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
	. "github.com/onsi/gomega/gstruct"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"

	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	networkv1alpha1 "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
)

var _ = Describe("subnet controller", func() {
	ns := SetupTest()

	const cidrAddress = "192.168.0.0"
	testCIDR := commonv1alpha1.MustParseIPPrefix(fmt.Sprintf("%s/24", cidrAddress))

	It("sets the owner Subnet as a Controller OwnerReference on the controlled IPAMRange", func() {
		By("creating a subnet")
		subnet := &networkv1alpha1.Subnet{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-",
			},
			Spec: networkv1alpha1.SubnetSpec{
				Ranges: []networkv1alpha1.RangeType{
					{
						CIDR: &testCIDR,
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, subnet)).Should(Succeed())

		By("waiting for the ipam range to exist and contain the controller reference")
		ipamRangeKey := client.ObjectKey{Namespace: subnet.Namespace, Name: networkv1alpha1.SubnetIPAMName(subnet.Name)}
		ipamRange := &networkv1alpha1.IPAMRange{}
		Eventually(func(g Gomega) {
			err := k8sClient.Get(ctx, ipamRangeKey, ipamRange)
			Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(ipamRange.OwnerReferences).To(ContainElement(metav1.OwnerReference{
				APIVersion:         networkv1alpha1.GroupVersion.String(),
				Kind:               networkv1alpha1.SubnetGK.Kind,
				Name:               subnet.Name,
				UID:                subnet.UID,
				BlockOwnerDeletion: pointer.BoolPtr(true),
				Controller:         pointer.BoolPtr(true),
			}))
		}, timeout, interval).Should(Succeed())
	})

	It("reconciles a Subnet without parent", func() {
		By("creating a subnet")
		subnet := &networkv1alpha1.Subnet{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-",
			},
			Spec: networkv1alpha1.SubnetSpec{
				Ranges: []networkv1alpha1.RangeType{
					{
						CIDR: &testCIDR,
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, subnet)).Should(Succeed())

		By("waiting for the status of the ipam range to have an allocated CIDR")
		ipamRangeKey := client.ObjectKey{Namespace: subnet.Namespace, Name: networkv1alpha1.SubnetIPAMName(subnet.Name)}
		ipamRange := &networkv1alpha1.IPAMRange{}
		Eventually(func(g Gomega) []commonv1alpha1.IPPrefix {
			err := k8sClient.Get(ctx, ipamRangeKey, ipamRange)
			Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
			g.Expect(err).NotTo(HaveOccurred())
			return ipamRange.Spec.CIDRs
		}, timeout, interval).Should(ContainElement(testCIDR))

		By("waiting for the state of the subnet to be up")
		subnetKey := client.ObjectKeyFromObject(subnet)
		Eventually(func() networkv1alpha1.SubnetState {
			Expect(k8sClient.Get(ctx, subnetKey, subnet)).Should(Succeed())
			return subnet.Status.State
		}, timeout, interval).Should(Equal(networkv1alpha1.SubnetStateUp))
	})

	It("reconciles a subnet with parent", func() {
		By("creating a parent subnet")
		parentSubnet := &networkv1alpha1.Subnet{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-parent-",
			},
			Spec: networkv1alpha1.SubnetSpec{
				Ranges: []networkv1alpha1.RangeType{
					{
						CIDR: &testCIDR,
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, parentSubnet)).Should(Succeed())

		By("creating a child subnet")
		rangeSize := 28
		childSubnet := &networkv1alpha1.Subnet{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-child-",
			},
			Spec: networkv1alpha1.SubnetSpec{
				Parent: &corev1.LocalObjectReference{
					Name: parentSubnet.Name,
				},
				Ranges: []networkv1alpha1.RangeType{
					{
						Size: int32(rangeSize),
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, childSubnet)).Should(Succeed())

		By("waiting for the child ipam range to be created and have the correct requests in spec")
		ipamRangeKey := client.ObjectKey{Namespace: childSubnet.Namespace, Name: networkv1alpha1.SubnetIPAMName(childSubnet.Name)}
		ipamRange := &networkv1alpha1.IPAMRange{}
		Eventually(func(g Gomega) []networkv1alpha1.IPAMRangeRequest {
			err := k8sClient.Get(ctx, ipamRangeKey, ipamRange)
			Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
			g.Expect(err).NotTo(HaveOccurred())
			return ipamRange.Spec.Requests
		}, timeout, interval).Should(ContainElement(networkv1alpha1.IPAMRangeRequest{
			Size: int32(rangeSize),
		}))

		By("waiting for the status of the Subnet to be up")
		parsedCIDR := commonv1alpha1.MustParseIPPrefix(fmt.Sprintf("%s/%d", cidrAddress, rangeSize))
		childSubnetKey := client.ObjectKeyFromObject(childSubnet)
		Eventually(func() networkv1alpha1.SubnetStatus {
			Expect(k8sClient.Get(ctx, childSubnetKey, childSubnet)).Should(Succeed())
			return childSubnet.Status
		}, timeout, interval).Should(MatchFields(IgnoreMissing|IgnoreExtras, Fields{
			"State": Equal(networkv1alpha1.SubnetStateUp),
			"CIDRs": ContainElement(parsedCIDR),
		}))
	})

	It("reconciles two child-subnets with the same parent", func() {
		By("creating a parent-subnet")
		parentSubnet := &networkv1alpha1.Subnet{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "parent-subnet-",
			},
			Spec: networkv1alpha1.SubnetSpec{
				Ranges: []networkv1alpha1.RangeType{
					{
						CIDR: &testCIDR,
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, parentSubnet)).Should(Succeed())

		By("creating the first child-subnet")
		childRangeSize := 30
		firstChildSubnet := &networkv1alpha1.Subnet{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "child-subnet-",
			},
			Spec: networkv1alpha1.SubnetSpec{
				Parent: &corev1.LocalObjectReference{
					Name: parentSubnet.Name,
				},
				Ranges: []networkv1alpha1.RangeType{
					{
						Size: int32(childRangeSize),
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, firstChildSubnet)).Should(Succeed())

		By("creating the second child-subnet")
		secondChildSubnet := &networkv1alpha1.Subnet{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "child-subnet-",
			},
			Spec: networkv1alpha1.SubnetSpec{
				Parent: &corev1.LocalObjectReference{
					Name: parentSubnet.Name,
				},
				Ranges: []networkv1alpha1.RangeType{
					{
						Size: int32(childRangeSize),
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, secondChildSubnet)).Should(Succeed())

		By("waiting for the first child-ipam-range to be created and have the correct requests in spec")
		firstIPAMRangeKey := client.ObjectKey{Namespace: firstChildSubnet.Namespace, Name: networkv1alpha1.SubnetIPAMName(firstChildSubnet.Name)}
		firstIPAMRange := &networkv1alpha1.IPAMRange{}
		Eventually(func(g Gomega) []networkv1alpha1.IPAMRangeRequest {
			err := k8sClient.Get(ctx, firstIPAMRangeKey, firstIPAMRange)
			Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
			g.Expect(err).NotTo(HaveOccurred())
			return firstIPAMRange.Spec.Requests
		}, timeout, interval).Should(ContainElement(networkv1alpha1.IPAMRangeRequest{
			Size: int32(childRangeSize),
		}))

		By("waiting for the second-ipam-range to be created and have the correct requests in spec")
		secondIPAMRangeKey := client.ObjectKey{Namespace: secondChildSubnet.Namespace, Name: networkv1alpha1.SubnetIPAMName(secondChildSubnet.Name)}
		secondIPAMRange := &networkv1alpha1.IPAMRange{}
		Eventually(func(g Gomega) []networkv1alpha1.IPAMRangeRequest {
			err := k8sClient.Get(ctx, secondIPAMRangeKey, secondIPAMRange)
			Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
			g.Expect(err).NotTo(HaveOccurred())
			return secondIPAMRange.Spec.Requests
		}, timeout, interval).Should(ContainElement(networkv1alpha1.IPAMRangeRequest{
			Size: int32(childRangeSize),
		}))

		By("waiting for the status of the first child-subnet to be up")
		firstChildSubnetKey := client.ObjectKeyFromObject(firstChildSubnet)
		Eventually(func(g Gomega) networkv1alpha1.SubnetStatus {
			err := k8sClient.Get(ctx, firstChildSubnetKey, firstChildSubnet)
			Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
			g.Expect(err).ToNot(HaveOccurred())
			return firstChildSubnet.Status
		}, timeout, interval).Should(MatchFields(IgnoreMissing|IgnoreExtras, Fields{
			"State": Equal(networkv1alpha1.SubnetStateUp),
			"CIDRs": Not(BeEmpty()),
		}))

		By("waiting for the status of the second child-subnet to be up")
		secondChildSubnetKey := client.ObjectKeyFromObject(secondChildSubnet)
		Eventually(func(g Gomega) networkv1alpha1.SubnetStatus {
			err := k8sClient.Get(ctx, secondChildSubnetKey, secondChildSubnet)
			Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
			g.Expect(err).ToNot(HaveOccurred())
			return secondChildSubnet.Status
		}, timeout, interval).Should(MatchFields(IgnoreMissing|IgnoreExtras, Fields{
			"State": Equal(networkv1alpha1.SubnetStateUp),
			"CIDRs": Not(BeEmpty()),
		}))

		By("checking if the two child-subnets have non-overlapping CIDRs")
		firstSubenetIPRange := firstChildSubnet.Status.CIDRs[0].IPPrefix.Range()
		secondSubenetIPRange := secondChildSubnet.Status.CIDRs[0].IPPrefix.Range()
		Expect(firstSubenetIPRange.Overlaps(secondSubenetIPRange)).To(BeFalse(), "Two child-subnets should have non-overlapping CIDRs.")

		By("checking if both child-subnets originate from the parent-subnet")
		parentSubnetIPRange := testCIDR.IPPrefix.Range()
		Expect(parentSubnetIPRange.Contains(firstSubenetIPRange.From())).To(BeTrue(), "The parent-subnet contains the start of the child-subnet.")
		Expect(parentSubnetIPRange.Contains(firstSubenetIPRange.To())).To(BeTrue(), "The parent-subnet contains the end of the child-subnet.")
		Expect(parentSubnetIPRange.Contains(secondSubenetIPRange.From())).To(BeTrue(), "The parent-subnet contains the start of the child-subnet.")
		Expect(parentSubnetIPRange.Contains(secondSubenetIPRange.To())).To(BeTrue(), "The parent-subnet contains the end of the child-subnet.")
	})
})
