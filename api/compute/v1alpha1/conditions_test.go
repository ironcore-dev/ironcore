// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package v1alpha1_test

import (
	"time"

	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Conditions", func() {
	DescribeTable("FindMachinePoolCondition",
		func(conds []computev1alpha1.MachinePoolCondition, condType computev1alpha1.MachinePoolConditionType, match types.GomegaMatcher) {
			Expect(computev1alpha1.FindMachinePoolCondition(conds, condType)).To(match)
		},
		Entry("returns the matching condition",
			[]computev1alpha1.MachinePoolCondition{
				{Type: "Other", Status: corev1.ConditionTrue},
				{Type: computev1alpha1.MachinePoolReady, Status: corev1.ConditionFalse, Reason: "X"},
			},
			computev1alpha1.MachinePoolReady,
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":   Equal(computev1alpha1.MachinePoolReady),
				"Reason": Equal("X"),
			})),
		),
		Entry("returns nil when no condition of the given type is present",
			[]computev1alpha1.MachinePoolCondition{{Type: "Other"}},
			computev1alpha1.MachinePoolReady,
			BeNil(),
		),
		Entry("returns nil for an empty slice",
			[]computev1alpha1.MachinePoolCondition{},
			computev1alpha1.MachinePoolReady,
			BeNil(),
		),
	)

	Describe("SetMachinePoolCondition", func() {
		It("should append the condition when it is absent", func() {
			out := computev1alpha1.SetMachinePoolCondition(nil, computev1alpha1.MachinePoolCondition{
				Type:   computev1alpha1.MachinePoolReady,
				Status: corev1.ConditionTrue,
			})

			Expect(out).To(HaveLen(1))
			Expect(out[0].Type).To(Equal(computev1alpha1.MachinePoolReady))

			By("setting LastTransitionTime on first append")
			Expect(out[0].LastTransitionTime.IsZero()).To(BeFalse())

			By("setting LastUpdateTime on first append")
			Expect(out[0].LastUpdateTime.IsZero()).To(BeFalse())
		})

		It("should update in place without advancing LastTransitionTime when status is unchanged", func() {
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

			By("preserving LastTransitionTime when status is unchanged")
			Expect(out[0].LastTransitionTime.Equal(&earlier)).To(BeTrue())

			By("advancing LastUpdateTime")
			Expect(out[0].LastUpdateTime.IsZero()).To(BeFalse())
		})

		It("should advance LastTransitionTime when the status changes", func() {
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

			Expect(out[0].LastTransitionTime.Equal(&earlier)).To(BeFalse())
		})
	})
})
