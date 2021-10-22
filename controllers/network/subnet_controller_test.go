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
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	. "github.com/onsi/gomega"
)

var _ = Describe("subnet controller", func() {
	Context("reconcileExists", func() {
		It("finishes reconciliation early if the instance is being deleted", func() {
			subnet := newSubnet("early-finished")
			subnet.DeletionTimestamp = now()
			Expect(k8sClient.Create(ctx, subnet)).Should(Succeed())

			ipamRange := newIPAMRange(subnet)
			Expect(k8sClient.Create(ctx, ipamRange)).Should(Succeed())

			Eventually(func() bool {
				got := &nw.IPAMRange{}
				Expect(k8sClient.Get(ctx, objectKey(ipamRange), got))
				return got.Status.Allocations == nil
			}, timeout, interval).Should(BeTrue())
		})
	})

	Context("Reconcile", func() {
		It("patches the status of the related IPAMRange", func() {
			subnet := newSubnet("reconciled")
			Expect(k8sClient.Create(ctx, subnet)).Should(Succeed())

			ipamRange := newIPAMRange(subnet)
			Expect(k8sClient.Create(ctx, ipamRange)).Should(Succeed())

			Eventually(func() bool {
				got := &nw.IPAMRange{}
				if err := k8sClient.Get(ctx, objectKey(ipamRange), got); err != nil {
					return false
				}

				return func() bool {
					if got.Status.Allocations == nil {
						return false
					}
					return reflect.DeepEqual(got.Status.Allocations[0].CIDR, &subnet.Spec.Ranges[0].CIDR)
				}()
			}, timeout, interval).Should(BeTrue())
		})
	})
})

const (
	ns = "default" // namespace

	ipPrefix = "192.168.0.0/24"

	timeout  = time.Second * 10
	interval = time.Millisecond * 250
)

type object interface {
	GetNamespace() string
	GetName() string
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

func newIPAMRange(sub *nw.Subnet) *nw.IPAMRange {
	rng := &nw.IPAMRange{}
	rng.Namespace = sub.Namespace
	rng.Name = nw.SubnetIPAMName(sub.Name)
	return rng
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
