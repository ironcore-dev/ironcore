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
	"github.com/onmetal/onmetal-api/apis/network/v1alpha1"
	"github.com/onmetal/onmetal-api/pkg/ipam"
	"github.com/onmetal/onmetal-api/pkg/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	"net"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
	"time"
)

type IPAM struct {
	lock           sync.Mutex
	lockCount      int
	object         *v1alpha1.IPAMRange
	objectId       utils.ObjectId
	ipam           *ipam.IPAM
	error          string
	deleted        bool
	pendingRequest *PendingRequest
	lastUsage      time.Time
	requestSpecs   ipam.RequestSpecList
}

type IPAMCache struct {
	client.Client
	lock         sync.Mutex
	ipams        map[client.ObjectKey]*IPAM
	pendingIpams map[client.ObjectKey]*IPAM
	lockedIpams  map[client.ObjectKey]*IPAM
}

type PendingRequest struct {
	key   client.ObjectKey
	CIDRs ipam.CIDRList
}

func NewIPAMCache(clt client.Client) *IPAMCache {
	return &IPAMCache{
		Client:       clt,
		ipams:        map[client.ObjectKey]*IPAM{},
		pendingIpams: map[client.ObjectKey]*IPAM{},
		lockedIpams:  map[client.ObjectKey]*IPAM{},
	}
}

func (i *IPAMCache) release(key client.ObjectKey) {
	i.lock.Lock()
	defer i.lock.Unlock()
	ipr := i.lockedIpams[key]
	if ipr != nil {
		if ipr.lockCount == 0 {
			panic("corrupted ipam cache locks")
		}
		ipr.lockCount--
		if ipr.lockCount == 0 {
			delete(i.lockedIpams, key)
			if !ipr.deleted {
				if ipr.pendingRequest != nil {
					i.pendingIpams[key] = ipr
				} else {
					i.ipams[key] = ipr
				}
				ipr.lastUsage = time.Now()
			}
		}
		ipr.lock.Unlock()
	} else {
		panic("corrupted ipam cache locks")
	}
	if len(i.ipams) > 100 { // TODO: config parameter for max cache size
		var found *IPAM
		var objectKey client.ObjectKey
		for k, i := range i.ipams {
			if found == nil || i.lastUsage.Before(found.lastUsage) {
				found = i
				objectKey = k
			}
		}
		delete(i.ipams, objectKey)
	}
}

func (i *IPAMCache) getRange(ctx context.Context, name client.ObjectKey, obj *v1alpha1.IPAMRange) (*IPAM, error) {
	i.lock.Lock()
	ipr := i.lockedIpams[name]
	if ipr == nil {
		ipr = i.ipams[name]
		if ipr == nil {
			ipr = i.pendingIpams[name]
		}
		if ipr == nil {
			if obj == nil {
				var tempObj v1alpha1.IPAMRange
				if err := i.Client.Get(ctx, name, &tempObj); err != nil {
					i.lock.Unlock()
					if errors.IsNotFound(err) {
						return nil, nil
					}
					return nil, err
				}
				obj = &tempObj
			}
		}
		delete(i.ipams, name)
		delete(i.pendingIpams, name)
	}
	if obj != nil {
		if ipr == nil || ipr.ipam == nil {
			ipr = newIPAM(obj)
		} else {
			ipr.object = obj
		}
	}
	ipr.lockCount++
	i.lockedIpams[name] = ipr
	i.lock.Unlock()
	ipr.lock.Lock()
	return ipr, nil
}

func (i *IPAMCache) removeRange(key client.ObjectKey) {
	i.lock.Lock()
	defer i.lock.Unlock()
	old := i.lockedIpams[key]
	if old != nil {
		old.deleted = true
	}
	delete(i.pendingIpams, key)
	delete(i.ipams, key)
	delete(i.lockedIpams, key)
}

func newIPAM(obj *v1alpha1.IPAMRange) *IPAM {
	ranges, err := ipam.ParseIPRanges(obj.Status.CIDRs...)

	roundRobin := false
	if err == nil {
		switch obj.Spec.Mode {
		case "", v1alpha1.ModeFirstMatch:
			roundRobin = false
		case v1alpha1.ModeRoundRobin:
			roundRobin = true
		default:
			err = fmt.Errorf("invalid mode %q: use %s or %s", obj.Spec.Mode, v1alpha1.ModeFirstMatch, v1alpha1.ModeRoundRobin)
		}
	}

	var ipr *ipam.IPAM
	if err == nil {
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
			_, cidr, err := net.ParseCIDR(c)
			if err == nil {
				pending.CIDRs = append(pending.CIDRs, cidr)
			}
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
			err = ipr.SetState(obj.Status.AllocationState, roundRobinIPs)
		}
	}

	var specs []ipam.RequestSpec
	for i, c := range obj.Spec.CIDRs {
		var spec ipam.RequestSpec
		spec, err = ipam.ParseRequestSpec(c)
		if err != nil {
			err = fmt.Errorf("invalid request spec %d for cidr %s: %s", i, c, err)
		} else {
			specs = append(specs, spec)
		}
	}
	result := &IPAM{
		requestSpecs:   specs,
		object:         obj,
		objectId:       utils.NewObjectId(obj),
		ipam:           ipr,
		deleted:        false,
		pendingRequest: pending,
	}
	if err != nil {
		result.error = err.Error()
	}
	return result
}

func (i *IPAM) Alloc(ctx context.Context, log *utils.Logger, clt client.Client, reqSpecs ipam.RequestSpecList, requestKey client.ObjectKey) (ipam.CIDRList, error) {
	var allocated ipam.CIDRList
	log.Infof("allocating %s", reqSpecs)
	if len(reqSpecs) > 0 {
		for _, c := range reqSpecs {
			cidr, err := c.Alloc(i.ipam)
			if cidr == nil || err != nil {
				for _, a := range allocated {
					i.ipam.Free(a)
				}
				allocated = nil
				break
			} else {
				allocated = append(allocated, cidr)
			}
		}
	}
	if len(allocated) != 0 {
		log.Infof("allocated %s", allocated)
		if err := i.updateRange(ctx, clt, allocated, requestKey); err != nil {
			for _, a := range allocated {
				i.ipam.Free(a)
			}
			log.Infof("range update failed: %s (allocation reverted)", err)
			return nil, err
		}
	}
	log.Infof("%s allocated in range", allocated)
	return allocated, nil
}

func (i *IPAM) Free(ctx context.Context, log *utils.Logger, ctl client.Client, allocated ipam.CIDRList, requestKey client.ObjectKey) error {
	log.Infof("releasing %s", allocated)
	if len(allocated) != 0 {
		for _, a := range allocated {
			i.ipam.Free(a)
		}
		if err := i.updateRange(ctx, ctl, nil, requestKey); err != nil {
			for _, a := range allocated {
				i.ipam.Busy(a)
			}
			log.Infof("range update failed: %s (free reverted)", err)
			return err
		}
	}
	log.Infof("%s released in range", allocated)
	return nil
}

func (i *IPAM) setIPAMState(newIpr *v1alpha1.IPAMRange) {
	blocks, round := i.ipam.State()
	var state []string
	for i := 0; i < len(round); i++ {
		state = append(state, fmt.Sprintf("%s/%d", round[i], i))
	}
	newIpr.Status.RoundRobinState = state
	newIpr.Status.AllocationState = blocks
}

func (i *IPAM) updateRange(ctx context.Context, clt client.Client, allocated []*net.IPNet, requestKey client.ObjectKey) error {
	newIpr := i.object.DeepCopy()
	i.setIPAMState(newIpr)

	newIpr.Status.PendingRequest = &v1alpha1.IPAMPendingRequest{
		Name:      requestKey.Name,
		Namespace: requestKey.Namespace,
		CIDRs:     nil,
	}

	for _, a := range allocated {
		newIpr.Status.PendingRequest.CIDRs = append(newIpr.Status.PendingRequest.CIDRs, a.String())
	}
	err := clt.Status().Patch(ctx, newIpr, client.MergeFrom(i.object))
	if err == nil {
		i.object = newIpr
		i.pendingRequest = &PendingRequest{
			key:   requestKey,
			CIDRs: allocated,
		}
	}
	return err
}
