// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package v1alpha1_test

import (
	"testing"
	"time"

	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFindMachinePoolCondition_returnsMatch(t *testing.T) {
	conds := []computev1alpha1.MachinePoolCondition{
		{Type: "Other", Status: corev1.ConditionTrue},
		{Type: computev1alpha1.MachinePoolReady, Status: corev1.ConditionFalse, Reason: "X"},
	}
	got := computev1alpha1.FindMachinePoolCondition(conds, computev1alpha1.MachinePoolReady)
	if got == nil || got.Reason != "X" {
		t.Fatalf("expected Ready condition with reason X, got %+v", got)
	}
}

func TestFindMachinePoolCondition_returnsNilWhenMissing(t *testing.T) {
	conds := []computev1alpha1.MachinePoolCondition{{Type: "Other"}}
	if got := computev1alpha1.FindMachinePoolCondition(conds, computev1alpha1.MachinePoolReady); got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}

func TestSetMachinePoolCondition_appendsWhenAbsent(t *testing.T) {
	out := computev1alpha1.SetMachinePoolCondition(nil, computev1alpha1.MachinePoolCondition{
		Type:   computev1alpha1.MachinePoolReady,
		Status: corev1.ConditionTrue,
	})
	if len(out) != 1 || out[0].Type != computev1alpha1.MachinePoolReady {
		t.Fatalf("expected Ready appended, got %+v", out)
	}
	if out[0].LastTransitionTime.IsZero() {
		t.Fatal("expected LastTransitionTime to be set on first append")
	}
	if out[0].LastUpdateTime.IsZero() {
		t.Fatal("expected LastUpdateTime to be set on first append")
	}
}

func TestSetMachinePoolCondition_updatesInPlaceWithoutTransition(t *testing.T) {
	earlier := metav1.NewTime(time.Now().Add(-time.Hour))
	in := []computev1alpha1.MachinePoolCondition{{
		Type:               computev1alpha1.MachinePoolReady,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: earlier,
	}}
	out := computev1alpha1.SetMachinePoolCondition(in, computev1alpha1.MachinePoolCondition{
		Type:    computev1alpha1.MachinePoolReady,
		Status:  corev1.ConditionTrue, // same status
		Message: "still ready",
	})
	if !out[0].LastTransitionTime.Equal(&earlier) {
		t.Fatalf("expected LastTransitionTime preserved when status unchanged, got %v", out[0].LastTransitionTime)
	}
	if out[0].LastUpdateTime.IsZero() {
		t.Fatal("expected LastUpdateTime advanced")
	}
}

func TestSetMachinePoolCondition_advancesTransitionWhenStatusChanges(t *testing.T) {
	earlier := metav1.NewTime(time.Now().Add(-time.Hour))
	in := []computev1alpha1.MachinePoolCondition{{
		Type:               computev1alpha1.MachinePoolReady,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: earlier,
	}}
	out := computev1alpha1.SetMachinePoolCondition(in, computev1alpha1.MachinePoolCondition{
		Type:   computev1alpha1.MachinePoolReady,
		Status: corev1.ConditionFalse,
	})
	if out[0].LastTransitionTime.Equal(&earlier) {
		t.Fatal("expected LastTransitionTime to advance when status changes")
	}
}
