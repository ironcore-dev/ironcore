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
	common "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	api "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
)

var _ = OptionalDescribe("IPAMRange three level deletion", func() {
	Context("When extending a valid IPAMRange", func() {

		validRootRangeLookupKey := types.NamespacedName{
			Namespace: "default",
			Name:      "ctx40-valid-root",
		}
		validParentRangeLookupKey := types.NamespacedName{
			Namespace: "default",
			Name:      "ctx40-valid-parent",
		}
		validSubRangeLookupKey1 := types.NamespacedName{
			Namespace: "default",
			Name:      "ctx40-valid-subrange1",
		}
		validSubRangeLookupKey2 := types.NamespacedName{
			Namespace: "default",
			Name:      "ctx40-valid-subrange2",
		}
		validSubRangeLookupKey3 := types.NamespacedName{
			Namespace: "default",
			Name:      "ctx40-valid-subrange3",
		}
		validRootCidr := "10.0.0.0/8"
		validParentRequestCidr := "/16"
		allocatedParentCidr1 := "10.0.0.0/16"
		allocatedParentCidr2 := "10.1.0.0/16"
		validSubRangeCidr := "2/17"

		configuredRootCIDRStatus := []api.CIDRAllocationStatus{
			api.CIDRAllocationStatus{
				CIDRAllocation: api.CIDRAllocation{
					Request: validRootCidr,
					CIDR:    validRootCidr,
				},
				Status:  api.AllocationStateAllocated,
				Message: SuccessfulUsageMessage,
			},
		}
		allocatedCIDRParentStatus := []api.CIDRAllocationStatus{
			api.CIDRAllocationStatus{
				CIDRAllocation: api.CIDRAllocation{
					Request: validParentRequestCidr,
					CIDR:    allocatedParentCidr1,
				},
				Status:  api.AllocationStateAllocated,
				Message: SuccessfulAllocationMessage,
			},
			api.CIDRAllocationStatus{
				CIDRAllocation: api.CIDRAllocation{
					Request: validParentRequestCidr,
					CIDR:    allocatedParentCidr2,
				},
				Status:  api.AllocationStateAllocated,
				Message: SuccessfulAllocationMessage,
			},
		}
		allocatedCIDRParentStatus2 := []api.CIDRAllocationStatus{
			api.CIDRAllocationStatus{
				CIDRAllocation: api.CIDRAllocation{
					Request: validParentRequestCidr,
					CIDR:    allocatedParentCidr1,
				},
				Status:  api.AllocationStateAllocated,
				Message: SuccessfulAllocationMessage,
			},
		}
		pendingDeletionCIDRParentStatus2 := []api.CIDRAllocationStatus{
			api.CIDRAllocationStatus{
				CIDRAllocation: api.CIDRAllocation{
					Request: validParentRequestCidr,
					CIDR:    allocatedParentCidr2,
				},
				Status:  api.AllocationStateAllocated,
				Message: SuccessfulAllocationMessage,
			},
		}
		allocatedCIDRStatus1 := []api.CIDRAllocationStatus{
			api.CIDRAllocationStatus{
				CIDRAllocation: api.CIDRAllocation{
					Request: validSubRangeCidr,
					CIDR:    "10.1.0.0/17",
				},
				Status:  api.AllocationStateAllocated,
				Message: SuccessfulAllocationMessage,
			},
		}

		It("Should clean and create root range object", func() {
			cleanUp(validRootRangeLookupKey, validParentRangeLookupKey, validSubRangeLookupKey1, validSubRangeLookupKey2, validSubRangeLookupKey3)
			createObject(validRootRangeLookupKey, nil, validRootCidr)
			Eventually(func() *api.IPAMRangeStatus {
				obj := &api.IPAMRange{}
				Expect(k8sClient.Get(ctx, validRootRangeLookupKey, obj)).Should(Succeed())
				return &obj.Status
			}, timeout, interval).Should(Equal(&api.IPAMRangeStatus{
				StateFields: common.StateFields{
					State: common.StateReady,
				},
				CIDRs: configuredRootCIDRStatus,
			}))
		})

		It("Should create parent range", func() {
			createObject(validParentRangeLookupKey, &common.ScopedReference{
				Name: validRootRangeLookupKey.Name,
			}, validParentRequestCidr, validParentRequestCidr)
			Eventually(func() *api.IPAMRangeStatus {
				obj := &api.IPAMRange{}
				Expect(k8sClient.Get(ctx, validParentRangeLookupKey, obj)).Should(Succeed())
				return &obj.Status
			}, timeout, interval).Should(Equal(&api.IPAMRangeStatus{
				StateFields: common.StateFields{
					State: common.StateReady,
				},
				CIDRs: allocatedCIDRParentStatus,
			}))
			Eventually(func() *IPAMStatus {
				obj := &api.IPAMRange{}
				Expect(k8sClient.Get(ctx, validRootRangeLookupKey, obj)).Should(Succeed())
				return projectStatus(ctx, validRootRangeLookupKey)
			}, timeout, interval).Should(Equal(&IPAMStatus{
				State: common.StateReady,
				CIDRs: configuredRootCIDRStatus,
				AllocationState: []string{
					"10.0.0.0/15[busy]",
					"10.2.0.0/15[free]",
					"10.4.0.0/14[free]",
					"10.8.0.0/13[free]",
					"10.16.0.0/12[free]",
					"10.32.0.0/11[free]",
					"10.64.0.0/10[free]",
					"10.128.0.0/9[free]",
				},
			}))
		})

		It("Should create a valid IPAMRange with parent", func() {
			createObject(validSubRangeLookupKey1, &common.ScopedReference{
				Name: validParentRangeLookupKey.Name,
			}, validSubRangeCidr)
			Eventually(func() *IPAMStatus {
				return projectStatus(ctx, validParentRangeLookupKey)
			}, timeout, interval).Should(Equal(&IPAMStatus{
				State: common.StateReady,
				CIDRs: allocatedCIDRParentStatus,
				AllocationState: []string{
					"10.0.0.0/16[free]",
					"10.1.0.0/17[busy]",
					"10.1.128.0/17[free]",
				},
				PendingRequest: nil,
			}))
			Eventually(func() *IPAMStatus {
				return projectStatus(ctx, validSubRangeLookupKey1)
			}, timeout, interval).Should(Equal(&IPAMStatus{
				State: common.StateReady,
				CIDRs: allocatedCIDRStatus1,
			}))
		})

		It("Should remove second request from parent", func() {
			updateObject(validParentRangeLookupKey, &common.ScopedReference{
				Name: validRootRangeLookupKey.Name,
			}, validParentRequestCidr)
			Eventually(func() *IPAMStatus {
				return projectStatus(ctx, validParentRangeLookupKey)
			}, timeout, interval).Should(Equal(&IPAMStatus{
				State:            common.StateReady,
				CIDRs:            allocatedCIDRParentStatus2,
				PendingDeletions: pendingDeletionCIDRParentStatus2,
				AllocationState: []string{
					"10.0.0.0/16[free]",
					"10.1.0.0/17[busy]",
					"10.1.128.0/17[free]",
				},
				PendingRequest: nil,
			}))
			Eventually(func() *IPAMStatus {
				obj := &api.IPAMRange{}
				Expect(k8sClient.Get(ctx, validRootRangeLookupKey, obj)).Should(Succeed())
				return projectStatus(ctx, validRootRangeLookupKey)
			}, timeout, interval).Should(Equal(&IPAMStatus{
				State: common.StateReady,
				CIDRs: configuredRootCIDRStatus,
				AllocationState: []string{
					"10.0.0.0/15[busy]",
					"10.2.0.0/15[free]",
					"10.4.0.0/14[free]",
					"10.8.0.0/13[free]",
					"10.16.0.0/12[free]",
					"10.32.0.0/11[free]",
					"10.64.0.0/10[free]",
					"10.128.0.0/9[free]",
				},
			}))
		})

		It("Should remove subrange request", func() {
			var obj = &api.IPAMRange{}
			Expect(k8sClient.Get(ctx, validSubRangeLookupKey1, obj)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, obj)).Should(Succeed())
			Eventually(func() *IPAMStatus {
				return projectStatus(ctx, validParentRangeLookupKey)
			}, timeout, interval).Should(Equal(&IPAMStatus{
				State: common.StateReady,
				CIDRs: allocatedCIDRParentStatus2,
				AllocationState: []string{
					"10.0.0.0/16[free]",
				},
			}))
			Eventually(func() *IPAMStatus {
				obj := &api.IPAMRange{}
				Expect(k8sClient.Get(ctx, validRootRangeLookupKey, obj)).Should(Succeed())
				return projectStatus(ctx, validRootRangeLookupKey)
			}, timeout, interval).Should(Equal(&IPAMStatus{
				State: common.StateReady,
				CIDRs: configuredRootCIDRStatus,
				AllocationState: []string{
					"10.0.0.0/16[busy]",
					"10.1.0.0/16[free]",
					"10.2.0.0/15[free]",
					"10.4.0.0/14[free]",
					"10.8.0.0/13[free]",
					"10.16.0.0/12[free]",
					"10.32.0.0/11[free]",
					"10.64.0.0/10[free]",
					"10.128.0.0/9[free]",
				},
			}))
		})
	})
})
