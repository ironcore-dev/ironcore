// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package compute

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"

	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	coordinationv1 "k8s.io/api/coordination/v1"
)

var _ = Describe("machinepool lifecycle controller", func() {
	machinePool := SetupMachinePool()

	Context("when neither lease nor ready condition see progress within the grace period", func() {
		It("should set the ready condition to Unknown (Lease w/o RenewTime) and update Unknown only once", func(ctx SpecContext) {
			By("creating a lease for the machine pool without RenewTime")
			lease := &coordinationv1.Lease{
				ObjectMeta: metav1.ObjectMeta{
					Name:      machinePool.Name,
					Namespace: "ironcore-machinepool-lease",
				},
				Spec: coordinationv1.LeaseSpec{
					HolderIdentity:       ptr.To(machinePool.Name),
					LeaseDurationSeconds: ptr.To(int32(3600)),
				},
			}
			Expect(k8sClient.Create(ctx, lease)).To(Succeed(), "failed to create lease")
			DeferCleanup(func(ctx SpecContext) {
				Expect(client.IgnoreNotFound(k8sClient.Delete(ctx, lease))).To(Succeed())
			})

			By("checking that the MachinePool Ready condition is set to Unknown")
			Eventually(ctx, Object(machinePool)).WithTimeout(2 * machinePoolLifecycleGracePeriod).Should(
				HaveField("Status.Conditions", ContainElement(SatisfyAll(
					HaveField("Type", computev1alpha1.MachinePoolReady),
					HaveField("Status", corev1.ConditionUnknown),
				))),
			)

			By("verifying the Unknown status is only set once (no further status patches)")
			resourceVersion := machinePool.ResourceVersion

			Consistently(ctx, Object(machinePool)).WithTimeout(3 * machinePoolLifecycleGracePeriod).Should(
				HaveField("ResourceVersion", Equal(resourceVersion)),
			)
		})

		It("should set the ready condition to Unknown (Lease with RenewTime)", func(ctx SpecContext) {
			By("creating a lease for the machine pool")
			lease := &coordinationv1.Lease{
				ObjectMeta: metav1.ObjectMeta{
					Name:      machinePool.Name,
					Namespace: "ironcore-machinepool-lease",
				},
				Spec: coordinationv1.LeaseSpec{
					HolderIdentity:       ptr.To(machinePool.Name),
					LeaseDurationSeconds: ptr.To(int32(3600)),
					RenewTime:            ptr.To(metav1.NowMicro()),
				},
			}
			Expect(k8sClient.Create(ctx, lease)).To(Succeed(), "failed to create lease")
			DeferCleanup(func(ctx SpecContext) {
				Expect(client.IgnoreNotFound(k8sClient.Delete(ctx, lease))).To(Succeed())
			})

			By("checking that the MachinePool Ready condition is set to Unknown")
			Eventually(ctx, Object(machinePool)).Should(
				HaveField("Status.Conditions", ContainElement(SatisfyAll(
					HaveField("Type", computev1alpha1.MachinePoolReady),
					HaveField("Status", corev1.ConditionUnknown),
				))),
			)
		})

		It("should set the ready condition to Unknown (No lease)", func(ctx SpecContext) {
			By("checking that the MachinePool Ready condition is set to Unknown")
			Eventually(ctx, Object(machinePool)).Should(
				HaveField("Status.Conditions", ContainElement(SatisfyAll(
					HaveField("Type", computev1alpha1.MachinePoolReady),
					HaveField("Status", corev1.ConditionUnknown),
				))),
			)
		})
	})

	Context("when the lease is renewed frequently", func() {
		It("should not set the ready condition to Unknown when the lease is regularly renewed", func(ctx SpecContext) {
			By("creating a lease for the machine pool with a current renewTime")
			lease := &coordinationv1.Lease{
				ObjectMeta: metav1.ObjectMeta{
					Name:      machinePool.Name,
					Namespace: "ironcore-machinepool-lease",
				},
				Spec: coordinationv1.LeaseSpec{
					HolderIdentity:       ptr.To(machinePool.Name),
					LeaseDurationSeconds: ptr.To(int32(3600)),
					RenewTime:            ptr.To(metav1.NowMicro()),
				},
			}
			Expect(k8sClient.Create(ctx, lease)).To(Succeed())
			DeferCleanup(func(ctx SpecContext) {
				Expect(client.IgnoreNotFound(k8sClient.Delete(ctx, lease))).To(Succeed())
			})

			By("continuously renewing the lease well within the grace period")
			stop := startLeaseRenewer(lease)
			DeferCleanup(stop)

			By("verifying the ready condition never becomes Unknown over 15 seconds")
			Consistently(ctx, Object(machinePool)).WithTimeout(3 * machinePoolLifecycleGracePeriod).Should(
				Not(HaveField("Status.Conditions", ContainElement(SatisfyAll(
					HaveField("Type", computev1alpha1.MachinePoolReady),
					HaveField("Status", corev1.ConditionUnknown),
				)))),
			)
		})
	})

	Context("when the ready condition progresses", func() {
		DescribeTable("should not flip a freshly-progressing ready condition to Unknown",
			func(ctx SpecContext, initialStatus corev1.ConditionStatus, nextStatus corev1.ConditionStatus) {
				By("setting the initial ready condition on the machine pool")
				patchReadyCondition(ctx, machinePool, initialStatus, "Initial", "initial state")

				By("continuously refreshing the ready condition well within the grace period")
				stop := startReadyConditionRenewer(machinePool, nextStatus)
				DeferCleanup(stop)

				By("verifying the ready condition never becomes Unknown")
				Consistently(ctx, Object(machinePool)).WithTimeout(3 * machinePoolLifecycleGracePeriod).Should(
					Not(HaveField("Status.Conditions", ContainElement(SatisfyAll(
						HaveField("Type", computev1alpha1.MachinePoolReady),
						HaveField("Status", corev1.ConditionUnknown),
					)))),
				)
			},
			Entry("True remains True", corev1.ConditionTrue, corev1.ConditionTrue),
			Entry("False remains False", corev1.ConditionFalse, corev1.ConditionFalse),
			Entry("False progressing to True", corev1.ConditionFalse, corev1.ConditionTrue),
			Entry("True progressing to False", corev1.ConditionTrue, corev1.ConditionFalse),
		)
	})

	Context("when only one health signal stays fresh", func() {
		It("should keep the ready condition healthy when the lease is stale but the ready condition is refreshed", func(ctx SpecContext) {
			By("creating a lease with a stale RenewTime")
			lease := &coordinationv1.Lease{
				ObjectMeta: metav1.ObjectMeta{
					Name:      machinePool.Name,
					Namespace: "ironcore-machinepool-lease",
				},
				Spec: coordinationv1.LeaseSpec{
					HolderIdentity:       ptr.To(machinePool.Name),
					LeaseDurationSeconds: ptr.To(int32(3600)),
					RenewTime:            ptr.To(metav1.NewMicroTime(time.Now().Add(-time.Hour))),
				},
			}
			Expect(k8sClient.Create(ctx, lease)).To(Succeed())
			DeferCleanup(func(ctx SpecContext) {
				Expect(client.IgnoreNotFound(k8sClient.Delete(ctx, lease))).To(Succeed())
			})

			By("continuously refreshing the ready condition")
			patchReadyCondition(ctx, machinePool, corev1.ConditionTrue, "Healthy", "machinepoollet healthy")
			stop := startReadyConditionRenewer(machinePool, corev1.ConditionTrue)
			DeferCleanup(stop)

			By("verifying the ready condition never becomes Unknown")
			Consistently(ctx, Object(machinePool)).WithTimeout(3 * machinePoolLifecycleGracePeriod).Should(
				Not(HaveField("Status.Conditions", ContainElement(SatisfyAll(
					HaveField("Type", computev1alpha1.MachinePoolReady),
					HaveField("Status", corev1.ConditionUnknown),
				)))),
			)
		})

		It("should keep the ready condition healthy when the ready condition is stale but the lease is renewed", func(ctx SpecContext) {
			By("seeding a stale ready condition on the machine pool")
			patchReadyCondition(ctx, machinePool, corev1.ConditionTrue, "Healthy", "machinepoollet healthy")

			By("creating a lease for the machine pool")
			lease := &coordinationv1.Lease{
				ObjectMeta: metav1.ObjectMeta{
					Name:      machinePool.Name,
					Namespace: "ironcore-machinepool-lease",
				},
				Spec: coordinationv1.LeaseSpec{
					HolderIdentity:       ptr.To(machinePool.Name),
					LeaseDurationSeconds: ptr.To(int32(3600)),
					RenewTime:            ptr.To(metav1.NowMicro()),
				},
			}
			Expect(k8sClient.Create(ctx, lease)).To(Succeed())
			DeferCleanup(func(ctx SpecContext) {
				Expect(client.IgnoreNotFound(k8sClient.Delete(ctx, lease))).To(Succeed())
			})

			By("continuously renewing the lease, while the ready condition stays untouched")
			stop := startLeaseRenewer(lease)
			DeferCleanup(stop)

			By("verifying the ready condition never becomes Unknown")
			Consistently(ctx, Object(machinePool)).WithTimeout(3 * machinePoolLifecycleGracePeriod).Should(
				Not(HaveField("Status.Conditions", ContainElement(SatisfyAll(
					HaveField("Type", computev1alpha1.MachinePoolReady),
					HaveField("Status", corev1.ConditionUnknown),
				)))),
			)
		})
	})

	Context("when a fresh signal arrives after the controller marked the pool Unknown", func() {
		It("should stop patching the status once the machinepoollet posts a fresh ready condition", func(ctx SpecContext) {
			By("waiting for the controller to set the ready condition to Unknown (no lease, no progress)")
			Eventually(ctx, Object(machinePool)).WithTimeout(2 * machinePoolLifecycleGracePeriod).Should(
				HaveField("Status.Conditions", ContainElement(SatisfyAll(
					HaveField("Type", computev1alpha1.MachinePoolReady),
					HaveField("Status", corev1.ConditionUnknown),
				))),
			)

			By("posting a fresh ready=True condition as the machinepoollet would")
			patchReadyCondition(ctx, machinePool, corev1.ConditionTrue, "Healthy", "machinepoollet recovered")

			By("verifying the controller does not flip the fresh ready condition back to Unknown within the grace period")
			Consistently(ctx, Object(machinePool)).WithTimeout(machinePoolLifecycleGracePeriod / 2).Should(
				Not(HaveField("Status.Conditions", ContainElement(SatisfyAll(
					HaveField("Type", computev1alpha1.MachinePoolReady),
					HaveField("Status", corev1.ConditionUnknown),
				)))),
			)
		})
	})
})

func patchReadyCondition(ctx SpecContext, pool *computev1alpha1.MachinePool, status corev1.ConditionStatus, reason, message string) {
	GinkgoHelper()
	Eventually(ctx, UpdateStatus(pool, func() {
		pool.Status.Conditions = computev1alpha1.SetMachinePoolCondition(pool.Status.Conditions, computev1alpha1.MachinePoolCondition{
			Type:    computev1alpha1.MachinePoolReady,
			Status:  status,
			Reason:  reason,
			Message: message,
		})
	})).Should(Succeed())
}

func startReadyConditionRenewer(pool *computev1alpha1.MachinePool, status corev1.ConditionStatus) func(SpecContext) {
	stopCh := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()
		counter := 0
		for {
			select {
			case <-stopCh:
				return
			case <-ticker.C:
				counter++
				c := counter
				_ = UpdateStatus(pool, func() {
					pool.Status.Conditions = computev1alpha1.SetMachinePoolCondition(pool.Status.Conditions, computev1alpha1.MachinePoolCondition{
						Type:               computev1alpha1.MachinePoolReady,
						Status:             status,
						Reason:             "MachinePoolReadyChanged",
						Message:            fmt.Sprintf("machinepool ready changed: %s %d", status, c),
						ObservedGeneration: int64(c),
					})
				})()
			}
		}
	}()
	return func(_ SpecContext) {
		close(stopCh)
		<-done
	}
}

func startLeaseRenewer(lease *coordinationv1.Lease) func(SpecContext) {
	stopCh := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-stopCh:
				return
			case <-ticker.C:
				_ = Update(lease, func() {
					lease.Spec.RenewTime = ptr.To(metav1.NowMicro())
				})()
			}
		}
	}()
	return func(_ SpecContext) {
		close(stopCh)
		<-done
	}
}
