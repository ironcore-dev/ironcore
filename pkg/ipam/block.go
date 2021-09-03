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
	"strings"
)

////////////////////////////////////////////////////////////////////////////////

func NewBlock(cidr *net.IPNet, bitmap Bitmap) *Block {
	return &Block{
		busy: bitmap,
		cidr: cidr,
	}
}

func ParseBlock(s string) *Block {
	idx := strings.Index(s, "[")
	if idx <= 0 || !strings.HasSuffix(s, "]") {
		return nil
	}

	cidr, err := ParseCIDR(s[:idx])
	if err != nil {
		return nil
	}

	switch state := s[idx+1 : len(s)-1]; state {
	case "free":
		return &Block{
			busy: 0,
			cidr: cidr,
		}
	case "busy":
		return &Block{
			busy: BITMAP_BUSY,
			cidr: cidr,
		}
	default:
		mask := uint64(0)
		for _, c := range state {
			switch c {
			case ' ':
			case '1', '0':
				mask = (mask << 1) + uint64(c-'0')
			default:
				return nil
			}
		}
		return &Block{
			busy: Bitmap(mask),
			cidr: cidr,
		}
	}

}

type Block struct {
	busy Bitmap
	cidr *net.IPNet
	prev *Block
	next *Block
}

func (this *Block) canAlloc(next net.IP, reqsize int) bool {
	s, l := this.cidr.Mask.Size()
	if s > reqsize {
		return false
	}
	if next != nil && IPCmp(CIDRLastIP(this.cidr), next) < 0 {
		return false
	}

	if l-s <= MAX_BITMAP_NET {
		n := reqsize - l + MAX_BITMAP_NET
		start := 0
		if next != nil {
			start = int(IPDiff(next, this.cidr.IP).Int64())
		}
		f := this.busy.canAllocate2(start, n)
		return f >= 0 && f < 1<<(l-s)
	}
	return this.busy == 0
}

func (this *Block) canSplit() bool {
	s, l := this.cidr.Mask.Size()
	return l-s > MAX_BITMAP_NET
}

func (this *Block) matchSize(reqsize int) bool {
	if this.Size() == reqsize {
		return true
	}
	s, l := this.cidr.Mask.Size()
	if l-s != MAX_BITMAP_NET {
		return false
	}
	return this.busy != 0
}

func (this *Block) isBusy() bool {
	return this.busy != 0
}

func (this *Block) isCIDRBusy(cidr *net.IPNet) bool {
	s, l := this.cidr.Mask.Size()
	r := CIDRNetMaskSize(cidr)
	if s > r {
		return false
	}
	if !this.cidr.Contains(cidr.IP) {
		return false
	}

	if l-s <= MAX_BITMAP_NET && s < r {
		return this.busy.get(int(cidr.IP[len(cidr.IP)-1]&((1<<(l-s))-1)), r-l+MAX_BITMAP_NET) != 0
	}

	return this.busy != 0
}

func (this *Block) isCompletelyBusy() bool {
	s := this.HostSize()
	if s <= MAX_BITMAP_NET {
		return this.busy&hostmask[s] == hostmask[s]
	}
	return this.busy == BITMAP_BUSY
}

func (this *Block) matchState(b *Block) bool {
	return this.busy == b.busy && (this.busy == 0 || this.isCompletelyBusy())
}

func (this *Block) set(cidr *net.IPNet, busy bool) bool {
	s, l := this.cidr.Mask.Size()
	r := CIDRNetMaskSize(cidr)
	if s > r {
		return false
	}
	if !this.cidr.Contains(cidr.IP) {
		return false
	}

	if s < r {
		return this.busy.set(int(cidr.IP[len(cidr.IP)-1]&((1<<(l-s))-1)), r-l+MAX_BITMAP_NET, busy)
	}

	if this.isBusy() == busy {
		return false
	}
	if busy {
		this.busy = BITMAP_BUSY
	} else {
		this.busy = 0
	}
	return true
}

func (this *Block) alloc(next net.IP, reqsize int) *net.IPNet {
	s, l := this.cidr.Mask.Size()
	if s > reqsize {
		return nil
	}

	if l-s <= MAX_BITMAP_NET {
		n := reqsize - l + MAX_BITMAP_NET
		start := 0
		if next != nil {
			start = int(IPDiff(next, this.cidr.IP).Int64())
		}
		ip := this.busy.allocate2(start, n)

		c := &net.IPNet{
			IP:   CIDRSubIP(this.cidr, int64(ip)),
			Mask: net.CIDRMask(reqsize, l),
		}
		return c
	}

	if reqsize != s {
		return nil
	}
	cidr := this.cidr
	this.busy = BITMAP_BUSY
	return cidr
}

func (this *Block) split() *Block {
	ones, bits := this.cidr.Mask.Size()
	hostsize := bits - ones
	if hostsize == 0 {
		return nil
	}
	mask := net.CIDRMask(ones+1, bits)
	delta := sub(mask, this.cidr.Mask)
	busy := this.busy
	if hostsize <= MAX_BITMAP_NET {
		busy = this.busy >> (1 << (hostsize - 1))
		this.busy &= hostmask[hostsize-1]
	}
	upper := &Block{
		cidr: &net.IPNet{
			IP:   net.IP(or(this.cidr.IP, delta)),
			Mask: mask,
		},
		busy: busy,
		prev: this,
		next: this.next,
	}

	if this.next != nil {
		this.next.prev = upper
	}
	this.next = upper
	this.cidr = &net.IPNet{
		IP:   this.cidr.IP,
		Mask: mask,
	}
	return upper
}

func (this *Block) buddies() (*Block, *Block, int, int) {
	ones, bits := this.cidr.Mask.Size()
	var upper *Block
	var lower *Block
	var oones int
	if this.IsUpper() {
		if this.prev == nil {
			return nil, nil, ones, bits
		}
		upper = this
		lower = this.prev
		oones, _ = lower.cidr.Mask.Size()
	} else {
		if this.next == nil {
			return nil, nil, ones, bits
		}
		lower = this
		upper = this.next
		oones, _ = upper.cidr.Mask.Size()
	}
	if ones != oones {
		return nil, nil, ones, bits
	}
	if IPAdd(lower.cidr.IP, 1<<(bits-ones)).Equal(upper.cidr.IP) {
		return lower, upper, ones, bits
	}
	return nil, nil, ones, bits
}

func (this *Block) join() *Block {
	lower, upper, ones, bits := this.buddies()
	if lower == nil || upper == nil {
		return nil
	}

	hostsize := this.HostSize()
	if !lower.matchState(upper) && hostsize >= MAX_BITMAP_NET {
		return nil
	}

	if upper.next != nil {
		upper.next.prev = lower
	}
	lower.next = upper.next

	mask := net.CIDRMask(ones-1, bits)
	lower.cidr = &net.IPNet{
		IP:   lower.cidr.IP,
		Mask: mask,
	}
	if hostsize < MAX_BITMAP_NET {
		lower.busy = lower.busy | (upper.busy << (1 << hostsize))
	}
	return lower
}

func (this *Block) String() string {
	msg := "free"
	if this.busy != 0 {
		if CIDRHostMaskSize(this.cidr) <= MAX_BITMAP_NET {
			t := fmt.Sprintf("%064b", this.busy)
			msg = ""
			for i := 0; i < MAX_BITMAP_SIZE; i += 8 {
				if msg != "" || t[i:i+8] != "00000000" {
					msg += " " + t[i:i+8]
				}
			}
			msg = msg[1:]
		} else {
			msg = "busy"
		}
	}
	return fmt.Sprintf("%s[%s]", this.cidr.String(), msg)
}

func (this *Block) Next() *Block {
	return this.next
}

func (this *Block) Prev() *Block {
	return this.prev
}

func (this *Block) Size() int {
	return CIDRNetMaskSize(this.cidr)
}

func (this *Block) HostSize() int {
	return CIDRHostMaskSize(this.cidr)
}

func (this *Block) IsUpper() bool {
	return CIDRIsUpper(this.cidr)
}
