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
	"sigs.k8s.io/controller-runtime/pkg/client"

	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	networkv1alpha1 "github.com/onmetal/onmetal-api/apis/network/v1alpha1"

	. "github.com/onsi/ginkgo"

	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("MachineReconciler", func() {
	Context("Reconcile a Machine", func() {
		ns := SetupTest(ctx)

		It("should delete unused IPAMRanges for deleted interfaces", func() {
			if1 := computev1alpha1.Interface{Name: "test-if1"}
			if2 := computev1alpha1.Interface{Name: "test-if2"}
			machine := &computev1alpha1.Machine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "with-interfaces",
					Namespace: ns.Name,
				},
				Spec: computev1alpha1.MachineSpec{
					Interfaces: []computev1alpha1.Interface{if1, if2},
				},
			}

			By("creating the machine")
			Expect(k8sClient.Create(ctx, machine)).To(Succeed())

			By("checking that IPAMRanges are created")
			Eventually(func(g Gomega) {
				key1 := types.NamespacedName{
					Name:      computev1alpha1.MachineInterfaceIPAMRangeName(machine.Name, if1.Name),
					Namespace: ns.Name,
				}
				obj1 := &networkv1alpha1.IPAMRange{}
				err := k8sClient.Get(ctx, key1, obj1)
				Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
				g.Expect(err).NotTo(HaveOccurred())

				key2 := types.NamespacedName{
					Name:      computev1alpha1.MachineInterfaceIPAMRangeName(machine.Name, if2.Name),
					Namespace: ns.Name,
				}
				obj2 := &networkv1alpha1.IPAMRange{}
				err = k8sClient.Get(ctx, key2, obj2)
				Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
				g.Expect(err).NotTo(HaveOccurred())
			}, timeout, interval).Should(Succeed())

			By("checking the IPAMRange associated with deleted interface is also deleted")
			updMachine := &computev1alpha1.Machine{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      machine.Name,
				Namespace: machine.Namespace,
			}, updMachine)).ToNot(HaveOccurred())
			Expect(updMachine.Spec.Interfaces).To(Equal([]computev1alpha1.Interface{if1, if2}))
			updMachine.Spec.Interfaces = []computev1alpha1.Interface{if2}
			Expect(k8sClient.Update(ctx, updMachine)).To(Succeed())

			Eventually(func(g Gomega) {
				// One IPAMRange should be deleted
				key1 := types.NamespacedName{
					Name:      computev1alpha1.MachineInterfaceIPAMRangeName(machine.Name, if1.Name),
					Namespace: ns.Name,
				}
				obj1 := &networkv1alpha1.IPAMRange{}
				g.Expect(errors.IsNotFound(k8sClient.Get(ctx, key1, obj1))).To(BeTrue(), "IsNotFound error expected")

				// Another one should still exist
				key2 := types.NamespacedName{
					Name:      computev1alpha1.MachineInterfaceIPAMRangeName(machine.Name, if2.Name),
					Namespace: ns.Name,
				}
				obj2 := &networkv1alpha1.IPAMRange{}
				err := k8sClient.Get(ctx, key2, obj2)
				Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
				g.Expect(err).NotTo(HaveOccurred())
			}, timeout, interval).Should(Succeed())
		})
	})
})
