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
	networkv1alpha1 "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
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

		configuredCIDRStatus := []networkv1alpha1.CIDRAllocationStatus{
			{
				CIDRAllocation: networkv1alpha1.CIDRAllocation{
					Request: validParentCidr1,
					CIDR:    validParentCidr1,
				},
				Status:  networkv1alpha1.AllocationStateAllocated,
				Message: SuccessfulUsageMessage,
			},
		}
		configuredCIDRStatus2 := []networkv1alpha1.CIDRAllocationStatus{
			{
				CIDRAllocation: networkv1alpha1.CIDRAllocation{
					Request: validParentCidr1,
					CIDR:    validParentCidr1,
				},
				Status:  networkv1alpha1.AllocationStateAllocated,
				Message: SuccessfulUsageMessage,
			},
			{
				CIDRAllocation: networkv1alpha1.CIDRAllocation{
					Request: validParentCidr2,
					CIDR:    validParentCidr2,
				},
				Status:  networkv1alpha1.AllocationStateAllocated,
				Message: SuccessfulUsageMessage,
			},
		}

		allocatedCIDRStatus1 := []networkv1alpha1.CIDRAllocationStatus{
			{
				CIDRAllocation: networkv1alpha1.CIDRAllocation{
					Request: validSubRangeCidr,
					CIDR:    "10.0.0.0/17",
				},
				Status:  networkv1alpha1.AllocationStateAllocated,
				Message: SuccessfulAllocationMessage,
			},
		}
		allocatedCIDRStatus2 := []networkv1alpha1.CIDRAllocationStatus{
			{
				CIDRAllocation: networkv1alpha1.CIDRAllocation{
					Request: validSubRangeCidr,
					CIDR:    "10.0.128.0/17",
				},
				Status:  networkv1alpha1.AllocationStateAllocated,
				Message: SuccessfulAllocationMessage,
			},
		}
		allocatedCIDRStatus3 := []networkv1alpha1.CIDRAllocationStatus{
			{
				CIDRAllocation: networkv1alpha1.CIDRAllocation{
					Request: validSubRangeCidr,
					CIDR:    "10.1.0.0/17",
				},
				Status:  networkv1alpha1.AllocationStateAllocated,
				Message: SuccessfulAllocationMessage,
			},
		}
		failedCIDRStatus3 := []networkv1alpha1.CIDRAllocationStatus{
			{
				CIDRAllocation: networkv1alpha1.CIDRAllocation{
					Request: validSubRangeCidr,
					CIDR:    "",
				},
				Status:  networkv1alpha1.AllocationStateBusy,
				Message: FailBusyAllocationMessage(validSubRangeCidr),
			},
		}

		It("Should clean and create base objects", func() {
			cleanUp(validParentRangeLookupKey, validSubRangeLookupKey1, validSubRangeLookupKey2, validSubRangeLookupKey3)
			createObject(validParentRangeLookupKey, nil, validParentCidr1)
			Eventually(func() *networkv1alpha1.IPAMRangeStatus {
				obj := &networkv1alpha1.IPAMRange{}
				Expect(k8sClient.Get(ctx, validParentRangeLookupKey, obj)).Should(Succeed())
				return &obj.Status
			}, timeout, interval).Should(Equal(&networkv1alpha1.IPAMRangeStatus{
				State: networkv1alpha1.IPAMRangeReady,
				CIDRs: configuredCIDRStatus,
			}))
		})

		It("Should create a valid IPAMRange with parent", func() {
			createObject(validSubRangeLookupKey1, &corev1.LocalObjectReference{
				Name: validParentRangeLookupKey.Name,
			}, validSubRangeCidr)
			Eventually(func() *IPAMStatus {
				return projectStatus(ctx, validParentRangeLookupKey)
			}, timeout, interval).Should(Equal(&IPAMStatus{
				State: networkv1alpha1.IPAMRangeReady,
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
				State: networkv1alpha1.IPAMRangeReady,
				CIDRs: allocatedCIDRStatus1,
			}))
		})

		It("Should create a second valid IPAMRange with parent", func() {
			createObject(validSubRangeLookupKey2, &corev1.LocalObjectReference{
				Name: validParentRangeLookupKey.Name,
			}, validSubRangeCidr)
			Eventually(func() *IPAMStatus {
				return projectStatus(ctx, validParentRangeLookupKey)
			}, timeout, interval).Should(Equal(&IPAMStatus{
				State: networkv1alpha1.IPAMRangeReady,
				CIDRs: configuredCIDRStatus,
				AllocationState: []string{
					"10.0.0.0/16[busy]",
				},
				PendingRequest: nil,
			}))
			Eventually(func() *IPAMStatus {
				return projectStatus(ctx, validSubRangeLookupKey2)
			}, timeout, interval).Should(Equal(&IPAMStatus{
				State: networkv1alpha1.IPAMRangeReady,
				CIDRs: allocatedCIDRStatus2,
			}))
		})

		It("Should reject a third valid IPAMRange with parent", func() {
			createObject(validSubRangeLookupKey3, &corev1.LocalObjectReference{
				Name: validParentRangeLookupKey.Name,
			}, validSubRangeCidr)
			Eventually(func() *IPAMStatus {
				return projectStatus(ctx, validParentRangeLookupKey)
			}, timeout, interval).Should(Equal(&IPAMStatus{
				State: networkv1alpha1.IPAMRangeReady,
				CIDRs: configuredCIDRStatus,
				AllocationState: []string{
					"10.0.0.0/16[busy]",
				},
				PendingRequest: nil,
			}))
			Eventually(func() *IPAMStatus {
				return projectStatus(ctx, validSubRangeLookupKey3)
			}, timeout, interval).Should(Equal(&IPAMStatus{
				State: networkv1alpha1.IPAMRangeBusy,
				CIDRs: failedCIDRStatus3,
			}))
		})

		PIt("Should extend parent with new CIDR range", func() {
			updateObject(validParentRangeLookupKey, nil, validParentCidr1, validParentCidr2)
			Eventually(func() *IPAMStatus {
				return projectStatus(ctx, validParentRangeLookupKey)
			}, timeout, interval).Should(Equal(&IPAMStatus{
				State: networkv1alpha1.IPAMRangeReady,
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
				State: networkv1alpha1.IPAMRangeReady,
				CIDRs: allocatedCIDRStatus3,
			}))
		})

	})

})
