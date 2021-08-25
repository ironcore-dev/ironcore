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

var _ = Describe("Request Spec", func() {

	Context("parse", func() {
		It("cidr", func() {
			req, err := ParseRequestSpec("10.0.0.0/8")
			Expect(err).To(BeNil())
			Expect(req.String()).To(Equal("10.0.0.0/8"))
		})

		It("ip", func() {
			req, err := ParseRequestSpec("10.0.0.5")
			Expect(err).To(BeNil())
			Expect(req.String()).To(Equal("10.0.0.5/32"))
		})

		It("size", func() {
			req, err := ParseRequestSpec("24")
			Expect(err).To(BeNil())
			Expect(req.String()).To(Equal("24"))
		})
		It("sub size", func() {
			req, err := ParseRequestSpec("/24")
			Expect(err).To(BeNil())
			Expect(req.String()).To(Equal("24"))
		})
		It("sub cidr", func() {
			req, err := ParseRequestSpec("1/24")
			Expect(err).To(BeNil())
			Expect(req.String()).To(Equal("1/24"))
		})

		It("invalid size", func() {
			_, err := ParseRequestSpec("129")
			Expect(err).To(Equal(fmt.Errorf("invalid request spec: size must be between 0 and 128")))
		})
		It("invalid size", func() {
			_, err := ParseRequestSpec("/129")
			Expect(err).To(Equal(fmt.Errorf("invalid request spec: size must be between 0 and 128")))
		})
	})

	Context("alloc", func() {
		ranges, _ := ParseIPRanges("10.1.0.0/16", "10.10.0.0/16")
		var ipam *IPAM

		BeforeEach(func() {
			ipam, _ = NewIPAMForRanges(ranges)
		})

		It("0/24", func() {
			req, err := ParseRequestSpec("0/24")
			Expect(err).To(BeNil())
			cidr, err := req.Alloc(ipam)
			Expect(err).To(BeNil())
			Expect(cidr).NotTo(BeNil())
			Expect(cidr.String()).To(Equal("10.1.0.0/24"))
		})
		It("1/24", func() {
			req, err := ParseRequestSpec("1/24")
			Expect(err).To(BeNil())
			cidr, err := req.Alloc(ipam)
			Expect(err).To(BeNil())
			Expect(cidr).NotTo(BeNil())
			Expect(cidr.String()).To(Equal("10.1.1.0/24"))
		})
		It("255/24", func() {
			req, err := ParseRequestSpec("255/24")
			Expect(err).To(BeNil())
			cidr, err := req.Alloc(ipam)
			Expect(err).To(BeNil())
			Expect(cidr).NotTo(BeNil())
			Expect(cidr.String()).To(Equal("10.1.255.0/24"))
		})
		It("256/24", func() {
			req, err := ParseRequestSpec("256/24")
			Expect(err).To(BeNil())
			cidr, err := req.Alloc(ipam)
			Expect(err).To(BeNil())
			Expect(cidr).NotTo(BeNil())
			Expect(cidr.String()).To(Equal("10.10.0.0/24"))
		})
		It("512/24", func() {
			req, err := ParseRequestSpec("512/24")
			Expect(err).To(BeNil())
			_, err = req.Alloc(ipam)
			Expect(err).To(Equal(fmt.Errorf("invalid rquest spec 512/24 for ipam ranges: too small ranges")))
		})
	})
})
