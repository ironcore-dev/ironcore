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

package v1alpha1

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

const (
	anyExists = iota
	anyExistsNoSchedule
	anyExistsValANoSchedule
	keyAExists
	keyAExistsNoSchedule
	keyAEqualsValA
	keyARedundantlyEqualsValA
	keyAEqualsValANoSchedule
	keyARedundantlyEqualsValANoSchedule
	keyAEqualsValBNoSchedule
	keyARedundantlyEqualsValBNoSchedule
	keyBExistsNoSchedule
	keyBEqualsValBNoSchedule
)

var tolerationMap = map[int]Toleration{
	anyExists: { // tolerates any key and any effect
		Operator: TolerationOpExists,
	},
	anyExistsNoSchedule: { // tolerates any key with the effect `TaintEffectNoSchedule`
		Operator: TolerationOpExists,
		Effect:   TaintEffectNoSchedule,
	},
	anyExistsValANoSchedule: { // tolerates any key with the effect `TaintEffectNoSchedule` and ignores the value
		Operator: TolerationOpExists,
		Value:    "valA",
		Effect:   TaintEffectNoSchedule,
	},
	keyAExists: {
		Key:      "keyA",
		Operator: TolerationOpExists,
	},
	keyAExistsNoSchedule: {
		Key:      "keyA",
		Operator: TolerationOpExists,
		Effect:   TaintEffectNoSchedule,
	},
	keyAEqualsValA: {
		Key:   "keyA",
		Value: "valA",
	},
	keyARedundantlyEqualsValA: {
		Key:      "keyA",
		Operator: TolerationOpEqual,
		Value:    "valA",
	},
	keyAEqualsValANoSchedule: {
		Key:    "keyA",
		Value:  "valA",
		Effect: TaintEffectNoSchedule,
	},
	keyARedundantlyEqualsValANoSchedule: {
		Key:      "keyA",
		Operator: TolerationOpEqual,
		Value:    "valA",
		Effect:   TaintEffectNoSchedule,
	},
	keyAEqualsValBNoSchedule: {
		Key:    "keyA",
		Value:  "valB",
		Effect: TaintEffectNoSchedule,
	},
	keyARedundantlyEqualsValBNoSchedule: {
		Key:      "keyA",
		Operator: TolerationOpEqual,
		Value:    "valB",
		Effect:   TaintEffectNoSchedule,
	},
	keyBExistsNoSchedule: {
		Key:      "keyB",
		Operator: TolerationOpExists,
		Effect:   TaintEffectNoSchedule,
	},
	keyBEqualsValBNoSchedule: {
		Key:    "keyB",
		Value:  "valB",
		Effect: TaintEffectNoSchedule,
	},
}

var _ = Describe("TolerateTaints", func() {
	It("returns true upon empty taints", func() {
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

	It("returns false upon empty tolerations and non-empty taints", func() {
		taints := []Taint{
			{
				Key:    "key",
				Effect: TaintEffectNoSchedule,
			},
		}
		Expect(TolerateTaints(nil, taints)).To(BeFalse(), "expected empty tolerations don't tolerate non-empty taints")
	})

	It("returns false when tolerations don't cover all taints", func() {
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

	It("returns false when tolerations cover all taints", func() {
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

var _ = Describe("MergeTolerations", func() {
	toTolerations := func(names []int) []Toleration {
		ans := []Toleration{}
		for _, name := range names {
			ans = append(ans, tolerationMap[name])
		}
		return ans
	}

	DescribeTable("merges two tolerations into one",
		func(a, b, merged []int) {
			Expect(MergeTolerations(toTolerations(a), toTolerations(b))).To(Equal(toTolerations(merged)))
		},
		Entry(
			"disjointed",
			[]int{keyAEqualsValANoSchedule, keyBEqualsValBNoSchedule},
			[]int{keyAEqualsValBNoSchedule},
			[]int{keyAEqualsValANoSchedule, keyBEqualsValBNoSchedule, keyAEqualsValBNoSchedule},
		),
		Entry(
			"duplicate",
			[]int{keyAEqualsValANoSchedule, keyBEqualsValBNoSchedule},
			[]int{keyAEqualsValANoSchedule, keyBEqualsValBNoSchedule},
			[]int{keyAEqualsValANoSchedule, keyBEqualsValBNoSchedule},
		),
		Entry(
			"redundant",
			[]int{keyAEqualsValANoSchedule, keyAEqualsValBNoSchedule},
			[]int{keyAExistsNoSchedule, keyBEqualsValBNoSchedule},
			[]int{keyAExistsNoSchedule, keyBEqualsValBNoSchedule},
		),
	)
})

var _ = Describe("isSuperset", func() {
	contains := func(nums []int, target int) bool {
		for _, num := range nums {
			if num == target {
				return true
			}
		}
		return false
	}

	superSubPairs := []struct {
		superKey int
		subKeys  []int
	}{{
		anyExists,
		[]int{
			anyExistsNoSchedule,
			anyExistsValANoSchedule,
			keyAExists,
			keyAExistsNoSchedule,
			keyAEqualsValA,
			keyARedundantlyEqualsValA,
			keyAEqualsValANoSchedule,
			keyARedundantlyEqualsValANoSchedule,
			keyAEqualsValBNoSchedule,
			keyARedundantlyEqualsValBNoSchedule,
			keyBExistsNoSchedule,
			keyBEqualsValBNoSchedule,
		},
	}, {
		anyExistsNoSchedule,
		[]int{
			anyExistsValANoSchedule,
			keyAExistsNoSchedule,
			keyAEqualsValANoSchedule,
			keyARedundantlyEqualsValANoSchedule,
			keyAEqualsValBNoSchedule,
			keyARedundantlyEqualsValBNoSchedule,
			keyBExistsNoSchedule,
			keyBEqualsValBNoSchedule,
		},
	}, {
		keyAExists,
		[]int{
			keyAEqualsValA,
			keyARedundantlyEqualsValA,
			keyAExistsNoSchedule,
			keyAEqualsValANoSchedule,
			keyARedundantlyEqualsValANoSchedule,
			keyAEqualsValBNoSchedule,
			keyARedundantlyEqualsValBNoSchedule,
		},
	}, {
		keyAExistsNoSchedule,
		[]int{
			keyAEqualsValANoSchedule,
			keyARedundantlyEqualsValANoSchedule,
			keyAEqualsValBNoSchedule,
			keyARedundantlyEqualsValBNoSchedule,
		},
	}, {
		keyAEqualsValA,
		[]int{
			keyARedundantlyEqualsValA,
			keyAEqualsValANoSchedule,
			keyARedundantlyEqualsValANoSchedule,
		},
	}}

	It("confirms a toleration is a superset of itsself", func() {
		for _, toleration := range tolerationMap {
			Expect(isSuperset(&toleration, &toleration)).To(BeTrue(), "expected %v is a superset of itself", toleration)
		}
	})

	It("tells if a toleration is a superset of another toleration", func() {
		for subKey, sub := range tolerationMap {
			for _, pair := range superSubPairs {
				super := tolerationMap[pair.superKey]
				if contains(pair.subKeys, subKey) { // sub is among the subsets
					Expect(isSuperset(&super, &sub)).To(BeTrue(), "expected %v is a superset of %v", super, sub)
				} else if subKey != pair.superKey { // nothing to do with this pair
					Expect(isSuperset(&super, &sub)).To(BeFalse(), "expected %v is not a superset of %v", super, sub)
				}
			}
		}
	})
})
