/*
 * Copyright (c) 2021 by the OnMetal authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package v1alpha1

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var _ = Describe("machine validation webhook", func() {
	ns := SetupTest()
	Context("machine update", func() {
		It("should error if the machine pool is set and changed", func() {
			machine := &Machine{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    ns.Name,
					GenerateName: "machine-",
				},
				Spec: MachineSpec{
					MachinePool:  corev1.LocalObjectReference{Name: "my-pool"},
					MachineClass: corev1.LocalObjectReference{Name: "my-class"},
				},
			}
			Expect(k8sClient.Create(ctx, machine)).To(Succeed())

			machine.Spec.MachinePool.Name = "my-other-pool"
			path := field.NewPath("spec", "machinePool", "name")
			fieldErr := field.Invalid(path, "my-other-pool", "immutable")
			fieldErrList := field.ErrorList{fieldErr}
			Expect(k8sClient.Update(ctx, machine)).To(SatisfyAll(
				HaveOccurred(),
				WithTransform(func(err error) string {
					return err.Error()
				}, ContainSubstring(fieldErrList.ToAggregate().Error())),
			))
		})

		It("should allow updating from a zero machine pool ref to a valid ref", func() {
			By("creating a machine")
			machine := &Machine{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    ns.Name,
					GenerateName: "machine-",
				},
				Spec: MachineSpec{
					MachineClass: corev1.LocalObjectReference{Name: "my-class"},
				},
			}
			Expect(k8sClient.Create(ctx, machine)).To(Succeed())

			By("updating the machine pool")
			machine.Spec.MachinePool.Name = "my-pool"
			Expect(k8sClient.Update(ctx, machine)).To(Succeed())
		})

		It("should allow other updates", func() {
			machine := &Machine{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    ns.Name,
					GenerateName: "machine-",
				},
			}
			Expect(k8sClient.Create(ctx, machine)).To(Succeed())

			machine.Spec.SecurityGroups = []corev1.LocalObjectReference{{Name: "new-security-group"}}
			err := k8sClient.Update(ctx, machine)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
