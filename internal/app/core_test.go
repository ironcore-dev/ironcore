// Copyright 2023 OnMetal authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package app_test

import (
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	corev1alpha1 "github.com/onmetal/onmetal-api/api/core/v1alpha1"
	. "github.com/onmetal/onmetal-api/utils/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Core", func() {
	ctx := SetupContext()
	ns, machineClass := SetupTest(ctx)

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

			By("getting the resource quota")
			resourceQuotaKey := client.ObjectKeyFromObject(resourceQuota)
			Expect(k8sClient.Get(ctx, resourceQuotaKey, resourceQuota)).To(Succeed())

			By("inspecting the resource quota")
			Expect(resourceQuota.Status.Used).To(Equal(corev1alpha1.ResourceList{
				corev1alpha1.ResourceRequestsCPU: resource.MustParse("1"),
			}))
		})
	})
})
