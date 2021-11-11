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
package network

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"inet.af/netaddr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/onmetal/onmetal-api/apis/network/v1alpha1"

	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
)

var _ = Describe("IPAMRangeWebhook", func() {
	Context("Validate IPAMRange at creation", func() {
		ns := SetupTest()

		It("parent name shouldn't be empty string", func() {
			instance := &v1alpha1.IPAMRange{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: ns.Name,
				},
				Spec: v1alpha1.IPAMRangeSpec{
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
		It("parent CIDRs shouldn't overlap", func() {
			prefix1, err := netaddr.ParseIPPrefix("192.168.1.0/24")
			Expect(err).ToNot(HaveOccurred())
			instance := &v1alpha1.IPAMRange{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: ns.Name,
				},
				Spec: v1alpha1.IPAMRangeSpec{
					CIDRs: []commonv1alpha1.CIDR{
						commonv1alpha1.NewCIDR(prefix1),
					},
				},
			}
			Expect(k8sClient.Create(ctx, instance)).To(Succeed())

			prefix2, err := netaddr.ParseIPPrefix("192.168.1.0/25")
			Expect(err).ToNot(HaveOccurred())
			instance2 := &v1alpha1.IPAMRange{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test2",
					Namespace: ns.Name,
				},
				Spec: v1alpha1.IPAMRangeSpec{
					CIDRs: []commonv1alpha1.CIDR{
						commonv1alpha1.NewCIDR(prefix2),
					},
				},
			}

			Expect(k8sClient.Create(ctx, instance2)).To(
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
			instance := &v1alpha1.IPAMRange{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: ns.Name,
				},
				Spec: v1alpha1.IPAMRangeSpec{
					Parent: &corev1.LocalObjectReference{
						Name: "parent",
					},
					Requests: []v1alpha1.IPAMRangeRequest{
						{
							CIDR: &cidr1,
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, instance)).To(Succeed())

			prefix2, err := netaddr.ParseIPPrefix("192.168.1.0/25")
			Expect(err).ToNot(HaveOccurred())
			cidr2 := commonv1alpha1.NewCIDR(prefix2)
			instance2 := &v1alpha1.IPAMRange{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test2",
					Namespace: ns.Name,
				},
				Spec: v1alpha1.IPAMRangeSpec{
					Parent: &corev1.LocalObjectReference{
						Name: "parent",
					},
					Requests: []v1alpha1.IPAMRangeRequest{
						{
							CIDR: &cidr2,
						},
					},
				},
			}

			Expect(k8sClient.Create(ctx, instance2)).To(
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
})
