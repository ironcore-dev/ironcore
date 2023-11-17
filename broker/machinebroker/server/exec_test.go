// Copyright 2022 IronCore authors
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

package server_test

import (
	"net/url"

	ori "github.com/ironcore-dev/ironcore/ori/apis/machine/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Exec", func() {
	_, srv := SetupTest()

	It("should return an exec-url with a token", func(ctx SpecContext) {
		By("issuing exec for an arbitrary machine id")
		res, err := srv.Exec(ctx, &ori.ExecRequest{MachineId: "my-machine"})
		Expect(err).NotTo(HaveOccurred())

		By("inspecting the result")
		u, err := url.ParseRequestURI(res.Url)
		Expect(err).NotTo(HaveOccurred(), "url is invalid: %q", res.Url)
		Expect(u.Host).To(Equal("localhost:8080"))
		Expect(u.Scheme).To(Equal("http"))
		Expect(u.Path).To(MatchRegexp(`/exec/[^/?&]{8}`))
	})
})
