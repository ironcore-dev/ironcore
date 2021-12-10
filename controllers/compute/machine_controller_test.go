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

package compute

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	networkv1alpha1 "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
)

var _ = Describe("machine controller", func() {
	ns := SetupTest(ctx)

	It("should delete unused IPAMRanges for deleted interfaces", func() {
		By("creating a subnet")
		subnet := &networkv1alpha1.Subnet{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "subnet-",
			},
			Spec: networkv1alpha1.SubnetSpec{
				Ranges: []networkv1alpha1.RangeType{
					{
						CIDR: commonv1alpha1.PtrToCIDR(commonv1alpha1.MustParseCIDR("192.168.0.0/24")),
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, subnet)).To(Succeed())

		By("creating a machine")
		iface1 := computev1alpha1.Interface{Name: "test-if1", Target: corev1.LocalObjectReference{Name: subnet.Name}}
		iface2 := computev1alpha1.Interface{Name: "test-if2", Target: corev1.LocalObjectReference{Name: subnet.Name}}
		machine := &computev1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "interfaces-",
				Namespace:    ns.Name,
			},
			Spec: computev1alpha1.MachineSpec{
				Interfaces: []computev1alpha1.Interface{iface1, iface2},
			},
		}
		Expect(k8sClient.Create(ctx, machine)).To(Succeed())

		By("waiting for the corresponding IPAMRanges to be created")
		ipamRange1 := &networkv1alpha1.IPAMRange{}
		ipamRangeKey1 := client.ObjectKey{
			Namespace: ns.Name,
			Name:      computev1alpha1.MachineInterfaceIPAMRangeName(machine.Name, iface1.Name),
		}
		ipamRange2 := &networkv1alpha1.IPAMRange{}
		ipamRangeKey2 := client.ObjectKey{
			Namespace: ns.Name,
			Name:      computev1alpha1.MachineInterfaceIPAMRangeName(machine.Name, iface2.Name),
		}
		Eventually(func(g Gomega) {
			err := k8sClient.Get(ctx, ipamRangeKey1, ipamRange1)
			Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
			g.Expect(err).NotTo(HaveOccurred())

			err = k8sClient.Get(ctx, ipamRangeKey2, ipamRange2)
			Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
			g.Expect(err).NotTo(HaveOccurred())
		}, timeout, interval).Should(Succeed())

		By("removing the first interface from the machine")
		machineBase := machine.DeepCopy()
		machine.Spec.Interfaces = []computev1alpha1.Interface{iface2}
		Expect(k8sClient.Patch(ctx, machine, client.MergeFrom(machineBase))).To(Succeed())

		By("waiting for the corresponding IPAMRange of the interface to be gone")
		Eventually(func(g Gomega) {
			err := k8sClient.Get(ctx, ipamRangeKey1, ipamRange1)
			g.Expect(errors.IsNotFound(err)).To(BeTrue(), "expected not found but got %v", err)

			err = k8sClient.Get(ctx, ipamRangeKey2, ipamRange2)
			g.Expect(err).NotTo(HaveOccurred())
		}, timeout, interval).Should(Succeed())
	})

	It("should reconcile a machine without any interface", func() {
		By("creating a machine")
		machine := &computev1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "machine-",
			},
		}
		Expect(k8sClient.Create(ctx, machine)).To(Succeed())

		By("asserting the interfaces in the machine's status stay empty")
		key := client.ObjectKeyFromObject(machine)
		Consistently(func() []computev1alpha1.InterfaceStatus {
			Expect(k8sClient.Get(ctx, key, machine)).To(Succeed())
			return machine.Status.Interfaces
		}, timeout, interval).Should(BeEmpty())
	})

	It("reconciles a machine owning interfaces with IP", func() {
		By("creating a subnet")
		subnet := &networkv1alpha1.Subnet{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "subnet-",
			},
			Spec: networkv1alpha1.SubnetSpec{
				Ranges: []networkv1alpha1.RangeType{
					{
						CIDR: commonv1alpha1.PtrToCIDR(commonv1alpha1.MustParseCIDR("192.168.0.0/24")),
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, subnet)).To(Succeed())

		By("creating a machine")
		iface1 := computev1alpha1.Interface{
			Name:     "iface-1",
			IP:       commonv1alpha1.PtrToIPAddr(commonv1alpha1.MustParseIPAddr("192.168.0.0")),
			Priority: 0,
			Target:   corev1.LocalObjectReference{Name: subnet.Name},
		}
		iface2 := computev1alpha1.Interface{
			Name:     "iface-2",
			IP:       commonv1alpha1.PtrToIPAddr(commonv1alpha1.MustParseIPAddr("192.168.0.1")),
			Priority: 1,
			Target:   corev1.LocalObjectReference{Name: subnet.Name},
		}
		machine := &computev1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "machine-",
			},
			Spec: computev1alpha1.MachineSpec{
				Interfaces: []computev1alpha1.Interface{iface1, iface2},
			},
		}
		Expect(k8sClient.Create(ctx, machine)).To(Succeed())

		By("waiting for the IPAMRanges of the interfaces to be ready")
		ipamRange1 := &networkv1alpha1.IPAMRange{}
		ipamRangeKey1 := client.ObjectKey{Namespace: ns.Name, Name: computev1alpha1.MachineInterfaceIPAMRangeName(machine.Name, iface1.Name)}
		ipamRange2 := &networkv1alpha1.IPAMRange{}
		ipamRangeKey2 := client.ObjectKey{Namespace: ns.Name, Name: computev1alpha1.MachineInterfaceIPAMRangeName(machine.Name, iface2.Name)}
		Eventually(func(g Gomega) {
			err := k8sClient.Get(ctx, ipamRangeKey1, ipamRange1)
			Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(ipamRange1.Spec.Parent.Name).To(Equal(networkv1alpha1.SubnetIPAMName(iface1.Target.Name)))
			g.Expect(ipamRange1.Spec.Items).To(ConsistOf(networkv1alpha1.IPAMRangeItem{
				IPs: commonv1alpha1.PtrToIPRange(commonv1alpha1.IPRangeFrom(*iface1.IP, *iface1.IP)),
			}))

			err = k8sClient.Get(ctx, ipamRangeKey2, ipamRange2)
			Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(ipamRange2.Spec.Parent.Name).To(Equal(networkv1alpha1.SubnetIPAMName(iface2.Target.Name)))
			g.Expect(ipamRange2.Spec.Items).To(ConsistOf(networkv1alpha1.IPAMRangeItem{
				IPs: commonv1alpha1.PtrToIPRange(commonv1alpha1.IPRangeFrom(*iface2.IP, *iface2.IP)),
			}))
		}, timeout, interval).Should(Succeed())

		By("waiting for the machine's status to be updated")
		machineKey := client.ObjectKeyFromObject(machine)
		Eventually(func() []computev1alpha1.InterfaceStatus {
			Expect(k8sClient.Get(ctx, machineKey, machine)).To(Succeed())
			return machine.Status.Interfaces
		}, timeout, interval).Should(ConsistOf(
			computev1alpha1.InterfaceStatus{
				Name:     "iface-1",
				IP:       commonv1alpha1.MustParseIPAddr("192.168.0.0"),
				Priority: 0,
			},
			computev1alpha1.InterfaceStatus{
				Name:     "iface-2",
				IP:       commonv1alpha1.MustParseIPAddr("192.168.0.1"),
				Priority: 1,
			},
		))
	})

	It("should reconcile machines having interfaces with no IP", func() {
		By("creating a subnet")
		subnet := &networkv1alpha1.Subnet{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "subnet-",
			},
			Spec: networkv1alpha1.SubnetSpec{
				Ranges: []networkv1alpha1.RangeType{
					{
						CIDR: commonv1alpha1.PtrToCIDR(commonv1alpha1.MustParseCIDR("192.168.0.0/24")),
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, subnet)).To(Succeed())

		By("creating a machine")
		iface1 := computev1alpha1.Interface{
			Name:   "iface-1",
			Target: corev1.LocalObjectReference{Name: subnet.Name},
		}
		iface2 := computev1alpha1.Interface{
			Name:   "iface-2",
			Target: corev1.LocalObjectReference{Name: subnet.Name},
		}
		machine := &computev1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "machine-",
			},
			Spec: computev1alpha1.MachineSpec{
				Interfaces: []computev1alpha1.Interface{iface1, iface2},
			},
		}
		Expect(k8sClient.Create(ctx, machine)).To(Succeed())

		By("waiting for the IPAMRanges to be created and up-to-date")
		ipamRange1 := &networkv1alpha1.IPAMRange{}
		ipamRangeKey1 := client.ObjectKey{Namespace: machine.Namespace, Name: computev1alpha1.MachineInterfaceIPAMRangeName(machine.Name, iface1.Name)}
		var ifaceIP1 commonv1alpha1.IPAddr
		ipamRange2 := &networkv1alpha1.IPAMRange{}
		ipamRangeKey2 := client.ObjectKey{Namespace: machine.Namespace, Name: computev1alpha1.MachineInterfaceIPAMRangeName(machine.Name, iface2.Name)}
		var ifaceIP2 commonv1alpha1.IPAddr

		Eventually(func(g Gomega) {
			By("getting the first IPAMRange")
			err := k8sClient.Get(ctx, ipamRangeKey1, ipamRange1)
			Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
			g.Expect(err).NotTo(HaveOccurred())

			By("verifying the spec of the first IPAMRange")
			g.Expect(ipamRange1.Spec).To(Equal(networkv1alpha1.IPAMRangeSpec{
				Parent: &corev1.LocalObjectReference{
					Name: networkv1alpha1.SubnetIPAMName(subnet.Name),
				},
				Items: []networkv1alpha1.IPAMRangeItem{
					{
						IPCount: 1,
					},
				},
			}))

			By("verifying the first IPAMRange has allocated an IP")
			g.Expect(ipamRange1.Status.Allocations).To(HaveLen(1))
			allocation := ipamRange1.Status.Allocations[0]
			g.Expect(allocation.IPs).NotTo(BeNil())
			ifaceIP1 = allocation.IPs.From

			By("getting the second IPAMRange")
			err = k8sClient.Get(ctx, ipamRangeKey2, ipamRange2)
			Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
			g.Expect(err).NotTo(HaveOccurred())

			By("verifying the spec of the second IPAMRange")
			g.Expect(ipamRange2.Spec).To(Equal(networkv1alpha1.IPAMRangeSpec{
				Parent: &corev1.LocalObjectReference{
					Name: networkv1alpha1.SubnetIPAMName(subnet.Name),
				},
				Items: []networkv1alpha1.IPAMRangeItem{
					{
						IPCount: 1,
					},
				},
			}))

			By("verifying the second IPAMRange has allocated an IP")
			g.Expect(ipamRange2.Status.Allocations).To(HaveLen(1))
			allocation = ipamRange2.Status.Allocations[0]
			g.Expect(allocation.IPs).NotTo(BeNil())
			ifaceIP2 = allocation.IPs.From
		}, timeout, interval).Should(Succeed())

		By("waiting for the machine status interfaces to contain the IPs")
		machineKey := client.ObjectKeyFromObject(machine)
		Eventually(func(g Gomega) []computev1alpha1.InterfaceStatus {
			Expect(k8sClient.Get(ctx, machineKey, machine)).To(Succeed())

			return machine.Status.Interfaces
		}, timeout, interval).Should(ConsistOf(
			computev1alpha1.InterfaceStatus{
				Name:     "iface-1",
				IP:       ifaceIP1,
				Priority: 0,
			},
			computev1alpha1.InterfaceStatus{
				Name:     "iface-2",
				IP:       ifaceIP2,
				Priority: 0,
			},
		))
	})
})
