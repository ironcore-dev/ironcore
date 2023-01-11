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
	"github.com/onmetal/controller-utils/set"
	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	machinebrokerv1alpha1 "github.com/onmetal/onmetal-api/broker/machinebroker/api/v1alpha1"
	"github.com/onmetal/onmetal-api/broker/machinebroker/apiutils"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	orimeta "github.com/onmetal/onmetal-api/ori/apis/meta/v1alpha1"
	. "github.com/onmetal/onmetal-api/utils/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("CreateNetworkInterface", func() {
	ctx := SetupContext()
	ns, srv := SetupTest(ctx)

	It("should correctly create a network interface", func() {
		By("creating a network interface")
		const (
			ip             = "192.168.178.1"
			prefix         = "192.168.178.1/24"
			virtualIP      = "10.0.0.1"
			loadBalancerIP = "10.0.0.2"
			networkHandle  = "foo"
		)
		res, err := srv.CreateNetworkInterface(ctx, &ori.CreateNetworkInterfaceRequest{
			NetworkInterface: &ori.NetworkInterface{
				Metadata: &orimeta.ObjectMetadata{},
				Spec: &ori.NetworkInterfaceSpec{
					Network: &ori.NetworkSpec{
						Handle: networkHandle,
					},
					Ips:       []string{ip},
					VirtualIp: &ori.VirtualIPSpec{Ip: virtualIP},
					Prefixes:  []string{prefix},
					LoadBalancerTargets: []*ori.LoadBalancerTargetSpec{
						{
							Ip: loadBalancerIP,
							Ports: []*ori.LoadBalancerPort{
								{
									Protocol: ori.Protocol_TCP,
									Port:     80,
									EndPort:  80,
								},
							},
						},
					},
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())
		networkInterface := res.NetworkInterface
		networkInterfaceID := networkInterface.Metadata.Id

		By("inspecting the kubernetes network interface")
		k8sNetworkInterface := &networkingv1alpha1.NetworkInterface{}
		k8sNetworkInterfaceKey := client.ObjectKey{Namespace: ns.Name, Name: networkInterfaceID}
		Expect(k8sClient.Get(ctx, k8sNetworkInterfaceKey, k8sNetworkInterface)).To(Succeed())

		Expect(k8sNetworkInterface.Spec.IPs).To(Equal([]networkingv1alpha1.IPSource{
			{
				Value: commonv1alpha1.MustParseNewIP(ip),
			},
		}))

		By("getting the referenced kubernetes network")
		k8sNetwork := &networkingv1alpha1.Network{}
		k8sNetworkKey := client.ObjectKey{Namespace: ns.Name, Name: k8sNetworkInterface.Spec.NetworkRef.Name}
		Expect(k8sClient.Get(ctx, k8sNetworkKey, k8sNetwork)).To(Succeed())

		By("inspecting the referenced kubernetes network")
		Expect(apiutils.GetDependents(k8sNetwork)).To(ContainElement(networkInterfaceID))
		Expect(k8sNetwork.Spec.Handle).To(Equal(networkHandle))

		By("inspecting the virtual ip reference")
		virtualIPRef := k8sNetworkInterface.Spec.VirtualIP.VirtualIPRef
		Expect(virtualIPRef).NotTo(BeNil())

		By("getting the referenced kubernetes virtual ip")
		k8sVirtualIP := &networkingv1alpha1.VirtualIP{}
		k8sVirtualIPKey := client.ObjectKey{Namespace: ns.Name, Name: virtualIPRef.Name}
		Expect(k8sClient.Get(ctx, k8sVirtualIPKey, k8sVirtualIP)).To(Succeed())

		By("inspecting the referenced kubernetes virtual ip")
		Expect(k8sVirtualIP.Status.IP).To(Equal(commonv1alpha1.MustParseNewIP(virtualIP)))

		By("listing alias prefixes for network interface")
		aliasPrefixes, err := srv.AliasPrefixes().ListByDependent(ctx, networkInterfaceID)
		Expect(err).NotTo(HaveOccurred())

		By("inspecting the alias prefix list")
		Expect(aliasPrefixes).To(HaveLen(1))
		aliasPrefix := aliasPrefixes[0]
		Expect(aliasPrefix.NetworkHandle).To(Equal(networkHandle))
		Expect(aliasPrefix.Prefix).To(Equal(commonv1alpha1.MustParseIPPrefix(prefix)))
		Expect(aliasPrefix.Destinations.Equal(set.New(networkInterfaceID))).To(BeTrue())

		By("listing load balancers for network interface")
		loadBalancers, err := srv.LoadBalancers().ListByDependent(ctx, networkInterfaceID)
		Expect(err).NotTo(HaveOccurred())

		By("inspecting the load balancer list")
		Expect(loadBalancers).To(ConsistOf(machinebrokerv1alpha1.LoadBalancer{
			NetworkHandle: networkHandle,
			IP:            commonv1alpha1.MustParseIP(loadBalancerIP),
			Ports: []machinebrokerv1alpha1.LoadBalancerPort{
				{
					Protocol: corev1.ProtocolTCP,
					Port:     80,
					EndPort:  80,
				},
			},
			Destinations: []string{networkInterfaceID},
		}))
	})

	It("should correctly create a network with a nat gateway target", func() {
		By("creating a network interface")
		const (
			ip                = "192.168.178.1"
			networkHandle     = "foo"
			natGatewayIP      = "10.0.0.1"
			natGatewayPort    = 80
			natGatewayEndPort = 8000
		)
		res, err := srv.CreateNetworkInterface(ctx, &ori.CreateNetworkInterfaceRequest{
			NetworkInterface: &ori.NetworkInterface{
				Metadata: &orimeta.ObjectMetadata{},
				Spec: &ori.NetworkInterfaceSpec{
					Network: &ori.NetworkSpec{
						Handle: networkHandle,
					},
					Ips: []string{ip},
					Nats: []*ori.NATSpec{
						{
							Ip:      natGatewayIP,
							Port:    natGatewayPort,
							EndPort: natGatewayEndPort,
						},
					},
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())
		networkInterface := res.NetworkInterface
		networkInterfaceID := networkInterface.Metadata.Id

		By("inspecting the kubernetes network interface")
		k8sNetworkInterface := &networkingv1alpha1.NetworkInterface{}
		k8sNetworkInterfaceKey := client.ObjectKey{Namespace: ns.Name, Name: networkInterfaceID}
		Expect(k8sClient.Get(ctx, k8sNetworkInterfaceKey, k8sNetworkInterface)).To(Succeed())

		Expect(k8sNetworkInterface.Spec.IPs).To(Equal([]networkingv1alpha1.IPSource{
			{
				Value: commonv1alpha1.MustParseNewIP(ip),
			},
		}))

		By("listing nat gateways for network interface")
		natGateways, err := srv.NATGateways().ListByDependent(ctx, networkInterfaceID)
		Expect(err).NotTo(HaveOccurred())

		By("inspecting the nat gateways")
		Expect(natGateways).To(ConsistOf(machinebrokerv1alpha1.NATGateway{
			NetworkHandle: networkHandle,
			IP:            commonv1alpha1.MustParseIP(natGatewayIP),
			Destinations: []machinebrokerv1alpha1.NATGatewayDestination{
				{
					ID:      networkInterfaceID,
					Port:    natGatewayPort,
					EndPort: natGatewayEndPort,
				},
			},
		}))
	})

	It("should re-use kubernetes networks and delete them only if no dependents exist", func() {
		const handle = "foo"
		const noOfNetworkInterfaces = 6

		iterate := func(f func(i int)) {
			for i := 0; i < noOfNetworkInterfaces; i++ {
				f(i)
			}
		}

		By("creating multiple network interfaces with the same network handle")
		networkInterfaces := make([]*ori.NetworkInterface, noOfNetworkInterfaces)
		iterate(func(i int) {
			res, err := srv.CreateNetworkInterface(ctx, &ori.CreateNetworkInterfaceRequest{
				NetworkInterface: &ori.NetworkInterface{
					Metadata: &orimeta.ObjectMetadata{},
					Spec: &ori.NetworkInterfaceSpec{
						Network: &ori.NetworkSpec{
							Handle: handle,
						},
						Ips: []string{"10.0.0.1"},
					},
				},
			})
			Expect(err).NotTo(HaveOccurred())
			networkInterfaces[i] = res.NetworkInterface
		})

		Expect(networkInterfaces).To(HaveEach(HaveField("Spec.Network.Handle", handle)))

		By("getting the corresponding kubernetes network interfaces")
		k8sNetworkInterfaces := make([]networkingv1alpha1.NetworkInterface, noOfNetworkInterfaces)
		iterate(func(i int) {
			k8sNetworkInterface := &k8sNetworkInterfaces[i]
			k8sNetworkInterfaceKey := client.ObjectKey{Namespace: ns.Name, Name: networkInterfaces[i].Metadata.Id}
			Expect(k8sClient.Get(ctx, k8sNetworkInterfaceKey, k8sNetworkInterface)).To(Succeed())
		})

		By("asserting all reference the same kubernetes network")
		k8sNetworkRef := k8sNetworkInterfaces[0].Spec.NetworkRef
		Expect(k8sNetworkInterfaces).To(HaveEach(HaveField("Spec.NetworkRef", k8sNetworkRef)))

		By("getting the referenced kubernetes network")
		k8sNetwork := &networkingv1alpha1.Network{}
		k8sNetworkKey := client.ObjectKey{Namespace: ns.Name, Name: k8sNetworkRef.Name}
		Expect(k8sClient.Get(ctx, k8sNetworkKey, k8sNetwork)).To(Succeed())

		By("inspecting the network dependents")
		Expect(apiutils.GetDependents(k8sNetwork)).To(HaveLen(noOfNetworkInterfaces))

		By("deleting the first half of the networks")
		const firstHalfNoOfNetworkInterfaces = noOfNetworkInterfaces / 2
		const secondHalfNoOfNetworkInterfaces = noOfNetworkInterfaces - firstHalfNoOfNetworkInterfaces
		for i := 0; i < firstHalfNoOfNetworkInterfaces; i++ {
			_, err := srv.DeleteNetworkInterface(ctx, &ori.DeleteNetworkInterfaceRequest{
				NetworkInterfaceId: networkInterfaces[i].Metadata.Id,
			})
			Expect(err).NotTo(HaveOccurred())
		}

		By("getting the kubernetes network again")
		Expect(k8sClient.Get(ctx, k8sNetworkKey, k8sNetwork)).To(Succeed())

		By("inspecting the network dependents")
		Expect(apiutils.GetDependents(k8sNetwork)).To(HaveLen(secondHalfNoOfNetworkInterfaces))

		By("deleting the remaining half of the networks")
		for i := firstHalfNoOfNetworkInterfaces; i < noOfNetworkInterfaces; i++ {
			_, err := srv.DeleteNetworkInterface(ctx, &ori.DeleteNetworkInterfaceRequest{
				NetworkInterfaceId: networkInterfaces[i].Metadata.Id,
			})
			Expect(err).NotTo(HaveOccurred())
		}

		By("getting the referenced network and asserting it's gone")
		Expect(k8sClient.Get(ctx, k8sNetworkKey, k8sNetwork)).To(Satisfy(apierrors.IsNotFound))
	})
})
