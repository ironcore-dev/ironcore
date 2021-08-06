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

package ipam

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("INT", func() {
	Context("int", func() {
		It("add", func() {
			r := Int64(2).Add(Int64(8))
			Expect(r).To(Equal(Int64(10)))
		})
		It("sub", func() {
			r := Int64(8).Sub(Int64(2))
			Expect(r).To(Equal(Int64(6)))
		})
		It("mul", func() {
			r := Int64(3).Mul(Int64(15))
			Expect(r).To(Equal(Int64(45)))
		})
		It("div", func() {
			r := Int64(15).Div(Int64(3))
			Expect(r).To(Equal(Int64(5)))
		})
		It("mod", func() {
			r := Int64(15).Mod(Int64(4))
			Expect(r).To(Equal(Int64(3)))
		})
		It("<<", func() {
			r := Int64(1).LShift(128)
			Expect(r).To(Equal(Int64(64).LShift(122)))
		})
		It("and", func() {
			r := Int64(1024 + 126).And(IntTwoFiveFive)
			Expect(r).To(Equal(Int64(126)))
		})
		It("or", func() {
			r := Int64(1024 + 126).Or(IntTwoFiveFive)
			Expect(r).To(Equal(Int64(1024 + 255)))
		})
	})
})
