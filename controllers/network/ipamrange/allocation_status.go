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

type AllocationStatus struct {
	Allocation
	Status  string
	Message string
}

func ParseAllocationStatus(allocation *api.CIDRAllocationStatus) *AllocationStatus {
	a, err := ParseAllocation(&allocation.CIDRAllocation)
	if err != nil {
		return &AllocationStatus{
			Allocation: *a,
			Status:     api.AllocationStateFailed,
			Message:    "invalid cidr",
		}
	}
	return &AllocationStatus{
		Allocation: *a,

		Status:  allocation.Status,
		Message: allocation.Message,
	}
}

func (a *AllocationStatus) IsValid() bool {
	return a.CIDR != nil && a.Status == api.AllocationStateAllocated
}

func (a *AllocationStatus) String() string {
	return fmt.Sprintf("%s:%s", a.Request, a.CIDR)
}

type AllocationStatusList []*AllocationStatus

func NewAllocationStatusListFromAllocations(allocations AllocationList) AllocationStatusList {
	status := make(AllocationStatusList, len(allocations))
	for i, a := range allocations {
		status[i] = &AllocationStatus{
			Allocation: *a,
			Status:     api.AllocationStateAllocated,
			Message:    SuccessfulAllocationMessage,
		}
	}
	return status
}

func (l *AllocationStatusList) Add(allocation ...*AllocationStatus) {
	(*l) = append(*l, allocation...)
}

func (l AllocationStatusList) HasBusy() bool {
	for _, a := range l {
		if a.Status == api.AllocationStateBusy {
			return true
		}
	}
	return false
}

func (l AllocationStatusList) AsRanges() ipam.IPRanges {
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

func (l AllocationStatusList) Match(list AllocationList) bool {
	notFound := append(list[:0:0], list...)
	for _, e := range l {
		for i, o := range notFound {
			if o.Request == e.Request {
				notFound = append(notFound[:i], notFound[i+1:]...)
				break
			}
		}
	}
	return len(notFound) == 0
}

func (l AllocationStatusList) String() string {
	s := "["
	sep := ""
	for _, spec := range l {
		s = fmt.Sprintf("%s%s%s", s, sep, spec)
		sep = ","
	}
	return s + "]"
}

func (l AllocationStatusList) Copy() AllocationStatusList {
	list := make(AllocationStatusList, len(l))
	copy(list, l)
	return list
}

func (l AllocationStatusList) LookUp(raw string) int {
	for i, a := range l {
		if a.Request == raw {
			return i
		}
	}
	return -1
}

func (l AllocationStatusList) Allocations() AllocationList {
	var list AllocationList
	for _, a := range l {
		if a.IsValid() {
			list.Add(&a.Allocation)
		}
	}
	return list
}

func OptinalCIDRToString(cidr *net.IPNet) string {
	if cidr == nil {
		return ""
	}
	return cidr.String()
}

func (l AllocationStatusList) GetAllocationStatusList() []api.CIDRAllocationStatus {
	status := make([]api.CIDRAllocationStatus, len(l))
	for i, a := range l {
		status[i] = api.CIDRAllocationStatus{
			CIDRAllocation: api.CIDRAllocation{
				Request: a.Request,
				CIDR:    OptinalCIDRToString(a.CIDR),
			},
			Status:  a.Status,
			Message: a.Message,
		}
	}
	return status
}
