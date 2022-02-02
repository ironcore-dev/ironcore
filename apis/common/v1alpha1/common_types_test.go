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
	. "github.com/onsi/gomega"
)

var tolerationMap = map[string]Toleration{ // map key format: key-value-effect
	"all": {Operator: TolerationOpExists},
	"all-nosched": {
		Operator: TolerationOpExists,
		Effect:   TaintEffectNoSchedule,
	},
	"foo": {
		Key:      "foo",
		Operator: TolerationOpExists,
	},
	"foo-bar": {
		Key:      "foo",
		Operator: TolerationOpEqual,
		Value:    "bar",
	},
	"foo-nosched": {
		Key:      "foo",
		Operator: TolerationOpExists,
		Effect:   TaintEffectNoSchedule,
	},
	"foo-bar-nosched": {
		Key:      "foo",
		Operator: TolerationOpEqual,
		Value:    "bar",
		Effect:   TaintEffectNoSchedule,
	},
	"foo-baz-nosched": {
		Key:      "foo",
		Operator: TolerationOpEqual,
		Value:    "baz",
		Effect:   TaintEffectNoSchedule,
	},
	"faz-nosched": {
		Key:      "faz",
		Operator: TolerationOpExists,
		Effect:   TaintEffectNoSchedule,
	},
	"faz-baz-nosched": {
		Key:      "faz",
		Operator: TolerationOpEqual,
		Value:    "baz",
		Effect:   TaintEffectNoSchedule,
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
	It("merges two tolerations into one", func() {
		toTolerations := func(names []string) []Toleration {
			ans := []Toleration{}
			for _, name := range names {
				ans = append(ans, tolerationMap[name])
			}
			return ans
		}

		tests := []struct {
			name     string
			a, b     []string
			expected []string
		}{{
			name:     "disjoint",
			a:        []string{"foo-bar-nosched", "faz-baz-nosched"},
			b:        []string{"foo-baz-nosched"},
			expected: []string{"foo-bar-nosched", "faz-baz-nosched", "foo-baz-nosched"},
		}, {
			name:     "duplicate",
			a:        []string{"foo-bar-nosched", "faz-baz-nosched"},
			b:        []string{"foo-bar-nosched", "faz-baz-nosched"},
			expected: []string{"foo-bar-nosched", "faz-baz-nosched"},
		}, {
			name:     "merge redundant",
			a:        []string{"foo-bar-nosched", "foo-baz-nosched"},
			b:        []string{"foo-nosched", "faz-baz-nosched"},
			expected: []string{"foo-nosched", "faz-baz-nosched"},
		}}

		for _, test := range tests {
			Expect(MergeTolerations(toTolerations(test.a), toTolerations(test.b))).To(Equal(toTolerations(test.expected)))
		}
	})
})

var _ = Describe("isSuperset", func() {
	It("tells if a toleration is a superset of another toleration", func() {
		superSubPairs := []struct {
			superset string
			subsets  []string
		}{{
			"all",
			[]string{"all-nosched", "foo", "foo-bar", "foo-nosched", "foo-bar-nosched", "foo-baz-nosched", "faz-nosched", "faz-baz-nosched"},
		}, {
			"all-nosched",
			[]string{"foo-nosched", "foo-bar-nosched", "foo-baz-nosched", "faz-nosched", "faz-baz-nosched"},
		}, {
			"foo",
			[]string{"foo-bar", "foo-nosched", "foo-bar-nosched", "foo-baz-nosched"},
		}, {
			"foo-nosched",
			[]string{"foo-bar-nosched", "foo-baz-nosched"},
		}, {
			"foo-bar",
			[]string{"foo-bar-nosched"},
		}}

		contains := func(ss []string, target string) bool {
			for _, s := range ss {
				if s == target {
					return true
				}
			}
			return false
		}

		for key := range tolerationMap {
			for _, pair := range superSubPairs {
				super := tolerationMap[pair.superset]
				sub := tolerationMap[key]
				if key == pair.superset || contains(pair.subsets, key) { // tolerations[key] is the superset or it's among the subsets
					Expect(isSuperset(&super, &sub)).To(BeTrue(), "expected %v is a superset of %v", super, sub)
				} else { // nothing to do with this pair
					Expect(isSuperset(&super, &sub)).To(BeFalse(), "expected %v is not a superset of %v", super, sub)
				}
			}
		}
	})
})
