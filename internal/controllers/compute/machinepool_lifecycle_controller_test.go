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
			machinePoolKey := client.ObjectKeyFromObject(machinePool)
			Eventually(func(g Gomega) {
				err := k8sClient.Get(ctx, machinePoolKey, machinePool)
				g.Expect(err).NotTo(HaveOccurred())

				readyCondition := computev1alpha1.FindMachinePoolCondition(machinePool.Status.Conditions, computev1alpha1.MachinePoolReady)

				g.Expect(readyCondition).NotTo(BeNil())
				g.Expect(readyCondition.Status).To(Equal(corev1.ConditionUnknown))
			}).WithTimeout(2 * machinePoolLifecycleGracePeriod).Should(Succeed())

			By("verifying the Unknown status is only set once (no further status patches)")
			Expect(k8sClient.Get(ctx, machinePoolKey, machinePool)).To(Succeed())

			resourceVersion := machinePool.ResourceVersion

			Consistently(func(g Gomega) {
				err := k8sClient.Get(ctx, machinePoolKey, machinePool)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(machinePool.ResourceVersion).To(Equal(resourceVersion))
			}).WithTimeout(3 * machinePoolLifecycleGracePeriod).Should(Succeed())
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
			machinePoolKey := client.ObjectKeyFromObject(machinePool)
			Eventually(func(g Gomega) {
				err := k8sClient.Get(ctx, machinePoolKey, machinePool)
				g.Expect(err).NotTo(HaveOccurred())

				readyCondition := computev1alpha1.FindMachinePoolCondition(machinePool.Status.Conditions, computev1alpha1.MachinePoolReady)

				g.Expect(readyCondition).NotTo(BeNil())
				g.Expect(readyCondition.Status).To(Equal(corev1.ConditionUnknown))
			}).Should(Succeed())
		})

		It("should set the ready condition to Unknown (No lease)", func(ctx SpecContext) {
			By("checking that the MachinePool Ready condition is set to Unknown")
			machinePoolKey := client.ObjectKeyFromObject(machinePool)
			Eventually(func(g Gomega) {
				err := k8sClient.Get(ctx, machinePoolKey, machinePool)
				g.Expect(err).NotTo(HaveOccurred())

				readyCondition := computev1alpha1.FindMachinePoolCondition(machinePool.Status.Conditions, computev1alpha1.MachinePoolReady)

				g.Expect(readyCondition).NotTo(BeNil())
				g.Expect(readyCondition.Status).To(Equal(corev1.ConditionUnknown))
			}).Should(Succeed())
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
						fresh := &coordinationv1.Lease{}
						if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(lease), fresh); err != nil {
							return
						}
						fresh.Spec.RenewTime = ptr.To(metav1.NowMicro())
						_ = k8sClient.Update(ctx, fresh)
					}
				}
			}()
			DeferCleanup(func(_ SpecContext) {
				close(stopCh)
				<-done
			})

			By("verifying the ready condition never becomes Unknown over 15 seconds")
			machinePoolKey := client.ObjectKeyFromObject(machinePool)
			Consistently(func(g Gomega) {
				err := k8sClient.Get(ctx, machinePoolKey, machinePool)
				g.Expect(err).NotTo(HaveOccurred())

				readyCondition := computev1alpha1.FindMachinePoolCondition(machinePool.Status.Conditions, computev1alpha1.MachinePoolReady)
				if readyCondition != nil {
					g.Expect(readyCondition.Status).NotTo(Equal(corev1.ConditionUnknown))
				}
			}).WithTimeout(3 * machinePoolLifecycleGracePeriod).Should(Succeed())
		})

	})

	Context("when the ready condition progresses", func() {
		DescribeTable("should not flip a freshly-progressing ready condition to Unknown",
			func(ctx SpecContext, initialStatus corev1.ConditionStatus, nextStatus corev1.ConditionStatus) {
				By("setting the initial ready condition on the machine pool")
				machinePoolKey := client.ObjectKeyFromObject(machinePool)
				patchReadyCondition(ctx, machinePoolKey, initialStatus, "Initial", "initial state")

				By("continuously refreshing the ready condition well within the grace period")
				stop := startReadyConditionRenewer(ctx, machinePoolKey, nextStatus)
				DeferCleanup(stop)

				By("verifying the ready condition never becomes Unknown")
				Consistently(func(g Gomega) {
					pool := &computev1alpha1.MachinePool{}
					g.Expect(k8sClient.Get(ctx, machinePoolKey, pool)).To(Succeed())

					readyCondition := computev1alpha1.FindMachinePoolCondition(pool.Status.Conditions, computev1alpha1.MachinePoolReady)
					g.Expect(readyCondition).NotTo(BeNil())
					g.Expect(readyCondition.Status).NotTo(Equal(corev1.ConditionUnknown))
				}).WithTimeout(3 * machinePoolLifecycleGracePeriod).Should(Succeed())
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
			machinePoolKey := client.ObjectKeyFromObject(machinePool)
			patchReadyCondition(ctx, machinePoolKey, corev1.ConditionTrue, "Healthy", "machinepoollet healthy")
			stop := startReadyConditionRenewer(ctx, machinePoolKey, corev1.ConditionTrue)
			DeferCleanup(stop)

			By("verifying the ready condition never becomes Unknown")
			Consistently(func(g Gomega) {
				pool := &computev1alpha1.MachinePool{}
				g.Expect(k8sClient.Get(ctx, machinePoolKey, pool)).To(Succeed())

				readyCondition := computev1alpha1.FindMachinePoolCondition(pool.Status.Conditions, computev1alpha1.MachinePoolReady)
				g.Expect(readyCondition).NotTo(BeNil())
				g.Expect(readyCondition.Status).NotTo(Equal(corev1.ConditionUnknown))
			}).WithTimeout(3 * machinePoolLifecycleGracePeriod).Should(Succeed())
		})

		It("should keep the ready condition healthy when the ready condition is stale but the lease is renewed", func(ctx SpecContext) {
			By("seeding a stale ready condition on the machine pool")
			machinePoolKey := client.ObjectKeyFromObject(machinePool)
			patchReadyCondition(ctx, machinePoolKey, corev1.ConditionTrue, "Healthy", "machinepoollet healthy")

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
			stop := startLeaseRenewer(ctx, client.ObjectKeyFromObject(lease))
			DeferCleanup(stop)

			By("verifying the ready condition never becomes Unknown")
			Consistently(func(g Gomega) {
				pool := &computev1alpha1.MachinePool{}
				g.Expect(k8sClient.Get(ctx, machinePoolKey, pool)).To(Succeed())

				readyCondition := computev1alpha1.FindMachinePoolCondition(pool.Status.Conditions, computev1alpha1.MachinePoolReady)
				g.Expect(readyCondition).NotTo(BeNil())
				g.Expect(readyCondition.Status).NotTo(Equal(corev1.ConditionUnknown))
			}).WithTimeout(3 * machinePoolLifecycleGracePeriod).Should(Succeed())
		})
	})

	Context("when a fresh signal arrives after the controller marked the pool Unknown", func() {
		It("should stop patching the status once the machinepoollet posts a fresh ready condition", func(ctx SpecContext) {
			By("waiting for the controller to set the ready condition to Unknown (no lease, no progress)")
			machinePoolKey := client.ObjectKeyFromObject(machinePool)
			Eventually(func(g Gomega) {
				pool := &computev1alpha1.MachinePool{}
				g.Expect(k8sClient.Get(ctx, machinePoolKey, pool)).To(Succeed())

				readyCondition := computev1alpha1.FindMachinePoolCondition(pool.Status.Conditions, computev1alpha1.MachinePoolReady)
				g.Expect(readyCondition).NotTo(BeNil())
				g.Expect(readyCondition.Status).To(Equal(corev1.ConditionUnknown))
			}).WithTimeout(2 * machinePoolLifecycleGracePeriod).Should(Succeed())

			By("posting a fresh ready=True condition as the machinepoollet would")
			patchReadyCondition(ctx, machinePoolKey, corev1.ConditionTrue, "Healthy", "machinepoollet recovered")

			By("verifying the controller does not flip the fresh ready condition back to Unknown within the grace period")
			Consistently(func(g Gomega) {
				current := &computev1alpha1.MachinePool{}
				g.Expect(k8sClient.Get(ctx, machinePoolKey, current)).To(Succeed())

				readyCondition := computev1alpha1.FindMachinePoolCondition(current.Status.Conditions, computev1alpha1.MachinePoolReady)
				g.Expect(readyCondition).NotTo(BeNil())
				g.Expect(readyCondition.Status).NotTo(Equal(corev1.ConditionUnknown))
			}).WithTimeout(machinePoolLifecycleGracePeriod / 2).Should(Succeed())
		})
	})
})

func patchReadyCondition(ctx SpecContext, key client.ObjectKey, status corev1.ConditionStatus, reason, message string) {
	GinkgoHelper()
	pool := &computev1alpha1.MachinePool{}
	Expect(k8sClient.Get(ctx, key, pool)).To(Succeed())

	patch := client.MergeFrom(pool.DeepCopy())
	pool.Status.Conditions = computev1alpha1.SetMachinePoolCondition(pool.Status.Conditions, computev1alpha1.MachinePoolCondition{
		Type:    computev1alpha1.MachinePoolReady,
		Status:  status,
		Reason:  reason,
		Message: message,
	})
	Expect(k8sClient.Status().Patch(ctx, pool, patch)).To(Succeed())
}

func startReadyConditionRenewer(ctx SpecContext, key client.ObjectKey, status corev1.ConditionStatus) func(SpecContext) {
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
				pool := &computev1alpha1.MachinePool{}
				if err := k8sClient.Get(ctx, key, pool); err != nil {
					return
				}
				patch := client.MergeFrom(pool.DeepCopy())
				counter++
				pool.Status.Conditions = computev1alpha1.SetMachinePoolCondition(pool.Status.Conditions, computev1alpha1.MachinePoolCondition{
					Type:    computev1alpha1.MachinePoolReady,
					Status:  status,
					Reason:  "MachinePoolReadyChanged",
					Message: fmt.Sprintf("machinepool ready changed: %s %d", status, counter),
				})
				_ = k8sClient.Status().Patch(ctx, pool, patch)
			}
		}
	}()
	return func(_ SpecContext) {
		close(stopCh)
		<-done
	}
}

func startLeaseRenewer(ctx SpecContext, key client.ObjectKey) func(SpecContext) {
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
				fresh := &coordinationv1.Lease{}
				if err := k8sClient.Get(ctx, key, fresh); err != nil {
					return
				}
				fresh.Spec.RenewTime = ptr.To(metav1.NowMicro())
				_ = k8sClient.Update(ctx, fresh)
			}
		}
	}()
	return func(_ SpecContext) {
		close(stopCh)
		<-done
	}
}
