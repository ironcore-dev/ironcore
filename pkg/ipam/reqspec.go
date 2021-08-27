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
	"strconv"
	"strings"
)

// RequestSpec represents a dedicate allocation type for a cidr from an IPAM.
type RequestSpec interface {
	IsCIDR() bool
	Bits() int
	Alloc(ipam *IPAM) (*net.IPNet, error)
	String() string
}

type RequestSpecList []RequestSpec

func (r RequestSpecList) String() string {
	s := "["
	sep := ""
	for _, spec := range r {
		s = fmt.Sprintf("%s%s%s", s, sep, spec)
		sep = ","
	}
	return s + "]"
}

func ParseRequestSpec(s string) (RequestSpec, error) {
	if strings.HasPrefix(s, "#") {
		n, err := ParseInt(s[1:])
		if err != nil {
			return nil, err
		}
		return &amountSpec{size: n}, nil
	}
	_, cidr, err := net.ParseCIDR(s)
	if err == nil {
		return &cidrSpec{cidr: cidr}, nil
	}

	ip := ParseIP(s)
	if ip != nil {
		return &cidrSpec{cidr: IPtoCIDR(ip)}, nil
	}

	size, err := strconv.ParseInt(s, 10, 32)
	if err == nil {
		if size < 0 || size > 128 {
			return nil, fmt.Errorf("invalid request spec: size must be between 0 and 128")
		}
		return &sizeSpec{size: int(size)}, nil
	}
	if ip != nil {
		return &cidrSpec{cidr: IPtoCIDR(ip)}, nil
	}

	idx := strings.Index(s, "/")
	if idx >= 0 {
		size, err := strconv.ParseInt(s[idx+1:], 10, 32)
		if err == nil {
			if size < 0 || size > 128 {
				return nil, fmt.Errorf("invalid request spec: size must be between 0 and 128")
			}
			if idx == 0 {
				return &sizeSpec{size: int(size)}, nil
			}
			i, err := strconv.ParseInt(s[:idx], 10, 32)
			if err == nil {
				if size < 0 {
					return nil, fmt.Errorf("invalid request spec: index must not be negative")
				}
				return &subSpec{size: int(size), index: int(i)}, nil
			}
		}
	}
	return nil, fmt.Errorf("invalid request spec: must be CIDR, IP, size or sub cidr (n/size)")
}

////////////////////////////////////////////////////////////////////////////////
// sizeSpec is a RequestSpec based of a netmask size

type sizeSpec struct {
	specsupport
	size int
}

func (this *sizeSpec) String() string {
	return strconv.Itoa(this.size)
}

func (this *sizeSpec) IsCIDR() bool {
	return false
}

func (this *sizeSpec) Bits() int {
	return this.size
}

func (this *sizeSpec) Alloc(ipam *IPAM) (*net.IPNet, error) {
	return this.alloc(ipam, this.size)
}

////////////////////////////////////////////////////////////////////////////////
// amountSpec is a RequestSpec based of a netmask size

type amountSpec struct {
	specsupport
	size Int
}

func (this *amountSpec) String() string {
	return "#" + this.size.String()
}

func (this *amountSpec) IsCIDR() bool {
	return false
}

func (this *amountSpec) Bits() int {
	d := IntOne
	c := 0
	for d.Cmp(this.size) < 0 {
		d = d.LShift(1)
		c++
	}
	return c
}

func (this *amountSpec) Alloc(ipam *IPAM) (*net.IPNet, error) {
	bits := this.Bits()
	if ipam.Bits() < bits {
		return nil, fmt.Errorf("IPAM too small for %d bit hostnet size", bits)
	}
	return this.alloc(ipam, ipam.Bits()-bits)
}

////////////////////////////////////////////////////////////////////////////////
// cidrSpec is a RequestSpec based of a dedired CIDR

type cidrSpec struct {
	specsupport
	cidr *net.IPNet
}

func (this *cidrSpec) String() string {
	return this.cidr.String()
}

func (this *cidrSpec) IsCIDR() bool {
	return true
}

func (this *cidrSpec) Bits() int {
	return CIDRBits(this.cidr)
}

func (this *cidrSpec) Alloc(ipam *IPAM) (*net.IPNet, error) {
	first := CIDRFirstIP(this.cidr)
	last := CIDRLastIP(this.cidr)
	for _, r := range ipam.ranges {
		if r.Contains(first) && r.Contains(last) {
			if ipam.Busy(this.cidr) {
				return this.cidr, nil
			}
			return nil, nil
		}
	}
	return nil, fmt.Errorf("cidr %s not included in IPAM ranges", this.cidr)
}

////////////////////////////////////////////////////////////////////////////////
// subSpec is a RequestSpec requesting the nth cidr of a dedicated size

type subSpec struct {
	specsupport
	size  int
	index int
}

func (this *subSpec) String() string {
	return fmt.Sprintf("%d/%d", this.index, this.size)
}

func (this *subSpec) IsCIDR() bool {
	return false
}

func (this *subSpec) Bits() int {
	return this.size
}

func (this *subSpec) Alloc(ipam *IPAM) (*net.IPNet, error) {
	index := this.index
	var cidr *net.IPNet
	for _, r := range ipam.ranges {
		if CIDRNetMaskSize(r) > this.size {
			return nil, fmt.Errorf("invalid rquest spec %s for ipam ranges", this)
		}
		num := 1 << (this.size - CIDRNetMaskSize(r))
		if index < num {
			_, bits := r.Mask.Size()
			size := IntOne.LShift(uint(bits - this.size)).Mul(Int64(int64(index)))
			cidr = &net.IPNet{
				IP:   IPAddInt(r.IP, size),
				Mask: net.CIDRMask(this.size, bits),
			}
			break
		}
		index -= num
	}
	if cidr == nil {
		return nil, fmt.Errorf("invalid rquest spec %s for ipam ranges: too small ranges", this)
	}
	if ipam.Busy(cidr) {
		return cidr, nil
	}
	return nil, nil
}

////////////////////////////////////////////////////////////////////////////////
// specsupport is a helper class for RequestSpec implementations

type specsupport struct{}

func (this specsupport) alloc(ipam *IPAM, size int) (*net.IPNet, error) {
	if err := this.checkForHostMaskSize(ipam, size); err != nil {
		return nil, err
	}
	return ipam.Alloc(size), nil
}

func (this specsupport) checkForHostMaskSize(ipam *IPAM, size int) error {
	for _, r := range ipam.ranges {
		if CIDRNetMaskSize(r) > size {
			return fmt.Errorf("ipam ranges too small for requested host netmask")
		}
	}
	return nil
}
