// Copyright 2022 IronCore authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package device_test

import (
	"fmt"

	. "github.com/ironcore-dev/ironcore/internal/admission/plugin/machinevolumedevices/device"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

var _ = Describe("Device", func() {
	DescribeTable("ParseName",
		func(name string, matchers ...types.GomegaMatcher) {
			prefix, idx, err := ParseName(name)
			switch len(matchers) {
			case 1:
				Expect(err).To(matchers[0])
			case 2:
				Expect(err).NotTo(HaveOccurred())
				Expect(prefix).To(matchers[0])
				Expect(idx).To(matchers[1])
			default:
				Fail(fmt.Sprintf("invalid number of matchers: %d, expected 1 (error case) / 2 (success case)", len(matchers)))
			}
		},
		Entry("minimum index", "sda", Equal("sd"), Equal(0)),
		Entry("other prefix", "fda", Equal("fd"), Equal(0)),
		Entry("simple valid name", "sdb", Equal("sd"), Equal(1)),
		Entry("two-letter name", "sdbb", Equal("sd"), Equal(53)),
		Entry("maximum index", "sdzz", Equal("sd"), Equal(MaxIndex)),
		Entry("more than two letters", "sdaaa", HaveOccurred()),
		Entry("invalid prefix", "f_aa", HaveOccurred()),
	)

	DescribeTable("Name",
		func(prefix string, index int, matchers ...types.GomegaMatcher) {
			if len(matchers) == 2 {
				Expect(func() { Name(prefix, index) }).To(matchers[1])
			} else {
				Expect(Name(prefix, index)).To(matchers[0])
			}
		},
		Entry("minimum index", "sd", 0, Equal("sda")),
		Entry("other index", "sd", 22, Equal("sdw")),
		Entry("maximum index", "sd", MaxIndex, Equal("sdzz")),
		Entry("negative index", "sd", -1, nil, Panic()),
		Entry("too large index", "sd", MaxIndex+1, nil, Panic()),
		Entry("invalid prefix", "f_a", 1, nil, Panic()),
	)

	Context("Namer", func() {
		var namer *Namer
		BeforeEach(func() {
			namer = NewNamer()
		})

		Describe("Observe", func() {
			It("should allow claiming names", func() {
				By("claiming some names")
				Expect(namer.Observe("foo")).To(Succeed())
				Expect(namer.Observe("bar")).To(Succeed())
				Expect(namer.Observe("baz")).To(Succeed())

				By("claiming a name that already has been claimed")
				Expect(namer.Observe("foo")).To(HaveOccurred())
			})

			It("should error on invalid names", func() {
				Expect(namer.Observe("invalid")).To(HaveOccurred())
			})
		})

		Describe("Free", func() {
			It("should allow free names, making them reclaimable", func() {
				By("claiming some names")
				Expect(namer.Observe("foo")).To(Succeed())
				Expect(namer.Observe("bar")).To(Succeed())

				By("freeing the names")
				Expect(namer.Free("foo")).To(Succeed())
				Expect(namer.Free("bar")).To(Succeed())

				By("claiming the names again")
				Expect(namer.Observe("foo")).To(Succeed())
				Expect(namer.Observe("bar")).To(Succeed())

				By("freeing a name that has not been claimed")
				Expect(namer.Free("qux")).To(HaveOccurred())
			})

			It("should error on invalid names", func() {
				Expect(namer.Free("invalid")).To(HaveOccurred())
			})
		})

		Describe("Generate", func() {
			It("should generate names with the desired prefix", func() {
				Expect(namer.Generate("sd")).To(Equal("sda"))
				Expect(namer.Generate("sd")).To(Equal("sdb"))
				Expect(namer.Generate("vd")).To(Equal("vda"))
			})

			It("should error if all names have been claimed", func() {
				By("claiming all available names")
				for i := 0; i < MaxIndex; i++ {
					_, err := namer.Generate("sd")
					Expect(err).NotTo(HaveOccurred())
				}

				By("trying to generate again")
				_, err := namer.Generate("sd")
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
