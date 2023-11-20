/*
 * Copyright (c) 2022 by the IronCore authors.
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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

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
