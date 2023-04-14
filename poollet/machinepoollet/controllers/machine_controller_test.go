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

package controllers_test

import (
	"fmt"

	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	corev1alpha1 "github.com/onmetal/onmetal-api/api/core/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	testingmachine "github.com/onmetal/onmetal-api/ori/testing/machine"
	machinepoolletmachine "github.com/onmetal/onmetal-api/poollet/machinepoollet/machine"
	. "github.com/onmetal/onmetal-api/utils/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

var _ = Describe("MachineController", func() {
	ctx := SetupContext()
	ns, mp, mc, srv := SetupTest(ctx)

	It("should create a machine", func() {
		By("creating a network")
		network := &networkingv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "network-",
			},
			Spec: networkingv1alpha1.NetworkSpec{
				Handle: "foo",
			},
		}
		Expect(k8sClient.Create(ctx, network)).To(Succeed())

		By("patching the network to be available")
		baseNetwork := network.DeepCopy()
		network.Status.State = networkingv1alpha1.NetworkStateAvailable
		Expect(k8sClient.Status().Patch(ctx, network, client.MergeFrom(baseNetwork))).To(Succeed())

		By("creating a network interface")
		networkInterface := &networkingv1alpha1.NetworkInterface{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "networkinterface-",
			},
			Spec: networkingv1alpha1.NetworkInterfaceSpec{
				NetworkRef: corev1.LocalObjectReference{Name: network.Name},
				IPs: []networkingv1alpha1.IPSource{
					{Value: corev1alpha1.MustParseNewIP("10.0.0.1")},
				},
			},
		}
		Expect(k8sClient.Create(ctx, networkInterface)).To(Succeed())

		By("creating a volume")
		volume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{},
		}
		Expect(k8sClient.Create(ctx, volume)).To(Succeed())

		By("patching the volume to be available")
		baseVolume := volume.DeepCopy()
		volume.Status.State = storagev1alpha1.VolumeStateAvailable
		volume.Status.Access = &storagev1alpha1.VolumeAccess{
			Driver: "test",
			Handle: "testhandle",
		}
		Expect(k8sClient.Status().Patch(ctx, volume, client.MergeFrom(baseVolume))).To(Succeed())

		By("creating a machine")
		machine := &computev1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "machine-",
			},
			Spec: computev1alpha1.MachineSpec{
				MachineClassRef: corev1.LocalObjectReference{Name: mc.Name},
				MachinePoolRef:  &corev1.LocalObjectReference{Name: mp.Name},
				Volumes: []computev1alpha1.Volume{
					{
						Name: "primary",
						VolumeSource: computev1alpha1.VolumeSource{
							VolumeRef: &corev1.LocalObjectReference{Name: volume.Name},
						},
					},
				},
				NetworkInterfaces: []computev1alpha1.NetworkInterface{
					{
						Name: "primary",
						NetworkInterfaceSource: computev1alpha1.NetworkInterfaceSource{
							NetworkInterfaceRef: &corev1.LocalObjectReference{Name: networkInterface.Name},
						},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, machine)).To(Succeed())

		By("waiting for the runtime to report the machine, volume and network interface")
		Eventually(srv).Should(SatisfyAll(
			HaveField("Machines", HaveLen(1)),
			HaveField("Volumes", HaveLen(1)),
			HaveField("NetworkInterfaces", HaveLen(1)),
		))
		_, oriMachine := GetSingleMapEntry(srv.Machines)
		_, oriVolume := GetSingleMapEntry(srv.Volumes)
		_, oriNetworkInterface := GetSingleMapEntry(srv.NetworkInterfaces)

		By("inspecting the ori machine")
		Expect(oriMachine.Spec.Class).To(Equal(mc.Name))
		Expect(oriMachine.Spec.Power).To(Equal(ori.Power_POWER_ON))
		Expect(oriMachine.Spec.Volumes).To(ConsistOf(&ori.VolumeAttachment{
			Name:     "primary",
			Device:   "oda",
			VolumeId: oriVolume.Metadata.Id,
		}))
		Expect(oriMachine.Spec.NetworkInterfaces).To(ConsistOf(&ori.NetworkInterfaceAttachment{
			Name:               "primary",
			NetworkInterfaceId: oriNetworkInterface.Metadata.Id,
		}))

		By("inspecting the ori volume")
		Expect(oriVolume.Spec).To(Equal(&ori.VolumeSpec{
			Driver: "test",
			Handle: "testhandle",
		}))

		By("inspecting the ori network interface")
		Expect(oriNetworkInterface.Spec).To(Equal(&ori.NetworkInterfaceSpec{
			Network:             &ori.NetworkSpec{Handle: "foo"},
			Ips:                 []string{"10.0.0.1"},
			Prefixes:            []string{},
			LoadBalancerTargets: []*ori.LoadBalancerTargetSpec{},
		}))

		By("waiting for the onmetal machine status to be up-to-date")
		expectedMachineID := machinepoolletmachine.MakeID(testingmachine.FakeRuntimeName, oriMachine.Metadata.Id)
		Eventually(Object(machine)).Should(SatisfyAll(
			HaveField("Status.MachineID", expectedMachineID.String()),
			HaveField("Status.MachinePoolObservedGeneration", machine.Generation),
		))
	})

	It("should correctly manage the power state of a machine", func() {
		By("creating a machine")
		machine := &computev1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "machine-",
			},
			Spec: computev1alpha1.MachineSpec{
				MachineClassRef: corev1.LocalObjectReference{Name: mc.Name},
				MachinePoolRef:  &corev1.LocalObjectReference{Name: mp.Name},
			},
		}
		Expect(k8sClient.Create(ctx, machine)).To(Succeed())

		By("waiting for the machine to be created")
		Eventually(srv).Should(HaveField("Machines", HaveLen(1)))

		By("inspecting the machine")
		_, oriMachine := GetSingleMapEntry(srv.Machines)
		Expect(oriMachine.Spec.Power).To(Equal(ori.Power_POWER_ON))

		By("updating the machine power")
		base := machine.DeepCopy()
		machine.Spec.Power = computev1alpha1.PowerOff
		Expect(k8sClient.Patch(ctx, machine, client.MergeFrom(base))).To(Succeed())

		By("waiting for the ori machine to be updated")
		Eventually(oriMachine).Should(HaveField("Spec.Power", Equal(ori.Power_POWER_OFF)))
	})

	It("should correctly reconcile alias prefixes", func() {
		By("creating a network")
		network := &networkingv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "network-",
			},
			Spec: networkingv1alpha1.NetworkSpec{
				Handle: "foo",
			},
		}
		Expect(k8sClient.Create(ctx, network)).To(Succeed())

		By("patching the network to be available")
		baseNetwork := network.DeepCopy()
		network.Status.State = networkingv1alpha1.NetworkStateAvailable
		Expect(k8sClient.Status().Patch(ctx, network, client.MergeFrom(baseNetwork))).To(Succeed())

		By("creating a machine")
		machine := &computev1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "machine-",
			},
			Spec: computev1alpha1.MachineSpec{
				MachineClassRef: corev1.LocalObjectReference{Name: mc.Name},
				MachinePoolRef:  &corev1.LocalObjectReference{Name: mp.Name},
				NetworkInterfaces: []computev1alpha1.NetworkInterface{
					{
						Name: "primary",
						NetworkInterfaceSource: computev1alpha1.NetworkInterfaceSource{
							Ephemeral: &computev1alpha1.EphemeralNetworkInterfaceSource{
								NetworkInterfaceTemplate: &networkingv1alpha1.NetworkInterfaceTemplateSpec{
									ObjectMeta: metav1.ObjectMeta{
										Labels: map[string]string{"foo": "bar"},
									},
									Spec: networkingv1alpha1.NetworkInterfaceSpec{
										NetworkRef: corev1.LocalObjectReference{Name: network.Name},
										IPs: []networkingv1alpha1.IPSource{
											{Value: corev1alpha1.MustParseNewIP("10.0.0.1")},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, machine)).To(Succeed())

		By("creating an alias prefix targeting the network interface")
		aliasPrefix := &networkingv1alpha1.AliasPrefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "aliasprefix-",
			},
			Spec: networkingv1alpha1.AliasPrefixSpec{
				NetworkRef: corev1.LocalObjectReference{Name: network.Name},
				Prefix: networkingv1alpha1.PrefixSource{
					Value: corev1alpha1.MustParseNewIPPrefix("10.0.1.0/24"),
				},
				NetworkInterfaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"foo": "bar"},
				},
			},
		}
		Expect(k8sClient.Create(ctx, aliasPrefix)).To(Succeed())

		By("waiting for the runtime to report the machine and network interface")
		Eventually(func(g Gomega) {
			g.Expect(srv.Machines).To(HaveLen(1))
			g.Expect(srv.NetworkInterfaces).To(HaveLen(1))

			_, oriMachine := GetSingleMapEntry(srv.Machines)
			_, oriNetworkInterface := GetSingleMapEntry(srv.NetworkInterfaces)

			g.Expect(oriMachine.Spec.NetworkInterfaces).To(ConsistOf(&ori.NetworkInterfaceAttachment{
				Name:               "primary",
				NetworkInterfaceId: oriNetworkInterface.Metadata.Id,
			}))

			g.Expect(oriNetworkInterface.Spec.Prefixes).To(ConsistOf("10.0.1.0/24"))
		}).Should(Succeed())
	})

	It("should correctly reconcile load balancer targets", func() {
		By("creating a network")
		network := &networkingv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "network-",
			},
			Spec: networkingv1alpha1.NetworkSpec{
				Handle: "foo",
			},
		}
		Expect(k8sClient.Create(ctx, network)).To(Succeed())

		By("patching the network to be available")
		baseNetwork := network.DeepCopy()
		network.Status.State = networkingv1alpha1.NetworkStateAvailable
		Expect(k8sClient.Status().Patch(ctx, network, client.MergeFrom(baseNetwork))).To(Succeed())

		By("creating a machine")
		machine := &computev1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "machine-",
			},
			Spec: computev1alpha1.MachineSpec{
				MachineClassRef: corev1.LocalObjectReference{Name: mc.Name},
				MachinePoolRef:  &corev1.LocalObjectReference{Name: mp.Name},
				NetworkInterfaces: []computev1alpha1.NetworkInterface{
					{
						Name: "primary",
						NetworkInterfaceSource: computev1alpha1.NetworkInterfaceSource{
							Ephemeral: &computev1alpha1.EphemeralNetworkInterfaceSource{
								NetworkInterfaceTemplate: &networkingv1alpha1.NetworkInterfaceTemplateSpec{
									ObjectMeta: metav1.ObjectMeta{
										Labels: map[string]string{"foo": "bar"},
									},
									Spec: networkingv1alpha1.NetworkInterfaceSpec{
										NetworkRef: corev1.LocalObjectReference{Name: network.Name},
										IPs: []networkingv1alpha1.IPSource{
											{Value: corev1alpha1.MustParseNewIP("192.168.178.1")},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, machine)).To(Succeed())

		By("creating a load balancer targeting the network interface")
		loadBalancer := &networkingv1alpha1.LoadBalancer{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "loadbalancer-",
			},
			Spec: networkingv1alpha1.LoadBalancerSpec{
				Type:       networkingv1alpha1.LoadBalancerTypePublic,
				NetworkRef: corev1.LocalObjectReference{Name: network.Name},
				IPFamilies: []corev1.IPFamily{corev1.IPv4Protocol},
				NetworkInterfaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"foo": "bar"},
				},
				Ports: []networkingv1alpha1.LoadBalancerPort{
					{
						Port: 80,
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, loadBalancer)).To(Succeed())

		By("adding an ip to the load balancer")
		baseLoadBalancer := loadBalancer.DeepCopy()
		loadBalancer.Status.IPs = []corev1alpha1.IP{corev1alpha1.MustParseIP("10.0.0.1")}
		Expect(k8sClient.Status().Patch(ctx, loadBalancer, client.MergeFrom(baseLoadBalancer))).To(Succeed())

		By("waiting for the runtime to report the machine and network interface")
		Eventually(func(g Gomega) {
			g.Expect(srv.Machines).To(HaveLen(1))
			g.Expect(srv.NetworkInterfaces).To(HaveLen(1))

			_, oriMachine := GetSingleMapEntry(srv.Machines)
			_, oriNetworkInterface := GetSingleMapEntry(srv.NetworkInterfaces)

			g.Expect(oriMachine.Spec.NetworkInterfaces).To(ConsistOf(&ori.NetworkInterfaceAttachment{
				Name:               "primary",
				NetworkInterfaceId: oriNetworkInterface.Metadata.Id,
			}))

			g.Expect(oriNetworkInterface.Spec.LoadBalancerTargets).To(ConsistOf(&ori.LoadBalancerTargetSpec{
				Ip: "10.0.0.1",
				Ports: []*ori.LoadBalancerPort{
					{
						Protocol: ori.Protocol_TCP,
						Port:     80,
						EndPort:  80,
					},
				},
			}))
		}).Should(Succeed())
	})

	It("should correctly reconcile nats", func() {
		By("creating a network")
		network := &networkingv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "network-",
			},
			Spec: networkingv1alpha1.NetworkSpec{
				Handle: "foo",
			},
		}
		Expect(k8sClient.Create(ctx, network)).To(Succeed())

		By("patching the network to be available")
		baseNetwork := network.DeepCopy()
		network.Status.State = networkingv1alpha1.NetworkStateAvailable
		Expect(k8sClient.Status().Patch(ctx, network, client.MergeFrom(baseNetwork))).To(Succeed())

		By("creating a machine")
		machine := &computev1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "machine-",
			},
			Spec: computev1alpha1.MachineSpec{
				MachineClassRef: corev1.LocalObjectReference{Name: mc.Name},
				MachinePoolRef:  &corev1.LocalObjectReference{Name: mp.Name},
				NetworkInterfaces: []computev1alpha1.NetworkInterface{
					{
						Name: "primary",
						NetworkInterfaceSource: computev1alpha1.NetworkInterfaceSource{
							Ephemeral: &computev1alpha1.EphemeralNetworkInterfaceSource{
								NetworkInterfaceTemplate: &networkingv1alpha1.NetworkInterfaceTemplateSpec{
									ObjectMeta: metav1.ObjectMeta{
										Labels: map[string]string{"foo": "bar"},
									},
									Spec: networkingv1alpha1.NetworkInterfaceSpec{
										NetworkRef: corev1.LocalObjectReference{Name: network.Name},
										IPs: []networkingv1alpha1.IPSource{
											{Value: corev1alpha1.MustParseNewIP("192.168.178.1")},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, machine)).To(Succeed())

		By("creating a nat gateway targeting the network interface")
		natGateway := &networkingv1alpha1.NATGateway{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "natgateway-",
			},
			Spec: networkingv1alpha1.NATGatewaySpec{
				Type:       networkingv1alpha1.NATGatewayTypePublic,
				IPFamilies: []corev1.IPFamily{corev1.IPv4Protocol},
				NetworkRef: corev1.LocalObjectReference{Name: network.Name},
				IPs:        []networkingv1alpha1.NATGatewayIP{{Name: "primary"}},
				NetworkInterfaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"foo": "bar"},
				},
			},
		}
		Expect(k8sClient.Create(ctx, natGateway)).To(Succeed())

		By("adding an ip to the nat gateway")
		baseNATGateway := natGateway.DeepCopy()
		natGateway.Status.IPs = []networkingv1alpha1.NATGatewayIPStatus{
			{
				Name: "primary",
				IP:   corev1alpha1.MustParseIP("10.0.0.1"),
			},
		}
		Expect(k8sClient.Status().Patch(ctx, natGateway, client.MergeFrom(baseNATGateway))).To(Succeed())

		By("waiting for the runtime to report the machine and network interface")
		Eventually(func(g Gomega) {
			g.Expect(srv.Machines).To(HaveLen(1))
			g.Expect(srv.NetworkInterfaces).To(HaveLen(1))

			_, oriMachine := GetSingleMapEntry(srv.Machines)
			_, oriNetworkInterface := GetSingleMapEntry(srv.NetworkInterfaces)

			g.Expect(oriMachine.Spec.NetworkInterfaces).To(ConsistOf(&ori.NetworkInterfaceAttachment{
				Name:               "primary",
				NetworkInterfaceId: oriNetworkInterface.Metadata.Id,
			}))

			g.Expect(oriNetworkInterface.Spec.Nats).To(ConsistOf(&ori.NATSpec{
				Ip:      "10.0.0.1",
				Port:    1024,
				EndPort: 1024 + networkingv1alpha1.DefaultPortsPerNetworkInterface - 1,
			}))
		}).Should(Succeed())
	})

	It("should correctly tear down network interfaces of machines", func() {
		By("creating a network")
		network := &networkingv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "network-",
			},
			Spec: networkingv1alpha1.NetworkSpec{
				Handle: "foo",
			},
		}
		Expect(k8sClient.Create(ctx, network)).To(Succeed())

		By("patching the network to be available")
		baseNetwork := network.DeepCopy()
		network.Status.State = networkingv1alpha1.NetworkStateAvailable
		Expect(k8sClient.Status().Patch(ctx, network, client.MergeFrom(baseNetwork))).To(Succeed())

		By("creating two machines")
		machine1 := &computev1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "machine-",
			},
			Spec: computev1alpha1.MachineSpec{
				MachineClassRef: corev1.LocalObjectReference{Name: mc.Name},
				MachinePoolRef:  &corev1.LocalObjectReference{Name: mp.Name},
				NetworkInterfaces: []computev1alpha1.NetworkInterface{
					{
						Name: "primary",
						NetworkInterfaceSource: computev1alpha1.NetworkInterfaceSource{
							Ephemeral: &computev1alpha1.EphemeralNetworkInterfaceSource{
								NetworkInterfaceTemplate: &networkingv1alpha1.NetworkInterfaceTemplateSpec{
									Spec: networkingv1alpha1.NetworkInterfaceSpec{
										NetworkRef: corev1.LocalObjectReference{Name: network.Name},
										IPs: []networkingv1alpha1.IPSource{
											{Value: corev1alpha1.MustParseNewIP("192.168.178.1")},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, machine1)).To(Succeed())

		machine2 := &computev1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "machine-",
			},
			Spec: computev1alpha1.MachineSpec{
				MachineClassRef: corev1.LocalObjectReference{Name: mc.Name},
				MachinePoolRef:  &corev1.LocalObjectReference{Name: mp.Name},
				NetworkInterfaces: []computev1alpha1.NetworkInterface{
					{
						Name: "primary",
						NetworkInterfaceSource: computev1alpha1.NetworkInterfaceSource{
							Ephemeral: &computev1alpha1.EphemeralNetworkInterfaceSource{
								NetworkInterfaceTemplate: &networkingv1alpha1.NetworkInterfaceTemplateSpec{
									Spec: networkingv1alpha1.NetworkInterfaceSpec{
										NetworkRef: corev1.LocalObjectReference{Name: network.Name},
										IPs: []networkingv1alpha1.IPSource{
											{Value: corev1alpha1.MustParseNewIP("192.168.178.2")},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, machine2)).To(Succeed())

		By("waiting for the runtime to report the network interfaces")
		Eventually(srv).Should(SatisfyAll(
			HaveField("NetworkInterfaces", HaveLen(2)),
		))

		By("deleting the first machine")
		Expect(k8sClient.Delete(ctx, machine1)).To(Succeed())

		By("waiting for the runtime to report a single network interface")
		Eventually(srv).Should(SatisfyAll(
			HaveField("NetworkInterfaces", HaveLen(1)),
		))

		By("asserting it stays that way")
		Consistently(srv).Should(SatisfyAll(
			HaveField("NetworkInterfaces", HaveLen(1)),
		))
	})
})

func GetSingleMapEntry[K comparable, V any](m map[K]V) (K, V) {
	if n := len(m); n != 1 {
		Fail(fmt.Sprintf("Expected for map to have a single entry but got %d", n), 1)
	}
	for k, v := range m {
		return k, v
	}
	panic("unreachable")
}
