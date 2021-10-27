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
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"inet.af/netaddr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	networkv1alpha1 "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
)

var _ = Describe("subnet controller", func() {
	Context("Reconcile", func() {
		It("sets the owner Subnet as a Controller OwnerReference on the controlled IPAMRange", func() {
			subnet := newSubnet("owner")
			ipamRng := newIPAMRange(subnet)
			ipamRngKey := objectKey(ipamRng)

			Expect(k8sClient.Create(ctx, subnet)).Should(Succeed())
			Eventually(func() error {
				return notFoundOrSucceed(k8sClient.Get(ctx, ipamRngKey, ipamRng))
			}, timeout, interval).Should(Succeed())

			Expect(ipamRng.OwnerReferences).To(ContainElement(controllerReference(subnet)))
		})

		It("reconciles a Subnet without parent", func() {
			subnet := newSubnet("no-parant")
			ipamRng := newIPAMRange(subnet)

			subnetKey := objectKey(subnet)
			ipamRngKey := objectKey(ipamRng)

			By("creating a Subnet without parent")
			Expect(k8sClient.Create(ctx, subnet)).Should(Succeed())
			Eventually(func() error {
				return notFoundOrSucceed(k8sClient.Get(ctx, ipamRngKey, ipamRng))
			}, timeout, interval).Should(Succeed())

			By("wating for the status of the owned IPAMRange to have an allocated CIDR")
			Eventually(func() []v1alpha1.CIDR {
				Expect(k8sClient.Get(ctx, ipamRngKey, ipamRng)).Should(Succeed())
				return ipamRng.Spec.CIDRs
			}, timeout, interval).Should(ContainElement(subnet.Spec.Ranges[0].CIDR))

			By("waiting for the status of the Subnet to become up")
			Eventually(func() networkv1alpha1.SubnetState {
				Expect(k8sClient.Get(ctx, subnetKey, subnet)).Should(Succeed())
				return subnet.Status.State
			}, timeout, interval).Should(Equal(networkv1alpha1.SubnetStateUp))
		})

		It("reconciles a subnet with parent", func() {
			parentNet := newSubnet("parent")
			childNet := newSubnetWithParent("child", "parent")
			childRng := newIPAMRange(childNet)

			childNetKey := objectKey(childNet)
			childRngKey := objectKey(childRng)

			By("creating a pair of parent and child Subnet")
			Expect(k8sClient.Create(ctx, parentNet)).Should(Succeed())
			Expect(k8sClient.Create(ctx, childNet)).Should(Succeed())
			Eventually(func() error {
				return notFoundOrSucceed(k8sClient.Get(ctx, childRngKey, childRng))
			}, timeout, interval).Should(Succeed())

			By("wating for the spec of the child IPAMRange to be patched")
			Eventually(func() *networkv1alpha1.IPAMRangeSpec {
				Expect(k8sClient.Get(ctx, childRngKey, childRng)).Should(Succeed())
				return &childRng.Spec
			}, timeout, interval).Should(Equal(ipamRangeSpec(childNet)))

			By("waiting for the status of the Subnet to be become up")
			Eventually(func() networkv1alpha1.SubnetState {
				Expect(k8sClient.Get(ctx, childNetKey, childNet)).Should(Succeed())
				return childNet.Status.State
			}, timeout, interval).Should(Equal(networkv1alpha1.SubnetStateUp))
		})
	})
})

const (
	// test data
	ipPrefix = "192.168.0.0/24"
	ns       = "default" // namespace

	// ginkgo
	interval = time.Millisecond * 250
	timeout  = time.Second * 10
)

var objectKey = client.ObjectKeyFromObject

func ipamRangeSpec(subnet *networkv1alpha1.Subnet) *networkv1alpha1.IPAMRangeSpec {
	rngSpec := &networkv1alpha1.IPAMRangeSpec{}
	rngSpec.Parent = &corev1.LocalObjectReference{Name: networkv1alpha1.SubnetIPAMName(subnet.Spec.Parent.Name)}
	rngSpec.Requests = []networkv1alpha1.IPAMRangeRequest{{CIDR: &subnet.Spec.Ranges[0].CIDR}}
	return rngSpec
}

func newIPAMRange(sub *networkv1alpha1.Subnet) *networkv1alpha1.IPAMRange {
	rng := &networkv1alpha1.IPAMRange{}
	rng.Namespace = sub.Namespace
	rng.Name = networkv1alpha1.SubnetIPAMName(sub.Name)
	return rng
}

func newSubnet(name string) *networkv1alpha1.Subnet {
	subnet := &networkv1alpha1.Subnet{}
	subnet.APIVersion = networkv1alpha1.GroupVersion.String()
	subnet.Kind = networkv1alpha1.SubnetGK.Kind
	subnet.Namespace = ns
	subnet.Name = name

	ipPrefix, err := netaddr.ParseIPPrefix(ipPrefix)
	Expect(err).ToNot(HaveOccurred())
	subnet.Spec.Ranges = []networkv1alpha1.RangeType{{CIDR: v1alpha1.CIDR{IPPrefix: ipPrefix}}}
	return subnet
}

func newSubnetWithParent(name, parentName string) *networkv1alpha1.Subnet {
	subnet := newSubnet(name)
	subnet.Spec.Parent = &corev1.LocalObjectReference{Name: parentName}
	return subnet
}

func notFoundOrSucceed(err error) error {
	Expect(apierrors.IsNotFound(err) || err == nil).To(BeTrue(), "error is `not found` or nil")
	return err
}

func controllerReference(subnet *networkv1alpha1.Subnet) metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion:         networkv1alpha1.GroupVersion.String(),
		Kind:               networkv1alpha1.SubnetGK.Kind,
		Name:               subnet.Name,
		UID:                subnet.UID,
		BlockOwnerDeletion: pointer.BoolPtr(true),
		Controller:         pointer.BoolPtr(true),
	}

}
