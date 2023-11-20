// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package compute

import (
	"github.com/ironcore-dev/ironcore/internal/apis/core"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient
// +genClient:nonNamespaced
// +genClient:noStatus

// MachineClass is the Schema for the machineclasses API
type MachineClass struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	// Capabilities describes the resources a machine class can provide.
	Capabilities core.ResourceList
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MachineClassList contains a list of MachineClass
type MachineClassList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []MachineClass
}
