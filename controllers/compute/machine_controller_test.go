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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	networkv1alpha1 "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
)

var _ = Describe("machine controller", func() {
	It("reconciles a machine with interfaces", func() {
		m := newMachineWithInterfaces("with-interface")
		rngs, rngKeys := newIPAMRanges(m)

		By("creating the machine")
		Expect(k8sClient.Create(ctx, m)).To(Succeed())

		for i := range rngs {
			By("fetching the corresponding IPAMRange")
			Eventually(func() error {
				return notFoundOrSucceed(k8sClient.Get(ctx, rngKeys[i], rngs[i]))
			}, timeout, interval).Should(Succeed())

			By("checking if the parent of the IPAMRange corresponds to the target of the interface")
			Expect(rngs[i].Spec.Parent.Name).To(Equal(networkv1alpha1.SubnetIPAMName(ifaceTargets[i])))

			By("checking if the request of the IPAMRange corresponds to the IP of the machine's interface")
			Expect(rngs[i].Spec.Requests[0]).To(Equal(*ipamRangeRequest(ifaceIPs[i])))

			By("checking the OwnerReferences of the IPAMRange contain the machine")
			Expect(rngs[i].OwnerReferences).To(ContainElement(controllerReference(m)))
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
	// test data
	ifaceTargets = []string{"0", "1"}
	ifaceIPs     = []string{"192.168.0.0", "192.168.0.1"}
	objectKey    = client.ObjectKeyFromObject
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

func ipamRangeRequest(ip string) *networkv1alpha1.IPAMRangeRequest {
	parsedIP, _ := netaddr.ParseIP(ip)
	return &networkv1alpha1.IPAMRangeRequest{
		IPs: v1alpha1.NewIPRangePtr(netaddr.IPRangeFrom(parsedIP, parsedIP)),
	}
}

func newInterface(name, ip, target string) *computev1alpha1.Interface {
	iface := &computev1alpha1.Interface{}
	iface.Name = name
	parsed, _ := netaddr.ParseIP(ip)
	iface.IP = v1alpha1.NewIPAddrPtr(parsed)
	iface.Target.Name = target
	return iface
}

func newInterfaces(m *computev1alpha1.Machine) []computev1alpha1.Interface {
	return []computev1alpha1.Interface{
		*newInterface(m.Name+"-0", ifaceIPs[0], ifaceTargets[0]),
		*newInterface(m.Name+"-1", ifaceIPs[1], ifaceTargets[1]),
	}
}

func newIPAMRanges(m *computev1alpha1.Machine) (rngs []*networkv1alpha1.IPAMRange, keys []types.NamespacedName) {
	for _, iface := range m.Spec.Interfaces {
		rng := &networkv1alpha1.IPAMRange{}
		rng.Namespace = m.Namespace
		rng.Name = computev1alpha1.MachineInterfaceIPAMRangeName(m.Name, iface.Name)
		rngs = append(rngs, rng)
		keys = append(keys, objectKey(rng))
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

func newMachineWithInterfaces(name string) *computev1alpha1.Machine {
	m := newMachine(name)
	m.Spec.Interfaces = newInterfaces(m)
	return m
}

func notFoundOrSucceed(err error) error {
	fmt.Fprintf(GinkgoWriter, "error in notFoundOrSucceed %#v\n\n", err)
	Expect(apierrors.IsNotFound(err) || err == nil).To(BeTrue(), "error is `not found` or nil")
	return err
}
