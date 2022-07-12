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
 * See the License for the specific language governing permissions and limitations under the License.
 */

package endpoints_test

import (
	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	. "github.com/onmetal/onmetal-api/machinepoollet/endpoints"
	"github.com/onmetal/onmetal-api/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/atomic"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("LoadBalancerService", func() {
	ctx := testutils.SetupContext()
	ns := SetupTest(ctx)

	It("should correctly report and update the endpoints", func() {
		By("creating a service")
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "lb-",
			},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeLoadBalancer,
				Ports: []corev1.ServicePort{
					{
						Port: 8080,
						Name: portName,
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, svc)).To(Succeed())

		By("creating the loadbalancerservice endpoints")
		eps, err := NewLoadBalancerServiceEndpoints(ctx, cfg, LoadBalancerServiceEndpointsOptions{
			Namespace:   ns.Name,
			ServiceName: svc.Name,
			PortName:    portName,
		})
		Expect(err).NotTo(HaveOccurred())

		By("inspecting the initial values")
		addresses, port := eps.GetEndpoints()
		Expect(addresses).To(BeEmpty())
		Expect(port).To(Equal(int32(8080)))

		By("creating a manager")
		mgr, err := ctrl.NewManager(cfg, ctrl.Options{
			MetricsBindAddress:     "0",
			HealthProbeBindAddress: "0",
		})
		Expect(err).NotTo(HaveOccurred())

		By("setting up the endpoints with the manager")
		Expect(eps.SetupWithManager(mgr)).To(Succeed())

		By("starting the manager")
		go func() {
			defer GinkgoRecover()
			Expect(mgr.Start(ctx)).To(Succeed())
		}()

		By("adding a listener")
		var notifyCount atomic.Int32
		eps.AddListener(ListenerFunc(func() {
			notifyCount.Inc()
		}))

		By("updating the service to include load balancer ingress items")
		baseSvc := svc.DeepCopy()
		svc.Status.LoadBalancer.Ingress = []corev1.LoadBalancerIngress{
			{IP: "127.0.0.1"},
			{Hostname: "foo.example.org"},
		}
		Expect(k8sClient.Status().Patch(ctx, svc, client.MergeFrom(baseSvc))).To(Succeed())

		By("waiting for the endpoint addresses to be updated")
		Eventually(func() []computev1alpha1.MachinePoolAddress {
			addresses, _ = (*eps).GetEndpoints()
			return addresses
		}).Should(Equal([]computev1alpha1.MachinePoolAddress{
			{Type: computev1alpha1.MachinePoolExternalDNS, Address: "foo.example.org"},
			{Type: computev1alpha1.MachinePoolExternalIP, Address: "127.0.0.1"},
		}))
		Expect(notifyCount.Load()).To(BeEquivalentTo(1), "notify count not updated")

		By("emptying the service load balancer ingress items")
		baseSvc = svc.DeepCopy()
		svc.Status.LoadBalancer.Ingress = nil
		Expect(k8sClient.Status().Patch(ctx, svc, client.MergeFrom(baseSvc))).To(Succeed())

		By("waiting for the address to be removed")
		By("waiting for the endpoint addresses to be updated")
		Eventually(func() []computev1alpha1.MachinePoolAddress {
			addresses, _ = (*eps).GetEndpoints()
			return addresses
		}).Should(BeEmpty())
		Expect(notifyCount.Load()).To(BeEquivalentTo(2), "notify count not updated")
	})
})
