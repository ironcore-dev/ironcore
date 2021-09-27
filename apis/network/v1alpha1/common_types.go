package v1alpha1

import corev1 "k8s.io/api/core/v1"

// Target is a target for network traffic.
// It may be either
// * a v1alpha1.Machine
// * a Gateway
// * a ReservedIP
// * a raw IP
// * a raw CIDR.
type Target struct {
	Machine    *MachineRouteTarget          `json:"machine,omitempty"`
	Gateway    *corev1.LocalObjectReference `json:"gateway,omitempty"`
	ReservedIP *corev1.LocalObjectReference `json:"reservedIP,omitempty"`
	IP         string                       `json:"ip,omitempty"`
	CIDR       string                       `json:"cidr,omitempty"`
}
