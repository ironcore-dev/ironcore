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
	Context("upon machine update", func() {
		It("signals error if machinepool is changed", func() {
			m := &Machine{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    ns.Name,
					GenerateName: "machine-",
				},
			}
			Expect(k8sClient.Create(ctx, m)).To(Succeed())

			newMachinePoolName := "another machinepool"
			m.Spec.MachinePool.Name = newMachinePoolName
			err := k8sClient.Update(ctx, m)
			Expect(err).To(HaveOccurred())

			path := field.NewPath("spec", "machinePool", "name")
			fieldErr := field.Invalid(path, newMachinePoolName, "machinepool should be immutable")
			fieldErrList := field.ErrorList{fieldErr}
			Expect(err.Error()).To(ContainSubstring(fieldErrList.ToAggregate().Error()))
		})

		It("keeps silent when machinepool isn't changed", func() {
			m := &Machine{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    ns.Name,
					GenerateName: "machine-",
				},
			}
			Expect(k8sClient.Create(ctx, m)).To(Succeed())

			m.Spec.SecurityGroups = []corev1.LocalObjectReference{{Name: "new-security-group"}}
			err := k8sClient.Update(ctx, m)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
