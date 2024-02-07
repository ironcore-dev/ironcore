// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
)

// MachineSpec defines the desired state of Machine
type MachineSpec struct {
	// MachineClassRef is a reference to the machine class/flavor of the machine.
	MachineClassRef corev1.LocalObjectReference `json:"machineClassRef"`
	// MachinePoolSelector selects a suitable MachinePoolRef by the given labels.
	MachinePoolSelector map[string]string `json:"machinePoolSelector,omitempty"`
	// MachinePoolRef defines machine pool to run the machine in.
	// If empty, a scheduler will figure out an appropriate pool to run the machine in.
	MachinePoolRef *corev1.LocalObjectReference `json:"machinePoolRef,omitempty"`
	// Power is the desired machine power state.
	// Defaults to PowerOn.
	Power Power `json:"power,omitempty"`
	// Image is the optional URL providing the operating system image of the machine.
	// +optional
	Image string `json:"image,omitempty"`
	// ImagePullSecretRef is an optional secret for pulling the image of a machine.
	ImagePullSecretRef *corev1.LocalObjectReference `json:"imagePullSecret,omitempty"`
	// NetworkInterfaces define a list of network interfaces present on the machine
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	NetworkInterfaces []NetworkInterface `json:"networkInterfaces,omitempty" patchStrategy:"merge,retainKeys" patchMergeKey:"name"`
	// Volumes are volumes attached to this machine.
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	Volumes []Volume `json:"volumes,omitempty" patchStrategy:"merge,retainKeys" patchMergeKey:"name"`
	// IgnitionRef is a reference to a secret containing the ignition YAML for the machine to boot up.
	// If key is empty, DefaultIgnitionKey will be used as fallback.
	IgnitionRef *commonv1alpha1.SecretKeySelector `json:"ignitionRef,omitempty"`
	// EFIVars are variables to pass to EFI while booting up.
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	EFIVars []EFIVar `json:"efiVars,omitempty" patchStrategy:"merge,retainKeys" patchMergeKey:"name"`
	// Tolerations define tolerations the Machine has. Only MachinePools whose taints
	// covered by Tolerations will be considered to run the Machine.
	Tolerations []commonv1alpha1.Toleration `json:"tolerations,omitempty"`
}

// Power is the desired power state of a Machine.
type Power string

const (
	// PowerOn indicates that a Machine should be powered on.
	PowerOn Power = "On"
	// PowerOff indicates that a Machine should be powered off.
	PowerOff Power = "Off"
)

// EFIVar is a variable to pass to EFI while booting up.
type EFIVar struct {
	// Name is the name of the EFIVar.
	Name string `json:"name"`
	// UUID is the uuid of the EFIVar.
	UUID string `json:"uuid"`
	// Value is the value of the EFIVar.
	Value string `json:"value"`
}

// DefaultIgnitionKey is the default key for MachineSpec.UserData.
const DefaultIgnitionKey = "ignition.yaml"

// NetworkInterface is the definition of a single interface
type NetworkInterface struct {
	// Name is the name of the network interface.
	Name string `json:"name"`
	// NetworkInterfaceSource is where to obtain the interface from.
	NetworkInterfaceSource `json:",inline"`
}

type NetworkInterfaceSource struct {
	// NetworkInterfaceRef instructs to use the NetworkInterface at the target reference.
	NetworkInterfaceRef *corev1.LocalObjectReference `json:"networkInterfaceRef,omitempty"`
	// Ephemeral instructs to create an ephemeral (i.e. coupled to the lifetime of the surrounding object)
	// NetworkInterface to use.
	Ephemeral *EphemeralNetworkInterfaceSource `json:"ephemeral,omitempty"`
}

// Volume defines a volume attachment of a machine
type Volume struct {
	// Name is the name of the Volume
	Name string `json:"name"`
	// Device is the device name where the volume should be attached.
	// Pointer to distinguish between explicit zero and not specified.
	// If empty, an unused device name will be determined if possible.
	Device *string `json:"device,omitempty"`
	// VolumeSource is the source where the storage for the Volume resides at.
	VolumeSource `json:",inline"`
}

// VolumeSource specifies the source to use for a Volume.
type VolumeSource struct {
	// VolumeRef instructs to use the specified Volume as source for the attachment.
	VolumeRef *corev1.LocalObjectReference `json:"volumeRef,omitempty"`
	// EmptyDisk instructs to use a Volume offered by the machine pool provider.
	EmptyDisk *EmptyDiskVolumeSource `json:"emptyDisk,omitempty"`
	// Ephemeral instructs to create an ephemeral (i.e. coupled to the lifetime of the surrounding object)
	// Volume to use.
	Ephemeral *EphemeralVolumeSource `json:"ephemeral,omitempty"`
}

// EmptyDiskVolumeSource is a volume that's offered by the machine pool provider.
// Usually ephemeral (i.e. deleted when the surrounding entity is deleted), with
// varying performance characteristics. Potentially not recoverable.
type EmptyDiskVolumeSource struct {
	// SizeLimit is the total amount of local storage required for this EmptyDisk volume.
	// The default is nil which means that the limit is undefined.
	SizeLimit *resource.Quantity `json:"sizeLimit,omitempty"`
}

// NetworkInterfaceStatus reports the status of an NetworkInterfaceSource.
type NetworkInterfaceStatus struct {
	// Name is the name of the NetworkInterface to whom the status belongs to.
	Name string `json:"name"`
	// Handle is the MachinePool internal handle of the NetworkInterface.
	Handle string `json:"handle,omitempty"`
	// IPs are the ips allocated for the network interface.
	IPs []commonv1alpha1.IP `json:"ips,omitempty"`
	// VirtualIP is the virtual ip allocated for the network interface.
	VirtualIP *commonv1alpha1.IP `json:"virtualIP,omitempty"`
	// State represents the attachment state of a NetworkInterface.
	State NetworkInterfaceState `json:"state,omitempty"`
	// LastStateTransitionTime is the last time the State transitioned.
	LastStateTransitionTime *metav1.Time `json:"lastStateTransitionTime,omitempty"`
}

// NetworkInterfaceState is the infrastructure attachment state a NetworkInterface can be in.
type NetworkInterfaceState string

const (
	// NetworkInterfaceStatePending indicates that the attachment of a network interface is pending.
	NetworkInterfaceStatePending NetworkInterfaceState = "Pending"
	// NetworkInterfaceStateAttached indicates that a network interface has been successfully attached.
	NetworkInterfaceStateAttached NetworkInterfaceState = "Attached"
)

// VolumeStatus is the status of a Volume.
type VolumeStatus struct {
	// Name is the name of a volume attachment.
	Name string `json:"name"`
	// Handle is the MachinePool internal handle of the volume.
	Handle string `json:"handle,omitempty"`
	// State represents the attachment state of a Volume.
	State VolumeState `json:"state,omitempty"`
	// LastStateTransitionTime is the last time the State transitioned.
	LastStateTransitionTime *metav1.Time `json:"lastStateTransitionTime,omitempty"`
}

// VolumeState is the infrastructure attachment state a Volume can be in.
type VolumeState string

const (
	// VolumeStatePending indicates that the attachment of a volume is pending.
	VolumeStatePending VolumeState = "Pending"
	// VolumeStateAttached indicates that a volume has been successfully attached.
	VolumeStateAttached VolumeState = "Attached"
)

// MachineStatus defines the observed state of Machine
type MachineStatus struct {
	// MachineID is the provider specific machine ID in the format '<type>://<machine_id>'.
	MachineID string `json:"machineID,omitempty"`
	// ObservedGeneration is the last generation the MachinePool observed of the Machine.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// State is the infrastructure state of the machine.
	State MachineState `json:"state,omitempty"`
	// NetworkInterfaces is the list of network interface states for the machine.
	NetworkInterfaces []NetworkInterfaceStatus `json:"networkInterfaces,omitempty"`
	// Volumes is the list of volume states for the machine.
	Volumes []VolumeStatus `json:"volumes,omitempty"`
}

// MachineState is the state of a machine.
// +enum
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
	// MachineStateTerminated means the machine has been permanently stopped and cannot be started.
	MachineStateTerminated MachineState = "Terminated"
	// MachineStateTerminating means the machine that is terminating.
	MachineStateTerminating MachineState = "Terminating"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient

// Machine is the Schema for the machines API
type Machine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MachineSpec   `json:"spec,omitempty"`
	Status MachineStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MachineList contains a list of Machine
type MachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Machine `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:conversion-gen:explicit-from=net/url.Values

// MachineExecOptions is the query options to a Machine's remote exec call
type MachineExecOptions struct {
	metav1.TypeMeta              `json:",inline"`
	InsecureSkipTLSVerifyBackend bool `json:"insecureSkipTLSVerifyBackend,omitempty"`
}
