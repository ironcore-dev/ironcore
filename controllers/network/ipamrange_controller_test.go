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
	"context"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"inet.af/netaddr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	commonv1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	nwapiv1 "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
)

var _ = Describe("IPAMRangeReconciler", func() {
	Context("Reconcile an IPAMRange", func() {
		It("should reconcile parent without children", func() {
			reconciler := &IPAMRangeReconciler{
				Client: k8sClient,
			}

			ctx := context.Background()
			cidr, err := netaddr.ParseIPPrefix("192.168.1.0/24")
			Expect(err).ToNot(HaveOccurred())

			instance := &nwapiv1.IPAMRange{
				Spec: nwapiv1.IPAMRangeSpec{
					CIDRs: []commonv1.CIDR{
						commonv1.NewCIDR(cidr),
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "parent",
					Namespace: "default",
				},
			}
			controllerutil.AddFinalizer(instance, ipamRangeFinalizerName)

			Expect(k8sClient.Create(ctx, instance)).To(Succeed())
			defer func() {
				err := k8sClient.Delete(ctx, instance)
				Expect(err).NotTo(HaveOccurred())
			}()

			result, err := reconciler.Reconcile(ctx, ctrl.Request{
				NamespacedName: client.ObjectKey{
					Namespace: instance.Namespace,
					Name:      instance.Name,
				},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(BeZero())
			Expect(result.Requeue).To(BeFalse())
		})
		It("should fail if parent doesn't exists", func() {
			reconciler := &IPAMRangeReconciler{
				Client: k8sClient,
			}

			ctx := context.Background()
			ipPref, err := netaddr.ParseIPPrefix("192.168.1.0/24")
			cidr := commonv1.NewCIDR(ipPref)
			Expect(err).ToNot(HaveOccurred())

			instance := &nwapiv1.IPAMRange{
				Spec: nwapiv1.IPAMRangeSpec{
					Parent: &corev1.LocalObjectReference{
						Name: "parent",
					},
					Requests: []nwapiv1.IPAMRangeRequest{
						{
							CIDR: &cidr,
						},
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "child",
					Namespace: "default",
				},
			}
			controllerutil.AddFinalizer(instance, ipamRangeFinalizerName)

			Expect(k8sClient.Create(ctx, instance)).To(Succeed())
			defer func() {
				err := k8sClient.Delete(ctx, instance)
				Expect(err).NotTo(HaveOccurred())
			}()

			result, err := reconciler.Reconcile(ctx, ctrl.Request{
				NamespacedName: client.ObjectKey{
					Namespace: instance.Namespace,
					Name:      instance.Name,
				},
			})

			Expect(err.Error()).To(ContainSubstring("could not retrieve parent"))
			Expect(result.RequeueAfter).To(BeZero())
			Expect(result.Requeue).To(BeFalse())
		})
		//
		//It("should update parent allocations to IPAMRangeAllocationUsed", func(){}),
		//
		//It("should update parent allocations to IPAMRangeAllocationFailed if size is too big", func(){}),
		//
		//It("should update parent allocations to IPAMRangeAllocationFailed if range is too big", func(){}),
		//
		//It("should update parent allocations to IPAMRangeAllocationFailed if request range is not in parent's range", func(){}),
		//
		//It("should update parent allocations when CIDR is changed", func(){}),
	})
})
