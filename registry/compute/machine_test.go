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
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"

	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	"github.com/onmetal/onmetal-api/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/net"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Machine", func() {
	ctx := testutils.SetupContext()
	ns := SetupTest(ctx)

	It("should allow exec-ing onto a machine", func() {
		By("starting a dummy http server")
		srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			_, _ = fmt.Fprint(w, "Hello World")
		}))
		defer srv.Close()

		By("parsing the dummy http server url")
		srvUrl, err := url.ParseRequestURI(srv.URL)
		Expect(err).NotTo(HaveOccurred())
		srvPort, err := net.ParsePort(srvUrl.Port(), false)
		Expect(err).NotTo(HaveOccurred())

		By("creating a machine pool")
		machinePool := &computev1alpha1.MachinePool{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "machinepool-",
			},
		}
		Expect(k8sClient.Create(ctx, machinePool)).To(Succeed())

		By("patching the machine pool to have the url of the server")
		baseMachinePool := machinePool.DeepCopy()
		machinePool.Status.Addresses = []computev1alpha1.MachinePoolAddress{
			{
				Type:    computev1alpha1.MachinePoolExternalDNS,
				Address: srvUrl.Hostname(),
			},
		}
		machinePool.Status.DaemonEndpoints.MachinepoolletEndpoint.Port = int32(srvPort)
		Expect(k8sClient.Status().Patch(ctx, machinePool, client.MergeFrom(baseMachinePool))).To(Succeed())

		By("creating a machine on that machine pool")
		machine := &computev1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "machine-",
			},
			Spec: computev1alpha1.MachineSpec{
				Image:           "image",
				MachineClassRef: corev1.LocalObjectReference{Name: "machine-class"},
				MachinePoolRef:  &corev1.LocalObjectReference{Name: machinePool.Name},
			},
		}
		Expect(k8sClient.Create(ctx, machine)).To(Succeed())

		By("making an exec-call to the machine")
		res, err := onmetalClientSet.ComputeV1alpha1().RESTClient().
			Post().
			Namespace(ns.Name).
			Resource("machines").
			Name(machine.Name).
			SubResource("exec").
			VersionedParams(&computev1alpha1.MachineExecOptions{InsecureSkipTLSVerifyBackend: true}, scheme.ParameterCodec).
			DoRaw(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(res)).To(Equal("Hello World"))
	})
})
