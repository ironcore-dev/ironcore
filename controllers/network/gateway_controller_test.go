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
	"sigs.k8s.io/controller-runtime/pkg/client"

	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	networkv1alpha1 "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
)

var _ = Describe("gateway controller", func() {
	ns := SetupTest()
	Context("creation", func() {
		It("adds the finalizer", func() {
			subnet := subnet(ns.Name, "192.168.0.0/24")
			gw := gateway(ns.Name, subnet)
			gwKey := objectKey(gw)
			Eventually(func() []string {
				err := k8sClient.Get(ctx, gwKey, gw)
				Expect(client.IgnoreNotFound(err)).To(Succeed())
				return gw.ObjectMeta.Finalizers
			}, timeout, interval).Should(ContainElement(networkv1alpha1.GatewayFinalizer))
		})
	})
})

func gateway(ns string, subnet *networkv1alpha1.Subnet) *networkv1alpha1.Gateway {
	gw := &networkv1alpha1.Gateway{
		TypeMeta: metav1.TypeMeta{
			APIVersion: networkv1alpha1.GroupVersion.String(),
			Kind:       "gateway",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    ns,
			GenerateName: "gateway-controller-",
		},
		Spec: networkv1alpha1.GatewaySpec{
			Subnet: v1.LocalObjectReference{Name: subnet.Name},
		},
	}
	Expect(k8sClient.Create(ctx, gw)).To(Succeed())
	return gw
}

func subnet(ns string, ipPrefix string) *networkv1alpha1.Subnet {
	parsedPrefix, err := netaddr.ParseIPPrefix(ipPrefix)
	Expect(err).ToNot(HaveOccurred())

	subnet := &networkv1alpha1.Subnet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: networkv1alpha1.GroupVersion.String(),
			Kind:       networkv1alpha1.SubnetGK.Kind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    ns,
			GenerateName: "gateway-controller-",
		},
		Spec: networkv1alpha1.SubnetSpec{
			Ranges: []networkv1alpha1.RangeType{
				{CIDR: &commonv1alpha1.CIDR{IPPrefix: parsedPrefix}},
			},
		},
	}

	Expect(k8sClient.Create(ctx, subnet)).To(Succeed())
	return subnet
}

var objectKey = client.ObjectKeyFromObject
