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
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
)

var ctx = context.Background()

var _ = Describe("machine controller", func() {
	Context("Reconcile", func() {
		It("sets the owner Machine as a Controller OwnerReference on the controlled IPAMRange", func() {
			m := &computev1alpha1.Machine{}
			m.Namespace = "default"
			m.Name = "test"
			Expect(k8sClient.Create(ctx, m)).To(Succeed())
		})
	})
})
