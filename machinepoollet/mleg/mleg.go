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
	"k8s.io/apiserver/pkg/server/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type MachineLifecycleEventType string

const (
	MachineCreated MachineLifecycleEventType = "MachineCreated"
	MachineStarted MachineLifecycleEventType = "MachineStarted"
	MachineStopped MachineLifecycleEventType = "MachineStopped"
	MachineRemoved MachineLifecycleEventType = "MachineRemoved"

	NetworkInterfaceAdded    MachineLifecycleEventType = "NetworkInterfaceAdded"
	NetworkInterfaceRemoved  MachineLifecycleEventType = "NetworkInterfaceRemoved"
	NetworkInterfaceAttached MachineLifecycleEventType = "NetworkInterfaceAttached"
	NetworkInterfaceDetached MachineLifecycleEventType = "NetworkInterfaceDetached"

	VolumeAdded    MachineLifecycleEventType = "VolumeAdded"
	VolumeRemoved  MachineLifecycleEventType = "VolumeRemoved"
	VolumeAttached MachineLifecycleEventType = "VolumeAttached"
	VolumeDetached MachineLifecycleEventType = "VolumeDetached"
)

type MachineLifecycleEventMetadata struct {
	Namespace string
	Name      string
	UID       string
}

// MachineLifecycleEvent is an event emitted for a machine.
type MachineLifecycleEvent struct {
	// ID is the id of the machine the event is for.
	ID string
	// Metadata is associated metadata for the machine the event is for.
	Metadata MachineLifecycleEventMetadata
	// Type is the emitted event type.
	Type MachineLifecycleEventType
	// Data is the corresponding data for the MachineLifecycleEventType. May be zero.
	Data any
}

type MachineLifecycleEventGenerator interface {
	manager.Runnable
	healthz.HealthChecker
	Watch() <-chan *MachineLifecycleEvent
}
