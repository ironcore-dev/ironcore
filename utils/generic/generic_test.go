// Copyright 2023 OnMetal authors
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

package generic_test

import (
	. "github.com/onmetal/onmetal-api/utils/generic"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Generic", func() {
	Describe("Identity", func() {
		It("should return the value it's given", func() {
			Expect(Identity("foo")).To(Equal("foo"))
			Expect(Identity(1)).To(Equal(1))
		})
	})

	Describe("Zero", func() {
		It("should return the zero value for the given type", func() {
			Expect(Zero[int]()).To(Equal(0))
			Expect(Zero[func()]()).To(BeNil())
		})
	})

	Describe("Pointer", func() {
		It("should return a pointer to the given value", func() {
			Expect(Pointer("foo")).To(PointTo(Equal("foo")))
			Expect(Pointer(1)).To(PointTo(Equal(1)))
		})
	})

	Describe("DerefFunc", func() {
		It("return the value if the pointer is non-nil", func() {
			Expect(DerefFunc(Pointer(42), func() int {
				Fail("should not be called")
				return 0
			})).To(Equal(42))
		})

		It("should call the function if the pointer is nil", func() {
			Expect(DerefFunc(nil, func() int { return 42 })).To(Equal(42))
		})
	})

	Describe("Deref", func() {
		It("return the value if the pointer is non-nil", func() {
			Expect(Deref(Pointer(42), 0)).To(Equal(42))
		})

		It("should call the function if the pointer is nil", func() {
			Expect(Deref(nil, 42)).To(Equal(42))
		})
	})
})
