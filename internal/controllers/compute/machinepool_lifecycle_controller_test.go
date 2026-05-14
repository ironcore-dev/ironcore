// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package compute

import (
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

// TODO Add tests for progressing ready conditions
var _ = FDescribe("machinepool lifecycle controller", func() {
	var machinePool *computev1alpha1.MachinePool

	BeforeEach(func(ctx SpecContext) {
		machinePool = &computev1alpha1.MachinePool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-pool-",
			},
		}
		Expect(k8sClient.Create(ctx, machinePool)).To(Succeed(), "failed to create machine pool")
	})

	Context("when neither lease nor ready condition see progress within the grace period", func() {
		It("should set the ready condition to Unknown (Lease w/o RenewTime) and update Unknwon only once", func(ctx SpecContext) {
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

				readyCondition := FindMachinePoolCondition(machinePool.Status.Conditions, computev1alpha1.MachinePoolReady)

				g.Expect(readyCondition).NotTo(BeNil())
				g.Expect(readyCondition.Status).To(Equal(corev1.ConditionUnknown))
			}).WithTimeout(30 * time.Second).Should(Succeed())

			By("verifying the Unknown status is only set once (no further status patches)")
			Expect(k8sClient.Get(ctx, machinePoolKey, machinePool)).To(Succeed())
			resourceVersion := machinePool.ResourceVersion

			Consistently(func(g Gomega) {
				err := k8sClient.Get(ctx, machinePoolKey, machinePool)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(machinePool.ResourceVersion).To(Equal(resourceVersion))
			}).WithTimeout(5 * time.Second).WithPolling(500 * time.Millisecond).Should(Succeed())
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

				readyCondition := FindMachinePoolCondition(machinePool.Status.Conditions, computev1alpha1.MachinePoolReady)

				g.Expect(readyCondition).NotTo(BeNil())
				g.Expect(readyCondition.Status).To(Equal(corev1.ConditionUnknown))
			}).WithTimeout(30 * time.Second).Should(Succeed())
		})

		It("should set the ready condition to Unknown (No lease)", func(ctx SpecContext) {
			By("checking that the MachinePool Ready condition is set to Unknown")
			machinePoolKey := client.ObjectKeyFromObject(machinePool)
			Eventually(func(g Gomega) {
				err := k8sClient.Get(ctx, machinePoolKey, machinePool)
				g.Expect(err).NotTo(HaveOccurred())

				readyCondition := FindMachinePoolCondition(machinePool.Status.Conditions, computev1alpha1.MachinePoolReady)

				g.Expect(readyCondition).NotTo(BeNil())
				g.Expect(readyCondition.Status).To(Equal(corev1.ConditionUnknown))
			}).WithTimeout(30 * time.Second).Should(Succeed())
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
				ticker := time.NewTicker(500 * time.Millisecond)
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

				readyCondition := FindMachinePoolCondition(machinePool.Status.Conditions, computev1alpha1.MachinePoolReady)
				if readyCondition != nil {
					g.Expect(readyCondition.Status).NotTo(Equal(corev1.ConditionUnknown))
				}
			}).WithTimeout(15 * time.Second).WithPolling(1 * time.Second).Should(Succeed())
		})

	})
})
