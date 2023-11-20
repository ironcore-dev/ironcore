// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	"net/url"

	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Exec", func() {
	_, srv := SetupTest()

	It("should return an exec-url with a token", func(ctx SpecContext) {
		By("issuing exec for an arbitrary machine id")
		res, err := srv.Exec(ctx, &iri.ExecRequest{MachineId: "my-machine"})
		Expect(err).NotTo(HaveOccurred())

		By("inspecting the result")
		u, err := url.ParseRequestURI(res.Url)
		Expect(err).NotTo(HaveOccurred(), "url is invalid: %q", res.Url)
		Expect(u.Host).To(Equal("localhost:8080"))
		Expect(u.Scheme).To(Equal("http"))
		Expect(u.Path).To(MatchRegexp(`/exec/[^/?&]{8}`))
	})
})
