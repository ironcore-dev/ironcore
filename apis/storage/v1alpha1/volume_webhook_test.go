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
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var _ = Describe("volume validation webhook", func() {
	ns := SetupTest()
	Context("upon volume update", func() {
		It("signals error if storagepool is changed", func() {
			m := &Volume{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    ns.Name,
					GenerateName: "volume-",
				},
			}
			Expect(k8sClient.Create(ctx, m)).To(Succeed())

			newStoragePoolName := "new_storagepool"
			m.Spec.StoragePool.Name = newStoragePoolName
			err := k8sClient.Update(ctx, m)
			Expect(err).To(HaveOccurred())

			path := field.NewPath("spec", "storagepool", "name")
			fieldErr := field.Invalid(path, newStoragePoolName, "immutable")
			fieldErrList := field.ErrorList{fieldErr}
			Expect(err.Error()).To(ContainSubstring(fieldErrList.ToAggregate().Error()))
		})
	})
})
