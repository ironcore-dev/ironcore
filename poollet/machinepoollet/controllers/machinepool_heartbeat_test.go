// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package controllers_test

import (
	"errors"

	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	"github.com/ironcore-dev/ironcore/poollet/machinepoollet/controllers"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("ComputeReadyCondition", func() {
	It("returns Ready=True with HeartbeatReceived when the probe succeeds", func() {
		got := controllers.ComputeReadyCondition(int64(7), nil)
		Expect(got.Type).To(Equal(computev1alpha1.MachinePoolReady))
		Expect(got.Status).To(Equal(corev1.ConditionTrue))
		Expect(got.Reason).To(Equal("HeartbeatReceived"))
		Expect(got.ObservedGeneration).To(Equal(int64(7)))
	})

	It("returns Ready=False with RuntimeUnreachable when the probe errors", func() {
		got := controllers.ComputeReadyCondition(int64(3), errors.New("boom"))
		Expect(got.Status).To(Equal(corev1.ConditionFalse))
		Expect(got.Reason).To(Equal("RuntimeUnreachable"))
		Expect(got.Message).To(Equal("boom"))
		Expect(got.ObservedGeneration).To(Equal(int64(3)))
	})
})

var _ = Describe("ReadyConditionsDiffer", func() {
	base := computev1alpha1.MachinePoolCondition{
		Type:               computev1alpha1.MachinePoolReady,
		Status:             corev1.ConditionTrue,
		Reason:             "HeartbeatReceived",
		Message:            "ok",
		ObservedGeneration: 5,
	}

	It("treats a nil existing as a diff", func() {
		desired := computev1alpha1.MachinePoolCondition{
			Type:   computev1alpha1.MachinePoolReady,
			Status: corev1.ConditionTrue,
		}
		Expect(controllers.ReadyConditionsDiffer(nil, desired)).To(BeTrue())
	})

	It("returns false for identical conditions", func() {
		Expect(controllers.ReadyConditionsDiffer(&base, base)).To(BeFalse())
	})

	It("ignores LastUpdateTime and LastTransitionTime", func() {
		desired := base // value copy; timestamps stay zero
		Expect(controllers.ReadyConditionsDiffer(&base, desired)).To(BeFalse())
	})

	DescribeTable("reports a diff when a meaningful field changes",
		func(modify func(*computev1alpha1.MachinePoolCondition)) {
			desired := base
			modify(&desired)
			Expect(controllers.ReadyConditionsDiffer(&base, desired)).To(BeTrue())
		},
		Entry("status", func(c *computev1alpha1.MachinePoolCondition) { c.Status = corev1.ConditionFalse }),
		Entry("reason", func(c *computev1alpha1.MachinePoolCondition) { c.Reason = "Other" }),
		Entry("message", func(c *computev1alpha1.MachinePoolCondition) { c.Message = "different" }),
		Entry("observedGeneration", func(c *computev1alpha1.MachinePoolCondition) { c.ObservedGeneration = 6 }),
	)
})

var _ = Describe("MachinePoolReadyState", func() {
	It("reports hasResult=false until the first Set", func() {
		s := controllers.NewMachinePoolReadyState()
		err, hasResult := s.Get()
		Expect(hasResult).To(BeFalse())
		Expect(err).ToNot(HaveOccurred())
	})

	It("reports changed=true on the first Set, even with a nil error", func() {
		s := controllers.NewMachinePoolReadyState()
		Expect(s.Set(nil)).To(BeTrue())
		err, hasResult := s.Get()
		Expect(hasResult).To(BeTrue())
		Expect(err).ToNot(HaveOccurred())
	})

	It("reports changed=false when the error string is unchanged", func() {
		s := controllers.NewMachinePoolReadyState()
		Expect(s.Set(nil)).To(BeTrue())
		Expect(s.Set(nil)).To(BeFalse())

		Expect(s.Set(errors.New("boom"))).To(BeTrue())
		Expect(s.Set(errors.New("boom"))).To(BeFalse())
	})

	It("reports changed=true when the error string changes", func() {
		s := controllers.NewMachinePoolReadyState()
		Expect(s.Set(errors.New("first"))).To(BeTrue())
		Expect(s.Set(errors.New("second"))).To(BeTrue())
	})

	It("reports changed=true when transitioning from success to failure and back", func() {
		s := controllers.NewMachinePoolReadyState()
		Expect(s.Set(nil)).To(BeTrue())
		Expect(s.Set(errors.New("boom"))).To(BeTrue())
		Expect(s.Set(nil)).To(BeTrue())
	})
})
