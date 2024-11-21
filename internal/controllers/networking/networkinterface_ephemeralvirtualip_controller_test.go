// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0
package networking

import (
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	"github.com/ironcore-dev/ironcore/utils/annotations"
	. "github.com/ironcore-dev/ironcore/utils/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

var _ = Describe("NetworkInterfaceEphemeralVirtualIP", func() {
	ns := SetupNamespace(&k8sClient)

	It("should create ephemeral virtual IPs for a network interface", func(ctx SpecContext) {
		By("creating a network interface that requires a virtual IP")
		vipSrc := networkingv1alpha1.VirtualIPSource{
			Ephemeral: &networkingv1alpha1.EphemeralVirtualIPSource{
				VirtualIPTemplate: &networkingv1alpha1.VirtualIPTemplateSpec{
					Spec: networkingv1alpha1.EphemeralVirtualIPSpec{
						VirtualIPSpec: networkingv1alpha1.VirtualIPSpec{
							Type:     networkingv1alpha1.VirtualIPTypePublic,
							IPFamily: corev1.IPv4Protocol,
						},
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
		By("Verifying OwnerRef is set for ephemeral volume")
		// Eventually(Object(vip)).Should(HaveField("ObjectMeta.OwnerReferences", Not(BeNil())))
		Eventually(Object(vip)).Should(HaveField("ObjectMeta.OwnerReferences", ConsistOf(MatchFields(IgnoreExtras, Fields{
			"APIVersion": Equal(networkingv1alpha1.SchemeGroupVersion.String()),
			"Kind":       Equal("NetworkInterface"),
			"Name":       Equal(nic.Name),
		})),
		))
	})
	It("should verify ownerRef is set for ephemeral virtual IPs based on ReclaimPolicy", func(ctx SpecContext) {
		By("creating a network interface with an ephemeral virtual IP")
		vipSrc := networkingv1alpha1.VirtualIPSource{
			Ephemeral: &networkingv1alpha1.EphemeralVirtualIPSource{
				VirtualIPTemplate: &networkingv1alpha1.VirtualIPTemplateSpec{
					Spec: networkingv1alpha1.EphemeralVirtualIPSpec{
						ReclaimPolicy: networkingv1alpha1.ReclaimPolicyTypeRetain,
						VirtualIPSpec: networkingv1alpha1.VirtualIPSpec{
							Type:     networkingv1alpha1.VirtualIPTypePublic,
							IPFamily: corev1.IPv4Protocol,
						},
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
		Eventually(Get(vip)).Should(Succeed())
		By("Verifying OwnerRef is not set for ephemeral volume when reclaim policy is retain")
		Eventually(Object(vip)).Should(HaveField("ObjectMeta.OwnerReferences", BeEmpty()))

		By("Updating reclaim policy to delete")
		baseNic := nic.DeepCopy()
		nic.Spec.VirtualIP.Ephemeral.VirtualIPTemplate.Spec.ReclaimPolicy = networkingv1alpha1.ReclaimPolicyTypeDelete
		Expect(k8sClient.Patch(ctx, nic, client.MergeFrom(baseNic))).To(Succeed())
		By("Verifying OwnerRef is updated for ephemeral volume")
		Eventually(Object(vip)).Should(HaveField("ObjectMeta.OwnerReferences", ConsistOf(MatchFields(IgnoreExtras, Fields{
			"APIVersion": Equal(networkingv1alpha1.SchemeGroupVersion.String()),
			"Kind":       Equal("NetworkInterface"),
			"Name":       Equal(nic.Name),
		})),
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
		undesiredVip := &networkingv1alpha1.VirtualIP{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "undesired-vip-",
			},
			Spec: networkingv1alpha1.VirtualIPSpec{
				Type:     networkingv1alpha1.VirtualIPTypePublic,
				IPFamily: corev1.IPv4Protocol,
			},
		}
		annotations.SetDefaultEphemeralManagedBy(undesiredVip)
		Expect(ctrl.SetControllerReference(nic, undesiredVip, k8sClient.Scheme())).To(Succeed())
		Expect(k8sClient.Create(ctx, undesiredVip)).To(Succeed())
		By("waiting for the undesired virtual IP to be gone")
		Eventually(Get(undesiredVip)).Should(Satisfy(apierrors.IsNotFound))
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
