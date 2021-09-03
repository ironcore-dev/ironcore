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
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"net"
	"sync"
	"time"

	common "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	api "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
	"github.com/onmetal/onmetal-api/pkg/ipam"
	"github.com/onmetal/onmetal-api/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	SuccessfulAllocationMessage = "allocation successful"
	SuccessfulUsageMessage      = "used as specified"
)

type IPAM struct {
	lock           sync.Mutex
	lockCount      int
	object         *api.IPAMRange
	objectId       utils.ObjectId
	ipam           *ipam.IPAM
	error          string
	deleted        bool
	pendingRequest *PendingRequest
	lastUsage      time.Time
	requestSpecs   RequestSpecList
	allocations    AllocationStatusList
}

type PendingRequest struct {
	key   client.ObjectKey
	CIDRs AllocationList
}

func newIPAM(log logr.Logger, obj *api.IPAMRange) *IPAM {
	result := &IPAM{
		objectId: utils.NewObjectId(obj),
	}
	result.updateFrom(log, obj)
	return result
}

func (i *IPAM) Alloc(ctx context.Context, log logr.Logger, clt client.Client, request *IPAM) (AllocationStatusList, error) {
	var allocated AllocationStatusList
	reqSpecs := request.requestSpecs.PendingSpecs(request.allocations)
	log.Info("allocating", "reqSpecs", reqSpecs)
	for _, c := range reqSpecs {
		if !c.IsValid() {
			continue
		}
		cidr, err := c.Spec.Alloc(i.ipam)
		if cidr == nil || err != nil {
			state := api.AllocationStateBusy
			if cidr != nil {
				state = api.AllocationStateFailed
			} else {
				err = fmt.Errorf("allocation %s not possible in given range", c.Request)
			}
			c.Error = err.Error()
			allocated = append(allocated, &AllocationStatus{
				Allocation: Allocation{
					Request: c.Request,
				},
				Status:  state,
				Message: c.Error,
			})
		} else {
			allocated = append(allocated, &AllocationStatus{
				Allocation: Allocation{
					Request: c.Request,
					CIDR:    cidr,
				},
				Status:  api.AllocationStateAllocated,
				Message: SuccessfulAllocationMessage,
			})
		}
	}
	if len(allocated) != 0 {
		log.Info("allocated", "allocated", allocated)
		if err := i.updateRange(ctx, clt, append(request.allocations, allocated...), request.objectId.ObjectKey); err != nil {
			log.Info("range update failed (allocation reverted)", "error", err)
			for _, a := range allocated {
				if a.IsValid() {
					i.ipam.Free(a.CIDR)
				}
			}
			return nil, err
		}
	}
	log.Info("allocated in range", "allocated", allocated)
	return allocated, nil
}

func (i *IPAM) Free(ctx context.Context, log logr.Logger, ctl client.Client, request *IPAM) error {
	log.Info("releasing", "allocations", request.allocations)
	allocations := request.allocations.Allocations()
	if len(allocations) != 0 {
		for _, a := range allocations {
			i.ipam.Free(a.CIDR)
		}
		if err := i.updateRange(ctx, ctl, nil, request.objectId.ObjectKey); err != nil {
			for _, a := range allocations {
				i.ipam.Busy(a.CIDR)
			}
			log.Info("range update failed (free reverted)", "error", err)
			return err
		}
	}
	log.Info("released in range", "allocation", allocations)
	return nil
}

func (i *IPAM) setIPAMState(newIpr *api.IPAMRange) {
	blocks, round := i.ipam.State()
	var state []string
	for i := 0; i < len(round); i++ {
		state = append(state, fmt.Sprintf("%s/%d", round[i], i))
	}
	newIpr.Status.RoundRobinState = state
	newIpr.Status.AllocationState = blocks
}

func (i *IPAM) updateRange(ctx context.Context, clt client.Client, allocated AllocationStatusList, requestKey client.ObjectKey) error {
	newIpr := i.object.DeepCopy()
	i.setIPAMState(newIpr)

	newIpr.Status.PendingRequest = &api.IPAMPendingRequest{
		Name:      requestKey.Name,
		Namespace: requestKey.Namespace,
		CIDRs:     nil,
	}

	var internalPending AllocationList
	for _, a := range allocated {
		if !a.IsValid() {
			continue
		}
		internalPending.Add(&Allocation{
			Request: a.Request,
			CIDR:    a.CIDR,
		})
		newIpr.Status.PendingRequest.CIDRs = append(newIpr.Status.PendingRequest.CIDRs, api.CIDRAllocation{
			Request: a.Request,
			CIDR:    a.CIDR.String(),
		})
	}
	err := clt.Status().Patch(ctx, newIpr, client.MergeFrom(i.object))
	if err == nil {
		i.object = newIpr
		i.pendingRequest = &PendingRequest{
			key:   requestKey,
			CIDRs: internalPending,
		}
	}
	return err
}

// determineState
// at least one ready -> state = ready
// all busy -> state = busy
// none ready + at least one failed -> state = failed
func (i *IPAM) determineState() *api.IPAMRange {
	newObj := i.object.DeepCopy()
	state := common.StatePending
	msg := ""
	for _, a := range i.object.Status.CIDRs {
		switch a.Status {
		case api.AllocationStateAllocated:
			state = common.StateReady
		case api.AllocationStateBusy:
			msg = fmt.Sprintf("%s, %s: %s", msg, a.Request, a.Message)
			if state == "" {
				state = common.StateBusy
			}
		case api.AllocationStateFailed:
			if state != common.StateReady {
				state = common.StateError
			}
			msg = fmt.Sprintf("%s, %s: %s", msg, a.Request, a.Message)
		}
	}
	if msg != "" {
		msg = msg[2:]
	}
	if len(i.object.Status.CIDRs) != len(i.object.Spec.CIDRs) {
		state = common.StatePending
	}
	if msg == "" {
		switch state {
		case common.StatePending:
			msg = "request is pending"
		case common.StateReady:
			msg = "request is ready for allocation"
		}
	}
	newObj.Status.State = state
	newObj.Status.Message = msg
	return newObj
}

func (i *IPAM) updateFrom(log logr.Logger, obj *api.IPAMRange) {
	found := true
	var allocations AllocationStatusList
	var err error
	for _, a := range obj.Status.CIDRs {
		allocation := ParseAllocationStatus(&a)
		allocations = append(allocations, allocation)
		if allocation.CIDR != nil && ipam.CIDRHostSize(allocation.CIDR).Cmp(ipam.Int64(2)) < 0 {
			found = false
			break
		}
	}
	ranges := allocations.AsRanges()
	roundRobin := false
	switch obj.Spec.Mode {
	case "", api.ModeFirstMatch:
		roundRobin = false
	case api.ModeRoundRobin:
		roundRobin = true
	default:
		err = fmt.Errorf("invalid mode %q: use %s or %s", obj.Spec.Mode, api.ModeFirstMatch, api.ModeRoundRobin)
	}

	var ipr *ipam.IPAM
	if err == nil && found {
		log.Info("ranges found", "ranges", ranges)
		if len(ranges) > 0 {
			ipr, err = ipam.NewIPAMForRanges(ranges)
		}
	}

	var pending *PendingRequest
	if obj.Status.PendingRequest != nil && obj.Status.PendingRequest.Name != "" {
		pending = &PendingRequest{
			key: client.ObjectKey{
				Namespace: obj.Status.PendingRequest.Namespace,
				Name:      obj.Status.PendingRequest.Name,
			},
		}
		for _, c := range obj.Status.PendingRequest.CIDRs {
			cidr, err := ParseAllocation(&c)
			if err != nil {
				continue
			}
			pending.CIDRs = append(pending.CIDRs, cidr)
		}
	}

	if ipr != nil {
		var roundRobinIPs []net.IP
		for _, s := range obj.Status.RoundRobinState {
			_, cidr, err := net.ParseCIDR(s)
			if err != nil {
				continue
			}
			ones, _ := cidr.Mask.Size()
			for len(roundRobinIPs) <= ones {
				roundRobinIPs = append(roundRobinIPs, nil)
			}
			roundRobinIPs[ones] = cidr.IP
		}

		if err == nil {
			ipr.SetRoundRobin(roundRobin)
			_, err = ipr.SetState(obj.Status.AllocationState, roundRobinIPs)
		}
	}

	var specs RequestSpecList
	for _, c := range obj.Spec.CIDRs {
		spec := ParseRequestSpec(c)
		specs = append(specs, spec)
	}
	if len(specs.InValidSpecs()) != 0 && len(specs.ValidSpecs()) == 0 {
		err = specs.Error()
	}
	i.requestSpecs = specs
	i.object = obj
	i.ipam = ipr
	i.deleted = false
	i.pendingRequest = pending
	i.allocations = allocations

	if err != nil {
		i.error = err.Error()
	}
}
