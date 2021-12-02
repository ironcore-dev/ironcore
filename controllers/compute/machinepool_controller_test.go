///*
// * Copyright (c) 2021 by the OnMetal authors.
// *
// * Licensed under the Apache License, Version 2.0 (the "License");
// * you may not use this file except in compliance with the License.
// * You may obtain a copy of the License at
// *
// *     http://www.apache.org/licenses/LICENSE-2.0
// *
// * Unless required by applicable law or agreed to in writing, software
// * distributed under the License is distributed on an "AS IS" BASIS,
// * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// * See the License for the specific language governing permissions and
// * limitations under the License.
// */
//
package compute

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
)

var _ = Describe("MachinePoolReconciler", func() {
	Context("Reconcile a MachinePool", func() {
		ns := SetupTest(ctx)

		It("should set state as Pending when Ready condition is not present", func() {
			machinePool := &computev1alpha1.MachinePool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "no-ready-condition",
					Namespace: ns.Name,
				},
			}
			Expect(k8sClient.Create(ctx, machinePool)).To(Succeed())

			machinePool.Status.Conditions = []computev1alpha1.MachinePoolCondition{
				{
					Type:               computev1alpha1.MachinePoolConditionTypeReady,
					Status:             corev1.ConditionFalse,
					LastUpdateTime:     metav1.Time{Time: time.Now().Add(time.Duration(-1) * machinePoolGracePeriod)},
					LastTransitionTime: metav1.Now(),
				},
			}
			Expect(k8sClient.Status().Update(ctx, machinePool)).To(Succeed())

			By("checking that MachinePool is in pending state")
			Eventually(func(g Gomega) {
				key := types.NamespacedName{
					Name:      machinePool.Name,
					Namespace: ns.Name,
				}
				obj := &computev1alpha1.MachinePool{}
				err := k8sClient.Get(ctx, key, obj)
				Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
				g.Expect(err).NotTo(HaveOccurred())

				g.Expect(obj.Status.State).To(Equal(computev1alpha1.MachinePoolStatePending))
			}, timeout, interval).Should(Succeed())
		})
		It("should set state as Pending when Ready condition is outdated", func() {
			machinePool := &computev1alpha1.MachinePool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ready-out-of-date",
					Namespace: ns.Name,
				},
			}
			Expect(k8sClient.Create(ctx, machinePool)).To(Succeed())

			machinePool.Status.Conditions = []computev1alpha1.MachinePoolCondition{
				{
					Type:               computev1alpha1.MachinePoolConditionTypeReady,
					Status:             corev1.ConditionTrue,
					LastUpdateTime:     metav1.Time{Time: time.Now().Add(time.Duration(-1) * machinePoolGracePeriod)},
					LastTransitionTime: metav1.Now(),
				},
			}
			Expect(k8sClient.Status().Update(ctx, machinePool)).To(Succeed())

			By("checking that MachinePool is in pending state")
			Eventually(func(g Gomega) {
				key := types.NamespacedName{
					Name:      machinePool.Name,
					Namespace: ns.Name,
				}
				obj := &computev1alpha1.MachinePool{}
				err := k8sClient.Get(ctx, key, obj)
				Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
				g.Expect(err).NotTo(HaveOccurred())

				g.Expect(obj.Status.State).To(Equal(computev1alpha1.MachinePoolStatePending))
			}, timeout, interval).Should(Succeed())
		})
		It("should set state to Ready when Ready condition is up-to-date", func() {
			machinePool := &computev1alpha1.MachinePool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ready",
					Namespace: ns.Name,
				},
			}
			Expect(k8sClient.Create(ctx, machinePool)).To(Succeed())

			machinePool.Status.Conditions = []computev1alpha1.MachinePoolCondition{
				{
					Type:               computev1alpha1.MachinePoolConditionTypeReady,
					Status:             corev1.ConditionTrue,
					LastUpdateTime:     metav1.Now(),
					LastTransitionTime: metav1.Now(),
				},
			}
			Expect(k8sClient.Status().Update(ctx, machinePool)).To(Succeed())

			By("checking that MachinePool is in ready state")
			Eventually(func(g Gomega) {
				key := types.NamespacedName{
					Name:      machinePool.Name,
					Namespace: ns.Name,
				}
				obj := &computev1alpha1.MachinePool{}
				err := k8sClient.Get(ctx, key, obj)
				Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
				g.Expect(err).NotTo(HaveOccurred())

				g.Expect(obj.Status.State).To(Equal(computev1alpha1.MachinePoolStateReady))
			}, timeout, interval).Should(Succeed())
		})
	})
})
