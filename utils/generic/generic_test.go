// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package generic_test

import (
	. "github.com/ironcore-dev/ironcore/utils/generic"
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
