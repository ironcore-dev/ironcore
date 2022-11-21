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
	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	machinebrokerv1alpha1 "github.com/onmetal/onmetal-api/machinebroker/api/v1alpha1"
	"github.com/onmetal/onmetal-api/machinebroker/apiutils"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"github.com/onmetal/onmetal-api/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("CreateMachine", func() {
	ctx := testutils.SetupContext()
	ns, srv := SetupTest(ctx)

	It("Should correctly create a machine", func() {
		By("creating a machine")

		var (
			annotations = map[string]string{"foo-ann": "bar-ann"}
			labels      = map[string]string{"foo-lbl": "bar-lbl"}
			metadata    = &ori.MachineMetadata{
				Namespace: "source-ns",
				Name:      "source-name",
				Uid:       "source-uid",
			}
			ignitionData     = []byte("ignition")
			volumeAttributes = map[string]string{"attri": "bute"}
			volumeSecretData = map[string][]byte{"secret": []byte("data")}
		)
		const (
			image         = "example.org/gardenlinux"
			class         = "my-class"
			networkHandle = "mynet"
			volumeDriver  = "ceph"
			volumeHandle  = "foobar"
		)
		res, err := srv.CreateMachine(ctx, &ori.CreateMachineRequest{
			Config: &ori.MachineConfig{
				Metadata: metadata,
				Image:    image,
				Class:    class,
				Ignition: &ori.IgnitionConfig{
					Data: ignitionData,
				},
				Annotations: annotations,
				Labels:      labels,
				NetworkInterfaces: []*ori.NetworkInterfaceConfig{
					{
						Name: "foo",
						Network: &ori.NetworkConfig{
							Handle: networkHandle,
						},
						Ips: []string{"192.168.178.1"},
						VirtualIp: &ori.VirtualIPConfig{
							Ip: "10.0.0.1",
						},
					},
				},
				Volumes: []*ori.VolumeConfig{
					{
						Name:   "foo",
						Device: "vdb",
						Access: &ori.VolumeAccessConfig{
							Driver:     volumeDriver,
							Handle:     volumeHandle,
							Attributes: volumeAttributes,
							SecretData: volumeSecretData,
						},
					},
					{
						Name:      "bar",
						Device:    "vdc",
						EmptyDisk: &ori.EmptyDiskConfig{},
					},
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(res.Machine.Id).To(HaveLen(63))
		Expect(res.Machine.Metadata).To(Equal(metadata))
		Expect(res.Machine.Annotations).To(Equal(annotations))
		Expect(res.Machine.Labels).To(Equal(labels))

		By("getting the created onmetal machine")
		onmetalMachine := &computev1alpha1.Machine{}
		onmetalMachineKey := client.ObjectKey{Namespace: ns.Name, Name: res.Machine.Id}
		Expect(k8sClient.Get(ctx, onmetalMachineKey, onmetalMachine)).To(Succeed())

		By("inspecting the created onmetal machine")
		Expect(onmetalMachine.Labels[machinebrokerv1alpha1.MachineIDLabel]).To(Equal(res.Machine.Id))
		Expect(onmetalMachine.Labels[machinebrokerv1alpha1.MachineManagerLabel]).To(Equal(machinebrokerv1alpha1.MachineBrokerManager))
		Expect(apiutils.GetMetadataAnnotation(onmetalMachine)).To(Equal(metadata))
		Expect(apiutils.GetAnnotationsAnnotation(onmetalMachine)).To(Equal(annotations))
		Expect(apiutils.GetLabelsAnnotation(onmetalMachine)).To(Equal(labels))
		Expect(onmetalMachine.Spec.MachineClassRef).To(Equal(corev1.LocalObjectReference{Name: class}))
		Expect(onmetalMachine.Spec.Image).To(Equal(image))
		Expect(onmetalMachine.Spec.IgnitionRef).NotTo(BeNil())
		Expect(onmetalMachine.Spec.Volumes).To(HaveLen(2))
		Expect(onmetalMachine.Spec.Volumes[0]).To(MatchFields(IgnoreExtras|IgnoreMissing, Fields{
			"Name":                 Equal("foo"),
			"Device":               Equal(pointer.String("vdb")),
			"LocalObjectReference": Not(BeNil()),
		}))
		Expect(onmetalMachine.Spec.Volumes[1]).To(MatchFields(IgnoreExtras|IgnoreMissing, Fields{
			"Name":      Equal("bar"),
			"Device":    Equal(pointer.String("vdc")),
			"EmptyDisk": Not(BeNil()),
		}))
		Expect(onmetalMachine.Spec.NetworkInterfaces).To(HaveLen(1))
		Expect(onmetalMachine.Spec.NetworkInterfaces[0]).To(MatchFields(IgnoreExtras|IgnoreMissing, Fields{
			"Name":                Equal("foo"),
			"NetworkInterfaceRef": Not(BeNil()),
		}))

		By("getting the referenced volume")
		onmetalVolume := &storagev1alpha1.Volume{}
		onmetalVolumeKey := client.ObjectKey{Namespace: ns.Name, Name: onmetalMachine.Spec.Volumes[0].VolumeRef.Name}
		Expect(k8sClient.Get(ctx, onmetalVolumeKey, onmetalVolume)).To(Succeed())

		By("inspecting the referenced volume")
		Expect(onmetalVolume.Labels[machinebrokerv1alpha1.MachineIDLabel]).To(Equal(res.Machine.Id))
		Expect(onmetalVolume.Labels[machinebrokerv1alpha1.VolumeNameLabel]).To(Equal("foo"))
		Expect(onmetalVolume.Annotations[commonv1alpha1.ManagedByAnnotation]).To(Equal(machinebrokerv1alpha1.MachineBrokerManager))
		onmetalVolumeStatus := onmetalVolume.Status
		Expect(onmetalVolumeStatus.State).To(Equal(storagev1alpha1.VolumeStateAvailable))
		onmetalVolumeAccess := onmetalVolumeStatus.Access
		Expect(onmetalVolumeAccess).NotTo(BeNil())
		Expect(onmetalVolumeAccess.Driver).To(Equal(volumeDriver))
		Expect(onmetalVolumeAccess.Handle).To(Equal(volumeHandle))
		Expect(onmetalVolumeAccess.VolumeAttributes).To(Equal(volumeAttributes))
		Expect(onmetalVolumeAccess.SecretRef).NotTo(BeNil())

		By("getting the referenced volume's secret")
		onmetalVolumeSecret := &corev1.Secret{}
		onmetalVolumeSecretKey := client.ObjectKey{Namespace: ns.Name, Name: onmetalVolumeAccess.SecretRef.Name}
		Expect(k8sClient.Get(ctx, onmetalVolumeSecretKey, onmetalVolumeSecret)).To(Succeed())

		By("inspecting the referenced volume's secret")
		Expect(onmetalVolumeSecret.Labels[machinebrokerv1alpha1.MachineIDLabel]).To(Equal(res.Machine.Id))
		Expect(onmetalVolumeSecret.Labels[machinebrokerv1alpha1.VolumeNameLabel]).To(Equal("foo"))
		Expect(onmetalVolumeSecret.Data).To(Equal(volumeSecretData))

		By("getting the referenced network interface")
		onmetalNetworkInterface := &networkingv1alpha1.NetworkInterface{}
		onmetalNetworkInterfaceKey := client.ObjectKey{Namespace: ns.Name, Name: onmetalMachine.Spec.NetworkInterfaces[0].NetworkInterfaceRef.Name}
		Expect(k8sClient.Get(ctx, onmetalNetworkInterfaceKey, onmetalNetworkInterface)).To(Succeed())

		By("inspecting the referenced network interface")
		Expect(onmetalNetworkInterface.Labels[machinebrokerv1alpha1.MachineIDLabel]).To(Equal(res.Machine.Id))
		Expect(onmetalNetworkInterface.Labels[machinebrokerv1alpha1.NetworkInterfaceNameLabel]).To(Equal("foo"))
		Expect(onmetalNetworkInterface.Spec.IPs).To(Equal([]networkingv1alpha1.IPSource{
			{
				Value: commonv1alpha1.MustParseNewIP("192.168.178.1"),
			},
		}))
		Expect(onmetalNetworkInterface.Spec.VirtualIP).NotTo(BeNil())
		Expect(onmetalNetworkInterface.Spec.VirtualIP.VirtualIPRef).NotTo(BeNil())

		By("getting the referenced network")
		onmetalNetwork := &networkingv1alpha1.Network{}
		onmetalNetworkKey := client.ObjectKey{Namespace: ns.Name, Name: onmetalNetworkInterface.Spec.NetworkRef.Name}
		Expect(k8sClient.Get(ctx, onmetalNetworkKey, onmetalNetwork)).To(Succeed())

		By("inspecting the referenced network")
		Expect(onmetalNetwork.Labels[machinebrokerv1alpha1.MachineIDLabel]).To(Equal(res.Machine.Id))
		Expect(onmetalNetwork.Labels[machinebrokerv1alpha1.NetworkInterfaceNameLabel]).To(Equal("foo"))
		Expect(onmetalNetwork.Annotations[commonv1alpha1.ManagedByAnnotation]).To(Equal(machinebrokerv1alpha1.MachineBrokerManager))
		Expect(onmetalNetwork.Spec.ProviderID).To(Equal(networkHandle))

		By("getting the referenced virtual ip")
		onmetalVirtualIP := &networkingv1alpha1.VirtualIP{}
		onmetalVirtualIPKey := client.ObjectKey{Namespace: ns.Name, Name: onmetalNetworkInterface.Spec.VirtualIP.VirtualIPRef.Name}
		Expect(k8sClient.Get(ctx, onmetalVirtualIPKey, onmetalVirtualIP)).To(Succeed())

		By("inspecting the referenced virtual ip")
		Expect(onmetalVirtualIP.Labels[machinebrokerv1alpha1.MachineIDLabel]).To(Equal(res.Machine.Id))
		Expect(onmetalVirtualIP.Labels[machinebrokerv1alpha1.NetworkInterfaceNameLabel]).To(Equal("foo"))
		Expect(onmetalVirtualIP.Labels[machinebrokerv1alpha1.IPFamilyLabel]).To(Equal(string(corev1.IPv4Protocol)))
		Expect(onmetalVirtualIP.Annotations[commonv1alpha1.ManagedByAnnotation]).To(Equal(machinebrokerv1alpha1.MachineBrokerManager))
		Expect(onmetalVirtualIP.Spec.Type).To(Equal(networkingv1alpha1.VirtualIPTypePublic))
		Expect(onmetalVirtualIP.Spec.IPFamily).To(Equal(corev1.IPv4Protocol))
		Expect(onmetalVirtualIP.Status.IP).To(Equal(commonv1alpha1.MustParseNewIP("10.0.0.1")))

		By("getting the referenced ignition secret")
		ignitionSecret := &corev1.Secret{}
		ignitionSecretKey := client.ObjectKey{Namespace: ns.Name, Name: onmetalMachine.Spec.IgnitionRef.Name}
		Expect(k8sClient.Get(ctx, ignitionSecretKey, ignitionSecret)).To(Succeed())

		By("inspecting the ignition secret")
		Expect(ignitionSecret.Labels[machinebrokerv1alpha1.MachineIDLabel]).To(Equal(res.Machine.Id))
		Expect(ignitionSecret.Data).To(Equal(map[string][]byte{computev1alpha1.DefaultIgnitionKey: ignitionData}))
	})
})
