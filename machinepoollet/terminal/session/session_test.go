// Copyright 2022 OnMetal authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package session_test

import (
	"io"

	"github.com/onmetal/onmetal-api/machinepoollet/terminal/fake"
	. "github.com/onmetal/onmetal-api/machinepoollet/terminal/session"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/tools/remotecommand"
)

var _ = Describe("Session", func() {
	It("should run the terminal", func() {
		var (
			term     = fake.NewFakeTerm()
			out, err fake.ThreadSafeBuffer
			resize   = make(chan remotecommand.TerminalSize)
		)
		defer term.Close()

		By("starting a session")
		s := Start(term, false)

		By("running a connection")
		runErr := make(chan error, 1)
		go func() {
			runErr <- s.Run(nil, &out, &err, resize)
		}()

		By("waiting for the terminal to be connected")
		Eventually(term.Connected).Should(BeTrue())

		By("requesting resizes")
		Eventually(resize).Should(BeSent(remotecommand.TerminalSize{Width: 10, Height: 20}))
		Eventually(resize).Should(BeSent(remotecommand.TerminalSize{Width: 100, Height: 200}))

		By("inspecting the resizes got accepted")
		Eventually(term.Resizes).Should(Equal([]remotecommand.TerminalSize{
			{Width: 10, Height: 20},
			{Width: 100, Height: 200},
		}))

		By("writing data to the terminal after the resizes, as only then we are sure the writers are attached")
		Expect(term.WriteOut([]byte("foo"))).To(Succeed())
		Expect(term.WriteErr([]byte("bar"))).To(Succeed())

		By("inspecting the data got written to the connection")
		Eventually(out.Bytes).Should(Equal([]byte("foo")))
		Eventually(err.Bytes).Should(Equal([]byte("bar")))

		By("closing the terminal")
		term.Close()

		By("waiting for the connection to exit")
		Eventually(runErr).Should(Receive(nil))

		By("waiting for the session to exit")
		Eventually(s.Exited).Should(BeClosed())

		By("checking a new run does not get accepted")
		Expect(s.Run(nil, &out, &err, resize)).To(MatchError(ErrStoppedAccepting))
	})

	It("should exit the connection if the detach signal is sent", func() {
		var (
			term        = fake.NewFakeTerm()
			in, writeIn = io.Pipe()
		)
		defer term.Close()

		By("starting a session")
		s := Start(term, false)

		By("running a connection")
		runErr := make(chan error, 1)
		go func() {
			runErr <- s.Run(in, nil, nil, nil)
		}()

		By("waiting for the terminal to be connected")
		Eventually(term.Connected).Should(BeTrue())

		By("writing the escape sequence to the terminal")
		_, err := writeIn.Write([]byte{16, 17})
		Expect(err).NotTo(HaveOccurred())

		By("waiting for the connection run to succeed")
		Eventually(runErr).Should(Receive(nil))
	})

	It("should multiplex writes", func() {
		var (
			term                   = fake.NewFakeTerm()
			out1, out2, err1, err2 fake.ThreadSafeBuffer
			resize1                = make(chan remotecommand.TerminalSize)
			resize2                = make(chan remotecommand.TerminalSize)
		)
		defer term.Close()

		By("starting the session")
		s := Start(term, false)

		By("running two connections")
		runErr1 := make(chan error, 1)
		go func() {
			runErr1 <- s.Run(nil, &out1, &err1, resize1)
		}()

		runErr2 := make(chan error, 1)
		go func() {
			runErr2 <- s.Run(nil, &out2, &err2, resize2)
		}()

		By("issuing a resize on both connections, as only then we are sure they are connected")
		Eventually(resize1).Should(BeSent(remotecommand.TerminalSize{}))
		Eventually(resize2).Should(BeSent(remotecommand.TerminalSize{}))
		Eventually(term.Resizes).Should(ConsistOf(remotecommand.TerminalSize{}, remotecommand.TerminalSize{}))

		By("writing something to out and err")
		Expect(term.WriteOut([]byte("foo"))).To(Succeed())
		Expect(term.WriteErr([]byte("bar"))).To(Succeed())

		By("inspecting both buffers received all values")
		Eventually(out1.Bytes).Should(Equal([]byte("foo")))
		Eventually(out2.Bytes).Should(Equal([]byte("foo")))
		Eventually(err1.Bytes).Should(Equal([]byte("bar")))
		Eventually(err2.Bytes).Should(Equal([]byte("bar")))
	})
})
