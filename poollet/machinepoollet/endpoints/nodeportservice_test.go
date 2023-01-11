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
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	"github.com/onmetal/onmetal-api/poollet/machinepoollet/endpoints"
	. "github.com/onmetal/onmetal-api/utils/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/atomic"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("NodePortService", func() {
	ctx := SetupContext()
	ns := SetupTest(ctx)

	It("should correctly report and update the endpoints", func() {
		By("creating a service")
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "nodeport-",
			},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeNodePort,
				Ports: []corev1.ServicePort{
					{
						Name: portName,
						Port: 8080,
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, svc)).To(Succeed())

		By("creating the node port endpoints")
		eps, err := endpoints.NewNodePortServiceEndpoints(ctx, cfg, endpoints.NodePortServiceEndpointsOptions{
			Namespace:   ns.Name,
			ServiceName: svc.Name,
			PortName:    portName,
		})
		Expect(err).NotTo(HaveOccurred())

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
		eps.AddListener(endpoints.ListenerFunc(func() {
			notifyCount.Inc()
		}))

		By("fetching the endpoints")
		addresses, port := eps.GetEndpoints()
		Expect(addresses).To(BeEmpty())
		Expect(port).To(Equal(svc.Spec.Ports[0].NodePort))

		By("creating a node")
		node := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "node-",
			},
		}
		Expect(k8sClient.Create(ctx, node)).To(Succeed())

		By("adding a hostname address to the node")
		baseNode := node.DeepCopy()
		node.Status.Addresses = []corev1.NodeAddress{{Type: corev1.NodeHostName, Address: "node-hostname"}}
		Expect(k8sClient.Status().Patch(ctx, node, client.MergeFrom(baseNode)))

		By("waiting for the endpoint addresses to be updated")
		Eventually(func() []computev1alpha1.MachinePoolAddress {
			addresses, _ = eps.GetEndpoints()
			return addresses
		}).Should(Equal([]computev1alpha1.MachinePoolAddress{{Type: computev1alpha1.MachinePoolHostName, Address: "node-hostname"}}))
		Expect(notifyCount.Load()).To(BeEquivalentTo(1), "notify count not updated")

		By("deleting the node")
		Expect(k8sClient.Delete(ctx, node)).To(Succeed())

		By("waiting for the address to be removed")
		By("waiting for the endpoint addresses to be updated")
		Eventually(func() []computev1alpha1.MachinePoolAddress {
			addresses, _ = eps.GetEndpoints()
			return addresses
		}).Should(BeEmpty())
		Expect(notifyCount.Load()).To(BeEquivalentTo(2), "notify count not updated")
	})
})
