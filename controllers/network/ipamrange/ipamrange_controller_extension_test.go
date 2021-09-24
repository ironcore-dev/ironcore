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

var _ = OptionalDescribe("IPAMRange extension", func() {
	Context("When extending a valid IPAMRange", func() {

		validParentRangeLookupKey := types.NamespacedName{
			Namespace: "default",
			Name:      "ctx22-valid-parent",
		}
		validSubRangeLookupKey1 := types.NamespacedName{
			Namespace: "default",
			Name:      "ctx22-valid-subrange1",
		}
		validSubRangeLookupKey2 := types.NamespacedName{
			Namespace: "default",
			Name:      "ctx22-valid-subrange2",
		}
		validSubRangeLookupKey3 := types.NamespacedName{
			Namespace: "default",
			Name:      "ctx22-valid-subrange3",
		}
		validParentCidr1 := "10.0.0.0/16"
		validParentCidr2 := "10.1.0.0/16"
		validSubRangeCidr := "/17"

		configuredCIDRStatus := []api.CIDRAllocationStatus{
			api.CIDRAllocationStatus{
				CIDRAllocation: api.CIDRAllocation{
					Request: validParentCidr1,
					CIDR:    validParentCidr1,
				},
				Status:  api.AllocationStateAllocated,
				Message: SuccessfulUsageMessage,
			},
		}
		configuredCIDRStatus2 := []api.CIDRAllocationStatus{
			api.CIDRAllocationStatus{
				CIDRAllocation: api.CIDRAllocation{
					Request: validParentCidr1,
					CIDR:    validParentCidr1,
				},
				Status:  api.AllocationStateAllocated,
				Message: SuccessfulUsageMessage,
			},
			api.CIDRAllocationStatus{
				CIDRAllocation: api.CIDRAllocation{
					Request: validParentCidr2,
					CIDR:    validParentCidr2,
				},
				Status:  api.AllocationStateAllocated,
				Message: SuccessfulUsageMessage,
			},
		}

		allocatedCIDRStatus1 := []api.CIDRAllocationStatus{
			api.CIDRAllocationStatus{
				CIDRAllocation: api.CIDRAllocation{
					Request: validSubRangeCidr,
					CIDR:    "10.0.0.0/17",
				},
				Status:  api.AllocationStateAllocated,
				Message: SuccessfulAllocationMessage,
			},
		}
		allocatedCIDRStatus2 := []api.CIDRAllocationStatus{
			api.CIDRAllocationStatus{
				CIDRAllocation: api.CIDRAllocation{
					Request: validSubRangeCidr,
					CIDR:    "10.0.128.0/17",
				},
				Status:  api.AllocationStateAllocated,
				Message: SuccessfulAllocationMessage,
			},
		}
		allocatedCIDRStatus3 := []api.CIDRAllocationStatus{
			api.CIDRAllocationStatus{
				CIDRAllocation: api.CIDRAllocation{
					Request: validSubRangeCidr,
					CIDR:    "10.1.0.0/17",
				},
				Status:  api.AllocationStateAllocated,
				Message: SuccessfulAllocationMessage,
			},
		}
		failedCIDRStatus3 := []api.CIDRAllocationStatus{
			api.CIDRAllocationStatus{
				CIDRAllocation: api.CIDRAllocation{
					Request: validSubRangeCidr,
					CIDR:    "",
				},
				Status:  api.AllocationStateBusy,
				Message: FailBusyAllocationMessage(validSubRangeCidr),
			},
		}

		It("Should clean and create base objects", func() {
			cleanUp(validParentRangeLookupKey, validSubRangeLookupKey1, validSubRangeLookupKey2, validSubRangeLookupKey3)
			createObject(validParentRangeLookupKey, nil, validParentCidr1)
			Eventually(func() *api.IPAMRangeStatus {
				obj := &api.IPAMRange{}
				Expect(k8sClient.Get(ctx, validParentRangeLookupKey, obj)).Should(Succeed())
				return &obj.Status
			}, timeout, interval).Should(Equal(&api.IPAMRangeStatus{
				StateFields: common.StateFields{
					State: common.StateReady,
				},
				CIDRs: configuredCIDRStatus,
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
				CIDRs: configuredCIDRStatus,
				AllocationState: []string{
					"10.0.0.0/17[busy]",
					"10.0.128.0/17[free]",
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

		It("Should create a second valid IPAMRange with parent", func() {
			createObject(validSubRangeLookupKey2, &common.ScopedReference{
				Name: validParentRangeLookupKey.Name,
			}, validSubRangeCidr)
			Eventually(func() *IPAMStatus {
				return projectStatus(ctx, validParentRangeLookupKey)
			}, timeout, interval).Should(Equal(&IPAMStatus{
				State: common.StateReady,
				CIDRs: configuredCIDRStatus,
				AllocationState: []string{
					"10.0.0.0/16[busy]",
				},
				PendingRequest: nil,
			}))
			Eventually(func() *IPAMStatus {
				return projectStatus(ctx, validSubRangeLookupKey2)
			}, timeout, interval).Should(Equal(&IPAMStatus{
				State: common.StateReady,
				CIDRs: allocatedCIDRStatus2,
			}))
		})

		It("Should reject a third valid IPAMRange with parent", func() {
			createObject(validSubRangeLookupKey3, &common.ScopedReference{
				Name: validParentRangeLookupKey.Name,
			}, validSubRangeCidr)
			Eventually(func() *IPAMStatus {
				return projectStatus(ctx, validParentRangeLookupKey)
			}, timeout, interval).Should(Equal(&IPAMStatus{
				State: common.StateReady,
				CIDRs: configuredCIDRStatus,
				AllocationState: []string{
					"10.0.0.0/16[busy]",
				},
				PendingRequest: nil,
			}))
			Eventually(func() *IPAMStatus {
				return projectStatus(ctx, validSubRangeLookupKey3)
			}, timeout, interval).Should(Equal(&IPAMStatus{
				State: common.StateBusy,
				CIDRs: failedCIDRStatus3,
			}))
		})

		It("Should extend parent with new CIDR range", func() {
			updateObject(validParentRangeLookupKey, nil, validParentCidr1, validParentCidr2)
			Eventually(func() *IPAMStatus {
				return projectStatus(ctx, validParentRangeLookupKey)
			}, timeout, interval).Should(Equal(&IPAMStatus{
				State: common.StateReady,
				CIDRs: configuredCIDRStatus2,
				AllocationState: []string{
					"10.0.0.0/16[busy]",
					"10.1.0.0/17[busy]",
					"10.1.128.0/17[free]",
				},
				PendingRequest: nil,
			}))
			Eventually(func() *IPAMStatus {
				return projectStatus(ctx, validSubRangeLookupKey3)
			}, timeout, interval).Should(Equal(&IPAMStatus{
				State: common.StateReady,
				CIDRs: allocatedCIDRStatus3,
			}))
		})

	})

})
