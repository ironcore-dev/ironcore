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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("machine validation webhook", func() {
	ns := SetupTest()
	Context("upon machine update", func() {
		It("validates if machinepool is changed", func() {
			m := &Machine{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    ns.Name,
					GenerateName: "machine-",
				},
			}
			Expect(k8sClient.Create(ctx, m)).To(Succeed())

			m.Spec.MachinePool.Name = "another machinepool"
			Expect(k8sClient.Update(ctx, m)).NotTo(Succeed())
		})
	})
})
