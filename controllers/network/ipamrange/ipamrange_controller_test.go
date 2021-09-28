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
	"fmt"
	networkv1alpha1 "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = OptionalDescribe("IPAMRange controller", func() {

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

		configuredCIDRStatus := []networkv1alpha1.CIDRAllocationStatus{
			networkv1alpha1.CIDRAllocationStatus{
				CIDRAllocation: networkv1alpha1.CIDRAllocation{
					Request: validParentCidrs,
					CIDR:    validParentCidrs,
				},
				Status:  networkv1alpha1.AllocationStateAllocated,
				Message: SuccessfulUsageMessage,
			},
		}

		allocatedCIDRStatus := []networkv1alpha1.CIDRAllocationStatus{
			networkv1alpha1.CIDRAllocationStatus{
				CIDRAllocation: networkv1alpha1.CIDRAllocation{
					Request: validSubRangeCidrs,
					CIDR:    "10.0.1.0/24",
				},
				Status:  networkv1alpha1.AllocationStateAllocated,
				Message: SuccessfulAllocationMessage,
			},
		}

		It("Should clean and create base objects", func() {
			cleanUp(validParentRangeLookupKey, validSubRangeLookupKey)
			createObject(validParentRangeLookupKey, nil, validParentCidrs)
		})

		It("Should set the State to Ready for valid IPAMRange", func() {
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
			createObject(validSubRangeLookupKey, &corev1.LocalObjectReference{
				Name: validParentRangeLookupKey.Name,
			}, validSubRangeCidrs)
		})

		It("Should set correct Status of valid IPAMRange", func() {
			Eventually(func() *IPAMStatus {
				return projectStatus(ctx, validParentRangeLookupKey)
			}, timeout, interval).Should(Equal(&IPAMStatus{
				State: networkv1alpha1.IPAMRangeReady,
				CIDRs: configuredCIDRStatus,
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
				PendingRequest: nil,
			}))
		})

		It("Should set correct status for valid sub IPAMRange", func() {
			Eventually(func() *IPAMStatus {
				return projectStatus(ctx, validSubRangeLookupKey)
			}, timeout, interval).Should(Equal(&IPAMStatus{
				State: networkv1alpha1.IPAMRangeReady,
				CIDRs: allocatedCIDRStatus,
			}))
		})

		PIt("Should release the allocated CIDR block on request deletion", func() {
			By("Deleting the request IPAMRange object")
			obj := &networkv1alpha1.IPAMRange{}
			Expect(k8sClient.Get(ctx, validSubRangeLookupKey, obj)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, obj)).Should(Succeed())
			By("Checking whether the allocation state is not containing the reserved range")
			Eventually(func() *IPAMStatus {
				return projectStatus(ctx, validParentRangeLookupKey)
			}, timeout, interval).Should(Equal(&IPAMStatus{
				State:           networkv1alpha1.IPAMRangeReady,
				CIDRs:           configuredCIDRStatus,
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
			createObject(invalidSubRangeLookupKey, &corev1.LocalObjectReference{
				Name: invalidParentRangeLookupKey.Name,
			}, invalidSubRangeCidrs)
		})

		It("Should set the State to Invalid", func() {
			Eventually(func() *IPAMStatus {
				return projectStatus(ctx, invalidParentRangeLookupKey)
			}, timeout, interval).Should(Equal(&IPAMStatus{
				State: networkv1alpha1.IPAMRangeInvalid,
			}))
		})

		It("Should set the State to Invalid for sub IPAMRange with Invalid parent", func() {
			Eventually(func() *IPAMStatus {
				return projectStatus(ctx, invalidSubRangeLookupKey)
			}, timeout, interval).Should(Equal(&IPAMStatus{
				State: networkv1alpha1.IPAMRangeError,
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

		configuredCIDRStatus := []networkv1alpha1.CIDRAllocationStatus{
			networkv1alpha1.CIDRAllocationStatus{
				CIDRAllocation: networkv1alpha1.CIDRAllocation{
					Request: validParentCidrs,
					CIDR:    validParentCidrs,
				},
				Status:  networkv1alpha1.AllocationStateAllocated,
				Message: SuccessfulUsageMessage,
			},
		}

		allocatedCIDRStatus := []networkv1alpha1.CIDRAllocationStatus{
			networkv1alpha1.CIDRAllocationStatus{
				CIDRAllocation: networkv1alpha1.CIDRAllocation{
					Request: validSubRangeCidrs,
					CIDR:    allocatedSubRangeCidr,
				},
				Status:  networkv1alpha1.AllocationStateAllocated,
				Message: SuccessfulAllocationMessage,
			},
		}

		It("Should clean and create base objects", func() {
			cleanUp(validParentRangeLookupKey, validSubRangeLookupKey)
			createObject(validParentRangeLookupKey, nil, validParentCidrs)
			createObject(validSubRangeLookupKey, &corev1.LocalObjectReference{
				Name: validParentRangeLookupKey.Name,
			}, validSubRangeCidrs)
		})

		It("Should check that parent and sub range have the correct status", func() {
			Eventually(func() *IPAMStatus {
				return projectStatus(ctx, validParentRangeLookupKey)
			}, timeout, interval).Should(Equal(&IPAMStatus{
				State: networkv1alpha1.IPAMRangeReady,
				CIDRs: configuredCIDRStatus,
				AllocationState: []string{
					"10.0.0.0/26[00000010]",
					"10.0.0.64/26[free]",
					"10.0.0.128/25[free]",
					"10.0.1.0/24[free]",
					"10.0.2.0/23[free]",
					"10.0.4.0/22[free]",
					"10.0.8.0/21[free]",
					"10.0.16.0/20[free]",
					"10.0.32.0/19[free]",
					"10.0.64.0/18[free]",
					"10.0.128.0/17[free]",
				},
			}))
			Eventually(func() *IPAMStatus {
				return projectStatus(ctx, validSubRangeLookupKey)
			}, timeout, interval).Should(Equal(&IPAMStatus{
				State: networkv1alpha1.IPAMRangeReady,
				CIDRs: allocatedCIDRStatus,
			}))
		})

		It("Should create a second subrange with the same request CIDR", func() {
			createObject(validSubRange2LookupKey, &corev1.LocalObjectReference{
				Name: validParentRangeLookupKey.Name,
			}, validSubRange2Cidrs)
		})

		It("Should set the second subrange to busy", func() {
			Eventually(func() []networkv1alpha1.CIDRAllocationStatus {
				obj := &networkv1alpha1.IPAMRange{}
				Expect(k8sClient.Get(ctx, validSubRange2LookupKey, obj)).Should(Succeed())
				return obj.Status.CIDRs
			}, timeout, interval).Should(Equal([]networkv1alpha1.CIDRAllocationStatus{
				{
					CIDRAllocation: networkv1alpha1.CIDRAllocation{
						Request: validSubRange2Cidrs,
						CIDR:    "",
					},
					Status:  networkv1alpha1.AllocationStateBusy,
					Message: FailBusyAllocationMessage(validSubRange2Cidrs),
				},
			}))
		})

		PIt("Should allocate a range for the second subrange when the first one is deleted", func() {
			obj := &networkv1alpha1.IPAMRange{}
			// get first one
			Expect(k8sClient.Get(ctx, validSubRangeLookupKey, obj)).Should(Succeed())
			// delete first one
			Expect(k8sClient.Delete(ctx, obj)).Should(Succeed())
			// check if second moved from busy -> ready
			Eventually(func() *IPAMStatus {
				return projectStatus(ctx, validSubRange2LookupKey)
			}, timeout, interval).Should(Equal(&IPAMStatus{
				State: networkv1alpha1.IPAMRangeReady,
				CIDRs: allocatedCIDRStatus,
			}))
		})
	})
})
