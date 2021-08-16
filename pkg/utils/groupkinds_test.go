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

package utils

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("GroupKinds", func() {
	var groupKinds GroupKinds
	idg1k1 := schema.GroupKind{
		Group: "g1",
		Kind:  "k1",
	}
	idg1k2 := schema.GroupKind{
		Group: "g1",
		Kind:  "k2",
	}
	idg2k1 := schema.GroupKind{
		Group: "g2",
		Kind:  "k1",
	}

	BeforeEach(func() {
		groupKinds = NewGroupKinds()
	})

	Context("Adding", func() {
		It("It should add one element", func() {
			groupKinds.Add(idg1k1)
			Expect(groupKinds.Contains(idg1k1)).Should(BeTrue())
		})
		It("It should add multiple elements", func() {
			groupKinds.Add(idg1k1)
			groupKinds.Add(idg1k2)
			Expect(groupKinds.Contains(idg1k1)).Should(BeTrue())
			Expect(groupKinds.Contains(idg1k2)).Should(BeTrue())
		})
		It("It should add multiple elements", func() {
			groupKindsTemp := NewGroupKinds(idg1k1, idg1k2)
			groupKinds.AddAll(groupKindsTemp)
			Expect(groupKinds.Contains(idg1k1)).Should(BeTrue())
			Expect(groupKinds.Contains(idg1k2)).Should(BeTrue())
		})
		It("It should join two GroupKinds", func() {
			groupKindsTemp := NewGroupKinds(idg1k1, idg1k2)
			groupKinds.Add(idg2k1)
			joined := groupKinds.Join(groupKindsTemp)
			Expect(joined.Contains(idg1k1)).Should(BeTrue())
			Expect(joined.Contains(idg1k2)).Should(BeTrue())
			Expect(joined.Contains(idg2k1)).Should(BeTrue())
		})
	})

	Context("Compare", func() {
		It("It should compare two equal GroupKinds", func() {
			groupKinds.Add(idg1k1)
			Expect(groupKinds.Equal(NewGroupKinds(idg1k1))).Should(BeTrue())
		})
		It("It should compare two unequal GroupKinds", func() {
			groupKinds.Add(idg1k1)
			Expect(groupKinds.Equal(NewGroupKinds(idg1k2))).Should(BeFalse())
		})
		It("It should compare two unequal GroupKinds", func() {
			groupKinds.Add(idg1k1)
			Expect(groupKinds.Equal(NewGroupKinds())).Should(BeFalse())
		})
		It("It should compare two empty GroupKinds", func() {
			Expect(groupKinds.Equal(NewGroupKinds())).Should(BeTrue())
		})
	})

	Context("Output", func() {
		It("It should render a correct string", func() {
			groupKinds.Add(idg1k1)
			output := "[k1.g1]"
			Expect(groupKinds.String()).Should(Equal(output))
		})
		It("It should render a correct string for multiple GroupKinds", func() {
			groupKinds.Add(idg1k1)
			groupKinds.Add(idg1k2)
			output1 := "[k1.g1,k2.g1]"
			output2 := "[k2.g1,k1.g1]"
			Expect(groupKinds.String()).Should(Or(Equal(output1), Equal(output2)))
		})
	})

	Context("Remove", func() {
		It("It should remove a single element", func() {
			groupKinds.Add(idg1k1)
			groupKinds.Add(idg1k2)
			Expect(groupKinds.Contains(idg1k1)).Should(BeTrue())
			Expect(groupKinds.Contains(idg1k2)).Should(BeTrue())
			groupKinds.Remove(idg1k2)
			Expect(groupKinds.Contains(idg1k2)).Should(BeFalse())
		})
	})
})
