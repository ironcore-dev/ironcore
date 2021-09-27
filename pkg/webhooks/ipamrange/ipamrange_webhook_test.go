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

package ipamrange

import (
	"context"
	api "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("IPAMRange webhook", func() {

	const (
		myIpamRange           = "my-ipamrange"
		myIpamRangeNamespace  = "default"
		myIpamRangeWithParent = "my-ipamrange-with-parent"
	)

	ipamRangeNoParent := &api.IPAMRange{
		ObjectMeta: metav1.ObjectMeta{
			Name:      myIpamRange,
			Namespace: myIpamRangeNamespace,
		},
		Spec: api.IPAMRangeSpec{
			Parent: nil,
			CIDRs:  []string{"10.0.0.0/16"},
		},
	}

	ipamRangeWithParent := ipamRangeNoParent.DeepCopy()
	ipamRangeWithParent.Name = myIpamRangeWithParent
	ipamRangeWithParent.Spec.Parent = &corev1.LocalObjectReference{
		Name: "a",
	}

	Context("When creating an IPAMRange object", func() {
		ctx := context.Background()
		It("It should pass on creation of an IPAMRange with no parent", func() {
			Expect(k8sClient.Create(ctx, ipamRangeNoParent)).Should(Succeed())
		})

		It("It should pass on creation of an IPAMRange with a parent", func() {
			Expect(k8sClient.Create(ctx, ipamRangeWithParent)).Should(Succeed())
		})

		It("It should reject a change from no parent to a parent", func() {
			ipamRangeWithNewParent := ipamRangeNoParent.DeepCopy()
			ipamRangeWithNewParent.Spec.Parent = &corev1.LocalObjectReference{
				Name: "a",
			}
			Expect(k8sClient.Update(ctx, ipamRangeWithNewParent)).Should(Not(Succeed()))
		})

		It("It should reject a change from a parent to a new parent", func() {
			ipamRangeWithNewParent := ipamRangeWithParent.DeepCopy()
			ipamRangeWithNewParent.Spec.Parent = &corev1.LocalObjectReference{
				Name: "b",
			}
			Expect(k8sClient.Update(ctx, ipamRangeWithNewParent)).Should(Not(Succeed()))
		})
	})
})
