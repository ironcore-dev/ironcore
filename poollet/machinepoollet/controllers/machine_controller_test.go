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

	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	. "github.com/onmetal/onmetal-api/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
					{Value: commonv1alpha1.MustParseNewIP("10.0.0.1")},
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
			Network:  &ori.NetworkSpec{Handle: "foo"},
			Ips:      []string{"10.0.0.1"},
			Prefixes: []string{},
		}))
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
											{Value: commonv1alpha1.MustParseNewIP("10.0.0.1")},
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
					Value: commonv1alpha1.MustParseNewIPPrefix("10.0.1.0/24"),
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
