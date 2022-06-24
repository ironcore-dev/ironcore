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
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/golang/mock/gomock"
	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	"github.com/onmetal/onmetal-api/machinepoollet/server"
	mockserver "github.com/onmetal/onmetal-api/mock/onmetal-api/machinepoollet/server"
	"github.com/onmetal/onmetal-api/terminal/fake"
	"github.com/onmetal/onmetal-api/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Machine", func() {
	ctx := testutils.SetupContext()
	ns := SetupTest(ctx)

	var (
		ctrl            *gomock.Controller
		machinePoolName string
	)
	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		DeferCleanup(ctrl.Finish)

		machinePoolName = fmt.Sprintf("machinepool-%s", testutils.RandomString(5))
	})

	It("should allow exec-ing onto a machine", func() {
		exec := mockserver.NewMockMachineExec(ctrl)
		term := fake.NewFakeTerm()
		defer term.Close()

		By("starting a machine pool server")
		srv, err := server.New(cfg, server.Options{
			MachineExec: exec,
			Host:        "localhost",
			CertDir:     machinePoolletCertDir,
			Auth: server.AuthOptions{
				MachinePoolName: machinePoolName,
				Authentication: server.AuthenticationOptions{
					ClientCAFile: machinePoolletCAFile,
				},
				Authorization: server.AuthorizationOptions{},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		go func() {
			_ = srv.Start(ctx)
		}()

		By("waiting for the server to have started")
		Eventually(srv.Started).Should(BeTrue())

		By("creating a machine pool")
		machinePool := &computev1alpha1.MachinePool{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns.Name,
				Name:      machinePoolName,
			},
		}
		Expect(k8sClient.Create(ctx, machinePool)).To(Succeed())

		By("patching the machine pool to have the url of the server")
		baseMachinePool := machinePool.DeepCopy()
		machinePool.Status.Addresses = []computev1alpha1.MachinePoolAddress{
			{
				Type:    computev1alpha1.MachinePoolHostName,
				Address: "localhost",
			},
		}
		machinePool.Status.DaemonEndpoints.MachinepoolletEndpoint.Port = int32(srv.Port())
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

		exec.EXPECT().Exec(gomock.Any(), ns.Name, machine.Name, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, namespace, name string, in io.Reader, out, err io.WriteCloser, resize <-chan remotecommand.TerminalSize) error {
				return term.Run(in, out, err, resize)
			})

		By("making an exec-call to the machine")
		req := onmetalClientSet.ComputeV1alpha1().RESTClient().
			Post().
			Namespace(ns.Name).
			Resource("machines").
			Name(machine.Name).
			SubResource("exec").
			VersionedParams(&computev1alpha1.MachineExecOptions{}, scheme.ParameterCodec)

		executor, err := remotecommand.NewSPDYExecutor(cfg, http.MethodPost, req.URL())
		Expect(err).NotTo(HaveOccurred())

		var (
			in, writeIn = io.Pipe()
			out         fake.ThreadSafeBuffer
			streamErr   = make(chan error)
		)
		go func() {
			err := executor.Stream(remotecommand.StreamOptions{
				Stdin:  in,
				Stdout: &out,
				Tty:    true,
			})
			streamErr <- err
		}()

		By("checking the stream does not crash and observing the terminal getting connected")
		Consistently(streamErr).ShouldNot(Receive())
		Eventually(term.Connected).Should(BeTrue())

		By("writing as client and observing the server reception")
		_, err = writeIn.Write([]byte("foo"))
		Expect(err).NotTo(HaveOccurred())
		Eventually(term.InBytes).Should(Equal([]byte("foo")))

		By("writing as server and observing the client reception")
		Expect(term.WriteOut([]byte("bar"))).To(Succeed())
		Eventually(out.Bytes).Should(Equal([]byte("bar")))

		By("closing the client reader")
		Expect(writeIn.Close()).Should(Succeed())

		By("waiting for the terminal to close")
		Eventually(term.Closed).Should(BeTrue())

		By("waiting for the executor to return")
		Eventually(streamErr).Should(Receive(BeNil()))
	})
})
