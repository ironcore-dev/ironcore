// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TolerateTaints", func() {
	It("should return true upon empty taints", func() {
		Expect(TolerateTaints(nil, nil)).To(BeTrue(), "expected empty tolerations tolerate empty taints")

		tolerations := []Toleration{
			{
				Key:      "key",
				Effect:   TaintEffectNoSchedule,
				Operator: TolerationOpExists,
			},
		}
		Expect(TolerateTaints(tolerations, nil)).To(BeTrue(), "expected non-empty tolerations tolerate empty taints")
	})

	It("should return false upon empty tolerations and non-empty taints", func() {
		taints := []Taint{
			{
				Key:    "key",
				Effect: TaintEffectNoSchedule,
			},
		}
		Expect(TolerateTaints(nil, taints)).To(BeFalse(), "expected empty tolerations don't tolerate non-empty taints")
	})

	It("should return false when tolerations don't cover all taints", func() {
		tolerations := []Toleration{
			{
				Key:      "key",
				Effect:   TaintEffectNoSchedule,
				Operator: TolerationOpExists,
			},
		}
		taints := []Taint{
			{
				Key:    "key",
				Effect: TaintEffectNoSchedule,
			},
			{
				Key:    "key1",
				Value:  "value1",
				Effect: TaintEffectNoSchedule,
			},
		}
		Expect(TolerateTaints(tolerations, taints)).To(BeFalse(), "expected the tolerations don't cover all the taints")
	})

	It("should return false when tolerations cover all taints", func() {
		tolerations := []Toleration{
			{
				Key:      "key",
				Effect:   TaintEffectNoSchedule,
				Operator: TolerationOpExists,
			},
			{
				Key:      "key1",
				Value:    "value1",
				Effect:   TaintEffectNoSchedule,
				Operator: TolerationOpEqual,
			},
		}
		taints := []Taint{
			{
				Key:    "key",
				Effect: TaintEffectNoSchedule,
			},
			{
				Key:    "key1",
				Value:  "value1",
				Effect: TaintEffectNoSchedule,
			},
		}
		Expect(TolerateTaints(tolerations, taints)).To(BeTrue(), "expected the tolerations cover all the taints")
	})
})
