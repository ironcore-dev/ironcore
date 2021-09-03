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

	////////////////////////////////////////////////////////////////////////////

	Context("using bitmaps", func() {
		Context("using bitmaps complete", func() {
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

				Expect(ipam.block.String()).To(Equal("10.0.0.0/26[11111111 11111111]"))
			})

			It("check 28/30/28", func() {
				ipam, _ := NewIPAM(cidr)

				r1 := ipam.Alloc(28)
				Expect(r1.String()).To(Equal("10.0.0.0/28"))
				r2 := ipam.Alloc(30)
				Expect(r2.String()).To(Equal("10.0.0.16/30"))
				r3 := ipam.Alloc(28)
				Expect(r3.String()).To(Equal("10.0.0.32/28"))

				Expect(ipam.String()).To(Equal("10.0.0.0/26[11111111 11111111 00000000 00001111 11111111 11111111]"))
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
				Expect(ipam.String()).To(Equal("10.0.0.0/26[11111111 00000000]"))

				r2 := MustParseCIDR("10.0.0.12/30")
				Expect(ipam.Busy(r2)).To(BeFalse())
				Expect(ipam.String()).To(Equal("10.0.0.0/26[11111111 00000000]"))

				r3 := MustParseCIDR("10.0.0.0/27")
				Expect(ipam.Busy(r3)).To(BeFalse())
				Expect(ipam.String()).To(Equal("10.0.0.0/26[11111111 00000000]"))

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
				Expect(ipam.String()).To(Equal("10.0.0.0/26[11111111 11111111 00000000 00001111 00000000 00000000]"))
				Expect(ipam.Free(r3)).To(BeTrue())
				Expect(ipam.String()).To(Equal("10.0.0.0/26[00001111 00000000 00000000]"))
				Expect(ipam.Free(r2)).To(BeTrue())
				Expect(ipam.String()).To(Equal("10.0.0.0/26[free]"))

			})
		})

		Context("using bitmap part", func() {
			_, cidr, _ := net.ParseCIDR("10.0.0.0/28")

			It("initializes ipam correctly", func() {
				ipam, _ := NewIPAM(cidr)

				Expect(ipam.block.next).To(BeNil())
				Expect(ipam.block.prev).To(BeNil())
			})

			It("check 30", func() {
				ipam, _ := NewIPAM(cidr)

				r := ipam.Alloc(30)
				Expect(r.String()).To(Equal("10.0.0.0/30"))

				Expect(ipam.String()).To(Equal("10.0.0.0/28[00001111]"))
			})

			It("check 30/32/30", func() {
				ipam, _ := NewIPAM(cidr)

				r1 := ipam.Alloc(30)
				Expect(r1.String()).To(Equal("10.0.0.0/30"))
				r2 := ipam.Alloc(32)
				Expect(r2.String()).To(Equal("10.0.0.4/32"))
				r3 := ipam.Alloc(30)
				Expect(r3.String()).To(Equal("10.0.0.8/30"))

				Expect(ipam.String()).To(Equal("10.0.0.0/28[00001111 00011111]"))
			})

			It("free 30", func() {
				ipam, _ := NewIPAM(cidr)

				r1 := ipam.Alloc(30)
				Expect(r1.String()).To(Equal("10.0.0.0/30"))

				Expect(ipam.Free(r1)).To(BeTrue())
				Expect(ipam.String()).To(Equal("10.0.0.0/28[free]"))
			})

			It("busy", func() {
				ipam, _ := NewIPAM(cidr)

				r1 := MustParseCIDR("10.0.0.8/30")
				Expect(ipam.Busy(r1)).To(BeTrue())
				Expect(ipam.String()).To(Equal("10.0.0.0/28[00001111 00000000]"))

				r2 := MustParseCIDR("10.0.0.8/32")
				Expect(ipam.Busy(r2)).To(BeFalse())
				Expect(ipam.String()).To(Equal("10.0.0.0/28[00001111 00000000]"))

				r3 := MustParseCIDR("10.0.0.0/27")
				Expect(ipam.Busy(r3)).To(BeFalse())
				Expect(ipam.String()).To(Equal("10.0.0.0/28[00001111 00000000]"))

				ipam.Free(r1)
				Expect(ipam.String()).To(Equal("10.0.0.0/28[free]"))
			})

			It("no round robin", func() {
				ipam, _ := NewIPAM(cidr)
				ipam.SetRoundRobin(false)

				r1 := ipam.Alloc(30)
				Expect(r1.String()).To(Equal("10.0.0.0/30"))
				ipam.Free(r1)
				r1 = ipam.Alloc(30)
				Expect(r1.String()).To(Equal("10.0.0.0/30"))
				ipam.Free(r1)
				Expect(ipam.String()).To(Equal("10.0.0.0/28[free]"))
			})

			It("round robin", func() {
				ipam, _ := NewIPAM(cidr)
				ipam.SetRoundRobin(true)

				r1 := ipam.Alloc(30)
				Expect(r1.String()).To(Equal("10.0.0.0/30"))
				ipam.Free(r1)
				r1 = ipam.Alloc(30)
				Expect(r1.String()).To(Equal("10.0.0.4/30"))
				ipam.Free(r1)
				r1 = ipam.Alloc(30)
				Expect(r1.String()).To(Equal("10.0.0.8/30"))
				ipam.Free(r1)
				Expect(ipam.String()).To(Equal("10.0.0.0/28[free]"))
			})

			It("scenario", func() {
				ipam, _ := NewIPAM(cidr)

				r1 := ipam.Alloc(30)
				Expect(r1.String()).To(Equal("10.0.0.0/30"))
				r2 := ipam.Alloc(32)
				Expect(r2.String()).To(Equal("10.0.0.4/32"))
				r3 := ipam.Alloc(30)
				Expect(r3.String()).To(Equal("10.0.0.8/30"))

				Expect(ipam.Free(r1)).To(BeTrue())
				Expect(ipam.String()).To(Equal("10.0.0.0/28[00001111 00010000]"))
				Expect(ipam.Free(r3)).To(BeTrue())
				Expect(ipam.String()).To(Equal("10.0.0.0/28[00010000]"))
				Expect(ipam.Free(r2)).To(BeTrue())
				Expect(ipam.String()).To(Equal("10.0.0.0/28[free]"))

			})
		})

		Context("using bitmap upper part", func() {
			_, cidr, _ := net.ParseCIDR("10.0.0.32/28")

			It("initializes ipam correctly", func() {
				ipam, _ := NewIPAM(cidr)

				Expect(ipam.block.next).To(BeNil())
				Expect(ipam.block.prev).To(BeNil())
			})

			It("check 30", func() {
				ipam, _ := NewIPAM(cidr)

				r := ipam.Alloc(30)
				Expect(r.String()).To(Equal("10.0.0.32/30"))

				Expect(ipam.String()).To(Equal("10.0.0.32/28[00001111]"))
			})

			It("check 30/32/30", func() {
				ipam, _ := NewIPAM(cidr)

				r1 := ipam.Alloc(30)
				Expect(r1.String()).To(Equal("10.0.0.32/30"))
				r2 := ipam.Alloc(32)
				Expect(r2.String()).To(Equal("10.0.0.36/32"))
				r3 := ipam.Alloc(30)
				Expect(r3.String()).To(Equal("10.0.0.40/30"))

				Expect(ipam.String()).To(Equal("10.0.0.32/28[00001111 00011111]"))
			})

			It("free 30", func() {
				ipam, _ := NewIPAM(cidr)

				r1 := ipam.Alloc(30)
				Expect(r1.String()).To(Equal("10.0.0.32/30"))

				Expect(ipam.Free(r1)).To(BeTrue())
				Expect(ipam.String()).To(Equal("10.0.0.32/28[free]"))
			})

			It("busy", func() {
				ipam, _ := NewIPAM(cidr)

				r1 := MustParseCIDR("10.0.0.40/30")
				Expect(ipam.Busy(r1)).To(BeTrue())
				Expect(ipam.String()).To(Equal("10.0.0.32/28[00001111 00000000]"))

				r2 := MustParseCIDR("10.0.0.40/32")
				Expect(ipam.Busy(r2)).To(BeFalse())
				Expect(ipam.String()).To(Equal("10.0.0.32/28[00001111 00000000]"))

				r3 := MustParseCIDR("10.0.0.32/27")
				Expect(ipam.Busy(r3)).To(BeFalse())
				Expect(ipam.String()).To(Equal("10.0.0.32/28[00001111 00000000]"))

				ipam.Free(r1)
				Expect(ipam.String()).To(Equal("10.0.0.32/28[free]"))
			})

			It("no round robin", func() {
				ipam, _ := NewIPAM(cidr)
				ipam.SetRoundRobin(false)

				r1 := ipam.Alloc(30)
				Expect(r1.String()).To(Equal("10.0.0.32/30"))
				ipam.Free(r1)
				r1 = ipam.Alloc(30)
				Expect(r1.String()).To(Equal("10.0.0.32/30"))
				ipam.Free(r1)
				Expect(ipam.String()).To(Equal("10.0.0.32/28[free]"))
			})

			It("round robin", func() {
				ipam, _ := NewIPAM(cidr)
				ipam.SetRoundRobin(true)

				r1 := ipam.Alloc(30)
				Expect(r1.String()).To(Equal("10.0.0.32/30"))
				ipam.Free(r1)
				r1 = ipam.Alloc(30)
				Expect(r1.String()).To(Equal("10.0.0.36/30"))
				ipam.Free(r1)
				r1 = ipam.Alloc(30)
				Expect(r1.String()).To(Equal("10.0.0.40/30"))
				ipam.Free(r1)
				Expect(ipam.String()).To(Equal("10.0.0.32/28[free]"))
			})

			It("scenario", func() {
				ipam, _ := NewIPAM(cidr)

				r1 := ipam.Alloc(30)
				Expect(r1.String()).To(Equal("10.0.0.32/30"))
				r2 := ipam.Alloc(32)
				Expect(r2.String()).To(Equal("10.0.0.36/32"))
				r3 := ipam.Alloc(30)
				Expect(r3.String()).To(Equal("10.0.0.40/30"))

				Expect(ipam.Free(r1)).To(BeTrue())
				Expect(ipam.String()).To(Equal("10.0.0.32/28[00001111 00010000]"))
				Expect(ipam.Free(r3)).To(BeTrue())
				Expect(ipam.String()).To(Equal("10.0.0.32/28[00010000]"))
				Expect(ipam.Free(r2)).To(BeTrue())
				Expect(ipam.String()).To(Equal("10.0.0.32/28[free]"))

			})
		})
	})

	////////////////////////////////////////////////////////////////////////////

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
			Expect(ipam.String()).To(Equal("10.0.0.0/26[00000001], 10.0.0.64/26[free], 10.0.0.128/25[free]"))

			r2 := ipam.Alloc(25)
			Expect(r2.String()).To(Equal("10.0.0.128/25"))
			Expect(ipam.String()).To(Equal("10.0.0.0/26[00000001], 10.0.0.64/26[free], 10.0.0.128/25[busy]"))
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
			Expect(ipam.String()).To(Equal("10.0.0.0/26[00000001], 10.0.0.64/26[free], 10.0.0.128/25[free]"))

			r2 := ipam.Alloc(25)
			Expect(r2.String()).To(Equal("10.0.0.128/25"))
			Expect(ipam.String()).To(Equal("10.0.0.0/26[00000001], 10.0.0.64/26[free], 10.0.0.128/25[busy]"))

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
			Expect(ipam.String()).To(Equal("10.0.0.0/26[11111111], 10.0.0.64/26[free], 10.0.0.128/25[free]"))
			ipam.Busy(MustParseCIDR("10.0.0.8/32"))
			Expect(ipam.String()).To(Equal("10.0.0.0/26[00000001 11111111], 10.0.0.64/26[free], 10.0.0.128/25[free]"))
			ipam.Busy(MustParseCIDR("10.0.0.128/25"))
			Expect(ipam.String()).To(Equal("10.0.0.0/26[00000001 11111111], 10.0.0.64/26[free], 10.0.0.128/25[busy]"))
			ipam.Busy(MustParseCIDR("10.0.0.127/32"))
			Expect(ipam.String()).To(Equal("10.0.0.0/26[00000001 11111111], 10.0.0.64/26[10000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000], 10.0.0.128/25[busy]"))
		})
	})

	Context("sub range", func() {
		_, cidr, _ := net.ParseCIDR("10.0.0.0/24")

		It("initializes ipam correctly", func() {
			ipam, _ := NewIPAM(cidr, MustParseIPRange("10.0.0.10-10.0.0.250"))

			Expect(ipam.String()).To(Equal("10.0.0.0/26[00000011 11111111], 10.0.0.64/26[free], 10.0.0.128/26[free], 10.0.0.192/26[11111000 00000000 00000000 00000000 00000000 00000000 00000000 00000000]"))
		})

		It("initializes ipam correctly with sparse range", func() {
			ipam, _ := NewIPAM(cidr, MustParseIPRange("10.0.0.10-10.0.0.126"))

			s := ipam.String()
			Expect(s).To(Equal("10.0.0.0/26[00000011 11111111], 10.0.0.64/26[10000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000], 10.0.0.128/25[busy]"))
		})
	})

	Context("ranges", func() {
		It("initializes ipam correctly", func() {
			ipam, err := NewIPAMForRanges(MustParseIPRanges("10.0.8.0-10.0.247.0"))

			Expect(err).To(BeNil())
			Expect(ipam.String()).To(Equal("10.0.8.0/21[free], 10.0.16.0/20[free], 10.0.32.0/19[free], 10.0.64.0/18[free], 10.0.128.0/18[free], 10.0.192.0/19[free], 10.0.224.0/20[free], 10.0.240.0/22[free], 10.0.244.0/23[free], 10.0.246.0/24[free], 10.0.247.0/32[free]"))

		})

		It("initializes ipam correctly 2", func() {
			ipam, err := NewIPAMForRanges(MustParseIPRanges("10.0.0.0/25", "10.0.0.128/25"))

			Expect(err).To(BeNil())
			Expect(ipam.String()).To(Equal("10.0.0.0/24[free]"))
		})

		It("initializes ipam correctly 3", func() {
			ipam, err := NewIPAMForRanges(MustParseIPRanges("10.0.0.0/25", "10.0.0.128/28", "10.0.0.160/28"))

			Expect(err).To(BeNil())
			Expect(ipam.String()).To(Equal("10.0.0.0/25[free], 10.0.0.128/28[free], 10.0.0.160/28[free]"))
		})

		It("allocates /30", func() {
			ipam, err := NewIPAMForRanges(MustParseIPRanges("10.0.0.0/25", "10.0.0.128/28", "10.0.0.160/28"))

			Expect(err).To(BeNil())
			r1 := ipam.Alloc(30)
			Expect(r1.String()).To(Equal("10.0.0.128/30"))
			Expect(ipam.String()).To(Equal("10.0.0.0/25[free], 10.0.0.128/28[00001111], 10.0.0.160/28[free]"))
			ipam.Free(r1)
			Expect(ipam.String()).To(Equal("10.0.0.0/25[free], 10.0.0.128/28[free], 10.0.0.160/28[free]"))
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
			Expect(b.String()).To(Equal("10.0.0.0/26[11111111]"))
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
	Context("small range", func() {
		_, cidr, _ := net.ParseCIDR("10.0.0.0/28")

		It("initializes ipam correctly", func() {
			ipam, _ := NewIPAM(cidr)

			Expect(ipam.block.next).To(BeNil())
			Expect(ipam.block.prev).To(BeNil())

			Expect(ipam.String()).To(Equal("10.0.0.0/28[free]"))
		})
		It("allocates 8", func() {
			ipam, _ := NewIPAM(cidr)

			cidr := ipam.Alloc(29)
			Expect(cidr).NotTo(BeNil())
			Expect(cidr.String()).To(Equal("10.0.0.0/29"))

			Expect(ipam.String()).To(Equal("10.0.0.0/28[11111111]"))

			Expect(ipam.Free(cidr)).To(BeTrue())
			Expect(ipam.String()).To(Equal("10.0.0.0/28[free]"))
		})
		It("allocates 8 a frees", func() {
			ipam, _ := NewIPAM(cidr)

			cidr := ipam.Alloc(29)
			Expect(cidr).NotTo(BeNil())
			Expect(cidr.String()).To(Equal("10.0.0.0/29"))

			Expect(ipam.String()).To(Equal("10.0.0.0/28[11111111]"))
		})
		It("allocates 2*8", func() {
			ipam, _ := NewIPAM(cidr)

			cidr := ipam.Alloc(29)
			Expect(cidr).NotTo(BeNil())
			Expect(cidr.String()).To(Equal("10.0.0.0/29"))

			cidr = ipam.Alloc(29)
			Expect(cidr).NotTo(BeNil())
			Expect(cidr.String()).To(Equal("10.0.0.8/29"))

			Expect(ipam.String()).To(Equal("10.0.0.0/28[11111111 11111111]"))
		})

		It("allocates 2*8 + fails third alloc", func() {
			ipam, _ := NewIPAM(cidr)

			cidr := ipam.Alloc(29)
			Expect(cidr).NotTo(BeNil())
			Expect(cidr.String()).To(Equal("10.0.0.0/29"))

			cidr = ipam.Alloc(29)
			Expect(cidr).NotTo(BeNil())
			Expect(cidr.String()).To(Equal("10.0.0.8/29"))

			cidr = ipam.Alloc(29)
			Expect(cidr).To(BeNil())

			Expect(ipam.String()).To(Equal("10.0.0.0/28[11111111 11111111]"))
		})

		It("round robin", func() {
			ipam, _ := NewIPAM(cidr)
			ipam.SetRoundRobin(true)

			cidr := ipam.Alloc(29)
			Expect(cidr).NotTo(BeNil())
			Expect(cidr.String()).To(Equal("10.0.0.0/29"))
			Expect(ipam.Free(cidr)).To(BeTrue())

			cidr = ipam.Alloc(29)
			Expect(cidr).NotTo(BeNil())
			Expect(cidr.String()).To(Equal("10.0.0.8/29"))
			Expect(ipam.Free(cidr)).To(BeTrue())

			cidr = ipam.Alloc(29)
			Expect(cidr).NotTo(BeNil())
			Expect(cidr.String()).To(Equal("10.0.0.0/29"))

			Expect(ipam.String()).To(Equal("10.0.0.0/28[11111111]"))
		})
	})

	Context("extend ipam", func() {
		_, cidr, _ := net.ParseCIDR("10.9.0.0/16")

		It("adds buddy on empty small ipam", func() {
			_, cidr, _ := net.ParseCIDR("10.0.0.32/29")
			ipam, _ := NewIPAM(cidr)

			_, add, _ := net.ParseCIDR("10.0.0.40/29")
			_, exp, _ := net.ParseCIDR("10.0.0.32/28")

			ipam.AddCIDRs(CIDRList{add})

			Expect(ipam.Ranges()).To(Equal(CIDRList{exp}))

			blocks, _ := ipam.State()
			Expect(blocks).To(Equal([]string{"10.0.0.32/28[free]"}))
		})

		It("adds buddy on empty ipam before existing", func() {
			ipam, _ := NewIPAM(cidr)

			_, add, _ := net.ParseCIDR("10.8.0.0/16")
			_, exp, _ := net.ParseCIDR("10.8.0.0/15")

			ipam.AddCIDRs(CIDRList{add})

			Expect(ipam.Ranges()).To(Equal(CIDRList{exp}))

			blocks, _ := ipam.State()
			Expect(blocks).To(Equal([]string{"10.8.0.0/15[free]"}))
		})
		It("adds buddy on empty ipam after existing", func() {
			_, add, _ := net.ParseCIDR("10.8.0.0/16")
			ipam, _ := NewIPAM(add)

			_, exp, _ := net.ParseCIDR("10.8.0.0/15")

			ipam.AddCIDRs(CIDRList{cidr})

			Expect(ipam.Ranges()).To(Equal(CIDRList{exp}))

			blocks, _ := ipam.State()
			Expect(blocks).To(Equal([]string{"10.8.0.0/15[free]"}))
		})

		It("adds buddy on non empty ipam", func() {
			ipam, _ := NewIPAM(cidr)

			Expect(ipam.Alloc(17)).NotTo(BeNil())
			_, add, _ := net.ParseCIDR("10.8.0.0/16")
			_, exp, _ := net.ParseCIDR("10.8.0.0/15")

			ipam.AddCIDRs(CIDRList{add})

			Expect(ipam.Ranges()).To(Equal(CIDRList{exp}))

			blocks, _ := ipam.State()
			Expect(blocks).To(Equal([]string{"10.8.0.0/16[free]", "10.9.0.0/17[busy]", "10.9.128.0/17[free]"}))
		})

		It("add joinable intermediate cidr", func() {
			ranges, _ := ParseIPRanges("10.8.0.0/16", "10.10.0.0/16", "10.11.0.0/16")
			ipam, _ := NewIPAMForRanges(ranges)

			r1 := MustParseCIDR("10.8.0.0/16")
			r2 := MustParseCIDR("10.10.0.0/15")
			r3 := MustParseCIDR("10.8.0.0/14")
			Expect(ipam.Ranges()).To(Equal(CIDRList{r1, r2}))

			_, add, _ := net.ParseCIDR("10.9.0.0/16")

			ipam.AddCIDRs(CIDRList{add})
			Expect(ipam.Ranges()).To(Equal(CIDRList{r3}))

			blocks, _ := ipam.State()
			Expect(blocks).To(Equal([]string{"10.8.0.0/14[free]"}))

		})

		It("add joinable intermediate cidr after allocation", func() {
			ranges, _ := ParseIPRanges("10.8.0.0/16", "10.10.0.0/16", "10.11.0.0/16")
			ipam, _ := NewIPAMForRanges(ranges)

			a1 := MustParseCIDR("10.8.0.0/17")

			r1 := MustParseCIDR("10.8.0.0/16")
			r2 := MustParseCIDR("10.10.0.0/15")
			r3 := MustParseCIDR("10.8.0.0/14")
			Expect(ipam.Ranges()).To(Equal(CIDRList{r1, r2}))

			Expect(ipam.Busy(a1)).To(BeTrue())

			_, add, _ := net.ParseCIDR("10.9.0.0/16")

			ipam.AddCIDRs(CIDRList{add})
			Expect(ipam.Ranges()).To(Equal(CIDRList{r3}))

			blocks, _ := ipam.State()
			Expect(blocks).To(Equal([]string{"10.8.0.0/17[busy]", "10.8.128.0/17[free]", "10.9.0.0/16[free]", "10.10.0.0/15[free]"}))

		})

		It("add joinable intermediate cidr after state recovery", func() {
			ranges, err := ParseIPRanges("10.8.0.0/16", "10.10.0.0/16", "10.11.0.0/16")
			Expect(err).To(Succeed())
			ipam, err := NewIPAMForRanges(ranges)
			Expect(err).To(Succeed())

			a1 := MustParseCIDR("10.8.0.0/17")

			r1 := MustParseCIDR("10.8.0.0/16")
			r2 := MustParseCIDR("10.10.0.0/15")
			r3 := MustParseCIDR("10.8.0.0/14")
			Expect(ipam.Ranges()).To(Equal(CIDRList{r1, r2}))

			Expect(ipam.Busy(a1)).To(BeTrue())

			blocks, round := ipam.State()

			_, add, _ := net.ParseCIDR("10.9.0.0/16")
			ranges = append(ranges, CIDRRange(add))
			ipam, err = NewIPAMForRanges(ranges)
			Expect(err).To(Succeed())
			additional, err := ipam.SetState(blocks, round)
			Expect(err).To(Succeed())
			Expect(additional).To(Equal(CIDRList{add}))
			Expect(ipam.Ranges()).To(Equal(CIDRList{r3}))

			blocks, _ = ipam.State()
			Expect(blocks).To(Equal([]string{"10.8.0.0/17[busy]", "10.8.128.0/17[free]", "10.9.0.0/16[free]", "10.10.0.0/15[free]"}))

		})

		///////////////////////////////////////// remove //////////////////////////

		It("remove buddy of empty small ipam", func() {
			_, cidr, _ := net.ParseCIDR("10.0.0.32/28")
			ipam, _ := NewIPAM(cidr)

			_, del, _ := net.ParseCIDR("10.0.0.40/29")
			_, exp, _ := net.ParseCIDR("10.0.0.32/29")

			ipam.DeleteCIDRs(CIDRList{del})

			Expect(ipam.Ranges()).To(Equal(CIDRList{exp}))

			blocks, _ := ipam.State()
			Expect(blocks).To(Equal([]string{"10.0.0.32/29[free]"}))
		})

		It("remove upper buddy of used small ipam", func() {
			_, cidr, _ := net.ParseCIDR("10.0.0.32/28")
			ipam, _ := NewIPAM(cidr)

			Expect(ipam.Alloc(30)).To(Equal(MustParseCIDR("10.0.0.32/30")))

			blocks, _ := ipam.State()
			Expect(blocks).To(Equal([]string{"10.0.0.32/28[00001111]"}))

			_, del, _ := net.ParseCIDR("10.0.0.40/29")
			_, exp, _ := net.ParseCIDR("10.0.0.32/29")

			ipam.DeleteCIDRs(CIDRList{del})

			Expect(ipam.Ranges()).To(Equal(CIDRList{exp}))

			blocks, _ = ipam.State()
			Expect(blocks).To(Equal([]string{"10.0.0.32/29[00001111]"}))
		})

		It("remove lower buddy of used small ipam", func() {
			cidr := MustParseCIDR("10.0.0.32/28")
			req := MustParseCIDR("10.0.0.40/30")
			ipam, _ := NewIPAM(cidr)

			Expect(ipam.Busy(req)).To(BeTrue())

			blocks, _ := ipam.State()
			Expect(blocks).To(Equal([]string{"10.0.0.32/28[00001111 00000000]"}))

			_, del, _ := net.ParseCIDR("10.0.0.32/29")
			_, exp, _ := net.ParseCIDR("10.0.0.40/29")

			ipam.DeleteCIDRs(CIDRList{del})

			Expect(ipam.Ranges()).To(Equal(CIDRList{exp}))

			blocks, _ = ipam.State()
			Expect(blocks).To(Equal([]string{"10.0.0.40/29[00001111]"}))
		})

		///////////////////////////////////////// remove by restore ///////////

		It("handle remove unused range", func() {
			ranges, err := ParseIPRanges("10.8.0.0/16", "10.10.0.0/16", "10.11.0.0/16")
			Expect(err).To(Succeed())
			ipam, err := NewIPAMForRanges(ranges)
			Expect(err).To(Succeed())

			r1 := MustParseCIDR("10.8.0.0/16")
			r2 := MustParseCIDR("10.10.0.0/15")
			r3 := MustParseCIDR("10.10.0.0/16")
			Expect(ipam.Ranges()).To(Equal(CIDRList{r1, r2}))

			blocks, round := ipam.State()

			ranges, err = ParseIPRanges("10.8.0.0/16", "10.10.0.0/16")
			Expect(err).To(Succeed())
			ipam, err = NewIPAMForRanges(ranges)
			Expect(err).To(Succeed())

			additional, err := ipam.SetState(blocks, round)
			Expect(err).To(Succeed())
			Expect(additional).To(Equal(CIDRList(nil)))
			Expect(ipam.Ranges()).To(Equal(CIDRList{r1, r3}))

			blocks, _ = ipam.State()
			Expect(blocks).To(Equal([]string{"10.8.0.0/16[free]", "10.10.0.0/16[free]"}))
		})

		It("handle remove range", func() {
			ranges, err := ParseIPRanges("10.8.0.0/16", "10.10.0.0/16", "10.11.0.0/16")
			Expect(err).To(Succeed())
			ipam, err := NewIPAMForRanges(ranges)
			Expect(err).To(Succeed())

			a1 := MustParseCIDR("10.8.0.0/17")

			r1 := MustParseCIDR("10.8.0.0/16")
			r2 := MustParseCIDR("10.10.0.0/15")
			r3 := MustParseCIDR("10.8.0.0/14")
			Expect(ipam.Ranges()).To(Equal(CIDRList{r1, r2}))

			Expect(ipam.Busy(a1)).To(BeTrue())

			blocks, round := ipam.State()

			_, add, _ := net.ParseCIDR("10.9.0.0/16")
			ranges = append(ranges, CIDRRange(add))
			ipam, err = NewIPAMForRanges(ranges)
			Expect(err).To(Succeed())
			additional, err := ipam.SetState(blocks, round)
			Expect(err).To(Succeed())
			Expect(additional).To(Equal(CIDRList{add}))
			Expect(ipam.Ranges()).To(Equal(CIDRList{r3}))

			blocks, _ = ipam.State()
			Expect(blocks).To(Equal([]string{"10.8.0.0/17[busy]", "10.8.128.0/17[free]", "10.9.0.0/16[free]", "10.10.0.0/15[free]"}))

		})

		It("handle remove busy range an small block", func() {
			ranges, err := ParseIPRanges("10.0.0.0/27")
			Expect(err).To(Succeed())
			ipam, err := NewIPAMForRanges(ranges)
			Expect(err).To(Succeed())

			a1 := MustParseCIDR("10.0.0.0/29")
			d1 := MustParseCIDR("10.0.0.0/28")

			r1 := MustParseCIDR("10.0.0.0/27")
			//r2 := MustParseCIDR("10.0.0.8/29")
			r3 := MustParseCIDR("10.0.0.16/28")
			Expect(ipam.Ranges()).To(Equal(CIDRList{r1}))

			Expect(ipam.Busy(a1)).To(BeTrue())

			blocks, _ := ipam.State()
			Expect(blocks).To(Equal([]string{"10.0.0.0/27[11111111]"}))

			ipam.DeleteCIDRs(CIDRList{d1})

			Expect(ipam.Ranges()).To(Equal(CIDRList{r3}))
			Expect(ipam.PendingDeleted()).To(Equal(CIDRList{d1}))
			blocks, _ = ipam.State()
			Expect(blocks).To(Equal([]string{"10.0.0.0/27[11111111]"}))

			Expect(ipam.Free(a1)).To(BeTrue())
			Expect(ipam.PendingDeleted()).To(Equal(CIDRList(nil)))
			blocks, _ = ipam.State()
			Expect(blocks).To(Equal([]string{"10.0.0.16/28[free]"}))
		})

		It("handle remove busy range on split small block", func() {
			ranges, err := ParseIPRanges("10.0.0.0/27")
			Expect(err).To(Succeed())
			ipam, err := NewIPAMForRanges(ranges)
			Expect(err).To(Succeed())

			a1 := MustParseCIDR("10.0.0.0/29")
			a2 := MustParseCIDR("10.0.0.16/29")
			d1 := MustParseCIDR("10.0.0.0/28")
			d2 := MustParseCIDR("10.8.0.0/28")

			r1 := MustParseCIDR("10.0.0.0/27")
			//r2 := MustParseCIDR("10.0.0.8/29")
			r3 := MustParseCIDR("10.0.0.16/28")
			Expect(ipam.Ranges()).To(Equal(CIDRList{r1}))

			Expect(ipam.Busy(a1)).To(BeTrue())
			Expect(ipam.Busy(a2)).To(BeTrue())

			blocks, _ := ipam.State()
			Expect(blocks).To(Equal([]string{"10.0.0.0/27[11111111 00000000 11111111]"}))

			ipam.DeleteCIDRs(CIDRList{d1, d2})

			Expect(ipam.Ranges()).To(Equal(CIDRList{r3}))
			Expect(ipam.PendingDeleted()).To(Equal(CIDRList{d1}))
			blocks, _ = ipam.State()
			Expect(blocks).To(Equal([]string{"10.0.0.0/27[11111111 00000000 11111111]"}))

			Expect(ipam.Free(a1)).To(BeTrue())
			Expect(ipam.PendingDeleted()).To(Equal(CIDRList(nil)))
			blocks, _ = ipam.State()
			Expect(blocks).To(Equal([]string{"10.0.0.16/28[11111111]"}))
		})
	})
})
