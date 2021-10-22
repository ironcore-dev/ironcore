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

	"github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	nw "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
	. "github.com/onsi/ginkgo"
	"inet.af/netaddr"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	. "github.com/onsi/gomega"
)

var _ = Describe("subnet controller", func() {
	Context("reconcileExists", func() {
		It("finishes reconciliation early if the instance is being deleted", func() {
			subnet := newSubnet("early-finished")
			subnet.DeletionTimestamp = now()
			ipamRange := newIPAMRange(subnet)

			Expect(k8sClient.Create(ctx, ipamRange)).Should(Succeed())
			Expect(k8sClient.Create(ctx, subnet)).Should(Succeed())

			Eventually(func() bool {
				rngGot := &nw.IPAMRange{}
				Expect(k8sClient.Get(ctx, objectKey(ipamRange), rngGot))
				return rngGot.Spec.CIDRs == nil
			}, timeout, interval).Should(BeTrue())
		})
	})

	Context("Reconcile", func() {
		It("reconciles a subnet without parent", func() {
			subnet := newSubnet("no-parant")
			ipamRange := newIPAMRange(subnet)

			Expect(k8sClient.Create(ctx, ipamRange)).Should(Succeed())
			Expect(k8sClient.Create(ctx, subnet)).Should(Succeed())

			By("patching the spec. of the related IPAMRange")
			Eventually(func() bool {
				rngGot := &nw.IPAMRange{}
				Expect(k8sClient.Get(ctx, objectKey(ipamRange), rngGot)).Should(Succeed())

				return func() bool {
					if rngGot.Status.Allocations == nil {
						return false
					}

					return reflect.DeepEqual(rngGot.Spec.CIDRs[0], subnet.Spec.Ranges[0].CIDR)
				}()
			}, timeout, interval).Should(BeTrue())

			By("patching the status of the Subnet")
			Eventually(func() bool {
				netGot := &nw.Subnet{}
				Expect(k8sClient.Get(ctx, objectKey(subnet), netGot)).Should(Succeed())

				return func() bool {
					return netGot.Status.State == nw.SubnetStateUp
				}()
			}, timeout, interval).Should(BeTrue())
		})

		It("reconciles a subnet with parent", func() {
			parent := newSubnet("parent")
			child := newSubnetWithParent("child", "parent")
			ipamRng := newIPAMRange(child)

			Expect(k8sClient.Create(ctx, ipamRng)).Should(Succeed())
			Expect(k8sClient.Create(ctx, parent)).Should(Succeed())
			Expect(k8sClient.Create(ctx, child)).Should(Succeed())

			By("patching the spec of the owned IPAMRange")
			Eventually(func() bool {
				rngGot := &nw.IPAMRange{}
				Expect(k8sClient.Get(ctx, objectKey(ipamRng), rngGot)).Should(Succeed())

				return func() bool {
					rngParent := rngGot.Spec.Parent
					childNetParent := child.Spec.Parent
					if rngParent == nil || childNetParent == nil {
						return false
					}

					// Check if the IPAMRange is patched
					return rngParent.Name == nw.SubnetIPAMName(childNetParent.Name) &&
						rngGot.Spec.CIDRs == nil &&
						reflect.DeepEqual(rngGot.Spec.Requests[0], nw.IPAMRangeRequest{CIDR: &child.Spec.Ranges[0].CIDR})

				}()
			}, timeout, interval).Should(BeTrue())

			By("patching the status of the Subnet")
			Eventually(func() bool {
				netGot := &nw.Subnet{}
				Expect(k8sClient.Get(ctx, objectKey(child), netGot)).Should(Succeed())

				return func() bool {
					return netGot.Status.State == nw.SubnetStateUp
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

type object interface {
	GetNamespace() string
	GetName() string
}

func newIPAMRange(sub *nw.Subnet) *nw.IPAMRange {
	rng := &nw.IPAMRange{}
	rng.Namespace = sub.Namespace
	rng.Name = nw.SubnetIPAMName(sub.Name)
	return rng
}

func newSubnet(name string) *nw.Subnet {
	subnet := &nw.Subnet{}
	subnet.Namespace = ns
	subnet.Name = name

	ipPrefix, err := netaddr.ParseIPPrefix(ipPrefix)
	Expect(err).ToNot(HaveOccurred())
	subnet.Spec.Ranges = []nw.RangeType{{CIDR: v1alpha1.CIDR{IPPrefix: ipPrefix}}}
	return subnet
}

func newSubnetWithParent(name, parentName string) *nw.Subnet {
	subnet := newSubnet(name)
	subnet.Spec.Parent = &core.LocalObjectReference{Name: parentName}
	return subnet
}

func now() *meta.Time {
	now := meta.NewTime(time.Now())
	return &now
}

func objectKey(o object) types.NamespacedName {
	return types.NamespacedName{
		Namespace: o.GetNamespace(),
		Name:      o.GetName(),
	}
}
