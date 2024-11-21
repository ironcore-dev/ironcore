// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package networking

import (
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VirtualIPSpec defines the desired state of VirtualIP
type VirtualIPSpec struct {
	// Type is the type of VirtualIP.
	Type VirtualIPType
	// IPFamily is the ip family of the VirtualIP.
	IPFamily corev1.IPFamily

	// TargetRef references the target for this VirtualIP (currently only NetworkInterface).
	TargetRef *commonv1alpha1.LocalUIDReference
}

// VirtualIPType is a type of VirtualIP.
type VirtualIPType string

const (
	// VirtualIPTypePublic is a VirtualIP that allocates and routes a stable public IP.
	VirtualIPTypePublic VirtualIPType = "Public"
)

// ReclaimPolicyType is the ironcore ReclaimPolicy of a VirtualIP.
type ReclaimPolicyType string

const (
	// ReclaimPolicyTypeRetain is used for any VirtualIP that is retained when the claim of VirtualIP is released.
	ReclaimPolicyTypeRetain ReclaimPolicyType = "Retain"
	// ReclaimPolicyTypeDelete is used for any VirtualIP that is deleted when the claim of VirtualIP is released.
	ReclaimPolicyTypeDelete ReclaimPolicyType = "Delete"
)

// VirtualIPStatus defines the observed state of VirtualIP
type VirtualIPStatus struct {
	// IP is the allocated IP, if any.
	IP *commonv1alpha1.IP
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VirtualIP is the Schema for the virtualips API
type VirtualIP struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   VirtualIPSpec
	Status VirtualIPStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VirtualIPList contains a list of VirtualIP
type VirtualIPList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []VirtualIP
}

type EphemeralVirtualIPSpec struct {
	// VirtualIPSpec defines the desired state of VirtualIP
	VirtualIPSpec
	// ReclaimPolicy is the ReclaimPolicyType of virtualIP
	ReclaimPolicy ReclaimPolicyType `json:"reclaimPolicy,omitempty"`
}

// VirtualIPTemplateSpec is the specification of a VirtualIP template.
type VirtualIPTemplateSpec struct {
	metav1.ObjectMeta
	Spec EphemeralVirtualIPSpec
}
