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
package ipam

import (
	"github.com/onmetal/controller-utils/clientutils"
	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	ipamv1alpha1 "github.com/onmetal/onmetal-api/apis/ipam/v1alpha1"
	"github.com/onmetal/onmetal-api/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("PrefixReconciler", func() {
	ctx := testutils.SetupContext()
	ns := SetupTest(ctx)

	It("should mark root prefixes as ready", func() {
		By("creating a root prefix")
		prefixValue := commonv1alpha1.MustParseNewIPPrefix("10.0.0.0/24")
		prefix := &ipamv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-root-",
			},
			Spec: ipamv1alpha1.PrefixSpec{
				Prefix: prefixValue,
			},
		}
		Expect(k8sClient.Create(ctx, prefix)).To(Succeed())

		By("asserting there is no allocation for root prefixes")
		Consistently(func(g Gomega) {
			list := &ipamv1alpha1.PrefixAllocationList{}
			g.Expect(clientutils.ListAndFilterControlledBy(ctx, k8sClient, prefix, list, client.InNamespace(ns.Name))).To(Succeed())
			g.Expect(list.Items).To(BeEmpty())
		}).Should(Succeed())

		By("waiting for the prefix to be marked as ready and report its available ranges")
		prefixKey := client.ObjectKeyFromObject(prefix)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, prefixKey, prefix)).To(Succeed())
			g.Expect(prefix.Status.Phase).To(Equal(ipamv1alpha1.PrefixPhaseAllocated))
			g.Expect(prefix.Status.Used).To(BeEmpty())
		}).Should(Succeed())
	})

	It("should allocate child prefixes", func() {
		By("creating a root prefix")
		prefixValue := commonv1alpha1.MustParseNewIPPrefix("10.0.0.0/24")
		rootPrefix := &ipamv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-root-",
			},
			Spec: ipamv1alpha1.PrefixSpec{
				Prefix: prefixValue,
			},
		}
		Expect(k8sClient.Create(ctx, rootPrefix)).To(Succeed())

		By("creating a child prefix")
		childPrefix := &ipamv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-child-",
			},
			Spec: ipamv1alpha1.PrefixSpec{
				IPFamily:     corev1.IPv4Protocol,
				PrefixLength: 28,
				ParentRef: &corev1.LocalObjectReference{
					Name: rootPrefix.Name,
				},
			},
		}
		Expect(k8sClient.Create(ctx, childPrefix)).To(Succeed())

		By("waiting for the child prefix to be marked as ready and report its available ranges")
		expectedChildPrefix := commonv1alpha1.MustParseIPPrefix("10.0.0.0/28")
		childPrefixKey := client.ObjectKeyFromObject(childPrefix)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, childPrefixKey, childPrefix)).To(Succeed())
			g.Expect(childPrefix.Status.Phase).To(Equal(ipamv1alpha1.PrefixPhaseAllocated))
			g.Expect(childPrefix.Spec.ParentRef).To(Equal(&corev1.LocalObjectReference{
				Name: rootPrefix.Name,
			}))
			g.Expect(childPrefix.Status.Used).To(BeEmpty())
		}).Should(Succeed())

		By("asserting the parent's available ranges have been updated")
		rootPrefixKey := client.ObjectKeyFromObject(rootPrefix)
		Expect(k8sClient.Get(ctx, rootPrefixKey, rootPrefix)).To(Succeed())
		Expect(rootPrefix.Status.Used).To(ConsistOf(expectedChildPrefix))

		By("asserting there is a single valid allocation")
		list := &ipamv1alpha1.PrefixAllocationList{}
		Expect(clientutils.ListAndFilterControlledBy(ctx, k8sClient, childPrefix, list, client.InNamespace(ns.Name))).To(Succeed())
		Expect(list.Items).To(HaveLen(1))
		allocation := list.Items[0]
		Expect(allocation.Status.Phase).To(Equal(ipamv1alpha1.PrefixAllocationPhaseAllocated))
		Expect(allocation.Spec).To(Equal(ipamv1alpha1.PrefixAllocationSpec{
			IPFamily:     corev1.IPv4Protocol,
			PrefixLength: 28,
			PrefixRef: &corev1.LocalObjectReference{
				Name: rootPrefix.Name,
			},
		}))
		Expect(allocation.Status.Prefix).To(HaveValue(Equal(expectedChildPrefix)))
	})

	It("should leave prefixes in pending state when they can't be allocated", func() {
		By("creating a root prefix")
		prefixValue := commonv1alpha1.MustParseNewIPPrefix("10.0.0.0/24")
		rootPrefix := &ipamv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-root-",
			},
			Spec: ipamv1alpha1.PrefixSpec{
				Prefix: prefixValue,
			},
		}
		Expect(k8sClient.Create(ctx, rootPrefix)).To(Succeed())

		By("creating a child prefix")
		childPrefix := &ipamv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-child-",
			},
			Spec: ipamv1alpha1.PrefixSpec{
				IPFamily:     corev1.IPv4Protocol,
				PrefixLength: 8,
				ParentRef: &corev1.LocalObjectReference{
					Name: rootPrefix.Name,
				},
			},
		}
		Expect(k8sClient.Create(ctx, childPrefix)).To(Succeed())

		By("asserting there is a single, non-succeeded allocation and the child does not become ready")
		childPrefixKey := client.ObjectKeyFromObject(childPrefix)
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, childPrefixKey, childPrefix)).Should(Succeed())
			g.Expect(childPrefix.Status.Phase).To(Equal(ipamv1alpha1.PrefixPhasePending))

			list := &ipamv1alpha1.PrefixAllocationList{}
			g.Expect(clientutils.ListAndFilterControlledBy(ctx, k8sClient, childPrefix, list, client.InNamespace(ns.Name))).To(Succeed())
			g.Expect(list.Items).To(HaveLen(1))
			allocation := list.Items[0]
			g.Expect(allocation.Status.Phase).To(Equal(ipamv1alpha1.PrefixAllocationPhaseFailed))
		}).Should(Succeed())
	})

	It("should assign a prefix on matching parents", func() {
		By("creating a child prefix")
		childPrefix := &ipamv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-child-",
			},
			Spec: ipamv1alpha1.PrefixSpec{
				IPFamily:     corev1.IPv4Protocol,
				PrefixLength: 28,
				ParentSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"foo": "bar",
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, childPrefix)).To(Succeed())

		By("checking that the child prefix is not being assigned")
		childPrefixKey := client.ObjectKeyFromObject(childPrefix)
		Consistently(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, childPrefixKey, childPrefix)).Should(Succeed())
			g.Expect(childPrefix.Status.Phase).To(Or(
				BeEquivalentTo(""),
				Equal(ipamv1alpha1.PrefixPhasePending),
			))
			g.Expect(childPrefix.Spec.ParentRef).To(BeNil())
		}).Should(Succeed())

		By("creating a root prefix that would fit but does not match")
		prefixValue := commonv1alpha1.MustParseNewIPPrefix("10.0.0.0/24")
		notMatchingRootPrefix := &ipamv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "not-matching-root-",
			},
			Spec: ipamv1alpha1.PrefixSpec{
				Prefix: prefixValue,
			},
		}
		Expect(k8sClient.Create(ctx, notMatchingRootPrefix)).To(Succeed())

		By("checking that the child prefix is not being assigned")
		Consistently(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, childPrefixKey, childPrefix)).Should(Succeed())
			g.Expect(childPrefix.Status.Phase).To(Equal(ipamv1alpha1.PrefixPhasePending))
			g.Expect(childPrefix.Spec.ParentRef).To(BeNil())
		}).Should(Succeed())

		By("creating a root prefix that fits and matches")
		rootPrefix := &ipamv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "matching-root-",
				Labels: map[string]string{
					"foo": "bar",
				},
			},
			Spec: ipamv1alpha1.PrefixSpec{
				Prefix: prefixValue,
			},
		}
		Expect(k8sClient.Create(ctx, rootPrefix)).To(Succeed())

		By("waiting for the child prefix to be marked as ready and report its available ranges")
		expectedChildPrefix := commonv1alpha1.MustParseIPPrefix("10.0.0.0/28")
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, childPrefixKey, childPrefix)).To(Succeed())
			g.Expect(childPrefix.Status.Phase).To(Equal(ipamv1alpha1.PrefixPhaseAllocated))
			g.Expect(childPrefix.Spec.ParentRef).To(Equal(&corev1.LocalObjectReference{
				Name: rootPrefix.Name,
			}))
			g.Expect(childPrefix.Spec.Prefix).To(HaveValue(Equal(expectedChildPrefix)))
		}).Should(Succeed())

		By("asserting the parent's used ranges have been updated")
		rootPrefixKey := client.ObjectKeyFromObject(rootPrefix)
		Expect(k8sClient.Get(ctx, rootPrefixKey, rootPrefix)).To(Succeed())
		Expect(rootPrefix.Status.Used).To(ConsistOf(expectedChildPrefix))

		By("asserting there is a single valid allocation")
		list := &ipamv1alpha1.PrefixAllocationList{}
		Expect(clientutils.ListAndFilterControlledBy(ctx, k8sClient, childPrefix, list, client.InNamespace(ns.Name))).To(Succeed())
		Expect(list.Items).To(HaveLen(1))
		allocation := &list.Items[0]
		Expect(allocation.Spec).To(Equal(ipamv1alpha1.PrefixAllocationSpec{
			IPFamily:     corev1.IPv4Protocol,
			PrefixLength: 28,
			PrefixRef: &corev1.LocalObjectReference{
				Name: rootPrefix.Name,
			},
			PrefixSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"foo": "bar",
				},
			},
		}))
		Expect(allocation.Status.Phase).To(Equal(ipamv1alpha1.PrefixAllocationPhaseAllocated))
		Expect(allocation.Status.Prefix).To(HaveValue(Equal(expectedChildPrefix)))
	})

	It("should allocate a static prefix from another prefix", func() {
		By("creating a root prefix")
		rootPrefix := &ipamv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "root-",
			},
			Spec: ipamv1alpha1.PrefixSpec{
				Prefix: commonv1alpha1.MustParseNewIPPrefix("10.0.0.0/8"),
			},
		}
		Expect(k8sClient.Create(ctx, rootPrefix)).To(Succeed())

		By("creating a child prefix referencing root allocating a static prefix")
		childPrefix := &ipamv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "child-",
			},
			Spec: ipamv1alpha1.PrefixSpec{
				Prefix:    commonv1alpha1.MustParseNewIPPrefix("10.0.0.0/9"),
				ParentRef: &corev1.LocalObjectReference{Name: rootPrefix.Name},
			},
		}
		Expect(k8sClient.Create(ctx, childPrefix)).To(Succeed())

		By("waiting for the child prefix to be allocated")
		childPrefixKey := client.ObjectKeyFromObject(childPrefix)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, childPrefixKey, childPrefix)).To(Succeed())
			g.Expect(childPrefix.Status.Phase).To(Equal(ipamv1alpha1.PrefixPhaseAllocated))
		}).Should(Succeed())
	})

	It("should not allocate a prefix equal to the prefix of the parent", func() {
		By("creating a root prefix")
		rootPrefix := &ipamv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "root-",
			},
			Spec: ipamv1alpha1.PrefixSpec{
				Prefix: commonv1alpha1.MustParseNewIPPrefix("10.0.0.0/8"),
			},
		}
		Expect(k8sClient.Create(ctx, rootPrefix)).To(Succeed())

		By("creating a child prefix referencing root allocating a static prefix")
		childPrefix := &ipamv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "child-",
			},
			Spec: ipamv1alpha1.PrefixSpec{
				Prefix:    commonv1alpha1.MustParseNewIPPrefix("10.0.0.0/8"),
				ParentRef: &corev1.LocalObjectReference{Name: rootPrefix.Name},
			},
		}
		Expect(k8sClient.Create(ctx, childPrefix)).To(Succeed())

		By("checking the child prefix does not get allocated")
		childPrefixKey := client.ObjectKeyFromObject(childPrefix)
		Consistently(func(g Gomega) {
			Expect(k8sClient.Get(ctx, childPrefixKey, childPrefix)).To(Succeed())
			g.Expect(childPrefix.Status.Phase).NotTo(Equal(ipamv1alpha1.PrefixPhaseAllocated))
		}).Should(Succeed())
	})

	It("should not allocate a prefix length equal to the prefix bits of the parent", func() {
		By("creating a root prefix")
		rootPrefix := &ipamv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "root-",
			},
			Spec: ipamv1alpha1.PrefixSpec{
				Prefix: commonv1alpha1.MustParseNewIPPrefix("10.0.0.0/8"),
			},
		}
		Expect(k8sClient.Create(ctx, rootPrefix)).To(Succeed())

		By("creating a child prefix referencing root allocating a static prefix")
		childPrefix := &ipamv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "child-",
			},
			Spec: ipamv1alpha1.PrefixSpec{
				IPFamily:     corev1.IPv4Protocol,
				PrefixLength: 8,
				ParentRef:    &corev1.LocalObjectReference{Name: rootPrefix.Name},
			},
		}
		Expect(k8sClient.Create(ctx, childPrefix)).To(Succeed())

		By("checking the child prefix does not get allocated")
		childPrefixKey := client.ObjectKeyFromObject(childPrefix)
		Consistently(func(g Gomega) {
			Expect(k8sClient.Get(ctx, childPrefixKey, childPrefix)).To(Succeed())
			g.Expect(childPrefix.Status.Phase).NotTo(Equal(ipamv1alpha1.PrefixPhaseAllocated))
		}).Should(Succeed())
	})
})
