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

var _ = Describe("Range", func() {

	Context("parse", func() {
		ip1 := ParseIP("10.0.1.0")
		ip2 := ParseIP("10.1.0.0")

		ip3 := ParseIP("10.0.0.0")
		ip4 := ParseIP("10.0.0.255")

		It("single ip", func() {
			Expect(ParseIPRange("10.0.1.0")).To(Equal(&IPRange{ip1, ip1}))
		})

		It("ip range", func() {
			Expect(ParseIPRange("10.0.1.0-10.1.0.0")).To(Equal(&IPRange{ip1, ip2}))
		})

		It("cidr", func() {
			r := &IPRange{ip3, ip4}
			Expect(ParseIPRange("10.0.0.0-10.0.0.255")).To(Equal(r))
			Expect(ParseIPRange("10.0.0.0/24")).To(Equal(r))
		})

		It("invalid ip", func() {
			_, err := ParseIPRange("10..2.3.4")
			Expect(err).To(HaveOccurred())
			_, err = ParseIPRange("10.0.0.0-10..2.3.4")
			Expect(err).To(HaveOccurred())
			_, err = ParseIPRange("10.0.0.0-10.2.3.4-1.2.3.4")
			Expect(err).To(HaveOccurred())
			_, err = ParseIPRange("10.3.0.0-10.2.3.4")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("range", func() {
		r := MustParseIPRange("10.0.1.0-10.0.3.0")
		It("contains border", func() {
			Expect(r.Contains(ParseIP("10.0.1.0"))).To(BeTrue())
			Expect(r.Contains(ParseIP("10.0.3.0"))).To(BeTrue())
		})
		It("contains", func() {
			Expect(r.Contains(ParseIP("10.0.2.0"))).To(BeTrue())
		})
		It("not contains", func() {
			Expect(r.Contains(ParseIP("10.0.0.255"))).To(BeFalse())
			Expect(r.Contains(ParseIP("10.0.3.1"))).To(BeFalse())
		})
	})
	Context("ranges", func() {
		It("sorts", func() {
			rs := MustParseIPRanges("10.0.1.0-10.0.1.10", "10.0.2.0-10.0.2.10", "10.0.0.0-10.0.0.10")

			Expect(NormalizeIPRanges(rs...).String()).To(Equal("[10.0.0.0-10.0.0.10, 10.0.1.0-10.0.1.10, 10.0.2.0-10.0.2.10]"))
		})

		It("joins overlapping", func() {
			rs := MustParseIPRanges("10.0.1.0-10.0.1.10", "10.0.1.3-10.0.2.10", "10.0.0.0-10.0.0.10")

			Expect(NormalizeIPRanges(rs...).String()).To(Equal("[10.0.0.0-10.0.0.10, 10.0.1.0-10.0.2.10]"))
		})

		It("skips included overlapping", func() {
			rs := MustParseIPRanges("10.0.1.0-10.0.1.10", "10.0.1.1-10.0.1.3", "10.0.0.0-10.0.0.10")

			Expect(NormalizeIPRanges(rs...).String()).To(Equal("[10.0.0.0-10.0.0.10, 10.0.1.0-10.0.1.10]"))
		})

		It("contains", func() {
			rs := MustParseIPRanges("10.0.1.0-10.0.1.10", "10.0.1.20-10.0.1.30")

			Expect(rs.Contains(ParseIP("10.0.1.5"))).To(BeTrue())
			Expect(rs.Contains(ParseIP("10.0.1.25"))).To(BeTrue())
		})
		It("not contains", func() {
			rs := MustParseIPRanges("10.0.1.0-10.0.1.10", "10.0.1.20-10.0.1.30")

			Expect(rs.Contains(ParseIP("10.0.0.0"))).To(BeFalse())
			Expect(rs.Contains(ParseIP("10.0.1.15"))).To(BeFalse())
			Expect(rs.Contains(ParseIP("10.0.1.35"))).To(BeFalse())
		})
	})

	Context("excludes", func() {
		_, cidr, _ := net.ParseCIDR("10.0.0.0/24")

		It("no includes", func() {

			Expect(Excludes(cidr)).To(BeNil())
		})

		It("all includes", func() {
			ranges := MustParseIPRanges("10.0.0.0/24")
			Expect(Excludes(cidr, ranges...)).To(BeNil())
		})
		It("indrect all includes", func() {
			ranges := MustParseIPRanges("10.0.0.0/25", "10.0.0.128/25")
			Expect(Excludes(cidr, ranges...)).To(BeNil())
		})

		It("single include", func() {
			ranges := MustParseIPRanges("10.0.0.8-10.0.0.15")
			excl, err := Excludes(cidr, ranges...)
			Expect(err).To(Not(HaveOccurred()))
			Expect(excl.String()).To(Equal("[10.0.0.0/29,10.0.0.128/25,10.0.0.64/26,10.0.0.32/27,10.0.0.16/28]"))
		})
		It("single include 2", func() {
			ranges := MustParseIPRanges("10.0.0.7-10.0.0.14")
			excl, err := Excludes(cidr, ranges...)
			Expect(err).To(Not(HaveOccurred()))
			Expect(excl.String()).To(Equal("[10.0.0.0/30,10.0.0.4/31,10.0.0.6/32,10.0.0.128/25,10.0.0.64/26,10.0.0.32/27,10.0.0.16/28,10.0.0.15/32]"))
		})
		It("single include 3", func() {
			ranges := MustParseIPRanges("10.0.0.9-10.0.0.16")
			excl, err := Excludes(cidr, ranges...)
			Expect(err).To(Not(HaveOccurred()))
			Expect(excl.String()).To(Equal("[10.0.0.0/29,10.0.0.8/32,10.0.0.128/25,10.0.0.64/26,10.0.0.32/27,10.0.0.24/29,10.0.0.20/30,10.0.0.18/31,10.0.0.17/32]"))
		})
		It("double include", func() {
			ranges := MustParseIPRanges("10.0.0.8-10.0.0.15", "10.0.0.128-10.0.0.135")
			excl, err := Excludes(cidr, ranges...)
			Expect(err).To(Not(HaveOccurred()))
			Expect(excl.String()).To(Equal("[10.0.0.0/29,10.0.0.64/26,10.0.0.32/27,10.0.0.16/28,10.0.0.192/26,10.0.0.160/27,10.0.0.144/28,10.0.0.136/29]"))
		})
	})

	Context("includes", func() {
		It("single cidr", func() {
			ranges := MustParseIPRanges("10.0.0.8-10.0.0.15")
			incl, err := Includes(ranges...)
			Expect(err).To(Not(HaveOccurred()))
			Expect(incl.String()).To(Equal("[10.0.0.8/29]"))
		})
		It("single range", func() {
			ranges := MustParseIPRanges("10.0.0.9-10.0.0.15")
			incl, err := Includes(ranges...)
			Expect(err).To(Not(HaveOccurred()))
			Expect(incl.String()).To(Equal("[10.0.0.9/32,10.0.0.10/31,10.0.0.12/30]"))
		})
		It("single range 2", func() {
			ranges := MustParseIPRanges("10.0.0.9-10.0.0.14")
			incl, err := Includes(ranges...)
			Expect(err).To(Not(HaveOccurred()))
			Expect(incl.String()).To(Equal("[10.0.0.9/32,10.0.0.10/31,10.0.0.12/31,10.0.0.14/32]"))
		})
	})
})
