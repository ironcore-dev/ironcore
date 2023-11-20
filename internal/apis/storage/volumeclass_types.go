// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"github.com/ironcore-dev/ironcore/internal/apis/core"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient
// +genclient:nonNamespaced
// +genclient:noStatus

// VolumeClass is the Schema for the volumeclasses API
type VolumeClass struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	// Capabilities describes the capabilities of a volume class
	Capabilities core.ResourceList
	// ResizePolicy describes the supported expansion policy of a VolumeClass.
	// If not set default to Static expansion policy.
	ResizePolicy ResizePolicy
}

// ResizePolicy is a type of policy.
type ResizePolicy string

const (
	// ResizePolicyStatic is a policy that does not allow the expansion of a Volume.
	ResizePolicyStatic ResizePolicy = "Static"
	// ResizePolicyExpandOnly is a policy that only allows the expansion of a Volume.
	ResizePolicyExpandOnly ResizePolicy = "ExpandOnly"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VolumeClassList contains a list of VolumeClass
type VolumeClassList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []VolumeClass
}
