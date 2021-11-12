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
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"inet.af/netaddr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"

	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	networkv1alpha1 "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
)

var _ = Describe("machine controller", func() {
	It("reconciles a machine without interface", func() {
		By("creating the machine")
		m := newMachine()
		Expect(k8sClient.Create(ctx, m)).To(Succeed())

		By("checking if the machine's status gets reconciled")
		key := objectKey(m)
		Consistently(func() []computev1alpha1.InterfaceStatus {
			Expect(k8sClient.Get(ctx, key, m)).To(Succeed())
			return m.Status.Interfaces
		}, timeout, interval).Should(BeEmpty())
	})

	It("reconciles a machine owning interfaces with IP", func() {
		By("creating the subnet")
		subnet := newSubnet(subnetName, "192.168.0.0/24")
		Expect(k8sClient.Create(ctx, subnet)).To(Succeed())

		m := newMachine()
		ifaces := []computev1alpha1.Interface{
			{
				Name:     "iface-0",
				IP:       ip("192.168.0.0"),
				Priority: 0,
				Target:   corev1.LocalObjectReference{Name: subnetName},
			},
			{
				Name:     "iface-1",
				IP:       ip("192.168.0.1"),
				Priority: 1,
				Target:   corev1.LocalObjectReference{Name: subnetName},
			},
		}
		m.Spec.Interfaces = ifaces

		By("creating the machine")
		Expect(k8sClient.Create(ctx, m)).To(Succeed())

		rngs := newIPAMRanges(m)
		for i, rng := range rngs {
			By("fetching the corresponding IPAMRange")
			key := objectKey(rng)
			Eventually(func() error {
				err := k8sClient.Get(ctx, key, rng)

				// Errors other than `not-found` shouldn't exist
				Expect(client.IgnoreNotFound(err)).To(Succeed())
				return err
			}, timeout, interval).Should(Succeed())

			By("checking if the parent of the IPAMRange corresponds to the target of the interface")
			Expect(rng.Spec.Parent.Name).To(Equal(networkv1alpha1.SubnetIPAMName(ifaces[i].Target.Name)))

			By("checking if the request of the IPAMRange corresponds to the IP of the machine's interface")
			Expect(rng.Spec.Requests[0]).To(Equal(ipamRangeRequest(ifaces[i].IP)))

			By("checking the OwnerReferences of the IPAMRange contain the machine")
			Expect(rng.OwnerReferences).To(ContainElement(controllerReference(m)))
		}

		expectedIfaceStatuses := []computev1alpha1.InterfaceStatus{}
		for i, rng := range rngs {
			By("waiting till the IPAMRange gets reconciled")
			key := objectKey(rng)
			Eventually(func() int {
				Expect(k8sClient.Get(ctx, key, rng)).To(Succeed())
				return len(rng.Status.Allocations)
			}, timeout, interval).ShouldNot(Equal(0))

			expectedIfaceStatuses = appendInterfaceStatuses(expectedIfaceStatuses, &ifaces[i], rng)
		}

		By("checking if the machine's status gets reconciled")
		key := objectKey(m)
		Eventually(func() []computev1alpha1.InterfaceStatus {
			Expect(k8sClient.Get(ctx, key, m)).To(Succeed())
			return m.Status.Interfaces
		}, timeout, interval).Should(Equal(expectedIfaceStatuses))
	})

	It("reconciles a machine owning interfaces without IP", func() {
		By("creating the subnet")
		subnet := newSubnet(subnetName, "192.168.0.0/24")
		Expect(k8sClient.Create(ctx, subnet)).To(Succeed())

		m := newMachine()
		ifaces := []computev1alpha1.Interface{
			{
				Name:   "iface-0",
				Target: corev1.LocalObjectReference{Name: subnetName},
			},
			{
				Name:   "iface-1",
				Target: corev1.LocalObjectReference{Name: subnetName},
			},
		}
		m.Spec.Interfaces = ifaces

		By("creating the machine")
		Expect(k8sClient.Create(ctx, m)).To(Succeed())

		rngs := newIPAMRanges(m)
		for i, rng := range rngs {
			By("fetching the corresponding IPAMRange")
			key := objectKey(rng)
			Eventually(func() error {
				err := k8sClient.Get(ctx, key, rng)

				// Errors other than `not-found` shouldn't exist
				Expect(client.IgnoreNotFound(err)).To(Succeed())
				return err
			}, timeout, interval).Should(Succeed())

			By("checking if the parent of the IPAMRange corresponds to the target of the interface")
			Expect(rng.Spec.Parent.Name).To(Equal(networkv1alpha1.SubnetIPAMName(ifaces[i].Target.Name)))

			By("checking if the IP count in the IPAMRange request equals 1")
			Expect(rng.Spec.Requests[0].IPCount).To(Equal(int32(1)))

			By("checking the OwnerReferences of the IPAMRange contain the machine")
			Expect(rng.OwnerReferences).To(ContainElement(controllerReference(m)))
		}

		expectedIfaceStatuses := []computev1alpha1.InterfaceStatus{}
		for i, rng := range rngs {
			By("waiting till the IPAMRange gets reconciled")
			key := objectKey(rng)
			Eventually(func() int {
				Expect(k8sClient.Get(ctx, key, rng)).To(Succeed())
				return len(rng.Status.Allocations)
			}, timeout, interval).ShouldNot(Equal(0))

			expectedIfaceStatuses = appendInterfaceStatuses(expectedIfaceStatuses, &ifaces[i], rng)
		}

		By("checking if the machine's status gets reconciled")
		key := objectKey(m)
		Eventually(func() []computev1alpha1.InterfaceStatus {
			Expect(k8sClient.Get(ctx, key, m)).To(Succeed())
			return m.Status.Interfaces
		}, timeout, interval).Should(Equal(expectedIfaceStatuses))
	})
})

const (
	// test data
	machineKind = "Machine"
	subnetName  = "sample"

	// ginkgo
	interval = time.Millisecond * 250
	timeout  = time.Second * 20
)

var (
	objectKey = client.ObjectKeyFromObject
)

func controllerReference(m *computev1alpha1.Machine) metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion:         computev1alpha1.GroupVersion.String(),
		Kind:               machineKind,
		Name:               m.Name,
		UID:                m.UID,
		BlockOwnerDeletion: pointer.BoolPtr(true),
		Controller:         pointer.BoolPtr(true),
	}
}

func ip(ip string) *commonv1alpha1.IPAddr {
	parsed, _ := netaddr.ParseIP(ip)
	return commonv1alpha1.NewIPAddrPtr(parsed)
}

func ipamRangeRequest(ip *commonv1alpha1.IPAddr) networkv1alpha1.IPAMRangeRequest {
	return networkv1alpha1.IPAMRangeRequest{
		IPs: commonv1alpha1.NewIPRangePtr(netaddr.IPRangeFrom(ip.IP, ip.IP)),
	}
}

func newIPAMRanges(m *computev1alpha1.Machine) (rngs []*networkv1alpha1.IPAMRange) {
	for _, iface := range m.Spec.Interfaces {
		rng := &networkv1alpha1.IPAMRange{}
		rng.Namespace = m.Namespace
		rng.Name = computev1alpha1.MachineInterfaceIPAMRangeName(m.Name, iface.Name)
		rngs = append(rngs, rng)
	}
	return
}

func newMachine() *computev1alpha1.Machine {
	m := &computev1alpha1.Machine{}
	m.APIVersion = computev1alpha1.GroupVersion.String()
	m.Kind = machineKind
	m.Namespace = ns.Name
	m.GenerateName = "machine-controller-test"
	return m
}

func newSubnet(name, ipPrefix string) *networkv1alpha1.Subnet {
	subnet := &networkv1alpha1.Subnet{}
	subnet.APIVersion = networkv1alpha1.GroupVersion.String()
	subnet.Kind = networkv1alpha1.SubnetGK.Kind
	subnet.Namespace = ns.Name
	subnet.Name = name

	parsed, err := netaddr.ParseIPPrefix(ipPrefix)
	Expect(err).ToNot(HaveOccurred())
	subnet.Spec.Ranges = []networkv1alpha1.RangeType{{CIDR: commonv1alpha1.CIDR{IPPrefix: parsed}}}
	return subnet
}

// func notFoundOrSucceed(err error) error {
// 	fmt.Fprintf(GinkgoWriter, "error in notFoundOrSucceed %#v\n\n", err)
// 	Expect(apierrors.IsNotFound(err) || err == nil).To(BeTrue(), "error is `not found` or nil")
// 	return err
// }
