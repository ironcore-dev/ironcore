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

package compute

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
)

var _ = Describe("machineclass controller", func() {
	ns := SetupTest(ctx)
	It("removes the finalizer from machineclass only if there's no machine still using the machineclass", func() {
		By("creating the machineclass consumed by the machine")
		machineClass := &computev1alpha1.MachineClass{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "machineclass-",
			},
			Spec: computev1alpha1.MachineClassSpec{},
		}
		Expect(k8sClient.Create(ctx, machineClass)).Should(Succeed())

		By("creating the machine")
		m := &computev1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "machine-",
			},
			Spec: computev1alpha1.MachineSpec{
				MachineClass: corev1.LocalObjectReference{
					Name: machineClass.Name,
				},
			},
		}
		Expect(k8sClient.Create(ctx, m)).Should(Succeed())

		By("checking the machineclass and its finalizer consistently exist upon deletion ")
		Expect(k8sClient.Delete(ctx, machineClass)).Should(Succeed())

		machineClassKey := client.ObjectKeyFromObject(machineClass)
		Consistently(func() []string {
			Expect(k8sClient.Get(ctx, machineClassKey, machineClass))
			return machineClass.Finalizers
		}, interval).Should(ContainElement(computev1alpha1.MachineClassFinalizer))

		By("checking the machineclass is eventually gone after the deletion of the machine")
		Expect(k8sClient.Delete(ctx, m)).Should(Succeed())
		Eventually(func() bool {
			err := k8sClient.Get(ctx, machineClassKey, machineClass)
			Expect(client.IgnoreNotFound(err)).To(BeEmpty(), "errors other than `not found` are not expected")
			return apierrors.IsNotFound(err)
		}, timeout, interval).Should(BeTrue(), "the error should be `not found`")
	})
})
