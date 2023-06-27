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

	"github.com/gogo/protobuf/proto"
	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	testingmachine "github.com/onmetal/onmetal-api/ori/testing/machine"
	machinepoolletv1alpha1 "github.com/onmetal/onmetal-api/poollet/machinepoollet/api/v1alpha1"
	machinepoolletmachine "github.com/onmetal/onmetal-api/poollet/machinepoollet/machine"
	. "github.com/onmetal/onmetal-api/utils/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
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
				ProviderID: "foo",
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
		const fooAnnotationValue = "bar"
		machine := &computev1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "machine-",
				Annotations: map[string]string{
					fooAnnotation: fooAnnotationValue,
				},
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
		))
		_, oriMachine := GetSingleMapEntry(srv.Machines)

		By("inspecting the ori machine")
		Expect(oriMachine.Metadata.Labels).To(HaveKeyWithValue(machinepoolletv1alpha1.DownwardAPILabel(fooDownwardAPILabel), fooAnnotationValue))
		Expect(oriMachine.Spec.Class).To(Equal(mc.Name))
		Expect(oriMachine.Spec.Power).To(Equal(ori.Power_POWER_ON))
		Expect(oriMachine.Spec.Volumes).To(ConsistOf(&ori.Volume{
			Name:   "primary",
			Device: "oda",
			Connection: &ori.VolumeConnection{
				Driver: "test",
				Handle: "testhandle",
			},
		}))
		Expect(oriMachine.Spec.NetworkInterfaces).To(ConsistOf(&ori.NetworkInterface{
			Name:      "primary",
			NetworkId: "foo",
			Ips:       []string{"10.0.0.1"},
		}))

		By("waiting for the onmetal machine status to be up-to-date")
		expectedMachineID := machinepoolletmachine.MakeID(testingmachine.FakeRuntimeName, oriMachine.Metadata.Id)
		Eventually(Object(machine)).Should(SatisfyAll(
			HaveField("Status.MachineID", expectedMachineID.String()),
			HaveField("Status.ObservedGeneration", machine.Generation),
		))

		By("setting the network interface id in the machine status")
		oriMachine = &testingmachine.FakeMachine{Machine: *proto.Clone(&oriMachine.Machine).(*ori.Machine)}
		oriMachine.Status.NetworkInterfaces = []*ori.NetworkInterfaceStatus{
			{
				Name:   "primary",
				Handle: "primary-handle",
				State:  ori.NetworkInterfaceState_NETWORK_INTERFACE_ATTACHED,
			},
		}
		srv.SetMachines([]*testingmachine.FakeMachine{oriMachine})

		By("waiting for the onmetal network interface to have a provider id set")
		Eventually(Object(machine)).Should(HaveField("Status.NetworkInterfaces", ConsistOf(MatchFields(IgnoreExtras, Fields{
			"Name":   Equal("primary"),
			"Handle": Equal("primary-handle"),
			"State":  Equal(computev1alpha1.NetworkInterfaceStateAttached),
		}))))
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
