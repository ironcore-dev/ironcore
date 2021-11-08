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
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"inet.af/netaddr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"

	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	networkv1alpha1 "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
)

var _ = Describe("machine controller", func() {
	It("reconciles a machine with interfaces", func() {
		m := newMachine("with-interfaces")
		ifaces := []computev1alpha1.Interface{
			{
				Name:   m.Name + "-0",
				IP:     ip("192.168.0.0"),
				Target: corev1.LocalObjectReference{Name: "target-0"},
			},
			{
				Name:   m.Name + "-1",
				IP:     ip("192.168.0.1"),
				Target: corev1.LocalObjectReference{Name: "target-1"},
			},
		}
		m.Spec.Interfaces = ifaces

		By("creating the machine")
		Expect(k8sClient.Create(ctx, m)).To(Succeed())

		rngs := newIPAMRanges(m)
		for i, rng := range rngs {
			By("fetching the corresponding IPAMRange")
			Eventually(func() error {
				return notFoundOrSucceed(k8sClient.Get(ctx, objectKey(rng), rng))
			}, timeout, interval).Should(Succeed())

			By("checking if the parent of the IPAMRange corresponds to the target of the interface")
			Expect(rng.Spec.Parent.Name).To(Equal(networkv1alpha1.SubnetIPAMName(ifaces[i].Target.Name)))

			By("checking if the request of the IPAMRange corresponds to the IP of the machine's interface")
			Expect(rng.Spec.Requests[0]).To(Equal(ipamRangeRequest(ifaces[i].IP)))

			By("checking the OwnerReferences of the IPAMRange contain the machine")
			Expect(rng.OwnerReferences).To(ContainElement(controllerReference(m)))
		}
	})
})

const (
	// test data
	kind = "Machine"

	// ginkgo
	interval = time.Millisecond * 250
	timeout  = time.Second * 10
)

var (
	objectKey = client.ObjectKeyFromObject
)

func controllerReference(m *computev1alpha1.Machine) metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion:         computev1alpha1.GroupVersion.String(),
		Kind:               kind,
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

func newMachine(name string) *computev1alpha1.Machine {
	m := &computev1alpha1.Machine{}
	m.APIVersion = computev1alpha1.GroupVersion.String()
	m.Kind = kind
	m.Namespace = ns.Name
	m.Name = name
	return m
}

func notFoundOrSucceed(err error) error {
	fmt.Fprintf(GinkgoWriter, "error in notFoundOrSucceed %#v\n\n", err)
	Expect(apierrors.IsNotFound(err) || err == nil).To(BeTrue(), "error is `not found` or nil")
	return err
}
