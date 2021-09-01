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

package utils

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"time"
)

var _ = Describe("Utils", func() {

	const (
		namespaceName = "mynamespace"
		finalizerName = "foo/bar"

		timeout  = time.Second * 30
		interval = time.Second * 1
	)

	var namespace *v1.Namespace
	var namespaceLookupKey types.NamespacedName

	Context("When handling Finalizers", func() {
		It("Should ensure that a finalizer is created and removed again", func() {
			ctx := context.Background()
			By("Creating a Namespace object")
			namespace = &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespaceName,
				},
			}
			namespaceLookupKey = types.NamespacedName{
				Name: namespaceName,
			}
			Expect(k8sClient.Create(ctx, namespace)).Should(Succeed())
			Expect(AssureFinalizer(ctx, k8sClient, finalizerName, namespace)).Should(BeNil())

			By("Expecting created")
			Eventually(func() bool {
				n := &v1.Namespace{}
				if err := k8sClient.Get(context.Background(), namespaceLookupKey, n); err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			By("Expecting Finalizer being created")
			Expect(AssureFinalizer(ctx, k8sClient, finalizerName, namespace)).Should(BeNil())

			By("Expecting Finalizer being removed")
			Expect(AssureFinalizerRemoved(ctx, k8sClient, finalizerName, namespace)).Should(BeNil())
		})
	})
})
