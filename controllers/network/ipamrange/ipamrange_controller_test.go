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
	common "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	api "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"time"
)

var _ = Describe("IPAMRange controller", func() {
	ctx := context.Background()

	const (
		validIPAMRangeName      = "valid-ipamrange"
		validIPAMRangeNamespace = "default"
		validIPAMRangeCIDR1     = "10.0.0.0/16"

		invalidIPAMRangeName      = "invalid-ipamrange"
		invalidIPAMRangeNamespace = "default"
		invalidIPAMRangeCIDR1     = "abc10.0.0.0/16"

		subrangeName      = "sub-range"
		subrangeNamespace = "default"

		timeout  = time.Second * 10
		interval = time.Second * 1
	)

	validIPAMRangeLookupKey := types.NamespacedName{
		Namespace: validIPAMRangeNamespace,
		Name:      validIPAMRangeName,
	}

	invalidIPAMRangeLookupKey := types.NamespacedName{
		Namespace: invalidIPAMRangeNamespace,
		Name:      invalidIPAMRangeName,
	}

	subrangeLookupKey := types.NamespacedName{
		Namespace: subrangeNamespace,
		Name:      subrangeName,
	}

	Context("When creating a valid IPAMRange", func() {
		It("Should create a valid IPPAMRange", func() {
			Expect(k8sClient.Create(ctx, &api.IPAMRange{
				ObjectMeta: metav1.ObjectMeta{
					Name:      validIPAMRangeName,
					Namespace: validIPAMRangeNamespace,
				},
				Spec: api.IPAMRangeSpec{
					Parent: nil,
					CIDRs:  []string{validIPAMRangeCIDR1},
				},
			})).Should(Succeed())
		})

		It("Should set the State to Ready", func() {
			Eventually(func() *api.IPAMRangeStatus {
				obj := &api.IPAMRange{}
				Expect(k8sClient.Get(ctx, validIPAMRangeLookupKey, obj)).Should(Succeed())
				return &obj.Status
			}, timeout, interval).Should(Equal(&api.IPAMRangeStatus{
				StateFields: common.StateFields{
					State: common.StateReady,
				},
				CIDRs: []string{validIPAMRangeCIDR1},
			}))
		})

		It("Should create a valid IPAMRange with parent", func() {
			Expect(k8sClient.Create(ctx, &api.IPAMRange{
				ObjectMeta: metav1.ObjectMeta{
					Name:      subrangeName,
					Namespace: subrangeNamespace,
				},
				Spec: api.IPAMRangeSpec{
					Parent: &common.ScopedReference{
						Name: validIPAMRangeName,
					},
					CIDRs: []string{"1/24"},
				},
			})).Should(Succeed())
		})

		It("Should set correct Status of valid IPAMRange with parent", func() {
			Eventually(func() *api.IPAMRangeStatus {
				obj := &api.IPAMRange{}
				Expect(k8sClient.Get(ctx, subrangeLookupKey, obj)).Should(Succeed())
				return &obj.Status
			}, timeout, interval).Should(Equal(&api.IPAMRangeStatus{
				StateFields: common.StateFields{
					State: common.StateReady,
				},
				CIDRs: []string{"10.0.1.0/24"},
			}))
			Eventually(func() []string {
				obj := &api.IPAMRange{}
				Expect(k8sClient.Get(ctx, validIPAMRangeLookupKey, obj)).Should(Succeed())
				return obj.Status.AllocationState
			}, timeout, interval).Should(ContainElements("10.0.1.0/24[busy]"))
		})

		It("Should release the allocated CIDR block on request deletion", func() {
			By("Deleting the request IPAMRange object")
			obj := &api.IPAMRange{}
			Expect(k8sClient.Get(ctx, subrangeLookupKey, obj)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, obj)).Should(Succeed())
			By("Checking whether the allocation state is not containing the reserved range")
			Eventually(func() []string {
				obj := &api.IPAMRange{}
				Expect(k8sClient.Get(ctx, validIPAMRangeLookupKey, obj)).Should(Succeed())
				return obj.Status.AllocationState
			}, timeout, interval).Should(Not(ContainElements("10.0.1.0/24[busy]")))
		})
	})

	Context("When creating a invalid IPAMRange", func() {
		It("Should create an invalid IPAMRange", func() {
			Expect(k8sClient.Create(ctx, &api.IPAMRange{
				ObjectMeta: metav1.ObjectMeta{
					Name:      invalidIPAMRangeName,
					Namespace: invalidIPAMRangeNamespace,
				},
				Spec: api.IPAMRangeSpec{
					Parent: nil,
					CIDRs:  []string{invalidIPAMRangeCIDR1},
				},
			})).Should(Succeed())
		})
		It("Should set the State to Invalid", func() {
			Eventually(func() string {
				obj := &api.IPAMRange{}
				Expect(k8sClient.Get(ctx, invalidIPAMRangeLookupKey, obj)).Should(Succeed())
				return obj.Status.State
			}, timeout, interval).Should(Equal(common.StateInvalid))
		})
	})
})
