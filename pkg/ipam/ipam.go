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
)

type IPAM struct {
	ranges        CIDRList
	block         *Block
	nextAlloc     []net.IP
	roundRobin    bool
	deletePending CIDRList
}

func NewIPAM(cidr *net.IPNet, ranges ...*IPRange) (*IPAM, error) {
	var nextAlloc []net.IP
	copy := *cidr
	if len(cidr.Mask) == net.IPv4len {
		copy.IP = cidr.IP.To4()
		nextAlloc = make([]net.IP, net.IPv4len*8+1)
	} else {
		copy.IP = cidr.IP.To16()
		nextAlloc = make([]net.IP, net.IPv6len*8+1)
	}
	block := &Block{
		cidr: &copy,
	}

	_ = nextAlloc
	ipam := &IPAM{
		ranges:    []*net.IPNet{&copy},
		block:     block,
		nextAlloc: nextAlloc,
	}
	if len(ranges) > 0 {
		cidrs, err := Excludes(cidr, ranges...)
		if err != nil {
			return nil, err
		}
		for _, c := range cidrs {
			ipam.Busy(c)
		}

		/*
			for b := ipam.block; b != nil; b = b.next {
				if b.isCompletelyBusy() {
					if b.prev != nil {
						b.prev.next = b.next
					}
					if b.next != nil {
						b.next.prev = b.prev
					}
					if b == ipam.block {
						ipam.block = b.next
					}
				}
			}
		*/

		if ipam.block == nil {
			return nil, fmt.Errorf("no available IP addresses")
		}
	}
	return ipam, nil
}

func NewIPAMForRanges(ranges IPRanges) (*IPAM, error) {
	var nextAlloc []net.IP
	if len(ranges) == 0 {
		return nil, fmt.Errorf("no ranges specified for IPAM")
	}
	cidrs, err := Includes(ranges...)
	if err != nil {
		return nil, err
	}
	cidrs.Sort()

	ipv4 := true
	for _, cidr := range cidrs {
		if cidr.IP.To4() == nil {
			ipv4 = false
		}
		break
	}

	if ipv4 {
		nextAlloc = make([]net.IP, net.IPv4len*8+1)
	} else {
		nextAlloc = make([]net.IP, net.IPv6len*8+1)
	}
	_ = nextAlloc
	ipam := &IPAM{
		ranges:    cidrs,
		nextAlloc: nextAlloc,
	}

	ipam.setupFor(ipv4, cidrs...)
	return ipam, nil
}

func (this *IPAM) setupFor(ipv4 bool, cidrs ...*net.IPNet) {
	last := this.block
	for _, cidr := range cidrs {
		var b *Block
		if ipv4 {
			cidr = CIDRto4(cidr)
		} else {
			cidr = CIDRto16(cidr)
		}

		b = &Block{cidr: cidr}
		b.prev = last
		if last != nil {
			last.next = b
		} else {
			this.block = b
		}
		last = b
	}
}

func (this *IPAM) AddCIDRs(list CIDRList) {
	this.insert(this.ranges.AddNormalized(list))
}

func (this *IPAM) DeleteCIDRs(list CIDRList) {
	this.delete(this.ranges.DeleteNormalized(list))
}

func (this *IPAM) insert(cidrs CIDRList) {
	for _, a := range cidrs {
		b := &Block{
			cidr: a,
		}

		var prev *Block = nil
		c := this.block
		for c != nil {
			if IPCmp(b.cidr.IP, c.cidr.IP) < 0 {
				break
			}
			prev = c
			c = c.next
		}
		this.insertBlock(prev, b)
	}
}

func (this *IPAM) delete(cidrs CIDRList) {
	i := 0
	for i < len(cidrs) {
		cidr := cidrs[i]
		b := this.block
		for b != nil {
			match, deleted := this.deletePart(b, cidr)
			if match {
				if deleted {
					cidrs.DeleteIndex(i)
				} else {
					i++
				}
				break
			}
			b = b.next
		}
		i++
	}
	this.deletePending = cidrs
}

func (this *IPAM) deletePart(b *Block, cidr *net.IPNet) (bool, bool) {
	if CIDRContains(b.cidr, cidr) {
		if !b.isCIDRBusy(cidr) {
			for b.Size() < CIDRNetMaskSize(cidr) {
				b.split()
				if !b.cidr.IP.Equal(cidr.IP) {
					b = b.next
				}
			}
			this.removeBlock(b)
			return true, true
		}
		return true, false
	}
	return false, false
}

func (this *IPAM) SetRoundRobin(b bool) {
	if !b && this.roundRobin {
		this.nextAlloc = make([]net.IP, len(this.nextAlloc), len(this.nextAlloc))
	}
	this.roundRobin = b
}

func (this *IPAM) Ranges() CIDRList {
	return this.ranges.Copy()
}

func (this *IPAM) PendingDeleted() CIDRList {
	return this.deletePending.Copy()
}

func (this *IPAM) IsCoveredCIDR(cidr *net.IPNet) bool {
	for _, r := range this.ranges {
		if CIDRContains(r, cidr) {
			return true
		}
	}
	return false
}

func (this *IPAM) State() ([]string, []net.IP) {
	state := []string{}
	b := this.block
	for b != nil {
		state = append(state, b.String())
		b = b.next
	}
	return state, this.nextAlloc
}

func (this *IPAM) IsRoundRobin() bool {
	return this.roundRobin
}

func (this *IPAM) SetState(blocks []string, next []net.IP) (CIDRList, error) {
	var additional CIDRList

	if len(next) > len(this.nextAlloc) {
		return nil, fmt.Errorf("invalid state")
	}
	ipv4 := this.Bits() == net.IPv4len*8
	for i := 0; i < len(this.nextAlloc); i++ {
		if i < len(next) && next[i] != nil {
			if ipv4 {
				this.nextAlloc[i] = next[i].To4()
			} else {
				this.nextAlloc[i] = next[i].To16()
			}
		} else {
			this.nextAlloc[i] = nil
		}
	}

	if blocks != nil {
		var block *Block
		var last *Block
		var ranges CIDRList
		for _, s := range blocks {
			b := ParseBlock(s)
			if b == nil {
				return nil, fmt.Errorf("invalid block state")
			}
			ranges.Add(b.cidr)
			b.prev = last
			if last == nil {
				block = b
			} else {
				last.next = b
			}
			last = b
		}
		this.block = block

		ranges.Normalize()
		required := this.ranges.Copy()
		required.Normalize()

		additional = ranges.additional(required)
		this.insert(additional)

		deleted := required.additional(ranges)
		this.delete(deleted)
	}
	return additional, nil
}

func (this *IPAM) Bits() int {
	return len(this.nextAlloc) - 1
	// return CIDRBits(this.block.cidr)
}

func (this *IPAM) String() string {
	s := ""
	sep := ""
	b := this.block
	for b != nil {
		s = fmt.Sprintf("%s%s%s", s, sep, b)
		sep = ", "
		b = b.next
	}
	return s
}

func (this *IPAM) getNext(reqsize int) net.IP {
	if this.roundRobin {
		return this.nextAlloc[reqsize]
	}
	return nil
}

func (this *IPAM) setNext(cidr *net.IPNet) {
	if this.roundRobin {
		this.nextAlloc[CIDRNetMaskSize(cidr)] = IPAddInt(cidr.IP, CIDRHostSize(cidr))
	}
}

func (this *IPAM) Alloc(reqsize int) *net.IPNet {
	var found *Block

	if reqsize < 0 || reqsize > this.Bits() {
		return nil
	}
	next := this.getNext(reqsize)

	for found == nil {
		for b := this.block; b != nil; b = b.next {
			s := b.Size()
			if next != nil {
				if IPCmp(CIDRLastIP(b.cidr), next) < 0 {
					continue
				}
			}

			if b.canAlloc(next, reqsize) && (len(this.deletePending) == 0 || this.IsCoveredCIDR(b.cidr)) {
				if found == nil || s > found.Size() {
					found = b
					if found.matchSize(reqsize) {
						break
					}
				}
			}
		}
		if next == nil || found != nil {
			break
		}
		next = nil
		this.nextAlloc[reqsize] = nil
	}
	if found == nil {
		return nil
	}
	found = this.split(found, reqsize)

	cidr := found.alloc(next, reqsize)
	if cidr != nil {
		this.setNext(cidr)
		this.join(found)
	}
	return cidr
}

func (this *IPAM) split(b *Block, reqsize int) *Block {
	next := this.nextAlloc[reqsize]
	for b.Size() < reqsize && b.canSplit() {
		b.split()
		if next != nil {
			if IPCmp(b.next.cidr.IP, next) <= 0 {
				b = b.next
			}
		}
	}
	return b
}

func (this *IPAM) join(b *Block) {
	for b != nil {
		if len(this.deletePending) != 0 {
			if CIDRHostMaskSize(b.cidr) <= MAX_BITMAP_NET {
				// remember check block area: [b.prev.next...b.next]
				// removing might create multiple splitted blocks in this area
				p := &this.block
				if b.prev != nil {
					p = &b.prev.next
				}
				n := b.next
				i := 0
			nextPending:
				for i < len(this.deletePending) {
					// always check complete splitted area
					b := *p
					for b != nil && b != n {
						_, deleted := this.deletePart(b, this.deletePending[i])
						if deleted {
							this.deletePending.DeleteIndex(i)
							continue nextPending
						}
						b = b.next
					}
					i++
				}
			} else {
				if !b.isBusy() {
					for i, d := range this.deletePending {
						if CIDREqual(d, b.cidr) {
							this.removeBlock(b)
							this.deletePending.DeleteIndex(i)
							return
						}
					}
				}
			}
		}
		b = b.join()
	}
}

func (this *IPAM) removeBlock(b *Block) {
	if b.prev == nil {
		this.block = b.next
	} else {
		b.prev.next = b.next
	}
	if b.next != nil {
		b.next.prev = b.prev
	}
}

func (this *IPAM) insertBlock(previous, b *Block) {
	b.prev = previous
	if previous == nil {
		b.next = this.block
		if this.block != nil {
			this.block.prev = b
		}
		this.block = b
	} else {
		b.next = previous.next
		if previous.next != nil {
			previous.next.prev = b
		}
		previous.next = b
	}
	this.join(b)
}

func (this *IPAM) Busy(cidr *net.IPNet) bool {
	cidr = CIDRAlign(cidr, this.Bits())
	if cidr == nil {
		return false
	}
	if len(this.deletePending) != 0 && !this.IsCoveredCIDR(cidr) {
		return false
	}
	return this.set(cidr, true)
}

func (this *IPAM) Free(cidr *net.IPNet) bool {
	cidr = CIDRAlign(cidr, this.Bits())
	if cidr == nil {
		return false
	}
	return this.set(cidr, false)
}

func (this *IPAM) set(cidr *net.IPNet, busy bool) bool {
	reqsize, _ := cidr.Mask.Size()
	b := this.block
	for b != nil && !b.cidr.Contains(cidr.IP) {
		b = b.next
	}
	if b == nil {
		return false
	}

	size := b.Size()
	if b.canSplit() {
		if b.isBusy() == busy {
			return false
		}
		for size < reqsize && b.canSplit() {
			upper := b.split()
			if upper.cidr.Contains(cidr.IP) {
				b = upper
			}
			size++
		}
	}

	if size > reqsize {
		return false
	}

	if !b.set(cidr, busy) {
		return false
	}
	this.join(b)
	return true
}
