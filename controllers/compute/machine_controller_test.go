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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	networkv1alpha1 "github.com/onmetal/onmetal-api/apis/network/v1alpha1"

	. "github.com/onsi/ginkgo"

	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("MachineReconciler", func() {
	Context("Reconcile a Machine", func() {
		ns := SetupTest(ctx)

		It("delete interface", func() {
			machine := &computev1alpha1.Machine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "with-interfaces",
					Namespace: ns.Name,
				},
				Spec: computev1alpha1.MachineSpec{
					Interfaces: []computev1alpha1.Interface{
						{Name: "test1"},
						{Name: "test2"},
					},
				},
			}

			By("creating the machine")
			Expect(k8sClient.Create(ctx, machine)).To(Succeed())

			By("check IPAMRanges are created")
			Eventually(func(g Gomega) {
				key1 := types.NamespacedName{
					Name:      computev1alpha1.MachineInterfaceIPAMRangeName(machine.Name, "test1"),
					Namespace: ns.Name,
				}
				obj1 := &networkv1alpha1.IPAMRange{}
				g.Expect(k8sClient.Get(ctx, key1, obj1)).To(Succeed())

				key2 := types.NamespacedName{
					Name:      computev1alpha1.MachineInterfaceIPAMRangeName(machine.Name, "test2"),
					Namespace: ns.Name,
				}
				obj2 := &networkv1alpha1.IPAMRange{}
				g.Expect(k8sClient.Get(ctx, key2, obj2)).To(Succeed())

				// Delete interface
				updMachine := &computev1alpha1.Machine{}
				Expect(k8sClient.Get(ctx, types.NamespacedName{
					Name:      machine.Name,
					Namespace: machine.Namespace,
				}, updMachine)).ToNot(HaveOccurred())

				for _, i := range updMachine.Spec.Interfaces {
					if i.Name == "test2" {
						updMachine.Spec.Interfaces = []computev1alpha1.Interface{i}
					}
				}
				Expect(k8sClient.Update(ctx, updMachine)).To(Succeed())
			}, timeout, interval).Should(Succeed())

			By("check IPAMRange associated with deleted interface is also deleted")
			Eventually(func(g Gomega) {
				// One IPAMRange should be deleted
				key1 := types.NamespacedName{
					Name:      computev1alpha1.MachineInterfaceIPAMRangeName(machine.Name, "test1"),
					Namespace: ns.Name,
				}
				obj1 := &networkv1alpha1.IPAMRange{}
				g.Expect(errors.IsNotFound(k8sClient.Get(ctx, key1, obj1))).To(BeTrue())

				// Another one should still exist
				key2 := types.NamespacedName{
					Name:      computev1alpha1.MachineInterfaceIPAMRangeName(machine.Name, "test2"),
					Namespace: ns.Name,
				}
				obj2 := &networkv1alpha1.IPAMRange{}
				g.Expect(k8sClient.Get(ctx, key2, obj2)).To(Succeed())
			}, timeout, interval).Should(Succeed())
		})
	})
})
