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
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	orimeta "github.com/onmetal/onmetal-api/ori/apis/meta/v1alpha1"
	"github.com/onmetal/onmetal-api/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("NetworkInterfaceDeleteLoadBalancer", func() {
	ctx := testutils.SetupContext()
	_, srv := SetupTest(ctx)

	It("should correctly delete a load balancer target for a network interface", func() {
		By("creating a network interface")
		res, err := srv.CreateNetworkInterface(ctx, &ori.CreateNetworkInterfaceRequest{
			NetworkInterface: &ori.NetworkInterface{
				Metadata: &orimeta.ObjectMetadata{},
				Spec: &ori.NetworkInterfaceSpec{
					Network: &ori.NetworkSpec{Handle: "foo"},
					Ips:     []string{"192.168.178.1"},
					LoadBalancerTargets: []*ori.LoadBalancerTargetSpec{
						{
							Ip: "10.0.0.1",
							Ports: []*ori.LoadBalancerPort{
								{
									Port:    80,
									EndPort: 8080,
								},
							},
						},
					},
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		By("inspecting the created network interface")
		networkInterface := res.NetworkInterface
		Expect(networkInterface.Spec.LoadBalancerTargets).To(ConsistOf(&ori.LoadBalancerTargetSpec{
			Ip: "10.0.0.1",
			Ports: []*ori.LoadBalancerPort{
				{
					Port:    80,
					EndPort: 8080,
				},
			},
		}))

		By("deleting the load balancer target")
		_, err = srv.DeleteNetworkInterfaceLoadBalancerTarget(ctx, &ori.DeleteNetworkInterfaceLoadBalancerTargetRequest{
			NetworkInterfaceId: networkInterface.Metadata.Id,
			LoadBalancerTarget: &ori.LoadBalancerTargetSpec{
				Ip: "10.0.0.1",
				Ports: []*ori.LoadBalancerPort{
					{
						Port:    80,
						EndPort: 8080,
					},
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		By("listing load balancers for the network interface")
		loadBalancers, err := srv.LoadBalancers().ListByDependent(ctx, networkInterface.Metadata.Id)
		Expect(err).NotTo(HaveOccurred())

		By("inspecting the retrieved list")
		Expect(loadBalancers).To(BeEmpty())
	})
})
