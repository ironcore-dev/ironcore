// Copyright 2022 OnMetal authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vleg

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/go-logr/logr"
	ori "github.com/onmetal/onmetal-api/ori/apis/storage/v1alpha1"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrl "sigs.k8s.io/controller-runtime"
)

type genericVolumeMetadata struct {
	namespace string
	name      string
	uid       string
}

type genericVolume struct {
	id       string
	metadata genericVolumeMetadata
	state    ori.VolumeState
}

func (g genericVolume) key() string { //nolint:unused
	return g.id
}

type Generic struct {
	// relistPeriod is the period for relisting.
	relistPeriod time.Duration
	// relistThreshold is the maximum threshold between two relists to become unhealthy.
	relistThreshold time.Duration
	// relistTime is the last time a relist happened.
	relistTime atomic.Pointer[time.Time]
	// runtime is the volume runtime.
	runtime ori.VolumeRuntimeClient
	// eventChannel is the channel from which the subscriber receives events.
	eventChannel chan *VolumeLifecycleEvent

	volumes oldNewMap[string, genericVolume]
}

type GenericOptions struct {
	ChannelCapacity int
	RelistPeriod    time.Duration
	RelistThreshold time.Duration
}

func setGenericMLEGOptionsDefaults(o *GenericOptions) {
	if o.ChannelCapacity <= 0 {
		o.ChannelCapacity = 1024
	}
	if o.RelistPeriod <= 0 {
		o.RelistPeriod = 1 * time.Second
	}
	if o.RelistThreshold <= 0 {
		o.RelistThreshold = 3 * time.Minute
	}
}

func NewGeneric(runtime ori.VolumeRuntimeClient, opts GenericOptions) VolumeLifecycleEventGenerator {
	setGenericMLEGOptionsDefaults(&opts)

	return &Generic{
		relistPeriod:    opts.RelistPeriod,
		relistThreshold: opts.RelistThreshold,
		runtime:         runtime,
		eventChannel:    make(chan *VolumeLifecycleEvent, opts.ChannelCapacity),
		volumes:         make(oldNewMap[string, genericVolume]),
	}
}

func (g *Generic) list(ctx context.Context) ([]*genericVolume, error) {
	listVolumesRes, err := g.runtime.ListVolumes(ctx, &ori.ListVolumesRequest{})
	if err != nil {
		return nil, fmt.Errorf("error listing volumes: %w", err)
	}

	volumesByID := make(map[string]*genericVolume)
	for _, volume := range listVolumesRes.Volumes {
		volumesByID[volume.Id] = &genericVolume{
			id: volume.Id,
			metadata: genericVolumeMetadata{
				uid:       volume.Metadata.Uid,
				namespace: volume.Metadata.Namespace,
				name:      volume.Metadata.Name,
			},
			state: volume.State,
		}
	}

	res := make([]*genericVolume, 0, len(volumesByID))
	for _, volume := range volumesByID {
		res = append(res, volume)
	}
	return res, nil
}

func (g *Generic) Name() string {
	return "vleg"
}

func (g *Generic) Check(_ *http.Request) error {
	relistTime := g.relistTime.Load()
	if relistTime == nil {
		return fmt.Errorf("vleg did not relist yet")
	}

	elapsed := time.Since(*relistTime)
	if elapsed > g.relistThreshold {
		return fmt.Errorf("vleg was last seen active %v ago, threshold is %v", elapsed, g.relistThreshold)
	}
	return nil
}

func (g *Generic) Start(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx).WithName("vleg")
	wait.UntilWithContext(ctx, func(ctx context.Context) {
		if err := g.relist(ctx, log); err != nil {
			log.Error(err, "Error relisting")
		}
	}, g.relistPeriod)
	return nil
}

func oriVolumeStateToVolumeLifecycleEventType(state ori.VolumeState) VolumeLifecycleEventType {
	switch state {
	case ori.VolumeState_VOLUME_AVAILABLE:
		return VolumeAvailable
	case ori.VolumeState_VOLUME_ERROR:
		return VolumeError
	case ori.VolumeState_VOLUME_PENDING:
		return VolumePending
	default:
		panic(fmt.Sprintf("unrecognized volume state %d", state))
	}
}

func (g *Generic) inferEvents(id string, metadata VolumeLifecycleEventMetadata, oldVolume, newVolume *genericVolume) []*VolumeLifecycleEvent {
	switch {
	case oldVolume == nil && newVolume != nil:
		return []*VolumeLifecycleEvent{{ID: id, Metadata: metadata, Type: oriVolumeStateToVolumeLifecycleEventType(newVolume.state)}}
	case oldVolume != nil && newVolume == nil:
		return []*VolumeLifecycleEvent{{ID: id, Metadata: metadata, Type: VolumeRemoved}}
	case oldVolume != nil && newVolume != nil:
		var events []*VolumeLifecycleEvent
		if oldVolume.state != newVolume.state {
			events = append(events, &VolumeLifecycleEvent{ID: id, Metadata: metadata, Type: oriVolumeStateToVolumeLifecycleEventType(newVolume.state)})
		}
		return events
	default:
		panic("unhandled case")
	}
}

func volumeLifecycleEventMetadata(old, current *genericVolume) VolumeLifecycleEventMetadata {
	switch {
	case current != nil:
		return VolumeLifecycleEventMetadata{
			Namespace: current.metadata.namespace,
			Name:      current.metadata.name,
			UID:       current.metadata.uid,
		}
	case old != nil:
		return VolumeLifecycleEventMetadata{
			Namespace: old.metadata.namespace,
			Name:      old.metadata.name,
			UID:       old.metadata.uid,
		}
	default:
		panic("both current and old are nil")
	}
}

func (g *Generic) relist(ctx context.Context, log logr.Logger) error {
	timestamp := time.Now()
	volumes, err := g.list(ctx)
	if err != nil {
		return fmt.Errorf("error listing volumes: %w", err)
	}
	g.relistTime.Store(&timestamp)

	g.volumes.setCurrent(volumes)

	eventsByVolumeID := make(map[string][]*VolumeLifecycleEvent)
	for volumeID := range g.volumes {
		oldVolume := g.volumes.getOld(volumeID)
		currentVolume := g.volumes.getCurrent(volumeID)
		volumeEventMetadata := volumeLifecycleEventMetadata(oldVolume, currentVolume)
		eventsByVolumeID[volumeID] = append(
			eventsByVolumeID[volumeID],
			g.inferEvents(volumeID, volumeEventMetadata, oldVolume, currentVolume)...,
		)
	}

	for volumeID, events := range eventsByVolumeID {
		g.volumes.update(volumeID)
		for i := range events {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case g.eventChannel <- events[i]:
			default:
				log.Info("Event channel is full, discarding event", "VolumeID", volumeID, "EventType", events[i].Type)
			}
		}
	}

	return nil
}

func (g *Generic) Watch() <-chan *VolumeLifecycleEvent {
	return g.eventChannel
}

type oldNewMapEntry[V any] struct {
	Old     V
	Current V
}

type keyed[K comparable] interface {
	key() K
}

type oldNewMap[K comparable, V keyed[K]] map[K]*oldNewMapEntry[*V]

func (m oldNewMap[K, V]) set(old, current []*V) { //nolint:unused
	for _, item := range old {
		key := (*item).key()
		m[key] = &oldNewMapEntry[*V]{
			Old: item,
		}
	}
	for _, item := range current {
		key := (*item).key()
		if r, ok := m[key]; ok {
			r.Current = item
		} else {
			m[key] = &oldNewMapEntry[*V]{
				Current: item,
			}
		}
	}
}

func (m oldNewMap[K, V]) setCurrent(current []*V) {
	for _, v := range m {
		v.Current = nil
	}

	for _, item := range current {
		key := (*item).key()
		if r, ok := m[key]; ok {
			r.Current = item
		} else {
			m[key] = &oldNewMapEntry[*V]{
				Current: item,
			}
		}
	}
}

func (m oldNewMap[K, V]) getCurrent(key K) *V {
	r, ok := m[key]
	if ok {
		return r.Current
	}
	return nil
}

func (m oldNewMap[K, V]) getOld(key K) *V {
	r, ok := m[key]
	if ok {
		return r.Old
	}
	return nil
}

func (m oldNewMap[K, V]) update(key K) {
	r, ok := m[key]
	if !ok {
		return
	}

	if r.Current == nil {
		delete(m, key)
		return
	}

	r.Old = r.Current
	r.Current = nil
}
