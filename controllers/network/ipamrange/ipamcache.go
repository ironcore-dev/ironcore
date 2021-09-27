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
	"github.com/go-logr/logr"
	networkv1alpha1 "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
	"time"
)

type IPAMCache struct {
	client.Client
	lock         sync.Mutex
	ipams        map[client.ObjectKey]*IPAM
	pendingIpams map[client.ObjectKey]*IPAM
	lockedIpams  map[client.ObjectKey]*IPAM
}

func NewIPAMCache(clt client.Client) *IPAMCache {
	return &IPAMCache{
		Client:       clt,
		ipams:        map[client.ObjectKey]*IPAM{},
		pendingIpams: map[client.ObjectKey]*IPAM{},
		lockedIpams:  map[client.ObjectKey]*IPAM{},
	}
}

func (i *IPAMCache) release(log logr.Logger, key client.ObjectKey) {
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
		log.V(2).Info("unlocking", "name", key)
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

func (i *IPAMCache) getRange(ctx context.Context, log logr.Logger, name client.ObjectKey, obj *networkv1alpha1.IPAMRange) (*IPAM, error) {
	i.lock.Lock()
	ipr := i.lockedIpams[name]
	if ipr == nil {
		ipr = i.ipams[name]
		if ipr == nil {
			ipr = i.pendingIpams[name]
		}
		if ipr == nil {
			if obj == nil {
				var tempObj networkv1alpha1.IPAMRange
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
			ipr = newIPAM(log, obj)
		} else {
			ipr.object = obj
			ipr.updateSpecFrom(obj)
			if obj.Status.State != networkv1alpha1.IPAMRangeReady {
				ipr.error = obj.Status.Message
			} else {
				ipr.error = ""
			}
		}
	}
	ipr.lockCount++
	i.lockedIpams[name] = ipr
	i.lock.Unlock()
	log.V(2).Info("locking", "name", name)
	ipr.lock.Lock()
	return ipr, nil
}
