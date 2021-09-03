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
	"math/big"
	"net"
	"sort"
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

func CIDRContains(a, b *net.IPNet) bool {
	if len(a.IP) != len(b.IP) {
		return false
	}
	sa := CIDRNetMaskSize(a)
	sb := CIDRNetMaskSize(b)
	if sa > sb {
		return false
	}
	return a.IP.Mask(a.Mask).Equal(b.IP.Mask(a.Mask))
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

func CIDRIsUpper(cidr *net.IPNet) bool {
	ones, bits := cidr.Mask.Size()
	delta := sub(cidr.Mask, net.CIDRMask(ones-1, bits))
	return !isZero(and(delta, cidr.IP))
}

////////////////////////////////////////////////////////////////////////////////

func IPMaskClone(mask net.IPMask) net.IPMask {
	return append(mask[:0:0], mask...)
}

////////////////////////////////////////////////////////////////////////////////

type CIDRList []*net.IPNet

func CIDRLess(a, b *net.IPNet) bool {
	d := IPCmp(a.IP, b.IP)
	switch {
	case d < 0:
		return true
	case d > 0:
		return false
	default:
		return CIDRNetMaskSize(a) > CIDRNetMaskSize(b)
	}
}

func (p CIDRList) Len() int           { return len(p) }
func (p CIDRList) Less(i, j int) bool { return CIDRLess(p[i], p[j]) }
func (p CIDRList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (this *CIDRList) String() string {
	sep := "["
	end := ""
	s := ""
	for _, c := range *this {
		s = fmt.Sprintf("%s%s%s", s, sep, c)
		sep = ","
		end = "]"
	}
	return s + end
}

func (this *CIDRList) Add(cidrs ...*net.IPNet) {
	*this = append(*this, cidrs...)
}

func (this *CIDRList) DeleteIndex(i int) {
	(*this) = append((*this)[:i], (*this)[i+1:]...)
}

func (this CIDRList) IsEmpty() bool {
	return len(this) == 0
}

func (this CIDRList) Contains(ip net.IP) bool {
	for _, c := range this {
		if c.Contains(ip) {
			return true
		}
	}
	return false
}

func (this CIDRList) ContainsCIDR(cidr *net.IPNet) bool {
	for _, c := range this {
		if CIDRContains(c, cidr) {
			return true
		}
	}
	return false
}

func (this CIDRList) Copy() CIDRList {
	var result CIDRList
	for _, c := range this {
		result = append(result, c)
	}
	return result
}

func (this CIDRList) Additional(b CIDRList) CIDRList {
	a := this.Copy()
	b = b.Copy()
	a.Normalize()
	b.Normalize()
	return a.additional(b)
}

func (this *CIDRList) AddNormalized(list CIDRList) CIDRList {
	toAdd := this.Additional(list)
	this.Add(toAdd...)
	this.Normalize()
	return toAdd
}

func (this *CIDRList) DeleteNormalized(list CIDRList) CIDRList {
	cur := *this
	*this = list.Additional(*this)
	toDel := this.Additional(cur)
	return toDel
}

func (this CIDRList) additional(b CIDRList) CIDRList {
	var result CIDRList

	for _, o := range this {
		fo := CIDRFirstIP(o)
		lo := CIDRLastIP(o)

		for len(b) > 0 {
			n := b[0]
			fn := CIDRFirstIP(n)

			if IPCmp(fn, lo) > 0 {
				break
			}

			b = b[1:]
			if CIDRContains(o, n) {
				continue
			}

			if CIDRContains(n, o) {
				var rest CIDRList
				addIncludes(&result, n, true, IPAdd(fo, -1))
				addIncludes(&rest, n, false, IPAdd(CIDRLastIP(o), 1))
				b = append(rest, b...)
				continue
			}
			result = append(result, n)
		}
	}
	result = append(result, b...)
	sort.Sort(result)
	return result
}

func (this CIDRList) Sort() {
	sort.Sort(this)
}

func (this *CIDRList) Normalize() {
	this.Sort()

	i := 0
	for i := range *this {
		(*this)[i] = CIDRNet((*this)[i])
	}
	for i < len(*this) {
		n, j := this.join(i)
		if n != nil {
			*this = append(append((*this)[:j], n), (*this)[j+2:]...)
			i = j
		} else {
			i++
		}
	}
}

func (this CIDRList) join(index int) (*net.IPNet, int) {
	cidr := this[index]
	ones, bits := cidr.Mask.Size()
	var upper int
	var lower int
	var oones int
	if index < len(this)-1 {
		if CIDRContains(this[index+1], this[index]) {
			return this[index+1], index
		}
	}
	if index > 0 {
		if CIDRContains(this[index-1], this[index]) {
			return this[index-1], index - 1
		}
	}
	if CIDRIsUpper(cidr) {
		if index == 0 {
			return nil, -1
		}
		upper = index
		lower = index - 1
		oones, _ = this[lower].Mask.Size()
	} else {
		if index == len(this)-1 {
			return nil, -1
		}
		lower = index
		upper = index + 1
		oones, _ = this[upper].Mask.Size()
	}
	if ones != oones {
		return nil, -1
	}
	if IPAdd(this[lower].IP, 1<<(bits-ones)).Equal(this[upper].IP) {
		return CIDRExtend(this[lower]), lower
	}
	return nil, -1
}

func (this CIDRList) Equal(list CIDRList) bool {
	if len(this) != len(list) {
		return false
	}
	for i, e := range this {
		if !CIDREqual(e, list[i]) {
			return false
		}
	}
	return true
}
