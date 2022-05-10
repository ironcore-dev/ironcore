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

package compute

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
)

// MachineSpec defines the desired state of Machine
type MachineSpec struct {
	// MachineClassRef is a reference to the machine class/flavor of the machine.
	MachineClassRef corev1.LocalObjectReference
	// MachinePoolSelector selects a suitable MachinePoolRef by the given labels.
	MachinePoolSelector map[string]string
	// MachinePoolRef defines machine pool to run the machine in.
	// If empty, a scheduler will figure out an appropriate pool to run the machine in.
	MachinePoolRef *corev1.LocalObjectReference
	// Image is the URL providing the operating system image of the machine.
	Image string
	// NetworkInterfaces define a list of network interfaces present on the machine
	NetworkInterfaces []NetworkInterface
	// Volumes are volumes attached to this machine.
	Volumes []Volume
	// IgnitionRef is a reference to a config map containing the ignition YAML for the machine to boot up.
	// If key is empty, DefaultIgnitionKey will be used as fallback.
	IgnitionRef *commonv1alpha1.ConfigMapKeySelector
	// EFIVars are variables to pass to EFI while booting up.
	EFIVars []EFIVar
	// Tolerations define tolerations the Machine has. Only MachinePools whose taints
	// covered by Tolerations will be considered to run the Machine.
	Tolerations []commonv1alpha1.Toleration
}

// EFIVar is a variable to pass to EFI while booting up.
type EFIVar struct {
	// Name is the name of the EFIVar.
	Name string
	// UUID is the uuid of the EFIVar.
	UUID string
	// Value is the value of the EFIVar.
	Value string
}

// DefaultIgnitionKey is the default key for MachineSpec.UserData.
const DefaultIgnitionKey = "ignition.yaml"

// NetworkInterface is the definition of a single interface
type NetworkInterface struct {
	// Name is the name of the network interface.
	Name string
	// NetworkInterfaceSource is where to obtain the interface from.
	NetworkInterfaceSource
}

type NetworkInterfaceSource struct {
	// NetworkInterfaceRef instructs to use the NetworkInterface at the target reference.
	NetworkInterfaceRef *corev1.LocalObjectReference
	// Ephemeral instructs to create an ephemeral (i.e. coupled to the lifetime of the surrounding object)
	// NetworkInterface to use.
	Ephemeral *EphemeralNetworkInterfaceSource
}

// Volume defines a volume attachment of a machine
type Volume struct {
	// Name is the name of the Volume
	Name string
	// VolumeSource is the source where the storage for the Volume resides at.
	VolumeSource
}

// VolumeSource specifies the source to use for a Volume.
type VolumeSource struct {
	// VolumeClaimRef instructs the Volume to use a VolumeClaimRef as source for the attachment.
	VolumeClaimRef *corev1.LocalObjectReference
}

// NetworkInterfaceStatus reports the status of an NetworkInterfaceSource.
type NetworkInterfaceStatus struct {
	// Name is the name of the NetworkInterface to whom the status belongs to.
	Name string
	// IPs are the ips allocated for the network interface.
	IPs []commonv1alpha1.IP
	// VirtualIP is the virtual ip allocated for the network interface.
	VirtualIP *commonv1alpha1.IP
}

// VolumeStatus is the status of a Volume.
type VolumeStatus struct {
	// Name is the name of a volume attachment.
	Name string
	// DeviceID is the disk device ID on the host.
	DeviceID string
}

// MachineStatus defines the observed state of Machine
type MachineStatus struct {
	// State is the state of the machine.
	State MachineState
	// Conditions are the conditions of the machines.
	Conditions []MachineCondition
	// NetworkInterfaces is the list of network interface states for the machine.
	NetworkInterfaces []NetworkInterfaceStatus
	// Volumes is the list of volume states for the machine.
	Volumes []VolumeStatus
}

// MachineState is the state of a machine.
//+enum
type MachineState string

const (
	// MachineStatePending means the Machine has been accepted by the system, but not yet completely started.
	// This includes time before being bound to a MachinePool, as well as time spent setting up the Machine on that
	// MachinePool.
	MachineStatePending MachineState = "Pending"
	// MachineStateRunning means the machine is running on a MachinePool.
	MachineStateRunning MachineState = "Running"
	// MachineStateShutdown means the machine is shut down.
	MachineStateShutdown MachineState = "Shutdown"
	// MachineStateError means the machine is in an error state.
	MachineStateError MachineState = "Error"
)

// MachineConditionType is a type a MachineCondition can have.
type MachineConditionType string

const (
	// MachineSynced represents the condition of a machine being synced with its backing resources
	MachineSynced MachineConditionType = "Synced"
)

// MachineCondition is one of the conditions of a volume.
type MachineCondition struct {
	// Type is the type of the condition.
	Type MachineConditionType
	// Status is the status of the condition.
	Status corev1.ConditionStatus
	// Reason is a machine-readable indication of why the condition is in a certain state.
	Reason string
	// Message is a human-readable explanation of why the condition has a certain reason / state.
	Message string
	// ObservedGeneration represents the .metadata.generation that the condition was set based upon.
	ObservedGeneration int64
	// LastTransitionTime is the last time the status of a condition has transitioned from one state to another.
	LastTransitionTime metav1.Time
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient

// Machine is the Schema for the machines API
type Machine struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   MachineSpec
	Status MachineStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MachineList contains a list of Machine
type MachineList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []Machine
}
