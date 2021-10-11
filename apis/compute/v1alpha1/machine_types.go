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
	"fmt"
	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func MachineInterfaceIPAMRangeName(machineName, ifaceName string) string {
	return fmt.Sprintf("machine-iface-%s-%s", machineName, ifaceName)
}

// MachineSpec defines the desired state of Machine
type MachineSpec struct {
	// Hostname is the hostname of the machine
	Hostname string `json:"hostname"`
	// MachineClass is a reference to the machine class/flavor of the machine.
	MachineClass corev1.LocalObjectReference `json:"machineClass"`
	// MachinePool defines machine pool to run the machine in.
	// If empty, a scheduler will figure out an appropriate pool to run the machine in.
	MachinePool corev1.LocalObjectReference `json:"machinePool,omitempty"`
	// Image is the URL providing the operating system image of the machine.
	Image string `json:"image"`
	// SSHPublicKeys is a list of SSH public key secret references of a machine.
	SSHPublicKeys []commonv1alpha1.SecretKeySelector `json:"sshPublicKeys,omitempty"`
	// Interfaces define a list of network interfaces present on the machine
	Interfaces []Interface `json:"interfaces,omitempty"`
	// SecurityGroups is a list of security groups of a machine
	SecurityGroups []corev1.LocalObjectReference `json:"securityGroups,omitempty"`
	// VolumeClaims are volumes claimed by this machine.
	VolumeClaims []VolumeClaim `json:"volumeClaims,omitempty"`
	// Ignition is a reference to a config map containing the ignition YAML for the machine to boot up.
	// If key is empty, DefaultIgnitionKey will be used as fallback.
	Ignition *commonv1alpha1.ConfigMapKeySelector `json:"ignition,omitempty"`
	// EFIVars are variables to pass to EFI while booting up.
	EFIVars []EFIVar `json:"efiVars,omitempty"`
}

// EFIVar is a variable to pass to EFI while booting up.
type EFIVar struct {
	Name  string `json:"name,omitempty"`
	UUID  string `json:"uuid,omitempty"`
	Value string `json:"value"`
}

// DefaultIgnitionKey is the default key for MachineSpec.UserData.
const DefaultIgnitionKey = "ignition.yaml"

// Interface is the definition of a single interface
type Interface struct {
	// Name is the name of the interface
	Name string `json:"name"`
	// Target is the referenced resource of this interface.
	Target corev1.LocalObjectReference `json:"target"`
	// Priority is the priority level of this interface
	Priority int32 `json:"priority,omitempty"`
	// IP specifies a concrete IP address which should be allocated from a Subnet
	IP *commonv1alpha1.IPAddr `json:"ip,omitempty"`
}

// VolumeClaim defines a volume claim of a machine
type VolumeClaim struct {
	// Name is the name of the VolumeClaim
	Name string `json:"name"`
	// Priority is the OS priority of the volume.
	Priority int32 `json:"priority,omitempty"`
	// RetainPolicy defines what should happen when the machine is being deleted
	RetainPolicy RetainPolicy `json:"retainPolicy"`
	// StorageClass describes the storage class of the volumes
	StorageClass corev1.LocalObjectReference `json:"storageClass"`
	// Size defines the size of the volume
	Size *resource.Quantity `json:"size,omitempty"`
	// Volume is a reference to an existing volume
	Volume corev1.LocalObjectReference `json:"volume,omitempty"`
}

type RetainPolicy string

const (
	RetainPolicyDeleteOnTermination RetainPolicy = "DeleteOnTermination"
	RetainPolicyPersistent          RetainPolicy = "Persistent"
)

// InterfaceStatus reports the status of an Interface.
type InterfaceStatus struct {
	// Name is the name of an interface.
	Name string `json:"name"`
	// IP is the IP allocated for an interface.
	IP commonv1alpha1.IPAddr `json:"ip"`
	// Priority is the OS priority of the interface.
	Priority int32 `json:"priority,omitempty"`
}

// VolumeClaimStatus is the status of a VolumeClaim.
type VolumeClaimStatus struct {
	// Name is the name of a volume claim.
	Name string `json:"name"`
	// Priority is the OS priority of the volume.
	Priority int32 `json:"priority,omitempty"`
}

// MachineStatus defines the observed state of Machine
type MachineStatus struct {
	State        MachineState        `json:"state,omitempty"`
	Conditions   []MachineCondition  `json:"conditions,omitempty"`
	Interfaces   []InterfaceStatus   `json:"interfaces,omitempty"`
	VolumeClaims []VolumeClaimStatus `json:"volumeClaims,omitempty"`
}

type MachineState string

const (
	MachineStateRunning  MachineState = "Running"
	MachineStateShutdown MachineState = "Shutdown"
	MachineStateError    MachineState = "Error"
	MachineStateInitial  MachineState = "Initial"
)

// MachineConditionType is a type a MachineCondition can have.
type MachineConditionType string

// MachineCondition is one of the conditions of a volume.
type MachineCondition struct {
	// Type is the type of the condition.
	Type MachineConditionType `json:"type"`
	// Status is the status of the condition.
	Status corev1.ConditionStatus `json:"status"`
	// Reason is a machine-readable indication of why the condition is in a certain state.
	Reason string `json:"reason"`
	// Message is a human-readable explanation of why the condition has a certain reason / state.
	Message string `json:"message"`
	// ObservedGeneration represents the .metadata.generation that the condition was set based upon.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// LastUpdateTime is the last time a condition has been updated.
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
	// LastTransitionTime is the last time the status of a condition has transitioned from one state to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Hostname",type=string,JSONPath=`.spec.hostname`
//+kubebuilder:printcolumn:name="MachineClass",type=string,JSONPath=`.spec.machineClass.name`
//+kubebuilder:printcolumn:name="Image",type=string,JSONPath=`.spec.image`
//+kubebuilder:printcolumn:name="MachinePool",type=string,JSONPath=`.spec.machinePool.name`
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
