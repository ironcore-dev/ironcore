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
	"net"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CIDR", func() {
	Context("ipv4", func() {
		cidr, _ := ParseCIDR("10.1.1.1/8")

		It("first", func() {
			Expect(CIDRFirstIP(cidr)).To(Equal(ParseIP("10.0.0.0")))
		})
		It("last", func() {
			Expect(CIDRLastIP(cidr)).To(Equal(ParseIP("10.255.255.255")))
		})
	})

	Context("cidr list utils", func() {
		c1, _ := ParseCIDR("10.0.0.0/16")
		c2, _ := ParseCIDR("10.8.0.0/16")

		c3, _ := ParseCIDR("10.8.8.0/24")
		o, _ := ParseCIDR("9.8.8.0/24")

		list := CIDRList{c1, c2}

		It("empty", func() {
			Expect(list.IsEmpty()).To(BeFalse())
		})

		It("contains ip", func() {
			Expect(list.Contains(c3.IP)).To(BeTrue())
		})
		It("no contains ip", func() {
			Expect(list.Contains(o.IP)).To(BeFalse())
		})

		It("contains cidr", func() {
			Expect(list.ContainsCIDR(c3)).To(BeTrue())
		})
		It("not contains cidr", func() {
			Expect(list.ContainsCIDR(o)).To(BeFalse())
		})

		It("ipmask ", func() {
			m := net.CIDRMask(10, 32)
			Expect(IPMaskClone(m)).To(Equal(m))
		})

		It("extend lower", func() {
			cidr := MustParseCIDR("10.8.0.0/16")
			r := MustParseCIDR("10.8.0.0/15")
			Expect(CIDRExtend(cidr)).To(Equal(r))
		})
		It("extend upper", func() {
			cidr := MustParseCIDR("10.9.0.0/16")
			r := MustParseCIDR("10.8.0.0/15")
			Expect(CIDRExtend(cidr)).To(Equal(r))
		})
		It("extend all", func() {
			cidr := MustParseCIDR("0.0.0.0/0")
			Expect(CIDRExtend(cidr)).To(BeNil())
		})
	})
	Context("cidr contains", func() {
		base, _ := ParseCIDR("10.0.0.0/8")
		lower, _ := ParseCIDR("10.0.0.0/9")
		middle, _ := ParseCIDR("10.5.0.0/24")
		upper, _ := ParseCIDR("10.128.0.0/9")
		above, _ := ParseCIDR("11.0.0.0/9")
		other, _ := ParseCIDR("192.168.0.0/24")

		It("lower", func() {
			Expect(CIDRContains(base, lower)).To(BeTrue())
		})
		It("middle", func() {
			Expect(CIDRContains(base, middle)).To(BeTrue())
		})
		It("upper", func() {
			Expect(CIDRContains(base, upper)).To(BeTrue())
		})
		It("upper", func() {
			Expect(CIDRContains(base, upper)).To(BeTrue())
		})
		It("self", func() {
			Expect(CIDRContains(base, base)).To(BeTrue())
		})
		It("above", func() {
			Expect(CIDRContains(base, above)).To(BeFalse())
		})
		It("other", func() {
			Expect(CIDRContains(base, other)).To(BeFalse())
		})
	})

	Context("cidr list delta", func() {
		a, _ := ParseCIDR("10.9.0.0/16")
		b, _ := ParseCIDR("10.192.0.0/16")

		s1, _ := ParseCIDR("10.8.0.0/15")
		s2, _ := ParseCIDR("10.8.0.0/14")
		s3, _ := ParseCIDR("10.8.0.0/13")

		i1, _ := ParseCIDR("10.9.0.0/18")
		i2, _ := ParseCIDR("10.9.128.0/18")
		i3, _ := ParseCIDR("10.9.192.0/18")

		a1, _ := ParseCIDR("10.8.0.0/16")
		a2, _ := ParseCIDR("10.10.0.0/16")
		a3, _ := ParseCIDR("10.10.0.0/15")

		a4, _ := ParseCIDR("10.12.0.0/14")

		s4, _ := ParseCIDR("10.0.0.0/9")
		a5, _ := ParseCIDR("10.0.0.0/13")
		a6, _ := ParseCIDR("10.12.0.0/14")
		a7, _ := ParseCIDR("10.16.0.0/12")
		a8, _ := ParseCIDR("10.32.0.0/11")
		a9, _ := ParseCIDR("10.64.0.0/10")

		s5, _ := ParseCIDR("10.128.0.0/9")
		a10, _ := ParseCIDR("10.128.0.0/10")
		a11, _ := ParseCIDR("10.193.0.0/16")
		a12, _ := ParseCIDR("10.194.0.0/15")
		a13, _ := ParseCIDR("10.196.0.0/14")
		a14, _ := ParseCIDR("10.200.0.0/13")
		a15, _ := ParseCIDR("10.208.0.0/12")
		a16, _ := ParseCIDR("10.224.0.0/11")

		s6, _ := ParseCIDR("10.0.0.0/8")

		base := CIDRList{a, b}
		base1 := CIDRList{a1, b}

		It("included lower", func() {
			Expect(base.Additional(CIDRList{i1})).To(Equal(CIDRList(nil)))
		})
		It("included middle", func() {
			Expect(base.Additional(CIDRList{i2})).To(Equal(CIDRList(nil)))
		})
		It("included upper", func() {
			Expect(base.Additional(CIDRList{i3})).To(Equal(CIDRList(nil)))
		})

		It("below", func() {
			Expect(base.Additional(CIDRList{a1})).To(Equal(CIDRList{a1}))
		})
		It("inbetween", func() {
			Expect(base.Additional(CIDRList{a2})).To(Equal(CIDRList{a2}))
		})

		It("including upper", func() {
			Expect(base.Additional(CIDRList{s1})).To(Equal(CIDRList{a1}))
		})
		It("including middle", func() {
			Expect(base.Additional(CIDRList{s2})).To(Equal(CIDRList{a1, a3}))
		})
		It("including lower", func() {
			Expect(base1.Additional(CIDRList{s1})).To(Equal(CIDRList{a}))
		})
		It("more", func() {
			Expect(base.Additional(CIDRList{s3})).To(Equal(CIDRList{a1, a3, a4}))
		})
		It("lower half", func() {
			Expect(base.Additional(CIDRList{s4})).To(Equal(CIDRList{a5, a1, a3, a6, a7, a8, a9}))
		})
		It("upper half", func() {
			Expect(base.Additional(CIDRList{s5})).To(Equal(CIDRList{a10, a11, a12, a13, a14, a15, a16}))
		})
		It("all", func() {
			Expect(base.Additional(CIDRList{s6})).To(Equal(CIDRList{a5, a1, a3, a6, a7, a8, a9, a10, a11, a12, a13, a14, a15, a16}))
		})
		It("all duplicates", func() {
			Expect(base.Additional(CIDRList{s6, s5, s4})).To(Equal(CIDRList{a5, a1, a3, a6, a7, a8, a9, a10, a11, a12, a13, a14, a15, a16}))
		})
	})

	Context("cidr list normalize", func() {
		c1 := MustParseCIDR("10.8.12.0/24")
		c2 := MustParseCIDR("10.9.12.0/24")
		c3 := MustParseCIDR("10.9.13.0/24")
		c4 := MustParseCIDR("10.9.8.0/22")
		c5 := MustParseCIDR("10.9.14.0/23")

		o1 := MustParseCIDR("10.9.8.0/21")
		//o2:= MustParseCIDR("10.8.0.0/16")

		r1 := MustParseCIDR("10.9.12.0/23")
		r2 := MustParseCIDR("10.9.8.0/21")

		s4 := MustParseCIDR("10.0.0.0/9")
		s5 := MustParseCIDR("10.128.0.0/9")
		s6 := MustParseCIDR("10.0.0.0/8")

		It("handles normalized list", func() {
			list := CIDRList{c1, c2}
			norm := list.Copy()
			norm.Normalize()
			Expect(norm).To(Equal(list))
		})

		It("handles simple list", func() {
			list := CIDRList{c2, c3}
			norm := list.Copy()
			norm.Normalize()
			Expect(norm).To(Equal(CIDRList{r1}))
		})
		It("handles successive joins", func() {
			list := CIDRList{c1, c4, c2, c3, c5}
			norm := list.Copy()
			norm.Normalize()
			Expect(norm).To(Equal(CIDRList{c1, r2}))
		})

		It("eliminated overlappig cidrs", func() {
			list := CIDRList{c1, c4, o1}
			norm := list.Copy()
			norm.Normalize()
			Expect(norm).To(Equal(CIDRList{c1, o1}))
		})
		It("eliminated overlappig cidrs successive", func() {
			list := CIDRList{s4, s6, s5}
			norm := list.Copy()
			norm.Normalize()
			Expect(norm).To(Equal(CIDRList{s6}))
		})
	})
})
