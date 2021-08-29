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
	"strings"
)

type IPRange struct {
	Start net.IP
	End   net.IP
}

func (this *IPRange) String() string {
	if this.Start.Equal(this.End) {
		return this.Start.String()
	}
	return fmt.Sprintf("%s-%s", this.Start, this.End)
}

func (this *IPRange) Contains(ip net.IP) bool {
	if IPCmp(this.Start, ip) > 0 {
		return false
	}
	return IPCmp(this.End, ip) >= 0
}

func ParseIPRange(str string) (*IPRange, error) {
	parts := strings.Split(str, "-")
	if len(parts) <= 2 {
		start := ParseIP(strings.ToLower(parts[0]))
		if start != nil {
			if len(parts) == 1 {
				return &IPRange{
					Start: start,
					End:   start,
				}, nil
			}
			end := ParseIP(strings.ToLower(parts[1]))
			if end != nil {
				if len(start) != len(end) {
					return nil, fmt.Errorf("invalid ip range %q: start and end ip have different version", str)
				}
				if IPCmp(end, start) < 0 {
					return nil, fmt.Errorf("invalid ip range %q: end of range before start", str)
				}
				return &IPRange{
					Start: start,
					End:   end,
				}, nil
			}
		} else {
			if len(parts) == 1 {
				cidr, err := ParseCIDR(parts[0])
				if err == nil {
					return &IPRange{
						Start: CIDRFirstIP(cidr),
						End:   CIDRLastIP(cidr),
					}, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("invalid ip range %q", str)
}

func MustParseIPRange(str string) *IPRange {
	r, err := ParseIPRange(str)
	if err != nil {
		panic(err)
	}
	return r
}

func ParseIPRanges(str ...string) (IPRanges, error) {
	r := IPRanges{}
	for _, s := range str {
		e, err := ParseIPRange(s)
		if err != nil {
			return nil, err
		}
		r = append(r, e)
	}
	return r, nil
}

func MustParseIPRanges(str ...string) IPRanges {
	r, err := ParseIPRanges(str...)
	if err != nil {
		panic(err)
	}
	return r
}

// IPRanges attaches the methods of Interface to []string, sorting in increasing order.
type IPRanges []*IPRange

func (p IPRanges) Len() int           { return len(p) }
func (p IPRanges) Less(i, j int) bool { return IPCmp(p[i].Start, p[j].Start) < 0 }
func (p IPRanges) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (this IPRanges) String() string {
	if len(this) == 0 {
		return "[]"
	}
	s := ""
	for _, r := range this {
		s = s + ", " + r.String()
	}
	return "[" + s[2:] + "]"
}

func (this IPRanges) Contains(ip net.IP) bool {
	for _, r := range this {
		if r.Contains(ip) {
			return true
		}
	}
	return false
}

func NormalizeIPRanges(ranges ...*IPRange) IPRanges {
	sort.Sort(IPRanges(ranges))

	for i := 0; i < len(ranges)-1; i++ {
		if IPCmp(ranges[i].End, ranges[i+1].Start) >= 0 {
			if IPCmp(ranges[i].End, ranges[i+1].End) >= 0 {
				ranges = append(ranges[:i+1], ranges[i+2:]...)
			} else {
				ranges = append(append(ranges[:i], &IPRange{ranges[i].Start, ranges[i+1].End}), ranges[i+2:]...)
			}
			i--
		}
	}
	return ranges
}

////////////////////////////////////////////////////////////////////////////////

type CIDRList []*net.IPNet

func (p CIDRList) Len() int           { return len(p) }
func (p CIDRList) Less(i, j int) bool { return IPCmp(p[i].IP, p[j].IP) < 0 }
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

func (this CIDRList) Copy() CIDRList {
	var result CIDRList
	for _, c := range this {
		result = append(result, c)
	}
	return result
}

func addExcludes(excls *CIDRList, cidr *net.IPNet, lower bool, border net.IP) bool {
	found := false
	for {
		l, u := CIDRSplit(cidr)
		if l == nil {
			return found
		}
		if lower {
			u, l = l, u
		}
		lastL := CIDRSubIP(l, (1<<CIDRHostMaskSize(l))-1)
		lastU := CIDRSubIP(u, (1<<CIDRHostMaskSize(u))-1)
		_ = lastU
		if l.Contains(border) {
			excls.Add(u)
			found = true
			if (lower && border.Equal(l.IP)) || (!lower && border.Equal(lastL)) {
				return found
			}
			cidr = l
		} else {
			cidr = u
		}
	}
}

func splitExcludes(cidr *net.IPNet, lower, upper net.IP) (*net.IPNet, *net.IPNet) {
	for {
		l, u := CIDRSplit(cidr)
		if l == nil {
			return nil, nil
		}
		if l.Contains(lower) {
			if u.Contains(upper) {
				return l, u
			}
			cidr = u
		} else {
			cidr = l
		}
	}
}

func Excludes(cidr *net.IPNet, ranges ...*IPRange) (CIDRList, error) {
	ranges = NormalizeIPRanges(ranges...)
	if len(ranges) == 0 {
		return nil, nil
	}
	cidr = CIDRNet(cidr)
	if IPCmp(cidr.IP, ranges[0].Start) > 0 {
		return nil, fmt.Errorf("range %s below cidr %s", ranges[0], cidr)
	}
	last := CIDRLastIP(cidr)
	if IPCmp(last, ranges[len(ranges)-1].End) < 0 {
		return nil, fmt.Errorf("range %s above cidr %s", ranges[len(ranges)-1], cidr)
	}

	var excl CIDRList

	if len(ranges) > 0 {
		addExcludes(&excl, cidr, true, ranges[0].Start)
		for i := 0; i < len(ranges)-1; i++ {
			l, u := splitExcludes(cidr, ranges[i].End, ranges[i+1].Start)
			addExcludes(&excl, l, false, ranges[i].End)
			addExcludes(&excl, u, true, ranges[i+1].Start)
		}
		addExcludes(&excl, cidr, false, ranges[len(ranges)-1].End)
	}
	return excl, nil
}

/*
func x() {
	for len(incl)>0 {
		last:=incl[len(incl)-1]
		if CIDRNetMaskSize(last)!=CIDRNetMaskSize(cidr) {
			break
		}
		cand:=CIDRExtend(cidr)
		if cand.IP.Equal(last.IP) {
			cidr=cand
			incl=incl[:len(incl)-1]
		}
	}
}

*/

func addIncludes(incls *CIDRList, cidr *net.IPNet, lower bool, border net.IP) bool {
	found := false
	for {
		if (lower && border.Equal(CIDRLastIP(cidr))) || (!lower && border.Equal(cidr.IP)) {
			incls.Add(cidr)
			return true
		}
		l, u := CIDRSplit(cidr)
		if l == nil {
			return found
		}
		if lower {
			u, l = l, u
		}
		if l.Contains(border) {
			incls.Add(u)
			found = true
			cidr = l
		} else {
			cidr = u
		}
	}
}

func Includes(ranges ...*IPRange) (CIDRList, error) {
	ranges = NormalizeIPRanges(ranges...)
	if len(ranges) == 0 {
		return nil, nil
	}

	incl := CIDRList{}

	if len(ranges) > 0 {
		for _, r := range ranges {
			cidr := IPtoCIDR(r.Start)

			for !cidr.Contains(r.End) {
				cidr = CIDRExtend(cidr)
			}
			l, u := CIDRSplit(cidr)
			addIncludes(&incl, l, false, r.Start)
			addIncludes(&incl, u, true, r.End)
		}

		if len(incl) > 1 {
			sort.Sort(incl)
			new := CIDRList{incl[0]}
			for i := 1; i < len(incl); i++ {
				cidr := incl[i]
				for len(new) > 0 {
					last := new[len(new)-1]
					if CIDRNetMaskSize(last) != CIDRNetMaskSize(cidr) {
						break
					}
					if cidr.IP.Equal(last.IP) {
						new = new[:len(new)-1]
					} else {
						cand := CIDRExtend(cidr)
						if cand.IP.Equal(last.IP) {
							cidr = cand
							new = new[:len(new)-1]
						} else {
							break
						}
					}
				}
				new.Add(cidr)
			}
			return new, nil
		}
	}
	return incl, nil
}
