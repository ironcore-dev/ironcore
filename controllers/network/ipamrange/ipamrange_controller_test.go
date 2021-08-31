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
	"fmt"
	common "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	api "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

var _ = Describe("IPAMRange controller", func() {
	ctx := context.Background()

	type IPAMStatus struct {
		State           string
		CIDRs           []string
		AllocationState []string
		RoundRobinState []string
		PendingRequest  *api.IPAMPendingRequest
	}

	const (
		timeout  = time.Second * 10
		interval = time.Second * 1
	)

	cleanUp := func(keys ...client.ObjectKey) {
		for _, key := range keys {
			ipamRange := &api.IPAMRange{}
			err := k8sClient.Get(ctx, types.NamespacedName{
				Namespace: key.Namespace,
				Name:      key.Name,
			}, ipamRange)
			if errors.IsNotFound(err) {
				return
			}
			Expect(err).Should(Succeed())
			Expect(k8sClient.Delete(ctx, ipamRange)).Should(Succeed())
			if len(ipamRange.GetFinalizers()) > 0 {
				newRange := ipamRange.DeepCopy()
				newRange.Finalizers = nil
				Expect(k8sClient.Patch(ctx, newRange, client.MergeFrom(ipamRange))).Should(Succeed())
			}
		}
	}

	createObject := func(key client.ObjectKey, parent *common.ScopedReference, cidrs ...string) {
		Expect(k8sClient.Create(ctx, &api.IPAMRange{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Spec: api.IPAMRangeSpec{
				Parent: parent,
				CIDRs:  cidrs,
			},
		})).Should(Succeed())
	}

	Context("When creating a valid IPAMRange", func() {

		validParentRangeLookupKey := types.NamespacedName{
			Namespace: "default",
			Name:      "ctx1-valid-parent",
		}
		validSubRangeLookupKey := types.NamespacedName{
			Namespace: "default",
			Name:      "ctx1-valid-subrange",
		}
		validParentCidrs := "10.0.0.0/16"
		validSubRangeCidrs := "1/24"

		It("Should clean and create base objects", func() {
			cleanUp(validParentRangeLookupKey, validSubRangeLookupKey)
			createObject(validParentRangeLookupKey, nil, validParentCidrs)
		})

		It("Should set the State to Ready for valid IPAMRange", func() {
			Eventually(func() *api.IPAMRangeStatus {
				obj := &api.IPAMRange{}
				Expect(k8sClient.Get(ctx, validParentRangeLookupKey, obj)).Should(Succeed())
				return &obj.Status
			}, timeout, interval).Should(Equal(&api.IPAMRangeStatus{
				StateFields: common.StateFields{
					State: common.StateReady,
				},
				CIDRs: []string{validParentCidrs},
			}))
		})

		It("Should create a valid IPAMRange with parent", func() {
			createObject(validSubRangeLookupKey, &common.ScopedReference{
				Name: validParentRangeLookupKey.Name,
			}, validSubRangeCidrs)
		})

		It("Should set correct Status of valid IPAMRange", func() {
			Eventually(func() *api.IPAMRangeStatus {
				obj := &api.IPAMRange{}
				Expect(k8sClient.Get(ctx, validParentRangeLookupKey, obj)).Should(Succeed())
				return &obj.Status
			}, timeout, interval).Should(Equal(&api.IPAMRangeStatus{
				StateFields: common.StateFields{
					State: common.StateReady,
				},
				CIDRs: []string{validParentCidrs},
				AllocationState: []string{
					"10.0.0.0/24[free]",
					"10.0.1.0/24[busy]",
					"10.0.2.0/23[free]",
					"10.0.4.0/22[free]",
					"10.0.8.0/21[free]",
					"10.0.16.0/20[free]",
					"10.0.32.0/19[free]",
					"10.0.64.0/18[free]",
					"10.0.128.0/17[free]",
				},
			}))
		})

		It("Should set correct status for valid sub IPAMRange", func() {
			Eventually(func() *IPAMStatus {
				obj := &api.IPAMRange{}
				Expect(k8sClient.Get(ctx, validSubRangeLookupKey, obj)).Should(Succeed())
				return &IPAMStatus{
					State: obj.Status.State,
					CIDRs: obj.Status.CIDRs,
				}
			}, timeout, interval).Should(Equal(&IPAMStatus{
				State: common.StateReady,
				CIDRs: []string{"10.0.1.0/24"},
			}))
		})

		It("Should release the allocated CIDR block on request deletion", func() {
			By("Deleting the request IPAMRange object")
			obj := &api.IPAMRange{}
			Expect(k8sClient.Get(ctx, validSubRangeLookupKey, obj)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, obj)).Should(Succeed())
			By("Checking whether the allocation state is not containing the reserved range")
			Eventually(func() *api.IPAMRangeStatus {
				obj := &api.IPAMRange{}
				Expect(k8sClient.Get(ctx, validParentRangeLookupKey, obj)).Should(Succeed())
				return &obj.Status
			}, timeout, interval).Should(Equal(&api.IPAMRangeStatus{
				StateFields: common.StateFields{
					State: common.StateReady,
				},
				CIDRs:           []string{validParentCidrs},
				AllocationState: []string{fmt.Sprintf("%s[free]", validParentCidrs)},
			}))
		})

	})

	Context("When creating a invalid IPAMRange", func() {

		invalidParentRangeLookupKey := types.NamespacedName{
			Namespace: "default",
			Name:      "ctx2-invalid-parent",
		}
		invalidSubRangeLookupKey := types.NamespacedName{
			Namespace: "default",
			Name:      "ctx2-invalid-subrange",
		}
		invalidParentCidrs := "abcd10.0.0.0/16"
		invalidSubRangeCidrs := "1/24"

		It("Should clean and create base objects", func() {
			cleanUp(invalidSubRangeLookupKey, invalidParentRangeLookupKey)
			createObject(invalidParentRangeLookupKey, nil, invalidParentCidrs)
			createObject(invalidSubRangeLookupKey, &common.ScopedReference{
				Name: invalidParentRangeLookupKey.Name,
			}, invalidSubRangeCidrs)
		})

		It("Should set the State to Invalid", func() {
			Eventually(func() *IPAMStatus {
				obj := &api.IPAMRange{}
				Expect(k8sClient.Get(ctx, invalidParentRangeLookupKey, obj)).Should(Succeed())
				return &IPAMStatus{
					State: obj.Status.State,
				}
			}, timeout, interval).Should(Equal(&IPAMStatus{
				State: common.StateInvalid,
			}))
		})

		It("Should set the State to Invalid for sub IPAMRange with Invalid parent", func() {
			Eventually(func() *IPAMStatus {
				obj := &api.IPAMRange{}
				Expect(k8sClient.Get(ctx, invalidSubRangeLookupKey, obj)).Should(Succeed())
				return &IPAMStatus{
					State: obj.Status.State,
				}
			}, timeout, interval).Should(Equal(&IPAMStatus{
				State: common.StateError,
			}))
		})

	})

	Context("When creating a valid IPAMRange", func() {

		validParentRangeLookupKey := types.NamespacedName{
			Namespace: "default",
			Name:      "ctx3-valid-parent",
		}
		validSubRangeLookupKey := types.NamespacedName{
			Namespace: "default",
			Name:      "ctx3-valid-subrange",
		}
		validSubRange2LookupKey := types.NamespacedName{
			Namespace: "default",
			Name:      "ctx3-valid-subrange2",
		}
		validParentCidrs := "10.0.0.0/16"
		validSubRangeCidrs := "1/32"
		validSubRange2Cidrs := "1/32"
		const allocatedSubRangeCidr = "10.0.0.1/32"

		It("Should clean and create base objects", func() {
			cleanUp(validParentRangeLookupKey, validSubRangeLookupKey)
			createObject(validParentRangeLookupKey, nil, validParentCidrs)
			createObject(validSubRangeLookupKey, &common.ScopedReference{
				Name: validParentRangeLookupKey.Name,
			}, validSubRangeCidrs)
		})

		It("Should check that parent and sub range have the correct status", func() {
			Eventually(func() *IPAMStatus {
				obj := &api.IPAMRange{}
				Expect(k8sClient.Get(ctx, validParentRangeLookupKey, obj)).Should(Succeed())
				return &IPAMStatus{
					State: obj.Status.State,
					CIDRs: obj.Status.CIDRs,
				}
			}, timeout, interval).Should(Equal(&IPAMStatus{
				State: common.StateReady,
				CIDRs: []string{validParentCidrs},
			}))
			Eventually(func() *IPAMStatus {
				obj := &api.IPAMRange{}
				Expect(k8sClient.Get(ctx, validSubRangeLookupKey, obj)).Should(Succeed())
				return &IPAMStatus{
					State: obj.Status.State,
					CIDRs: obj.Status.CIDRs,
				}
			}, timeout, interval).Should(Equal(&IPAMStatus{
				State: common.StateReady,
				CIDRs: []string{allocatedSubRangeCidr},
			}))
		})

		It("Should create a second subrange with the same request CIDR", func() {
			createObject(validSubRange2LookupKey, &common.ScopedReference{
				Name: validParentRangeLookupKey.Name,
			}, validSubRange2Cidrs)
		})

		It("Should set the second subrange to busy", func() {
			Eventually(func() *IPAMStatus {
				obj := &api.IPAMRange{}
				Expect(k8sClient.Get(ctx, validSubRange2LookupKey, obj)).Should(Succeed())
				return &IPAMStatus{
					State: obj.Status.State,
				}
			}, timeout, interval).Should(Equal(&IPAMStatus{
				State: common.StateBusy,
			}))
		})

		It("Should allocate a range for the second subrange when the first one is deleted", func() {
			obj := &api.IPAMRange{}
			// get first one
			Expect(k8sClient.Get(ctx, validSubRangeLookupKey, obj)).Should(Succeed())
			// delete first one
			Expect(k8sClient.Delete(ctx, obj)).Should(Succeed())
			// check if second moved from busy -> ready
			Eventually(func() *IPAMStatus {
				obj := &api.IPAMRange{}
				Expect(k8sClient.Get(ctx, validSubRange2LookupKey, obj)).Should(Succeed())
				return &IPAMStatus{
					State: obj.Status.State,
					CIDRs: obj.Status.CIDRs,
				}
			}, timeout, interval).Should(Equal(&IPAMStatus{
				State: common.StateReady,
				CIDRs: []string{allocatedSubRangeCidr},
			}))
		})

	})
})
