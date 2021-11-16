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
package v1alpha1

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"inet.af/netaddr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"

	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
)

var _ = Describe("IPAMRangeWebhook", func() {
	Context("Validate IPAMRange at creation", func() {
		ns := SetupTest()

		It("parent name shouldn't be empty string", func() {
			instance := &IPAMRange{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: ns.Name,
				},
				Spec: IPAMRangeSpec{
					Parent: &corev1.LocalObjectReference{
						Name: "",
					},
				},
			}

			Expect(k8sClient.Create(ctx, instance)).To(
				WithTransform(
					func(err error) string { return err.Error() },
					ContainSubstring(field.Required(
						field.
							NewPath("spec").
							Child("parent").
							Child("name"),
						"",
					).Error()),
				),
			)
		})
		It("cidrs should be empty for child IPAMRange", func() {
			prefix1, err := netaddr.ParseIPPrefix("192.168.1.0/24")
			Expect(err).ToNot(HaveOccurred())
			instance := &IPAMRange{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: ns.Name,
				},
				Spec: IPAMRangeSpec{
					Parent: &corev1.LocalObjectReference{
						Name: "parent",
					},
					CIDRs: []commonv1alpha1.CIDR{
						commonv1alpha1.NewCIDR(prefix1),
					},
				},
			}
			Expect(k8sClient.Create(ctx, instance)).To(
				WithTransform(
					func(err error) string { return err.Error() },
					ContainSubstring(field.Forbidden(
						field.
							NewPath("spec").
							Child("cidrs"),
						"CIDRs should be empty for child IPAMRange",
					).Error()),
				),
			)
		})
		It("parent CIDRs shouldn't overlap", func() {
			prefix1, err := netaddr.ParseIPPrefix("192.168.1.0/24")
			Expect(err).ToNot(HaveOccurred())
			prefix2, err := netaddr.ParseIPPrefix("192.168.1.0/25")
			Expect(err).ToNot(HaveOccurred())
			instance := &IPAMRange{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: ns.Name,
				},
				Spec: IPAMRangeSpec{
					CIDRs: []commonv1alpha1.CIDR{commonv1alpha1.NewCIDR(prefix1), commonv1alpha1.NewCIDR(prefix2)},
				},
			}
			Expect(k8sClient.Create(ctx, instance)).To(
				WithTransform(
					func(err error) string { return err.Error() },
					ContainSubstring(field.Duplicate(
						field.
							NewPath("spec").
							Child("cidrs"),
						prefix2.String(),
					).Error()),
				),
			)
		})
		It("requests shouldn't overlap", func() {
			prefix1, err := netaddr.ParseIPPrefix("192.168.1.0/24")
			Expect(err).ToNot(HaveOccurred())
			cidr1 := commonv1alpha1.NewCIDR(prefix1)
			prefix2, err := netaddr.ParseIPPrefix("192.168.1.0/25")
			Expect(err).ToNot(HaveOccurred())
			cidr2 := commonv1alpha1.NewCIDR(prefix2)
			instance := &IPAMRange{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: ns.Name,
				},
				Spec: IPAMRangeSpec{
					Parent: &corev1.LocalObjectReference{
						Name: "parent",
					},
					Requests: []IPAMRangeRequest{{CIDR: &cidr1}, {CIDR: &cidr2}},
				},
			}
			Expect(k8sClient.Create(ctx, instance)).To(
				WithTransform(
					func(err error) string { return err.Error() },
					ContainSubstring(field.Duplicate(
						field.
							NewPath("spec").
							Child("requests").
							Child("cidr"),
						prefix2.String(),
					).Error()),
				),
			)

		})
	})

	Context("Validate IPAMRange at update", func() {
		ns := SetupTest()

		It("parent CIDRs shouldn't overlap", func() {
			prefix1, err := netaddr.ParseIPPrefix("192.168.1.0/24")
			Expect(err).ToNot(HaveOccurred())
			instance := &IPAMRange{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: ns.Name,
				},
				Spec: IPAMRangeSpec{
					CIDRs: []commonv1alpha1.CIDR{
						commonv1alpha1.NewCIDR(prefix1),
					},
				},
			}
			Expect(k8sClient.Create(ctx, instance)).To(Succeed())

			By("updating parent with duplicate CIDR")
			prefix2, err := netaddr.ParseIPPrefix("192.168.1.0/25")
			Expect(err).ToNot(HaveOccurred())
			instance.Spec.CIDRs = append(instance.Spec.CIDRs, commonv1alpha1.NewCIDR(prefix2))
			Expect(k8sClient.Update(ctx, instance)).To(
				WithTransform(
					func(err error) string { return err.Error() },
					ContainSubstring(field.Duplicate(
						field.
							NewPath("spec").
							Child("cidrs"),
						prefix2.String(),
					).Error()),
				),
			)
		})
		It("requests shouldn't overlap", func() {
			prefix1, err := netaddr.ParseIPPrefix("192.168.1.0/24")
			Expect(err).ToNot(HaveOccurred())
			cidr1 := commonv1alpha1.NewCIDR(prefix1)
			instance := &IPAMRange{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: ns.Name,
				},
				Spec: IPAMRangeSpec{
					Parent: &corev1.LocalObjectReference{
						Name: "parent",
					},
					Requests: []IPAMRangeRequest{{CIDR: &cidr1}},
				},
			}
			Expect(k8sClient.Create(ctx, instance)).To(Succeed())

			By("updating IPAMRange with duplicate request")
			prefix2, err := netaddr.ParseIPPrefix("192.168.1.0/25")
			Expect(err).ToNot(HaveOccurred())
			cidr2 := commonv1alpha1.NewCIDR(prefix2)
			instance.Spec.Requests = append(instance.Spec.Requests, IPAMRangeRequest{CIDR: &cidr2})
			Expect(k8sClient.Update(ctx, instance)).To(
				WithTransform(
					func(err error) string { return err.Error() },
					ContainSubstring(field.Duplicate(
						field.
							NewPath("spec").
							Child("requests").
							Child("cidr"),
						prefix2.String(),
					).Error()),
				),
			)
		})
		It("forbidden to delete used CIDRs", func() {
			prefix1, err := netaddr.ParseIPPrefix("192.168.1.0/24")
			Expect(err).ToNot(HaveOccurred())
			cidr1 := commonv1alpha1.NewCIDR(prefix1)
			prefix2, err := netaddr.ParseIPPrefix("192.168.2.0/24")
			Expect(err).ToNot(HaveOccurred())
			cidr2 := commonv1alpha1.NewCIDR(prefix2)
			instance := &IPAMRange{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: ns.Name,
				},
				Spec: IPAMRangeSpec{
					CIDRs: []commonv1alpha1.CIDR{cidr1, cidr2},
				},
			}
			Expect(k8sClient.Create(ctx, instance)).To(Succeed())

			instance.Status = IPAMRangeStatus{
				Allocations: []IPAMRangeAllocationStatus{
					{State: IPAMRangeAllocationUsed, CIDR: &cidr2},
				},
			}
			Expect(k8sClient.Status().Update(ctx, instance)).To(Succeed())

			By("removing one of the CIDRs")
			instance.Spec.CIDRs = []commonv1alpha1.CIDR{cidr1}
			Expect(k8sClient.Update(ctx, instance)).To(
				WithTransform(
					func(err error) string { return err.Error() },
					ContainSubstring(field.Forbidden(
						field.
							NewPath("spec").
							Child("cidrs"),
						fmt.Sprintf("CIDR %s is used by child request", cidr2.String()),
					).Error()),
				),
			)
		})
	})

	Context("Validate IPAMRange at deletion", func() {
		ns := SetupTest()

		It("shouldn't allow to delete IPAMRange if it has allocations", func() {
			parentPrefix, err := netaddr.ParseIPPrefix("192.168.1.0/24")
			Expect(err).ToNot(HaveOccurred())
			cidr := commonv1alpha1.NewCIDR(parentPrefix)
			instance := &IPAMRange{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "parent",
					Namespace: ns.Name,
				},
				Spec: IPAMRangeSpec{
					CIDRs: []commonv1alpha1.CIDR{cidr},
				},
			}
			Expect(k8sClient.Create(ctx, instance)).To(Succeed())

			instance.Status = IPAMRangeStatus{
				Allocations: []IPAMRangeAllocationStatus{
					{State: IPAMRangeAllocationUsed, CIDR: &cidr},
				},
			}
			Expect(k8sClient.Status().Update(ctx, instance)).To(Succeed())

			Expect(k8sClient.Delete(ctx, instance)).To(
				WithTransform(
					func(err error) string { return err.Error() },
					ContainSubstring("there's still children that depend on this IPAMRange"),
				),
			)
		})
	})
})
