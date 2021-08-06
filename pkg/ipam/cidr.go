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
	"math/big"
	"net"
)

var intOne = big.NewInt(1)
var intZero = big.NewInt(0)

func ParseCIDR(s string) (*net.IPNet, error) {
	ip, cidr, err := net.ParseCIDR(s)
	if err != nil {
		return nil, err
	}
	if len(cidr.IP) == net.IPv4len {
		return &net.IPNet{
			IP:   ip.To4(),
			Mask: cidr.Mask,
		}, nil
	}

	if v4 := ip.To4(); v4 != nil {
		return &net.IPNet{
			IP:   v4,
			Mask: cidr.Mask[12:],
		}, nil
	}
	return &net.IPNet{
		IP:   ip,
		Mask: cidr.Mask,
	}, nil
}

func MustParseCIDR(s string) *net.IPNet {
	cidr, err := ParseCIDR(s)
	if err != nil {
		panic(err)
	}
	return cidr
}

func CIDRNetMaskSize(cidr *net.IPNet) int {
	s, _ := cidr.Mask.Size()
	return s
}

func CIDRHostMaskSize(cidr *net.IPNet) int {
	s, l := cidr.Mask.Size()
	return l - s
}

func CIDRHostSize(cidr *net.IPNet) Int {
	s, l := cidr.Mask.Size()

	return IntOne.LShift(uint(l - s))
}

func CIDRBits(cidr *net.IPNet) int {
	return len(cidr.Mask) * 8
}

func CIDRClone(cidr *net.IPNet) *net.IPNet {
	return &net.IPNet{
		IP:   IPClone(cidr.IP),
		Mask: IPMaskClone(cidr.Mask),
	}
}

func CIDREqual(a, b *net.IPNet) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if !a.IP.Equal(b.IP) {
		return false
	}
	if !net.IP(a.Mask).Equal(net.IP(b.Mask)) {
		return false
	}
	return true
}

func CIDRAlign(cidr *net.IPNet, bits int) *net.IPNet {
	if bits != CIDRBits(cidr) {
		if bits == net.IPv4len*8 {
			cidr = CIDRto4(cidr)
		} else {
			cidr = CIDRto16(cidr)
		}
	}
	return cidr
}

func CIDRto4(cidr *net.IPNet) *net.IPNet {
	if len(cidr.Mask) == net.IPv4len {
		return cidr
	}
	ip := cidr.IP.To4()
	if ip == nil {
		return nil
	}
	return &net.IPNet{
		IP:   ip,
		Mask: cidr.Mask[12:],
	}
}

func CIDRto16(cidr *net.IPNet) *net.IPNet {
	if len(cidr.Mask) == net.IPv6len {
		return cidr
	}
	return &net.IPNet{
		IP:   cidr.IP.To4(),
		Mask: net.CIDRMask(net.IPv6len-net.IPv4len+CIDRNetMaskSize(cidr), net.IPv6len*8),
	}
}

func CIDRNet(cidr *net.IPNet) *net.IPNet {
	net := *cidr
	net.IP = CIDRFirstIP(cidr)
	return &net
}

func CIDRSubIP(cidr *net.IPNet, n int64) net.IP {
	if n < 0 || CIDRHostSize(cidr).Cmp(Int64(n)) <= 0 {
		return nil
	}
	return IPAdd(CIDRFirstIP(cidr), n)
}

func CIDRSubIPInt(cidr *net.IPNet, n Int) net.IP {
	if n.Sgn() < 0 || CIDRHostSize(cidr).Cmp(n) <= 0 {
		return nil
	}
	return IPAddInt(CIDRFirstIP(cidr), n)
}

func CIDRFirstIP(cidr *net.IPNet) net.IP {
	return cidr.IP.Mask(cidr.Mask)
}

func CIDRLastIP(cidr *net.IPNet) net.IP {
	ip := CIDRFirstIP(cidr)
	return or(ip, inv(cidr.Mask))
}

////////////////////////////////////////////////////////////////////////////////

func CIDRSplit(cidr *net.IPNet) (lower, upper *net.IPNet) {
	ones, bits := cidr.Mask.Size()
	if ones == bits {
		return nil, nil
	}
	mask := net.CIDRMask(ones+1, bits)
	delta := sub(mask, cidr.Mask)
	upper = &net.IPNet{
		IP:   net.IP(or(cidr.IP, delta)),
		Mask: mask,
	}
	lower = &net.IPNet{
		IP:   cidr.IP,
		Mask: mask,
	}
	return
}

func CIDRExtend(cidr *net.IPNet) *net.IPNet {
	ones, bits := cidr.Mask.Size()
	if ones == 0 {
		return nil
	}
	mask := net.CIDRMask(ones-1, bits)
	return &net.IPNet{
		IP:   net.IP(and(cidr.IP, mask)),
		Mask: mask,
	}
}

////////////////////////////////////////////////////////////////////////////////

func IPMaskClone(mask net.IPMask) net.IPMask {
	return append(mask[:0:0], mask...)
}
