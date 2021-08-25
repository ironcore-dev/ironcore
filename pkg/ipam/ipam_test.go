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
	"net"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("IPAM", func() {
	Context("using complete blocks", func() {
		_, cidr, _ := net.ParseCIDR("10.0.0.0/8")

		It("initializes ipam correctly", func() {
			ipam, _ := NewIPAM(cidr)

			Expect(ipam.block.next).To(BeNil())
			Expect(ipam.block.prev).To(BeNil())
		})

		It("initializes splits", func() {
			ipam, _ := NewIPAM(cidr)

			r := ipam.Alloc(9)
			Expect(r.String()).To(Equal("10.0.0.0/9"))

			r = ipam.Alloc(10)
			Expect(r.String()).To(Equal("10.128.0.0/10"))

			Expect(ipam.String()).To(Equal("10.0.0.0/9[busy], 10.128.0.0/10[busy], 10.192.0.0/10[free]"))
		})

		It("free", func() {
			ipam, _ := NewIPAM(cidr)

			r1 := ipam.Alloc(9)
			Expect(r1.String()).To(Equal("10.0.0.0/9"))

			Expect(ipam.Free(r1)).To(BeTrue())

			Expect(ipam.block.next).To(BeNil())
			Expect(ipam.block.prev).To(BeNil())
		})

		It("busy", func() {
			ipam, _ := NewIPAM(cidr)

			r1 := MustParseCIDR("10.128.0.0/10")
			Expect(ipam.Busy(r1)).To(BeTrue())
			Expect(ipam.String()).To(Equal("10.0.0.0/9[free], 10.128.0.0/10[busy], 10.192.0.0/10[free]"))

			r2 := MustParseCIDR("10.128.1.0/24")
			Expect(ipam.Busy(r2)).To(BeFalse())
			Expect(ipam.String()).To(Equal("10.0.0.0/9[free], 10.128.0.0/10[busy], 10.192.0.0/10[free]"))

			ipam.Free(r1)
			Expect(ipam.String()).To(Equal("10.0.0.0/8[free]"))
		})

		It("no round robin", func() {
			ipam, _ := NewIPAM(cidr)
			ipam.SetRoundRobin(false)

			r1 := ipam.Alloc(9)
			Expect(r1.String()).To(Equal("10.0.0.0/9"))
			ipam.Free(r1)
			r1 = ipam.Alloc(9)
			Expect(r1.String()).To(Equal("10.0.0.0/9"))
			ipam.Free(r1)
			Expect(ipam.String()).To(Equal("10.0.0.0/8[free]"))
		})

		It("no round robin", func() {
			ipam, _ := NewIPAM(cidr)
			ipam.SetRoundRobin(false)

			r1 := ipam.Alloc(9)
			Expect(r1.String()).To(Equal("10.0.0.0/9"))
			ipam.Free(r1)
			r1 = ipam.Alloc(9)
			Expect(r1.String()).To(Equal("10.0.0.0/9"))
			ipam.Free(r1)
			Expect(ipam.String()).To(Equal("10.0.0.0/8[free]"))
		})
		It("round robin", func() {
			ipam, _ := NewIPAM(cidr)
			ipam.SetRoundRobin(true)

			r1 := ipam.Alloc(9)
			Expect(r1.String()).To(Equal("10.0.0.0/9"))
			ipam.Free(r1)
			r1 = ipam.Alloc(9)
			Expect(r1.String()).To(Equal("10.128.0.0/9"))
			ipam.Free(r1)
			r1 = ipam.Alloc(9)
			Expect(r1.String()).To(Equal("10.0.0.0/9"))
			ipam.Free(r1)
			Expect(ipam.String()).To(Equal("10.0.0.0/8[free]"))
		})
		It("scenario", func() {
			ipam, _ := NewIPAM(cidr)

			r1 := ipam.Alloc(9)
			Expect(r1.String()).To(Equal("10.0.0.0/9"))

			r2 := ipam.Alloc(10)
			Expect(r2.String()).To(Equal("10.128.0.0/10"))

			r3 := ipam.Alloc(12)
			Expect(r3.String()).To(Equal("10.192.0.0/12"))

			r4 := ipam.Alloc(11)
			Expect(r4.String()).To(Equal("10.224.0.0/11"))

			Expect(ipam.String()).To(Equal("10.0.0.0/9[busy], 10.128.0.0/10[busy], 10.192.0.0/12[busy], 10.208.0.0/12[free], 10.224.0.0/11[busy]"))

			Expect(ipam.Free(r1)).To(BeTrue())
			Expect(ipam.Free(r3)).To(BeTrue())
			Expect(ipam.Free(r2)).To(BeTrue())
			Expect(ipam.Free(r4)).To(BeTrue())

			Expect(ipam.block.next).To(BeNil())
			Expect(ipam.block.prev).To(BeNil())
		})
	})

	Context("using bitmaps", func() {
		_, cidr, _ := net.ParseCIDR("10.0.0.0/26")

		It("initializes ipam correctly", func() {
			ipam, _ := NewIPAM(cidr)

			Expect(ipam.block.next).To(BeNil())
			Expect(ipam.block.prev).To(BeNil())
		})

		It("check 28", func() {
			ipam, _ := NewIPAM(cidr)

			r := ipam.Alloc(28)
			Expect(r.String()).To(Equal("10.0.0.0/28"))

			Expect(ipam.block.String()).To(Equal("10.0.0.0/26[00000000 00000000 00000000 00000000 00000000 00000000 11111111 11111111]"))
		})

		It("check 28/30/28", func() {
			ipam, _ := NewIPAM(cidr)

			r1 := ipam.Alloc(28)
			Expect(r1.String()).To(Equal("10.0.0.0/28"))
			r2 := ipam.Alloc(30)
			Expect(r2.String()).To(Equal("10.0.0.16/30"))
			r3 := ipam.Alloc(28)
			Expect(r3.String()).To(Equal("10.0.0.32/28"))

			Expect(ipam.String()).To(Equal("10.0.0.0/26[00000000 00000000 11111111 11111111 00000000 00001111 11111111 11111111]"))
		})

		It("free 28", func() {
			ipam, _ := NewIPAM(cidr)

			r1 := ipam.Alloc(28)
			Expect(r1.String()).To(Equal("10.0.0.0/28"))

			Expect(ipam.Free(r1)).To(BeTrue())
			Expect(ipam.String()).To(Equal("10.0.0.0/26[free]"))
		})

		It("busy", func() {
			ipam, _ := NewIPAM(cidr)

			r1 := MustParseCIDR("10.0.0.8/29")
			Expect(ipam.Busy(r1)).To(BeTrue())
			Expect(ipam.String()).To(Equal("10.0.0.0/26[00000000 00000000 00000000 00000000 00000000 00000000 11111111 00000000]"))

			r2 := MustParseCIDR("10.0.0.12/30")
			Expect(ipam.Busy(r2)).To(BeFalse())
			Expect(ipam.String()).To(Equal("10.0.0.0/26[00000000 00000000 00000000 00000000 00000000 00000000 11111111 00000000]"))

			r3 := MustParseCIDR("10.0.0.0/27")
			Expect(ipam.Busy(r3)).To(BeFalse())
			Expect(ipam.String()).To(Equal("10.0.0.0/26[00000000 00000000 00000000 00000000 00000000 00000000 11111111 00000000]"))

			ipam.Free(r1)
			Expect(ipam.String()).To(Equal("10.0.0.0/26[free]"))
		})

		It("no round robin", func() {
			ipam, _ := NewIPAM(cidr)
			ipam.SetRoundRobin(false)

			r1 := ipam.Alloc(27)
			Expect(r1.String()).To(Equal("10.0.0.0/27"))
			ipam.Free(r1)
			r1 = ipam.Alloc(27)
			Expect(r1.String()).To(Equal("10.0.0.0/27"))
			ipam.Free(r1)
			Expect(ipam.String()).To(Equal("10.0.0.0/26[free]"))
		})

		It("round robin", func() {
			ipam, _ := NewIPAM(cidr)
			ipam.SetRoundRobin(true)

			r1 := ipam.Alloc(27)
			Expect(r1.String()).To(Equal("10.0.0.0/27"))
			ipam.Free(r1)
			r1 = ipam.Alloc(27)
			Expect(r1.String()).To(Equal("10.0.0.32/27"))
			ipam.Free(r1)
			r1 = ipam.Alloc(27)
			Expect(r1.String()).To(Equal("10.0.0.0/27"))
			ipam.Free(r1)
			Expect(ipam.String()).To(Equal("10.0.0.0/26[free]"))
		})

		It("scenario", func() {
			ipam, _ := NewIPAM(cidr)

			r1 := ipam.Alloc(28)
			Expect(r1.String()).To(Equal("10.0.0.0/28"))
			r2 := ipam.Alloc(30)
			Expect(r2.String()).To(Equal("10.0.0.16/30"))
			r3 := ipam.Alloc(28)
			Expect(r3.String()).To(Equal("10.0.0.32/28"))

			Expect(ipam.Free(r1)).To(BeTrue())
			Expect(ipam.String()).To(Equal("10.0.0.0/26[00000000 00000000 11111111 11111111 00000000 00001111 00000000 00000000]"))
			Expect(ipam.Free(r3)).To(BeTrue())
			Expect(ipam.String()).To(Equal("10.0.0.0/26[00000000 00000000 00000000 00000000 00000000 00001111 00000000 00000000]"))
			Expect(ipam.Free(r2)).To(BeTrue())
			Expect(ipam.String()).To(Equal("10.0.0.0/26[free]"))

		})
	})

	Context("mixed", func() {
		_, cidr, _ := net.ParseCIDR("10.0.0.0/24")

		It("initializes ipam correctly", func() {
			ipam, _ := NewIPAM(cidr)

			Expect(ipam.block.next).To(BeNil())
			Expect(ipam.block.prev).To(BeNil())

			Expect(ipam.String()).To(Equal("10.0.0.0/24[free]"))
		})

		It("check 32/25", func() {
			ipam, _ := NewIPAM(cidr)

			r1 := ipam.Alloc(32)
			Expect(r1.String()).To(Equal("10.0.0.0/32"))
			Expect(ipam.String()).To(Equal("10.0.0.0/26[00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000001], 10.0.0.64/26[free], 10.0.0.128/25[free]"))

			r2 := ipam.Alloc(25)
			Expect(r2.String()).To(Equal("10.0.0.128/25"))
			Expect(ipam.String()).To(Equal("10.0.0.0/26[00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000001], 10.0.0.64/26[free], 10.0.0.128/25[busy]"))
		})

		It("no round robin", func() {
			_, cidr, _ := net.ParseCIDR("10.0.0.0/25")
			ipam, _ := NewIPAM(cidr)
			ipam.SetRoundRobin(false)

			r1 := ipam.Alloc(27)
			Expect(r1.String()).To(Equal("10.0.0.0/27"))
			ipam.Free(r1)
			r1 = ipam.Alloc(27)
			Expect(r1.String()).To(Equal("10.0.0.0/27"))
			ipam.Free(r1)
			r1 = ipam.Alloc(27)
			Expect(r1.String()).To(Equal("10.0.0.0/27"))
			ipam.Free(r1)
			Expect(ipam.String()).To(Equal("10.0.0.0/25[free]"))

		})

		It("round robin", func() {
			_, cidr, _ := net.ParseCIDR("10.0.0.0/25")
			ipam, _ := NewIPAM(cidr)
			ipam.SetRoundRobin(true)

			r1 := ipam.Alloc(27)
			Expect(r1.String()).To(Equal("10.0.0.0/27"))
			ipam.Free(r1)
			r1 = ipam.Alloc(27)
			Expect(r1.String()).To(Equal("10.0.0.32/27"))
			ipam.Free(r1)
			r1 = ipam.Alloc(27)
			Expect(r1.String()).To(Equal("10.0.0.64/27"))
			ipam.Free(r1)
			r1 = ipam.Alloc(27)
			Expect(r1.String()).To(Equal("10.0.0.96/27"))
			ipam.Free(r1)
			r1 = ipam.Alloc(27)
			Expect(r1.String()).To(Equal("10.0.0.0/27"))
			ipam.Free(r1)
			Expect(ipam.String()).To(Equal("10.0.0.0/25[free]"))
		})

		It("scenario", func() {
			ipam, _ := NewIPAM(cidr)

			r1 := ipam.Alloc(32)
			Expect(r1.String()).To(Equal("10.0.0.0/32"))
			Expect(ipam.String()).To(Equal("10.0.0.0/26[00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000001], 10.0.0.64/26[free], 10.0.0.128/25[free]"))

			r2 := ipam.Alloc(25)
			Expect(r2.String()).To(Equal("10.0.0.128/25"))
			Expect(ipam.String()).To(Equal("10.0.0.0/26[00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000001], 10.0.0.64/26[free], 10.0.0.128/25[busy]"))

			Expect(ipam.Free(r1)).To(BeTrue())
			Expect(ipam.String()).To(Equal("10.0.0.0/25[free], 10.0.0.128/25[busy]"))

			Expect(ipam.Free(r2)).To(BeTrue())
			Expect(ipam.String()).To(Equal("10.0.0.0/24[free]"))
		})

		It("scenario 1", func() {
			ipam, _ := NewIPAM(cidr)
			ipam.Busy(MustParseCIDR("10.0.0.127/32"))
			Expect(ipam.String()).To(Equal("10.0.0.0/26[free], 10.0.0.64/26[10000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000], 10.0.0.128/25[free]"))
		})

		It("scenario 2", func() {
			ipam, _ := NewIPAM(cidr)
			ipam.Busy(MustParseCIDR("10.0.0.0/29"))
			Expect(ipam.String()).To(Equal("10.0.0.0/26[00000000 00000000 00000000 00000000 00000000 00000000 00000000 11111111], 10.0.0.64/26[free], 10.0.0.128/25[free]"))
			ipam.Busy(MustParseCIDR("10.0.0.8/32"))
			Expect(ipam.String()).To(Equal("10.0.0.0/26[00000000 00000000 00000000 00000000 00000000 00000000 00000001 11111111], 10.0.0.64/26[free], 10.0.0.128/25[free]"))
			ipam.Busy(MustParseCIDR("10.0.0.128/25"))
			Expect(ipam.String()).To(Equal("10.0.0.0/26[00000000 00000000 00000000 00000000 00000000 00000000 00000001 11111111], 10.0.0.64/26[free], 10.0.0.128/25[busy]"))
			ipam.Busy(MustParseCIDR("10.0.0.127/32"))
			Expect(ipam.String()).To(Equal("10.0.0.0/26[00000000 00000000 00000000 00000000 00000000 00000000 00000001 11111111], 10.0.0.64/26[10000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000], 10.0.0.128/25[busy]"))
		})
	})

	Context("sub range", func() {
		_, cidr, _ := net.ParseCIDR("10.0.0.0/24")

		It("initializes ipam correctly", func() {
			ipam, _ := NewIPAM(cidr, MustParseIPRange("10.0.0.10-10.0.0.250"))

			Expect(ipam.String()).To(Equal("10.0.0.0/26[00000000 00000000 00000000 00000000 00000000 00000000 00000011 11111111], 10.0.0.64/26[free], 10.0.0.128/26[free], 10.0.0.192/26[11111000 00000000 00000000 00000000 00000000 00000000 00000000 00000000]"))
		})

		It("initializes ipam correctly with sparse range", func() {
			ipam, _ := NewIPAM(cidr, MustParseIPRange("10.0.0.10-10.0.0.126"))

			Expect(ipam.String()).To(Equal("10.0.0.0/26[00000000 00000000 00000000 00000000 00000000 00000000 00000011 11111111], 10.0.0.64/26[10000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000]"))
		})
	})

	Context("ranges", func() {
		It("initializes ipam correctly", func() {
			ipam, err := NewIPAMForRanges(MustParseIPRanges("10.0.8.0-10.0.247.0"))

			Expect(err).To(BeNil())
			Expect(ipam.String()).To(Equal("10.0.8.0/21[free], 10.0.16.0/20[free], 10.0.32.0/19[free], 10.0.64.0/18[free], 10.0.128.0/18[free], 10.0.192.0/19[free], 10.0.224.0/20[free], 10.0.240.0/22[free], 10.0.244.0/23[free], 10.0.246.0/24[free], 10.0.247.0/26[11111111 11111111 11111111 11111111 11111111 11111111 11111111 11111110]"))

		})

		It("initializes ipam correctly 2", func() {
			ipam, err := NewIPAMForRanges(MustParseIPRanges("10.0.0.0/25", "10.0.0.128/25"))

			Expect(err).To(BeNil())
			Expect(ipam.String()).To(Equal("10.0.0.0/24[free]"))
		})

		It("initializes ipam correctly 3", func() {
			ipam, err := NewIPAMForRanges(MustParseIPRanges("10.0.0.0/25", "10.0.0.128/28", "10.0.0.160/28"))

			Expect(err).To(BeNil())
			Expect(ipam.String()).To(Equal("10.0.0.0/25[free], 10.0.0.128/26[11111111 11111111 00000000 00000000 11111111 11111111 00000000 00000000]"))
		})

		It("allocates /30", func() {
			ipam, err := NewIPAMForRanges(MustParseIPRanges("10.0.0.0/25", "10.0.0.128/28", "10.0.0.160/28"))

			Expect(err).To(BeNil())
			r1 := ipam.Alloc(30)
			Expect(r1.String()).To(Equal("10.0.0.128/30"))
			Expect(ipam.String()).To(Equal("10.0.0.0/25[free], 10.0.0.128/26[11111111 11111111 00000000 00000000 11111111 11111111 00000000 00001111]"))
			ipam.Free(r1)
			Expect(ipam.String()).To(Equal("10.0.0.0/25[free], 10.0.0.128/26[11111111 11111111 00000000 00000000 11111111 11111111 00000000 00000000]"))
		})
	})

	Context("serialize blocks", func() {
		_, cidr, _ := net.ParseCIDR("10.0.0.0/8")
		_, bitmap, _ := net.ParseCIDR("10.0.0.0/26")

		It("serializes free block", func() {
			b := &Block{
				busy: 0,
				cidr: cidr,
			}
			Expect(b.String()).To(Equal("10.0.0.0/8[free]"))
			Expect(ParseBlock(b.String())).To(Equal(b))
		})
		It("serializes busy block", func() {
			b := &Block{
				busy: BITMAP_BUSY,
				cidr: cidr,
			}
			Expect(b.String()).To(Equal("10.0.0.0/8[busy]"))
			Expect(ParseBlock(b.String())).To(Equal(b))
		})
		It("serializes bitmap block", func() {
			b := &Block{
				busy: 255,
				cidr: bitmap,
			}
			Expect(b.String()).To(Equal("10.0.0.0/26[00000000 00000000 00000000 00000000 00000000 00000000 00000000 11111111]"))
			Expect(ParseBlock(b.String())).To(Equal(b))
		})
		It("serializes bitmap block", func() {
			b := &Block{
				busy: 255 +
					254*256 +
					253*256*256 +
					252*256*256*256 +
					251*256*256*256*256 +
					256*256*256*256*256*256*256,
				cidr: bitmap,
			}
			Expect(b.String()).To(Equal("10.0.0.0/26[00000001 00000000 00000000 11111011 11111100 11111101 11111110 11111111]"))
			Expect(ParseBlock(b.String())).To(Equal(b))
		})
	})

	Context("serialize state", func() {
		_, cidr, _ := net.ParseCIDR("10.0.0.0/8")

		It("serializes empty ipam", func() {
			ipam, _ := NewIPAM(cidr)
			ipam.SetRoundRobin(true)
			blocks, round := ipam.State()
			fmt.Printf("%v\n", blocks)

			nipam, _ := NewIPAM(cidr)
			nipam.SetRoundRobin(true)
			nipam.SetState(blocks, round)
			Expect(nipam).To(Equal(ipam))
		})
		It("serializes busy ipam", func() {
			ipam, _ := NewIPAM(cidr)
			ipam.SetRoundRobin(true)
			ipam.Alloc(30)
			blocks, round := ipam.State()
			fmt.Printf("%v\n", blocks)

			nipam, _ := NewIPAM(cidr)
			nipam.SetRoundRobin(true)
			nipam.SetState(blocks, round)
			Expect(nipam).To(Equal(ipam))
		})
		It("serializes busy 2 ipam", func() {
			ipam, _ := NewIPAM(cidr)
			ipam.SetRoundRobin(true)
			ipam.Alloc(30)
			ipam.Alloc(20)
			blocks, round := ipam.State()
			fmt.Printf("%v\n", blocks)

			nipam, _ := NewIPAM(cidr)
			nipam.SetRoundRobin(true)
			nipam.SetState(blocks, round)
			Expect(nipam).To(Equal(ipam))
		})

	})
})
