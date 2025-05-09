// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package app_test

import (
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	. "github.com/ironcore-dev/ironcore/utils/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Core", func() {
	var (
		ctx          = SetupContext()
		ns           = SetupTest(ctx)
		machineClass = &computev1alpha1.MachineClass{}
	)

	BeforeEach(func() {
		*machineClass = computev1alpha1.MachineClass{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "machine-class-",
			},
			Capabilities: corev1alpha1.ResourceList{
				corev1alpha1.ResourceCPU:    resource.MustParse("1"),
				corev1alpha1.ResourceMemory: resource.MustParse("1Gi"),
			},
		}
		Expect(k8sClient.Create(ctx, machineClass)).To(Succeed(), "failed to create test machine class")
		DeferCleanup(k8sClient.Delete, ctx, machineClass)
	})

	Context("ResourceQuota", func() {
		It("should update the resource quota on creation of a new entity", func() {
			By("creating a resource quota")
			resourceQuota := &corev1alpha1.ResourceQuota{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    ns.Name,
					GenerateName: "resource-quota-",
				},
				Spec: corev1alpha1.ResourceQuotaSpec{
					Hard: corev1alpha1.ResourceList{
						corev1alpha1.ResourceRequestsCPU: resource.MustParse("2"),
					},
				},
			}
			Expect(k8sClient.Create(ctx, resourceQuota)).To(Succeed())

			By("manually updating the resource quota status")
			baseResourceQuota := resourceQuota.DeepCopy()
			resourceQuota.Status.Hard = resourceQuota.Spec.Hard
			resourceQuota.Status.Used = corev1alpha1.ResourceList{
				corev1alpha1.ResourceRequestsCPU: resource.MustParse("0"),
			}
			Expect(k8sClient.Status().Patch(ctx, resourceQuota, client.MergeFrom(baseResourceQuota))).To(Succeed())

			By("creating a machine")
			machine := &computev1alpha1.Machine{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    ns.Name,
					GenerateName: "machine-",
				},
				Spec: computev1alpha1.MachineSpec{
					MachineClassRef: corev1.LocalObjectReference{Name: machineClass.Name},
				},
			}
			Expect(k8sClient.Create(ctx, machine)).To(Succeed())

			By("waiting for the resource quota to be updated")
			resourceQuotaKey := client.ObjectKeyFromObject(resourceQuota)
			Eventually(ctx, func(g Gomega) {
				Expect(k8sClient.Get(ctx, resourceQuotaKey, resourceQuota)).To(Succeed())
				g.Expect(resourceQuota.Status.Used).To(Equal(corev1alpha1.ResourceList{
					corev1alpha1.ResourceRequestsCPU: resource.MustParse("1"),
				}))
			}).Should(Succeed())
		})
	})
})
