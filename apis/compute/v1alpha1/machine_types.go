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

package v1alpha1

import (
	common "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MachineSpec defines the desired state of Machine
type MachineSpec struct {
	// Hostname is the hostname of the machine
	Hostname string `json:"hostname"`
	// MachineClass is the machine class/flavor of the machine
	MachineClass common.ScopedReference `json:"machineClass"`
	// MachinePool defines the compute pool of the machine
	MachinePool common.ScopedReference `json:"machinePool,omitempty"`
	// Location is the physical location of the machine
	Location common.Location `json:"location"`
	// Image is the operating system image of the machine
	Image common.ScopedReference `json:"image"`
	// SSHPublicKeys is a list of SSH public keys of a machine
	SSHPublicKeys []SSHPublicKeyEntry `json:"sshPublicKeys"`
	// Interfaces define a list of network interfaces present on the machine
	Interfaces []Interface `json:"interfaces,omitempty"`
	// SecurityGroups is a list of security groups of a machine
	SecurityGroups []common.ScopedReference `json:"securityGroups"`
	// VolumeClaims
	VolumeClaims []VolumeClaim `json:"volumeClaims"`
	// UserData defines the ignition file
	UserData string `json:"userData,omitempty"`
}

// Interface is the definition of a single interface
type Interface struct {
	// Name is the name of the interface
	Name string `json:"name"`
	// Target is the referenced resource of this interface
	Target common.ScopedKindReference `json:"target"`
	// Priority is the priority level of this interface
	Priority int `json:"priority,omitempty"`
	// IP specifies a concrete IP address which should be allocated from a Subnet
	IP string `json:"ip,omitempty"`
	// RoutingOnly is a routing hint for this interface
	RoutingOnly bool `json:"routingOnly,omitempty"`
}

// VolumeClaim defines a volume claim of a machine
type VolumeClaim struct {
	// Name is the name of the VolumeClaim
	Name string `json:"name"`
	// RetainPolicy defines what should happen when the machine is being deleted
	RetainPolicy RetainPolicy `json:"retainPolicy"`
	// Device defines the device for a volume on the machine
	Device string `json:"device"`
	// StorageClass describes the storage class of the volumes
	StorageClass common.ScopedReference `json:"storageClass"`
	// Size defines the size of the volume
	Size *resource.Quantity `json:"size,omitempty"`
	// Volume is a reference to an existing volume
	Volume common.ScopedReference `json:"volume,omitempty"`
}

type RetainPolicy string

const (
	RetainPolicyDeleteOnTermination RetainPolicy = "DeleteOnTermination"
	RetainPolicyPersistent          RetainPolicy = "Persistent"
	MachineStateRunning                          = "Running"
	MachineStateShutdown                         = "Shutdown"
	MachineStateError                            = "Error"
	MachineStateInitial                          = "Initial"
)

// SSHPublicKeyEntry describes either a reference to a SSH public key or a selector
// to filter for a public key
type SSHPublicKeyEntry struct {
	// Scope is the scope of a SSH public key
	Scope string `json:"scope,omitempty"`
	// Name is the name of the SSH public key
	Name string `json:"name,omitempty"`
	// Selector defines a LabelSelector to filter for a public key
	Selector metav1.LabelSelector `json:"selector,omitempty"`
}

// MachineStatus defines the observed state of Machine
type MachineStatus struct {
	common.StateFields `json:",inline"`
	//TODO: define machine state fields
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Hostname",type=string,JSONPath=`.spec.hostname`
//+kubebuilder:printcolumn:name="MachineClass",type=string,JSONPath=`.spec.machineClass.name`
//+kubebuilder:printcolumn:name="Image",type=string,JSONPath=`.spec.image.name`
//+kubebuilder:printcolumn:name="Region",type=string,JSONPath=`.spec.location.region`
//+kubebuilder:printcolumn:name="AZ",type=string,JSONPath=`.spec.location.availabilityZone`
//+kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Machine is the Schema for the machines API
type Machine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MachineSpec   `json:"spec,omitempty"`
	Status MachineStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MachineList contains a list of Machine
type MachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Machine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Machine{}, &MachineList{})
}
