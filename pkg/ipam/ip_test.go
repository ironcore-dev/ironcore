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

var _ = Describe("IP", func() {
	Context("ipv4", func() {
		It("parse", func() {
			ip := ParseIP("10.0.0.0")
			Expect(ip).NotTo(BeNil())
			Expect(len(ip)).To(Equal(net.IPv4len))
			Expect(ip.String()).To(Equal("10.0.0.0"))
		})
		It("cmp1", func() {
			a := ParseIP("10.0.8.0")
			b := ParseIP("10.0.0.0")
			Expect(IPCmp(a, b)).To(Equal(1))
			Expect(IPCmp(b, a)).To(Equal(-1))
			Expect(IPCmp(a, a)).To(Equal(0))
		})
		It("cmp2", func() {
			a := ParseIP("10.1.0.0")
			b := ParseIP("10.0.1.0")
			Expect(IPCmp(a, b)).To(Equal(1))
			Expect(IPCmp(b, a)).To(Equal(-1))
			Expect(IPCmp(a, a)).To(Equal(0))
		})
		It("diff", func() {
			a := ParseIP("10.1.0.0")
			b := ParseIP("10.0.1.0")
			Expect(IPDiff(a, b)).To(Equal(Int64(65536 - 256)))
			Expect(IPDiff(b, a)).To(Equal(Int64(-(65536 - 256))))
		})
	})
})
