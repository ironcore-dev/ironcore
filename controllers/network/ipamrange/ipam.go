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

	"github.com/mandelsoft/kubipam/pkg/ipam"
	common "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	api "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
	"github.com/onmetal/onmetal-api/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	SuccessfulAllocationMessage = "allocation successful"
	SuccessfulUsageMessage      = "used as specified"
	failBusyAllocationMessage   = "allocation %s not possible in given range"
)

func FailBusyAllocationMessage(allocation string) string {
	return fmt.Sprintf(failBusyAllocationMessage, allocation)
}

type IPAM struct {
	lock      sync.Mutex
	lockCount int
	object    *api.IPAMRange
	objectId  utils.ObjectId
	ipam      *ipam.IPAM
	error     string
	deleted   bool
	lastUsage time.Time
	// two phase commit for range allocation
	pendingRequest *PendingRequest
	// requested ranges
	requestSpecs RequestSpecList
	// currently assigned ranges
	allocations AllocationStatusList
	// inflight deletions:
	// allocations of deleted requests which are not yet released in IPAM
	deletions AllocationStatusList
}

func newIPAM(log logr.Logger, obj *api.IPAMRange) *IPAM {
	result := &IPAM{
		objectId: utils.NewObjectId(obj),
	}
	result.updateFrom(log, obj)
	return result
}

func (i *IPAM) Alloc(ctx context.Context, log logr.Logger, clt client.Client, request *IPAM, reqSpecs RequestSpecList) (AllocationStatusList, error) {
	if len(reqSpecs) == 0 {
		log.Info("no outstanding allocations pending")
		return nil, nil
	}
	var allocated AllocationStatusList
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
				err = fmt.Errorf(failBusyAllocationMessage, c.Request)
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
		if err := i.updateRange(ctx, clt, append(request.allocations, allocated...), request.deletions, request.objectId.ObjectKey); err != nil {
			log.Info("range update failed (allocation reverted)", "error", err)
			for _, a := range allocated {
				if a.IsValid() {
					i.ipam.Free(a.CIDR)
				}
			}
			return nil, err
		}
		log.Info("allocated in range", "allocated", allocated)
	} else {
		log.Info("no new allocations")
	}
	return allocated, nil
}

// HandleRelease
// - find reqSpecs in allocations
// - move from allocations to deletions
// - free found reqSpecs
func (i *IPAM) HandleRelease(ctx context.Context, log logr.Logger, clt client.Client, request *IPAM, deleteList AllocationStatusList) error {
	if len(deleteList) == 0 {
		log.Info("no outstanding deletions pending")
		return nil
	}
	log.Info("releasing", "deleteList", deleteList)
	var cidrList ipam.CIDRList
	var deletions AllocationStatusList
	allocations := i.allocations.Copy()
	for _, d := range deleteList {
		for index := 0; index < len(allocations); index++ {
			if allocations[index] == d {
				allocations.RemoveIndex(index)
				index--
				break
			}
		}
		if !d.IsValid() {
			continue
		}
		deletions = append(deletions, d)
		cidrList = append(cidrList, d.CIDR)
	}

	if len(cidrList) != 0 {
		newObj := i.object.DeepCopy()
		newObj.Status.PendingDeletions = deletions.AsCIDRAllocationStatusList()
		newObj.Status.CIDRs = allocations.AsCIDRAllocationStatusList()
		if err := clt.Status().Patch(ctx, newObj, client.MergeFrom(i.object)); err != nil {
			return err
		}
		i.ipam.DeleteCIDRs(cidrList)
		i.deletions = deletions
		i.allocations = allocations
		log.Info("released from ipam", "deletions", cidrList)
	} else {
		log.Info("no new deletions")
	}
	return nil
}

func (i *IPAM) Free(ctx context.Context, log logr.Logger, ctl client.Client, request *IPAM) (AllocationStatusList, AllocationStatusList, error) {
	if len(request.deletions) == 0 {
		return nil, request.deletions, nil
	}
	pending := request.ipam.PendingDeleted()
	var deleted AllocationStatusList
	deletions := request.deletions.Copy()
outer:
	for index := 0; index < len(deletions); index++ {
		for _, p := range pending {
			if ipam.CIDROverlap(p, deletions[index].CIDR) {
				continue outer
			}
		}
		deleted.Add(request.deletions[index])
		deletions.RemoveIndex(index)
		index--
	}
	if len(deleted) != 0 {
		for _, d := range deleted {
			i.ipam.Free(d.CIDR)
		}
		if err := i.updateRange(ctx, ctl, request.allocations, deletions, request.objectId.ObjectKey); err != nil {
			for _, d := range deleted {
				i.ipam.Busy(d.CIDR)
			}
			log.Info("range update failed (free reverted)", "error", err)
			return request.deletions, deleted, err
		}
		return deletions, deleted, nil
	}
	return request.deletions, deleted, nil
}

func (i *IPAM) FreeAll(ctx context.Context, log logr.Logger, ctl client.Client, request *IPAM) error {
	log.Info("releasing", "allocations", request.allocations)
	allocations := append(request.allocations.Allocations(), request.deletions.Allocations()...)
	if len(allocations) != 0 {
		for _, a := range allocations {
			i.ipam.Free(a.CIDR)
		}
		if err := i.updateRange(ctx, ctl, nil, nil, request.objectId.ObjectKey); err != nil {
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

func (i *IPAM) updateRange(ctx context.Context, clt client.Client, allocated, pending AllocationStatusList, requestKey client.ObjectKey) error {
	newIpr := i.object.DeepCopy()
	i.setIPAMState(newIpr)

	newIpr.Status.PendingRequest = &api.IPAMPendingRequest{
		Name:      requestKey.Name,
		Namespace: requestKey.Namespace,
	}

	var internalAllocations AllocationList
	var internalDeletions AllocationList
	for _, a := range allocated {
		if !a.IsValid() {
			continue
		}
		internalAllocations.Add(&Allocation{
			Request: a.Request,
			CIDR:    a.CIDR,
		})
		newIpr.Status.PendingRequest.CIDRs = append(newIpr.Status.PendingRequest.CIDRs, api.CIDRAllocation{
			Request: a.Request,
			CIDR:    a.CIDR.String(),
		})

	}
	for _, a := range pending {
		if !a.IsValid() {
			continue
		}
		internalDeletions.Add(&Allocation{
			Request: a.Request,
			CIDR:    a.CIDR,
		})
		newIpr.Status.PendingRequest.Deletions = append(newIpr.Status.PendingRequest.Deletions, api.CIDRAllocation{
			Request: a.Request,
			CIDR:    a.CIDR.String(),
		})

	}
	err := clt.Status().Patch(ctx, newIpr, client.MergeFrom(i.object))
	if err == nil {
		i.object = newIpr
		i.pendingRequest = &PendingRequest{
			key:       requestKey,
			CIDRs:     internalAllocations,
			Deletions: internalDeletions,
		}
	}
	return err
}

// determineState
// at least one ready -> state = ready
// all busy -> state = busy
// none ready + at least one failed -> state = failed
func (i *IPAM) determineState(allocations, deletions AllocationStatusList) *api.IPAMRange {
	newObj := i.object.DeepCopy()
	newObj.Status.CIDRs = allocations.AsCIDRAllocationStatusList()
	newObj.Status.PendingDeletions = deletions.AsCIDRAllocationStatusList()
	state := ""
	msg := ""
	for _, a := range allocations {
		switch a.Status {
		case api.AllocationStateAllocated:
			state = common.StateReady
			msg = "request is ready for allocation"
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
	if state == "" {
		if len(allocations) == 0 {
			state = common.StateReady
			msg = "empty request"
		} else {
			state = common.StatePending
			msg = "request is pending"
		}
	}
	if msg != "" {
		msg = msg[2:]
	}
	newObj.Status.State = state
	newObj.Status.Message = msg
	return newObj
}

func (i *IPAM) updateFrom(log logr.Logger, obj *api.IPAMRange) {
	var err error
	allocations, found := parseAllocations(obj.Status.CIDRs)
	deletions, _ := parseAllocations(obj.Status.PendingDeletions)
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
		ranges := allocations.AsRanges()
		log.Info("ranges found", "ranges", ranges)
		if len(ranges) > 0 {
			ipr, err = ipam.NewIPAMForRanges(ranges)
		}
	}

	pending := ParsePendingRequest(obj)

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
	i.deletions = deletions
	if err != nil {
		i.error = err.Error()
	}
}

func parseAllocations(list []api.CIDRAllocationStatus) (AllocationStatusList, bool) {
	var allocations AllocationStatusList
	found := false
	for _, a := range list {
		allocation := ParseAllocationStatus(&a)
		allocations = append(allocations, allocation)
		if allocation.CIDR != nil {
			found = true
		}
	}
	return allocations, found
}

func (i *IPAM) updateSpecFrom(obj *api.IPAMRange) {
	var specs RequestSpecList
	var err error
	for _, c := range obj.Spec.CIDRs {
		spec := ParseRequestSpec(c)
		specs = append(specs, spec)
	}
	if len(specs.InValidSpecs()) != 0 && len(specs.ValidSpecs()) == 0 {
		err = specs.Error()
	}
	i.requestSpecs = specs
	if err != nil {
		i.error = err.Error()
	}
}

func (i *IPAM) updateAllocations(allocations, deletions AllocationStatusList) {
	if i.ipam == nil {
		ranges := allocations.AsRanges()
		if len(ranges) > 0 {
			i.ipam, _ = ipam.NewIPAMForRanges(ranges)
		}
	} else {
		i.ipam.AddCIDRs(allocations.AsCIDRList())
	}
	i.allocations = allocations
	i.deletions = deletions
}
