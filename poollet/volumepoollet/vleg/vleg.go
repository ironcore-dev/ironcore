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
	"k8s.io/apiserver/pkg/server/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type VolumeLifecycleEventType string

const (
	VolumePending   VolumeLifecycleEventType = "VolumePending"
	VolumeAvailable VolumeLifecycleEventType = "VolumeAvailable"
	VolumeError     VolumeLifecycleEventType = "VolumeError"
	VolumeRemoved   VolumeLifecycleEventType = "VolumeRemoved"
)

type VolumeLifecycleEventMetadata struct {
	Namespace string
	Name      string
	UID       string
}

// VolumeLifecycleEvent is an event emitted for a volume.
type VolumeLifecycleEvent struct {
	// ID is the id of the volume the event is for.
	ID string
	// Metadata is associated metadata for the volume the event is for.
	Metadata VolumeLifecycleEventMetadata
	// Type is the emitted event type.
	Type VolumeLifecycleEventType
	// Data is the corresponding data for the VolumeLifecycleEventType. May be zero.
	Data any
}

type VolumeLifecycleEventGenerator interface {
	manager.Runnable
	healthz.HealthChecker
	Watch() <-chan *VolumeLifecycleEvent
}
