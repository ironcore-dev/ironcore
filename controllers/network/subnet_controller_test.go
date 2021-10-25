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
	"reflect"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"inet.af/netaddr"
	corev1 "k8s.io/api/core/v1"
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
			Expect(k8sClient.Create(ctx, subnet)).Should(Succeed())

			ipamRng := newIPAMRange(subnet)
			Eventually(func() error {
				return k8sClient.Get(ctx, objectKey(ipamRng), ipamRng)
			}, timeout, interval).Should(BeNil())

			Expect(ipamRng.OwnerReferences).To(ContainElement(controllerReference(subnet)))
		})

		It("reconciles a subnet without parent", func() {
			subnet := newSubnet("no-parant")
			ipamRng := newIPAMRange(subnet)

			Expect(k8sClient.Create(ctx, subnet)).Should(Succeed())
			Eventually(func() error {
				return k8sClient.Get(ctx, objectKey(ipamRng), ipamRng)
			}, timeout, interval).Should(BeNil())

			By("patching the spec. of the related IPAMRange")
			Eventually(func() bool {
				rngGot := &networkv1alpha1.IPAMRange{}
				Expect(k8sClient.Get(ctx, objectKey(ipamRng), rngGot)).Should(Succeed())

				return func() bool {
					if rngGot.Status.Allocations == nil {
						return false
					}

					return reflect.DeepEqual(rngGot.Spec.CIDRs[0], subnet.Spec.Ranges[0].CIDR)
				}()
			}, timeout, interval).Should(BeTrue())

			By("patching the status of the Subnet")
			Eventually(func() bool {
				netGot := &networkv1alpha1.Subnet{}
				Expect(k8sClient.Get(ctx, objectKey(subnet), netGot)).Should(Succeed())

				return func() bool {
					return netGot.Status.State == networkv1alpha1.SubnetStateUp
				}()
			}, timeout, interval).Should(BeTrue())
		})

		It("reconciles a subnet with parent", func() {
			parent := newSubnet("parent")
			child := newSubnetWithParent("child", "parent")
			ipamRng := newIPAMRange(child)

			Expect(k8sClient.Create(ctx, parent)).Should(Succeed())
			Expect(k8sClient.Create(ctx, child)).Should(Succeed())
			Eventually(func() error {
				return k8sClient.Get(ctx, objectKey(ipamRng), ipamRng)
			}, timeout, interval).Should(BeNil())

			By("patching the spec of the owned IPAMRange")
			Eventually(func() bool {
				rngGot := &networkv1alpha1.IPAMRange{}
				Expect(k8sClient.Get(ctx, objectKey(ipamRng), rngGot)).Should(Succeed())

				return func() bool {
					parentRng := rngGot.Spec.Parent
					parentSubnet := child.Spec.Parent
					if parentRng == nil || parentSubnet == nil {
						return false
					}

					// Check if the IPAMRange is patched
					return parentRng.Name == networkv1alpha1.SubnetIPAMName(parentSubnet.Name) &&
						rngGot.Spec.CIDRs == nil &&
						reflect.DeepEqual(rngGot.Spec.Requests[0], networkv1alpha1.IPAMRangeRequest{CIDR: &child.Spec.Ranges[0].CIDR})

				}()
			}, timeout, interval).Should(BeTrue())

			By("patching the status of the Subnet")
			Eventually(func() bool {
				netGot := &networkv1alpha1.Subnet{}
				Expect(k8sClient.Get(ctx, objectKey(child), netGot)).Should(Succeed())

				return func() bool {
					return netGot.Status.State == networkv1alpha1.SubnetStateUp
				}()
			}, timeout, interval).Should(BeTrue())
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
