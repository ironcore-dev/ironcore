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

package mleg

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/go-logr/logr"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrl "sigs.k8s.io/controller-runtime"
)

type genericMachineMetadata struct {
	namespace string
	name      string
	uid       string
}

type genericMachine struct {
	id       string
	metadata genericMachineMetadata

	volumes           []*genericVolume
	networkInterfaces []*genericNetworkInterface
}

func (g genericMachine) key() string { //nolint:unused
	return g.id
}

type genericVolume struct {
	name string
}

func (g genericVolume) key() string { //nolint:unused
	return g.name
}

type genericNetworkInterface struct {
	name string
}

func (g genericNetworkInterface) key() string { //nolint:unused
	return g.name
}

type Generic struct {
	// relistPeriod is the period for relisting.
	relistPeriod time.Duration
	// relistThreshold is the maximum threshold between two relists to become unhealthy.
	relistThreshold time.Duration
	// relistTime is the last time a relist happened.
	relistTime atomic.Pointer[time.Time]
	// runtime is the machine runtime.
	runtime ori.MachineRuntimeClient
	// eventChannel is the channel from which the subscriber receives events.
	eventChannel chan *MachineLifecycleEvent

	machines oldNewMap[string, genericMachine]
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

func NewGeneric(runtime ori.MachineRuntimeClient, opts GenericOptions) MachineLifecycleEventGenerator {
	setGenericMLEGOptionsDefaults(&opts)

	return &Generic{
		relistPeriod:    opts.RelistPeriod,
		relistThreshold: opts.RelistThreshold,
		runtime:         runtime,
		eventChannel:    make(chan *MachineLifecycleEvent, opts.ChannelCapacity),
		machines:        make(oldNewMap[string, genericMachine]),
	}
}

func (g *Generic) list(ctx context.Context) ([]*genericMachine, error) {
	listMachinesRes, err := g.runtime.ListMachines(ctx, &ori.ListMachinesRequest{})
	if err != nil {
		return nil, fmt.Errorf("error listing machines: %w", err)
	}

	listNetworkInterfacesRes, err := g.runtime.ListNetworkInterfaces(ctx, &ori.ListNetworkInterfacesRequest{})
	if err != nil {
		return nil, fmt.Errorf("error listing network interfaces: %w", err)
	}

	listVolumesRes, err := g.runtime.ListVolumes(ctx, &ori.ListVolumesRequest{})
	if err != nil {
		return nil, fmt.Errorf("error listing volumes: %w", err)
	}

	machinesByID := make(map[string]*genericMachine)
	for _, machine := range listMachinesRes.Machines {
		machinesByID[machine.Id] = &genericMachine{
			id: machine.Id,
			metadata: genericMachineMetadata{
				uid:       machine.Metadata.Uid,
				namespace: machine.Metadata.Namespace,
				name:      machine.Metadata.Name,
			},
		}
	}

	for _, networkInterface := range listNetworkInterfacesRes.NetworkInterfaces {
		machine, ok := machinesByID[networkInterface.MachineId]
		if !ok {
			machine = &genericMachine{
				id: networkInterface.MachineId,
				metadata: genericMachineMetadata{
					uid:       networkInterface.MachineMetadata.Uid,
					namespace: networkInterface.MachineMetadata.Namespace,
					name:      networkInterface.MachineMetadata.Name,
				},
			}
			machinesByID[networkInterface.MachineId] = machine
		}

		machine.networkInterfaces = append(machine.networkInterfaces, &genericNetworkInterface{
			name: networkInterface.Name,
		})
	}

	for _, volume := range listVolumesRes.Volumes {
		machine, ok := machinesByID[volume.MachineId]
		if !ok {
			machine = &genericMachine{
				id: volume.MachineId,
				metadata: genericMachineMetadata{
					uid:       volume.MachineMetadata.Uid,
					namespace: volume.MachineMetadata.Namespace,
					name:      volume.MachineMetadata.Name,
				},
			}
			machinesByID[volume.MachineId] = machine
		}

		machine.volumes = append(machine.volumes, &genericVolume{
			name: volume.Name,
		})
	}

	res := make([]*genericMachine, 0, len(machinesByID))
	for _, machine := range machinesByID {
		res = append(res, machine)
	}
	return res, nil
}

func (g *Generic) Name() string {
	return "mleg"
}

func (g *Generic) Check(_ *http.Request) error {
	relistTime := g.relistTime.Load()
	if relistTime == nil {
		return fmt.Errorf("mleg did not relist yet")
	}

	elapsed := time.Since(*relistTime)
	if elapsed > g.relistThreshold {
		return fmt.Errorf("mleg was last seen active %v ago, threshold is %v", elapsed, g.relistThreshold)
	}
	return nil
}

func (g *Generic) Start(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx).WithName("mleg")
	wait.UntilWithContext(ctx, func(ctx context.Context) {
		if err := g.relist(ctx, log); err != nil {
			log.Error(err, "Error relisting")
		}
	}, g.relistPeriod)
	return nil
}

func (g *Generic) inferNetworkInterfaceEvents(
	machineID string,
	metadata MachineLifecycleEventMetadata,
	networkInterfaceName string,
	oldNetworkInterface, newNetworkInterface *genericNetworkInterface,
) []*MachineLifecycleEvent {
	switch {
	case oldNetworkInterface == nil && newNetworkInterface != nil:
		return []*MachineLifecycleEvent{{ID: machineID, Metadata: metadata, Type: NetworkInterfaceAdded, Data: networkInterfaceName}}
	case oldNetworkInterface != nil && newNetworkInterface == nil:
		return []*MachineLifecycleEvent{{ID: machineID, Metadata: metadata, Type: NetworkInterfaceRemoved, Data: networkInterfaceName}}
	case oldNetworkInterface != nil && newNetworkInterface != nil:
		return nil
	default:
		panic("unhandled case")
	}
}

func (g *Generic) inferVolumeEvents(
	machineID string,
	metadata MachineLifecycleEventMetadata,
	volumeName string,
	oldVolume, newVolume *genericVolume,
) []*MachineLifecycleEvent {
	switch {
	case oldVolume == nil && newVolume != nil:
		return []*MachineLifecycleEvent{{ID: machineID, Metadata: metadata, Type: VolumeAdded, Data: volumeName}}
	case oldVolume != nil && newVolume == nil:
		return []*MachineLifecycleEvent{{ID: machineID, Metadata: metadata, Type: VolumeRemoved, Data: volumeName}}
	case oldVolume != nil && newVolume != nil:
		return nil
	default:
		panic("unhandled case")
	}
}

func (g *Generic) inferEvents(id string, metadata MachineLifecycleEventMetadata, oldMachine, newMachine *genericMachine) []*MachineLifecycleEvent {
	switch {
	case oldMachine == nil && newMachine != nil:
		return []*MachineLifecycleEvent{{ID: id, Metadata: metadata, Type: MachineStarted}}
	case oldMachine != nil && newMachine == nil:
		return []*MachineLifecycleEvent{{ID: id, Metadata: metadata, Type: MachineStopped}, {ID: id, Metadata: metadata, Type: MachineRemoved}}
	case oldMachine != nil && newMachine != nil:
		var events []*MachineLifecycleEvent

		networkInterfaces := make(oldNewMap[string, genericNetworkInterface])
		networkInterfaces.set(oldMachine.networkInterfaces, newMachine.networkInterfaces)
		for name, entry := range networkInterfaces {
			events = append(events, g.inferNetworkInterfaceEvents(id, metadata, name, entry.Old, entry.Current)...)
		}

		volumes := make(oldNewMap[string, genericVolume])
		volumes.set(oldMachine.volumes, newMachine.volumes)
		for name, entry := range volumes {
			events = append(events, g.inferVolumeEvents(id, metadata, name, entry.Old, entry.Current)...)
		}

		return events
	default:
		panic("unhandled case")
	}
}

func machineLifecycleEventMetadata(old, current *genericMachine) MachineLifecycleEventMetadata {
	switch {
	case current != nil:
		return MachineLifecycleEventMetadata{
			Namespace: current.metadata.namespace,
			Name:      current.metadata.name,
			UID:       current.metadata.uid,
		}
	case old != nil:
		return MachineLifecycleEventMetadata{
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
	machines, err := g.list(ctx)
	if err != nil {
		return fmt.Errorf("error listing machines: %w", err)
	}
	g.relistTime.Store(&timestamp)

	g.machines.setCurrent(machines)

	eventsByMachineID := make(map[string][]*MachineLifecycleEvent)
	for machineID := range g.machines {
		oldMachine := g.machines.getOld(machineID)
		currentMachine := g.machines.getCurrent(machineID)
		machineEventMetadata := machineLifecycleEventMetadata(oldMachine, currentMachine)
		eventsByMachineID[machineID] = append(
			eventsByMachineID[machineID],
			g.inferEvents(machineID, machineEventMetadata, oldMachine, currentMachine)...,
		)
	}

	for machineID, events := range eventsByMachineID {
		g.machines.update(machineID)
		for i := range events {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case g.eventChannel <- events[i]:
			default:
				log.Info("Event channel is full, discarding event", "MachineID", machineID, "EventType", events[i].Type)
			}
		}
	}

	return nil
}

func (g *Generic) Watch() <-chan *MachineLifecycleEvent {
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

func (m oldNewMap[K, V]) set(old, current []*V) {
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
