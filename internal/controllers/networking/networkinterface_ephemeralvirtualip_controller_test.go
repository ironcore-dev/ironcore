/*
 * Copyright (c) 2021 by the IronCore authors.
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
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	"github.com/ironcore-dev/ironcore/utils/annotations"
	. "github.com/ironcore-dev/ironcore/utils/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

var _ = Describe("NetworkInterfaceEphemeralVirtualIP", func() {
	ns := SetupNamespace(&k8sClient)

	It("should create ephemeral virtual IPs for a network interface", func(ctx SpecContext) {
		By("creating a network interface that requires a virtual IP")
		vipSrc := networkingv1alpha1.VirtualIPSource{
			Ephemeral: &networkingv1alpha1.EphemeralVirtualIPSource{
				VirtualIPTemplate: &networkingv1alpha1.VirtualIPTemplateSpec{
					Spec: networkingv1alpha1.VirtualIPSpec{
						Type:     networkingv1alpha1.VirtualIPTypePublic,
						IPFamily: corev1.IPv4Protocol,
					},
				},
			},
		}
		nic := &networkingv1alpha1.NetworkInterface{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "nic-",
			},
			Spec: networkingv1alpha1.NetworkInterfaceSpec{
				NetworkRef: corev1.LocalObjectReference{Name: "my-network"},
				IPs: []networkingv1alpha1.IPSource{
					{
						Value: commonv1alpha1.MustParseNewIP("10.0.0.1"),
					},
				},
				VirtualIP: &vipSrc,
			},
		}
		Expect(k8sClient.Create(ctx, nic)).To(Succeed())

		By("waiting for the virtual IP to exist")
		vip := &networkingv1alpha1.VirtualIP{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns.Name,
				Name:      networkingv1alpha1.NetworkInterfaceVirtualIPName(nic.Name, vipSrc),
			},
		}
		Eventually(Object(vip)).Should(SatisfyAll(
			BeControlledBy(nic),
			HaveField("Spec", networkingv1alpha1.VirtualIPSpec{
				Type:     networkingv1alpha1.VirtualIPTypePublic,
				IPFamily: corev1.IPv4Protocol,
				TargetRef: &commonv1alpha1.LocalUIDReference{
					Name: nic.Name,
					UID:  nic.UID,
				},
			}),
		))
	})

	It("should delete undesired virtual IPs for a network interface", func(ctx SpecContext) {
		By("creating a network interface")
		nic := &networkingv1alpha1.NetworkInterface{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "nic-",
			},
			Spec: networkingv1alpha1.NetworkInterfaceSpec{
				NetworkRef: corev1.LocalObjectReference{Name: "my-network"},
				IPs:        []networkingv1alpha1.IPSource{{Value: commonv1alpha1.MustParseNewIP("10.0.0.1")}},
			},
		}
		Expect(k8sClient.Create(ctx, nic)).To(Succeed())

		By("creating an undesired virtual IP")
		vip := &networkingv1alpha1.VirtualIP{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "undesired-vip-",
			},
			Spec: networkingv1alpha1.VirtualIPSpec{
				Type:     networkingv1alpha1.VirtualIPTypePublic,
				IPFamily: corev1.IPv4Protocol,
			},
		}
		annotations.SetDefaultEphemeralManagedBy(vip)
		Expect(ctrl.SetControllerReference(nic, vip, k8sClient.Scheme())).To(Succeed())
		Expect(k8sClient.Create(ctx, vip)).To(Succeed())

		By("waiting for the virtual IP to be gone")
		Eventually(Get(vip)).Should(Satisfy(apierrors.IsNotFound))
	})

	It("should not delete externally managed virtual IPs for a network interface", func(ctx SpecContext) {
		By("creating a network interface")
		nic := &networkingv1alpha1.NetworkInterface{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "nic-",
			},
			Spec: networkingv1alpha1.NetworkInterfaceSpec{
				NetworkRef: corev1.LocalObjectReference{Name: "my-network"},
				IPs:        []networkingv1alpha1.IPSource{{Value: commonv1alpha1.MustParseNewIP("10.0.0.1")}},
			},
		}
		Expect(k8sClient.Create(ctx, nic)).To(Succeed())

		By("creating an undesired virtual IP")
		externalVip := &networkingv1alpha1.VirtualIP{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "external-vip-",
			},
			Spec: networkingv1alpha1.VirtualIPSpec{
				Type:     networkingv1alpha1.VirtualIPTypePublic,
				IPFamily: corev1.IPv4Protocol,
			},
		}
		Expect(ctrl.SetControllerReference(nic, externalVip, k8sClient.Scheme())).To(Succeed())
		Expect(k8sClient.Create(ctx, externalVip)).To(Succeed())

		By("asserting that the virtual IP is not being deleted")
		Eventually(Object(externalVip)).Should(HaveField("DeletionTimestamp", BeNil()))
	})
})
