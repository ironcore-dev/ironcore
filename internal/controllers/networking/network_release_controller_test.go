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
package networking

import (
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	. "github.com/onmetal/onmetal-api/utils/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

var _ = Describe("NetworkReleaseReconciler", func() {
	ns := SetupNamespace(&k8sClient)

	It("should release network interfaces whose owner is gone", func(ctx SpecContext) {
		By("creating a network having a peering claim that does not exist")
		network := &networkingv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "network-",
			},
			Spec: networkingv1alpha1.NetworkSpec{
				PeeringClaimRefs: []networkingv1alpha1.NetworkPeeringClaimRef{
					{
						Name: "should-not-exist",
						UID:  uuid.NewUUID(),
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, network)).To(Succeed())

		By("waiting for the network to have the peering claim released")
		Eventually(Object(network)).Should(HaveField("Spec.PeeringClaimRefs", BeEmpty()))
	})
})
