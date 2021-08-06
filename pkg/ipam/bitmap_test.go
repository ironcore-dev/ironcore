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
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bitmap", func() {

	Context("bitmaps", func() {
		It("initializes correctly", func() {
			fmt.Println()
			for i := 0; i <= 6; i++ {
				fmt.Printf("%2d: %d\n", i, hostmask[i])
			}

			Expect(hostmask[3]).To(Equal(Bitmap(255)))
		})

		It("check 4", func() {
			b := bitmapHostMask(4)
			Expect(b.isAllocated(0, 4)).To(BeTrue())
			Expect(b.isAllocated(4, 4)).To(BeFalse())

			Expect(b.isFree(4, 4)).To(BeTrue())
			Expect(b.isFree(0, 4)).To(BeFalse())

			Expect(b.canAllocate(4)).To(Equal(4))
			Expect(b.canAllocate(3)).To(Equal(8))
		})
		It("check 3", func() {
			b := bitmapHostMask(3)
			Expect(b.canAllocate(3)).To(Equal(8))
		})

		It("allocate 3", func() {
			b := bitmapHostMask(3)
			Expect(b.allocate(3)).To(Equal(8))
			Expect(b).To(Equal(Bitmap(bitmapHostMask(2))))
			Expect(b.canAllocate(3)).To(Equal(16))
		})
		It("check 4/3", func() {
			b := bitmapHostMask(4)
			Expect(b.allocate(3)).To(Equal(8))
			Expect(b).To(Equal(Bitmap(bitmapHostMask(4) + bitmapHostMask(3)<<8)))
			Expect(b.allocate(4)).To(Equal(4))
			Expect(b).To(Equal(Bitmap(bitmapHostMask(2))))
		})

		It("check MAX", func() {
			b := bitmapHostMask(4)
			Expect(b.allocate(MAX_BITMAP_NET)).To(Equal(4))
			Expect(b.isAllocated(0, 3)).To(BeFalse())
			Expect(b.isFree(0, 3)).To(BeFalse())
			Expect(b.isAllocated(4, MAX_BITMAP_NET-1)).To(BeFalse())
			Expect(b.allocate(MAX_BITMAP_NET)).To(Equal(5))
			Expect(b.isAllocated(4, MAX_BITMAP_NET-1)).To(BeTrue())
		})
	})
})
