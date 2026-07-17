// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"slices"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FindMachinePoolCondition returns a pointer to the condition of the given type,
// or nil if no condition of that type is present.
func FindMachinePoolCondition(conditions []MachinePoolCondition, typ MachinePoolConditionType) *MachinePoolCondition {
	idx := slices.IndexFunc(conditions, func(cond MachinePoolCondition) bool {
		return cond.Type == typ
	})
	if idx < 0 {
		return nil
	}
	return &conditions[idx]
}

// SetMachinePoolCondition inserts or updates a condition of the given type in the
// conditions slice. LastUpdateTime is always set to now. LastTransitionTime is set
// to now only when the condition is newly inserted or its Status differs from the
// previous value.
func SetMachinePoolCondition(conditions []MachinePoolCondition, cond MachinePoolCondition) []MachinePoolCondition {
	idx := slices.IndexFunc(conditions, func(c MachinePoolCondition) bool {
		return c.Type == cond.Type
	})

	cond.LastUpdateTime = metav1.Now()

	if idx < 0 || conditions[idx].Status != cond.Status {
		cond.LastTransitionTime = metav1.Now()
	} else {
		cond.LastTransitionTime = conditions[idx].LastTransitionTime
	}

	if idx < 0 {
		return append(conditions, cond)
	}
	conditions[idx] = cond
	return conditions
}
