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
	ipam           *ipam.IPAM
	error          string
	deleted        bool
	pendingRequest *PendingRequest
	lastUsage      time.Time
	requestSpecs   []ipam.RequestSpec
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
			if ipr.pendingRequest != nil {
				i.pendingIpams[key] = ipr
			} else {
				i.ipams[key] = ipr
			}
			ipr.lastUsage = time.Now()
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
		tempIpr := newIPAM(obj)
		if ipr == nil {
			ipr = tempIpr
		} else {
			ipr.object = obj
			ipr.pendingRequest = tempIpr.pendingRequest
			ipr.ipam = tempIpr.ipam
		}
	}
	ipr.lockCount++
	i.lockedIpams[name] = ipr
	i.lock.Unlock()
	ipr.lock.Lock()
	return ipr, nil
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
		ipam:           ipr,
		deleted:        false,
		pendingRequest: pending,
	}
	if err != nil {
		result.error = err.Error()
	}
	return result
}
