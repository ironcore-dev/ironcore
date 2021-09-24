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
	"github.com/go-logr/logr"
	api "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PendingRequest struct {
	key       client.ObjectKey
	CIDRs     AllocationList
	Deletions AllocationList
}

func ParsePendingRequest(obj *api.IPAMRange) *PendingRequest {
	var pending *PendingRequest
	if obj.Status.PendingRequest != nil && obj.Status.PendingRequest.Name != "" {
		pending = &PendingRequest{
			key: client.ObjectKey{
				Namespace: obj.Status.PendingRequest.Namespace,
				Name:      obj.Status.PendingRequest.Name,
			},
		}
		pending.CIDRs = parseAllocationList(obj.Status.PendingRequest.CIDRs)
		pending.Deletions = parseAllocationList(obj.Status.PendingRequest.Deletions)
	}
	return pending
}

func (p *PendingRequest) MatchState(log logr.Logger, ipam *IPAM) bool {
	allocations := ipam.allocations.Allocated().Allocations()
	deletions := ipam.deletions.Allocations()
	log.Info("found status for pending request", "allocations", allocations, "deletions", deletions)
	log.Info("expected status for pending request", "allocations", p.CIDRs, "deletions", p.Deletions)
	if !reflect.DeepEqual(p.CIDRs, allocations) || !reflect.DeepEqual(p.Deletions, deletions) {
		log.Info("expected status not yet set in pending request")
		return false
	}
	log.Info("expected status reached")
	return true
}

func (p *PendingRequest) MatchRequest(obj *api.IPAMRange) bool {
	if p == nil {
		return false
	}
	return p.key.Namespace == obj.Namespace && p.key.Name == obj.Name
}

func parseAllocationList(allocation []api.CIDRAllocation) AllocationList {
	var list AllocationList
	for _, a := range allocation {
		cidr, err := ParseAllocation(&a)
		if err != nil {
			continue
		}
		list = append(list, cidr)
	}
	return list
}
