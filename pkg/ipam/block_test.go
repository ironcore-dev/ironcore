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

var _ = Describe("Blocks", func() {
	Context("split", func() {
		It("split large free block", func() {
			cidr := MustParseCIDR("10.0.0.0/8")
			block := NewBlock(cidr, 0)

			split := block.split()

			Expect(split.cidr.String()).To(Equal("10.128.0.0/9"))
			Expect(block.cidr.String()).To(Equal("10.0.0.0/9"))
			Expect(split.busy).To(Equal(Bitmap(0)))
			Expect(block.busy).To(Equal(Bitmap(0)))
		})
		It("split large busy block", func() {
			cidr := MustParseCIDR("10.0.0.0/8")
			block := NewBlock(cidr, BITMAP_BUSY)

			split := block.split()

			Expect(split.cidr.String()).To(Equal("10.128.0.0/9"))
			Expect(block.cidr.String()).To(Equal("10.0.0.0/9"))
			Expect(split.busy).To(Equal(BITMAP_BUSY))
			Expect(block.busy).To(Equal(BITMAP_BUSY))
		})

		It("split small free block", func() {
			cidr := MustParseCIDR("10.0.0.0/26")
			block := NewBlock(cidr, 0)

			split := block.split()

			Expect(split.cidr.String()).To(Equal("10.0.0.32/27"))
			Expect(block.cidr.String()).To(Equal("10.0.0.0/27"))
			Expect(split.busy).To(Equal(Bitmap(0)))
			Expect(block.busy).To(Equal(Bitmap(0)))
		})
		It("split small busy block", func() {
			cidr := MustParseCIDR("10.0.0.0/26")
			block := NewBlock(cidr, BITMAP_BUSY)

			split := block.split()

			Expect(split.cidr.String()).To(Equal("10.0.0.32/27"))
			Expect(block.cidr.String()).To(Equal("10.0.0.0/27"))
			Expect(split.busy).To(Equal(hostmask[5]))
			Expect(block.busy).To(Equal(hostmask[5]))
		})
		It("split small partly busy block", func() {
			b1 := Bitmap(5)
			b2 := Bitmap(8)
			cidr := MustParseCIDR("10.0.0.0/26")
			block := NewBlock(cidr, b1<<32|b2)

			split := block.split()

			Expect(split.cidr.String()).To(Equal("10.0.0.32/27"))
			Expect(block.cidr.String()).To(Equal("10.0.0.0/27"))
			Expect(split.busy).To(Equal(b1))
			Expect(block.busy).To(Equal(b2))
		})
	})
	Context("split", func() {
		It("joins large free blocks", func() {
			cidr1 := MustParseCIDR("10.128.0.0/9")
			cidr2 := MustParseCIDR("10.0.0.0/9")
			block1 := NewBlock(cidr1, 0)
			block2 := NewBlock(cidr2, 0)
			block2.next = block1
			block1.prev = block2

			joined := block1.join()

			Expect(joined.cidr.String()).To(Equal("10.0.0.0/8"))
			Expect(joined.busy).To(Equal(Bitmap(0)))
		})
		It("joins large busy blocks", func() {
			cidr1 := MustParseCIDR("10.128.0.0/9")
			cidr2 := MustParseCIDR("10.0.0.0/9")
			block1 := NewBlock(cidr1, BITMAP_BUSY)
			block2 := NewBlock(cidr2, BITMAP_BUSY)
			block2.next = block1
			block1.prev = block2

			joined := block1.join()

			Expect(joined.cidr.String()).To(Equal("10.0.0.0/8"))
			Expect(joined.busy).To(Equal(BITMAP_BUSY))
		})

		It("joins small free blocks", func() {
			cidr1 := MustParseCIDR("10.0.0.32/27")
			cidr2 := MustParseCIDR("10.0.0.0/27")
			block1 := NewBlock(cidr1, 0)
			block2 := NewBlock(cidr2, 0)
			block2.next = block1
			block1.prev = block2

			joined := block1.join()

			Expect(joined.cidr.String()).To(Equal("10.0.0.0/26"))
			Expect(joined.busy).To(Equal(Bitmap(0)))
		})
		It("joins small busy blocks", func() {
			cidr1 := MustParseCIDR("10.0.0.32/27")
			cidr2 := MustParseCIDR("10.0.0.0/27")
			block1 := NewBlock(cidr1, hostmask[5])
			block2 := NewBlock(cidr2, hostmask[5])
			block2.next = block1
			block1.prev = block2

			joined := block1.join()

			Expect(joined.cidr.String()).To(Equal("10.0.0.0/26"))
			Expect(joined.busy).To(Equal(BITMAP_BUSY))
		})
		It("joins small partly busy blocks", func() {
			b1 := Bitmap(5)
			b2 := Bitmap(8)
			cidr1 := MustParseCIDR("10.0.0.32/27")
			cidr2 := MustParseCIDR("10.0.0.0/27")
			block1 := NewBlock(cidr1, b1)
			block2 := NewBlock(cidr2, b2)
			block2.next = block1
			block1.prev = block2

			joined := block1.join()

			Expect(joined.cidr.String()).To(Equal("10.0.0.0/26"))
			Expect(joined.busy).To(Equal(b1<<32 | b2))
		})
	})
})
