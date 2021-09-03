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

package ipamrange

import (
	"fmt"
	api "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
	"github.com/onmetal/onmetal-api/pkg/ipam"
	"net"
)

type Allocation struct {
	Request string
	CIDR    *net.IPNet
}

func ParseAllocation(allocation *api.CIDRAllocation) (*Allocation, error) {
	var cidr *net.IPNet
	var err error
	if allocation.CIDR != "" {
		cidr, err = ipam.ParseCIDR(allocation.CIDR)
	}
	a := &Allocation{
		Request: allocation.Request,
		CIDR:    cidr,
	}
	return a, err
}

func (a *Allocation) String() string {
	return fmt.Sprintf("%s:%s", a.Request, a.CIDR)
}

type AllocationList []*Allocation

func (l AllocationList) AsRanges() ipam.IPRanges {
	var list ipam.IPRanges
	for _, c := range l {
		if c.CIDR != nil {
			list = append(list, &ipam.IPRange{
				Start: ipam.CIDRFirstIP(c.CIDR),
				End:   ipam.CIDRLastIP(c.CIDR),
			})
		}
	}
	return list
}

func (l *AllocationList) Add(allocation ...*Allocation) {
	(*l) = append(*l, allocation...)
}

func (l AllocationList) String() string {
	s := "["
	sep := ""
	for _, spec := range l {
		s = fmt.Sprintf("%s%s%s", s, sep, spec)
		sep = ","
	}
	return s + "]"
}

func (l AllocationList) Copy() AllocationList {
	list := make(AllocationList, len(l))
	copy(list, l)
	return list
}

func (l AllocationList) LookUp(request string) int {
	for i, a := range l {
		if a.Request == request {
			return i
		}
	}
	return -1
}
