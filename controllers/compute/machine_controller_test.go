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
	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/apis/networking/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/apis/storage/v1alpha1"
	"github.com/onmetal/onmetal-api/controllers/shared"
	"github.com/onmetal/onmetal-api/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

var _ = Describe("MachineReconciler", func() {
	ctx := testutils.SetupContext()
	ns := SetupTest(ctx)

	It("should manage ephemeral objects", func() {
		By("creating a network")
		network := &networkingv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "network-",
			},
		}
		Expect(k8sClient.Create(ctx, network)).To(Succeed())

		By("creating a machine")
		machine := &computev1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "machine-",
			},
			Spec: computev1alpha1.MachineSpec{
				MachineClassRef: corev1.LocalObjectReference{Name: "machine-class"},
				Image:           "my-image:latest",
				NetworkInterfaces: []computev1alpha1.NetworkInterface{
					{
						Name: "interface",
						NetworkInterfaceSource: computev1alpha1.NetworkInterfaceSource{
							Ephemeral: &computev1alpha1.EphemeralNetworkInterfaceSource{
								NetworkInterfaceTemplate: &networkingv1alpha1.NetworkInterfaceTemplateSpec{
									Spec: networkingv1alpha1.NetworkInterfaceSpec{
										NetworkRef: corev1.LocalObjectReference{Name: network.Name},
										IPs:        []networkingv1alpha1.IPSource{{Value: commonv1alpha1.MustParseNewIP("10.0.0.1")}},
									},
								},
							},
						},
					},
				},
				Volumes: []computev1alpha1.Volume{
					{
						Name: "volume",
						VolumeSource: computev1alpha1.VolumeSource{
							Ephemeral: &computev1alpha1.EphemeralVolumeSource{
								VolumeTemplate: &storagev1alpha1.VolumeTemplateSpec{
									Spec: storagev1alpha1.VolumeSpec{
										VolumeClassRef: corev1.LocalObjectReference{Name: "my-class"},
										VolumePoolRef:  &corev1.LocalObjectReference{Name: "my-pool"},
										Resources: corev1.ResourceList{
											"storage": resource.MustParse("10Gi"),
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

		By("waiting for the network interface to exist")
		nic := &networkingv1alpha1.NetworkInterface{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns.Name,
				Name:      shared.MachineEphemeralNetworkInterfaceName(machine.Name, "interface"),
			},
		}
		Eventually(Object(nic)).Should(SatisfyAll(
			HaveField("ObjectMeta.OwnerReferences", ContainElement(metav1.OwnerReference{
				APIVersion:         computev1alpha1.SchemeGroupVersion.String(),
				Kind:               "Machine",
				Name:               machine.Name,
				UID:                machine.UID,
				Controller:         pointer.Bool(true),
				BlockOwnerDeletion: pointer.Bool(true),
			})),
			HaveField("Spec.MachineRef", &commonv1alpha1.LocalUIDReference{
				Name: machine.Name,
				UID:  machine.UID,
			}),
		))

		By("waiting for the volume to exist")
		volume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns.Name,
				Name:      shared.MachineEphemeralVolumeName(machine.Name, "volume"),
			},
		}
		Eventually(Object(volume)).Should(SatisfyAll(
			HaveField("ObjectMeta.OwnerReferences", ContainElement(metav1.OwnerReference{
				APIVersion:         computev1alpha1.SchemeGroupVersion.String(),
				Kind:               "Machine",
				Name:               machine.Name,
				UID:                machine.UID,
				Controller:         pointer.Bool(true),
				BlockOwnerDeletion: pointer.Bool(true),
			})),
			HaveField("Spec.ClaimRef", &commonv1alpha1.LocalUIDReference{
				Name: machine.Name,
				UID:  machine.UID,
			}),
		))

		By("waiting for the machine status to be updated")
		Eventually(Object(machine)).Should(SatisfyAll(
			HaveField("Status.NetworkInterfaces", []computev1alpha1.NetworkInterfaceStatus{
				{
					Name:  "interface",
					Phase: computev1alpha1.NetworkInterfacePhaseBound,
					IPs:   []commonv1alpha1.IP{commonv1alpha1.MustParseIP("10.0.0.1")},
				},
			}),
			HaveField("Status.Volumes", []computev1alpha1.VolumeStatus{
				{
					Name:  "volume",
					Phase: computev1alpha1.VolumePhaseBound,
				},
			}),
		))

		By("removing the ephemeral items from the machine")
		baseMachine := machine.DeepCopy()
		machine.Spec.NetworkInterfaces = nil
		machine.Spec.Volumes = nil
		Expect(k8sClient.Patch(ctx, machine, client.MergeFrom(baseMachine))).To(Succeed())

		By("waiting for the network interface to be gone")
		Eventually(Get(nic)).Should(Satisfy(apierrors.IsNotFound))

		By("waiting for the volume to be gone")
		Eventually(Get(volume)).Should(Satisfy(apierrors.IsNotFound))

		By("waiting for the machine status to be updated")
		Eventually(Object(machine)).Should(SatisfyAll(
			HaveField("Status.NetworkInterfaces", BeEmpty()),
			HaveField("Status.Volumes", BeEmpty()),
		))
	})
})
