// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package controllers_test

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	testingmachine "github.com/ironcore-dev/ironcore/iri/testing/machine"
	machinepoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/machinepoollet/api/v1alpha1"
	machinepoolletmachine "github.com/ironcore-dev/ironcore/poollet/machinepoollet/machine"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

var _ = Describe("MachineController", func() {
	ns, mp, mc, srv := SetupTest()

	It("should create a machine", func(ctx SpecContext) {
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
		Eventually(UpdateStatus(network, func() {
			network.Status.State = networkingv1alpha1.NetworkStateAvailable
		})).Should(Succeed())

		By("creating a network interface")
		nic := &networkingv1alpha1.NetworkInterface{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "nic-",
			},
			Spec: networkingv1alpha1.NetworkInterfaceSpec{
				NetworkRef: corev1.LocalObjectReference{Name: network.Name},
				IPs: []networkingv1alpha1.IPSource{
					{Value: commonv1alpha1.MustParseNewIP("10.0.0.1")},
				},
			},
		}
		Expect(k8sClient.Create(ctx, nic)).To(Succeed())

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
		Eventually(UpdateStatus(volume, func() {
			volume.Status.State = storagev1alpha1.VolumeStateAvailable
			volume.Status.Access = &storagev1alpha1.VolumeAccess{
				Driver: "test",
				Handle: "testhandle",
			}
		})).Should(Succeed())

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
							NetworkInterfaceRef: &corev1.LocalObjectReference{Name: nic.Name},
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
		_, iriMachine := GetSingleMapEntry(srv.Machines)

		By("inspecting the iri machine")
		Expect(iriMachine.Metadata.Labels).To(HaveKeyWithValue(machinepoolletv1alpha1.DownwardAPILabel(fooDownwardAPILabel), fooAnnotationValue))
		Expect(iriMachine.Spec.Class).To(Equal(mc.Name))
		Expect(iriMachine.Spec.Power).To(Equal(iri.Power_POWER_ON))
		Expect(iriMachine.Spec.Volumes).To(ConsistOf(&iri.Volume{
			Name:   "primary",
			Device: "oda",
			Connection: &iri.VolumeConnection{
				Driver: "test",
				Handle: "testhandle",
			},
		}))
		Expect(iriMachine.Spec.NetworkInterfaces).To(ConsistOf(&iri.NetworkInterface{
			Name:      "primary",
			NetworkId: "foo",
			Ips:       []string{"10.0.0.1"},
		}))

		By("waiting for the ironcore machine status to be up-to-date")
		expectedMachineID := machinepoolletmachine.MakeID(testingmachine.FakeRuntimeName, iriMachine.Metadata.Id)
		Eventually(Object(machine)).Should(SatisfyAll(
			HaveField("Status.MachineID", expectedMachineID.String()),
			HaveField("Status.ObservedGeneration", machine.Generation),
		))

		By("setting the network interface id in the machine status")
		iriMachine = &testingmachine.FakeMachine{Machine: *proto.Clone(&iriMachine.Machine).(*iri.Machine)}
		iriMachine.Metadata.Generation = 1
		iriMachine.Status.ObservedGeneration = 1
		iriMachine.Status.NetworkInterfaces = []*iri.NetworkInterfaceStatus{
			{
				Name:   "primary",
				Handle: "primary-handle",
				State:  iri.NetworkInterfaceState_NETWORK_INTERFACE_ATTACHED,
				Ips:    []string{"10.0.0.1"},
			},
		}
		srv.SetMachines([]*testingmachine.FakeMachine{iriMachine})

		By("waiting for the ironcore network interface to have a provider id set")
		Eventually(Object(nic)).Should(And(
			HaveField("Spec.ProviderID", Equal("primary-handle")),
			HaveField("Status", MatchFields(IgnoreExtras, Fields{
				"State": Equal(networkingv1alpha1.NetworkInterfaceStateAvailable),
				"IPs":   ContainElement(commonv1alpha1.MustParseIP("10.0.0.1")),
			})),
		))
		Eventually(Object(machine)).Should(HaveField("Status.NetworkInterfaces", ConsistOf(MatchFields(IgnoreExtras, Fields{
			"Name":   Equal("primary"),
			"Handle": Equal("primary-handle"),
			"State":  Equal(computev1alpha1.NetworkInterfaceStateAttached),
		}))))

		By("removing the network interface from the machine")
		Eventually(Update(machine, func() {
			machine.Spec.NetworkInterfaces = []computev1alpha1.NetworkInterface{}
		})).Should(Succeed())

		By("ensuring that the network interface has been removed from the iri machine")
		Eventually(srv.Machines[iriMachine.Metadata.Id]).Should(SatisfyAll(
			HaveField("Spec.NetworkInterfaces", BeEmpty()),
		))
	})

	It("should correctly manage the power state of a machine", func(ctx SpecContext) {
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
		_, iriMachine := GetSingleMapEntry(srv.Machines)
		Expect(iriMachine.Spec.Power).To(Equal(iri.Power_POWER_ON))

		By("updating the machine power")
		base := machine.DeepCopy()
		machine.Spec.Power = computev1alpha1.PowerOff
		Expect(k8sClient.Patch(ctx, machine, client.MergeFrom(base))).To(Succeed())

		By("waiting for the iri machine to be updated")
		Eventually(iriMachine).Should(HaveField("Spec.Power", Equal(iri.Power_POWER_OFF)))
	})

	It("should correctly manage state of a machine", func(ctx SpecContext) {
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

		By("inspecting the machine to be running")
		_, iriMachine := GetSingleMapEntry(srv.Machines)
		iriMachine = &testingmachine.FakeMachine{Machine: *proto.Clone(&iriMachine.Machine).(*iri.Machine)}
		iriMachine.Metadata.Generation = 1
		iriMachine.Status.ObservedGeneration = 1
		iriMachine.Status.State = iri.MachineState_MACHINE_RUNNING
		srv.SetMachines([]*testingmachine.FakeMachine{iriMachine})
		Eventually(Object(machine)).Should(HaveField("Status.State", Equal(computev1alpha1.MachineStateRunning)))

		By("inspecting the machine to be terminating")
		_, iriMachine = GetSingleMapEntry(srv.Machines)
		iriMachine = &testingmachine.FakeMachine{Machine: *proto.Clone(&iriMachine.Machine).(*iri.Machine)}
		iriMachine.Metadata.Generation = 2
		iriMachine.Status.ObservedGeneration = 2
		iriMachine.Status.State = iri.MachineState_MACHINE_TERMINATING
		srv.SetMachines([]*testingmachine.FakeMachine{iriMachine})
		Eventually(Object(machine)).Should(HaveField("Status.State", Equal(computev1alpha1.MachineStateTerminating)))
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
