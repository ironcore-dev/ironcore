/*
 * Copyright (c) 2022 by the OnMetal authors.
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

package validation

import (
	"github.com/onmetal/onmetal-api/internal/apis/networking"
	. "github.com/onmetal/onmetal-api/internal/testutils/validation"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

var _ = Describe("LoadBalancer", func() {
	DescribeTable("ValidateLoadBalancer",
		func(loadBalancer *networking.LoadBalancer, match types.GomegaMatcher) {
			errList := ValidateLoadBalancer(loadBalancer)
			Expect(errList).To(match)
		},
		Entry("missing name",
			&networking.LoadBalancer{},
			ContainElement(RequiredField("metadata.name")),
		),
		Entry("missing namespace",
			&networking.LoadBalancer{ObjectMeta: metav1.ObjectMeta{Name: "foo"}},
			ContainElement(RequiredField("metadata.namespace")),
		),
		Entry("bad name",
			&networking.LoadBalancer{ObjectMeta: metav1.ObjectMeta{Name: "foo*"}},
			ContainElement(InvalidField("metadata.name")),
		),
		Entry("no network ref",
			&networking.LoadBalancer{},
			ContainElement(InvalidField("spec.networkRef.name")),
		),
		Entry("invalid network ref name",
			&networking.LoadBalancer{
				Spec: networking.LoadBalancerSpec{
					NetworkRef: corev1.LocalObjectReference{Name: "foo*"},
				},
			},
			ContainElement(InvalidField("spec.networkRef.name")),
		),
		Entry("missing type",
			&networking.LoadBalancer{},
			ContainElement(RequiredField("spec.type")),
		),
		Entry("load balancer type public",
			&networking.LoadBalancer{Spec: networking.LoadBalancerSpec{Type: networking.LoadBalancerTypePublic}},
			Not(ContainElement(RequiredField("spec.type"))),
		),
		Entry("load balancer type internal",
			&networking.LoadBalancer{Spec: networking.LoadBalancerSpec{Type: networking.LoadBalancerTypeInternal}},
			Not(ContainElement(RequiredField("spec.type"))),
		),
		Entry("duplicate ip family",
			&networking.LoadBalancer{
				Spec: networking.LoadBalancerSpec{
					IPFamilies: []corev1.IPFamily{corev1.IPv4Protocol, corev1.IPv4Protocol},
				},
			},
			ContainElement(DuplicateField("spec.ipFamilies[1]")),
		),
		Entry("overlapping port ranges",
			&networking.LoadBalancer{
				Spec: networking.LoadBalancerSpec{
					Ports: []networking.LoadBalancerPort{
						{
							Port:    1,
							EndPort: pointer.Int32(3),
						},
						{
							Port:    2,
							EndPort: pointer.Int32(4),
						},
					},
				},
			},
			ContainElement(ForbiddenField("spec.ports[1]")),
		),
		Entry("overlapping port ranges but different protocol",
			&networking.LoadBalancer{
				Spec: networking.LoadBalancerSpec{
					Ports: []networking.LoadBalancerPort{
						{
							Port:    1,
							EndPort: pointer.Int32(3),
						},
						{
							Protocol: ProtocolPtr(corev1.ProtocolUDP),
							Port:     2,
							EndPort:  pointer.Int32(4),
						},
					},
				},
			},
			Not(ContainElement(ForbiddenField("spec.ports[1]"))),
		),
	)

	DescribeTable("ValidateLoadBalancerUpdate",
		func(newLoadBalancer, oldLoadBalancer *networking.LoadBalancer, match types.GomegaMatcher) {
			errList := ValidateLoadBalancerUpdate(newLoadBalancer, oldLoadBalancer)
			Expect(errList).To(match)
		},
		Entry("immutable networkRef",
			&networking.LoadBalancer{
				Spec: networking.LoadBalancerSpec{
					NetworkRef: corev1.LocalObjectReference{Name: "foo"},
				},
			},
			&networking.LoadBalancer{
				Spec: networking.LoadBalancerSpec{
					NetworkRef: corev1.LocalObjectReference{Name: "foo*"},
				},
			},
			ContainElement(ForbiddenField("spec.networkRef")),
		),
		Entry("mutable network interface selector",
			&networking.LoadBalancer{
				Spec: networking.LoadBalancerSpec{
					NetworkInterfaceSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"foo": "bar"},
					},
				},
			},
			&networking.LoadBalancer{
				Spec: networking.LoadBalancerSpec{
					NetworkInterfaceSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"bar": "baz"},
					},
				},
			},
			Not(ContainElement(ForbiddenField("spec"))),
		),
	)
})
