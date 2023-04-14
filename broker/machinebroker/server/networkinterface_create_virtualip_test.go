// Copyright 2022 OnMetal authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server_test

import (
	corev1alpha1 "github.com/onmetal/onmetal-api/api/core/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	orimeta "github.com/onmetal/onmetal-api/ori/apis/meta/v1alpha1"
	. "github.com/onmetal/onmetal-api/utils/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("NetworkInterfaceCreateVirtualIP", func() {
	ctx := SetupContext()
	ns, srv := SetupTest(ctx)

	It("should correctly create a virtual ip for a network interface", func() {
		By("creating a network interface")
		res, err := srv.CreateNetworkInterface(ctx, &ori.CreateNetworkInterfaceRequest{
			NetworkInterface: &ori.NetworkInterface{
				Metadata: &orimeta.ObjectMetadata{},
				Spec: &ori.NetworkInterfaceSpec{
					Network: &ori.NetworkSpec{Handle: "foo"},
					Ips:     []string{"192.168.178.1"},
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		By("inspecting the created network interface")
		networkInterface := res.NetworkInterface
		Expect(networkInterface.Spec.VirtualIp).To(BeNil())

		By("creating a virtual ip for the network interface")
		_, err = srv.CreateNetworkInterfaceVirtualIP(ctx, &ori.CreateNetworkInterfaceVirtualIPRequest{
			NetworkInterfaceId: networkInterface.Metadata.Id,
			VirtualIp: &ori.VirtualIPSpec{
				Ip: "10.0.0.1",
			},
		})
		Expect(err).NotTo(HaveOccurred())

		By("getting the kubernetes network interface")
		k8sNetworkInterface := &networkingv1alpha1.NetworkInterface{}
		k8sNetworkInterfaceKey := client.ObjectKey{Namespace: ns.Name, Name: networkInterface.Metadata.Id}
		Expect(k8sClient.Get(ctx, k8sNetworkInterfaceKey, k8sNetworkInterface)).To(Succeed())

		By("inspecting the kubernetes network interface")
		var ref corev1.LocalObjectReference
		assertVirtualIPSourceRef(k8sNetworkInterface.Spec.VirtualIP, &ref)

		By("getting the referenced kubernetes virtual ip")
		k8sVirtualIP := &networkingv1alpha1.VirtualIP{}
		k8sVirtualIPKey := client.ObjectKey{Namespace: ns.Name, Name: ref.Name}
		Expect(k8sClient.Get(ctx, k8sVirtualIPKey, k8sVirtualIP)).To(Succeed())

		By("inspecting the kubernetes virtual ip")
		Expect(metav1.IsControlledBy(k8sVirtualIP, k8sNetworkInterface)).To(BeTrue(), "kubernetes virtual ip should be controlled by network interface")
		Expect(k8sVirtualIP.Spec).To(Equal(networkingv1alpha1.VirtualIPSpec{
			Type:     networkingv1alpha1.VirtualIPTypePublic,
			IPFamily: corev1.IPv4Protocol,
		}))
		Expect(k8sVirtualIP.Status.IP).To(Equal(corev1alpha1.MustParseNewIP("10.0.0.1")))
	})
})

func assertVirtualIPSourceRef(src *networkingv1alpha1.VirtualIPSource, ref *corev1.LocalObjectReference) {
	ExpectWithOffset(1, src).NotTo(BeNil())
	ExpectWithOffset(1, src.VirtualIPRef).NotTo(BeNil())
	*ref = *src.VirtualIPRef
}
