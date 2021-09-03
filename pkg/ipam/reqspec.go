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
	"sort"
	"strconv"
	"strings"
	"sync"
)

type RequestSpecType func(s string) (RequestSpec, bool, error)

var lock sync.RWMutex
var requesttypes map[string]RequestSpecType = map[string]RequestSpecType{}
var reqsyntax string

func RegisterRequestType(name string, t RequestSpecType) {
	lock.Lock()
	defer lock.Unlock()
	requesttypes[name] = t

	var list []string
	reqsyntax = ""
	sep := ""
	for n := range requesttypes {
		list = append(list, n)
	}
	sort.Strings(list)
	for _, n := range list {
		reqsyntax = fmt.Sprintf("%s%s%s", reqsyntax, sep, n)
		sep = ", "
	}
}

func init() {
	RegisterRequestType("#<amount>", amountSpecType)
	RegisterRequestType("<cidr>", cidrSpecType)
	RegisterRequestType("<ip>", ipSpecType)
	RegisterRequestType("<netmasksize>", netmasksizeSpecType)
	RegisterRequestType("%<hostmasksize>", hostmasksizeSpecType)
	RegisterRequestType("[<n>]/[%]<masksize>", subSpecType)
}

// RequestSpec represents a dedicate allocation type for a cidr from an IPAM.
type RequestSpec interface {
	IsCIDR() bool
	Bits() int
	Alloc(ipam *IPAM) (*net.IPNet, error)
	String() string
}

type RequestSpecList []RequestSpec

func (list RequestSpecList) String() string {
	sep := ""
	s := "["
	for _, spec := range list {
		s = fmt.Sprintf("%s%s%s", s, sep, spec)
		sep = ", "
	}
	return s + "]"
}

func ParseRequestSpec(s string) (RequestSpec, error) {
	var err error
	for syn, parser := range requesttypes {
		spec, final, perr := parser(s)
		if spec != nil {
			return spec, err
		}
		if perr != nil {
			err = fmt.Errorf("%s: %s", syn, perr)
			if final {
				break
			}
		}
	}
	if err == nil {
		lock.RLock()
		defer lock.RUnlock()
		err = fmt.Errorf("invalid request spec: use one of %s", reqsyntax)
	}
	return nil, err
}

////////////////////////////////////////////////////////////////////////////////
// netmasksizeSpec is a RequestSpec based of a netmask size

type netmasksizeSpec struct {
	specsupport
	size int
}

func netmasksizeSpecType(s string) (RequestSpec, bool, error) {
	size, err := strconv.ParseInt(s, 10, 32)
	if err == nil {
		if size < 0 || size > 128 {
			return nil, true, fmt.Errorf("invalid request spec: size must be between 0 and 128")
		}
		return &netmasksizeSpec{size: int(size)}, true, nil
	}
	return nil, false, nil
}

func (this *netmasksizeSpec) String() string {
	return strconv.Itoa(this.size)
}

func (this *netmasksizeSpec) IsCIDR() bool {
	return false
}

func (this *netmasksizeSpec) Bits() int {
	return this.size
}

func (this *netmasksizeSpec) Alloc(ipam *IPAM) (*net.IPNet, error) {
	return this.alloc(ipam, this.size)
}

////////////////////////////////////////////////////////////////////////////////
// netmasksizeSpec is a RequestSpec based of a netmask size

type hostmasksizeSpec struct {
	netmasksizeSpec
}

func hostmasksizeSpecType(s string) (RequestSpec, bool, error) {
	if !strings.HasPrefix(s, "%") {
		return nil, false, nil
	}
	spec, _, err := netmasksizeSpecType(s[1:])
	if spec == nil || err != nil {
		return spec, true, err
	}
	return &hostmasksizeSpec{*(spec.(*netmasksizeSpec))}, false, nil
}

func (this *hostmasksizeSpec) String() string {
	return "%" + this.netmasksizeSpec.String()
}

func (this *hostmasksizeSpec) Alloc(ipam *IPAM) (*net.IPNet, error) {
	return this.alloc(ipam, ipam.Bits()-this.size)
}

////////////////////////////////////////////////////////////////////////////////
// amountSpec is a RequestSpec based of a netmask size

type amountSpec struct {
	specsupport
	size Int
}

func amountSpecType(s string) (RequestSpec, bool, error) {
	if strings.HasPrefix(s, "#") {
		n, err := ParseInt(s[1:])
		if err != nil {
			return nil, true, err
		}
		return &amountSpec{size: n}, true, nil
	}
	return nil, false, nil
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

func cidrSpecType(s string) (RequestSpec, bool, error) {
	_, cidr, err := net.ParseCIDR(s)
	if err == nil {
		return &cidrSpec{cidr: cidr}, true, nil
	}
	if !strings.Contains(s, "/") || !strings.ContainsAny(s, ".:") {
		err = nil
	}
	return nil, false, err
}

func ipSpecType(s string) (RequestSpec, bool, error) {
	ip := ParseIP(s)
	if ip != nil {
		return &cidrSpec{cidr: IPtoCIDR(ip)}, true, nil
	}
	var err error
	if !strings.Contains(s, "/") && strings.ContainsAny(s, ".:") {
		err = fmt.Errorf("invalid IP address: %s", s)
	}
	return nil, false, err
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
	host  bool
	size  int
	index int
}

func subSpecType(s string) (RequestSpec, bool, error) {
	idx := strings.Index(s, "/")
	if idx >= 0 {
		start := idx + 1
		host := false
		if s[start] == '%' {
			host = true
			start++
		}
		size, err := strconv.ParseInt(s[start:], 10, 32)
		if err == nil {
			if size < 0 || size > 128 {
				return nil, true, fmt.Errorf("invalid request spec: size must be between 0 and 128")
			}
			if idx == 0 {
				if host {
					return &hostmasksizeSpec{netmasksizeSpec{size: int(size)}}, true, nil
				}
				return &netmasksizeSpec{size: int(size)}, true, nil
			}
			i, err := strconv.ParseInt(s[:idx], 10, 32)
			if err == nil {
				if size < 0 {
					return nil, true, fmt.Errorf("invalid request spec: index must not be negative")
				}
				return &subSpec{host: host, size: int(size), index: int(i)}, true, nil
			}
		}
	}
	return nil, false, nil
}

func (this *subSpec) String() string {
	if this.host {
		return fmt.Sprintf("%d/%%%d", this.index, this.size)
	}
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
	size := this.size
	if this.host {
		size = ipam.Bits() - size
	}
	for _, r := range ipam.ranges {
		if CIDRNetMaskSize(r) > size {
			return nil, fmt.Errorf("invalid request spec %s for ipam ranges", this)
		}
		num := 1 << (size - CIDRNetMaskSize(r))
		if index < num {
			_, bits := r.Mask.Size()
			hostsize := IntOne.LShift(uint(bits - size)).Mul(Int64(int64(index)))
			cidr = &net.IPNet{
				IP:   IPAddInt(r.IP, hostsize),
				Mask: net.CIDRMask(size, bits),
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
