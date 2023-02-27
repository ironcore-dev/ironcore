// Copyright 2023 OnMetal authors
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

package config_test

import (
	"context"
	"crypto/x509"

	"github.com/go-logr/logr"
	utilcertificate "github.com/onmetal/onmetal-api/utils/certificate"
	"github.com/onmetal/onmetal-api/utils/client/config"
	utilrest "github.com/onmetal/onmetal-api/utils/rest"
	utilresttesting "github.com/onmetal/onmetal-api/utils/rest/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/rest"
)

var _ = Describe("Controller", func() {
	It("should persist the config", func(ctx SpecContext) {
		By("creating a fake config rotator")
		restConfigRotator := utilresttesting.NewFakeConfigRotator()

		By("creating a file-based store")
		store := new(config.MemoryStore)

		By("creating a controller")
		bootstrapCfg := &rest.Config{
			Host: "https://k8s.example.org",
		}
		c, err := config.NewController(ctx, store, bootstrapCfg, config.ControllerOptions{
			Name:           "sample-controller",
			SignerName:     "sample-signer",
			Template:       &x509.CertificateRequest{},
			GetUsages:      utilcertificate.DefaultKubeAPIServerClientGetUsages,
			LogConstructor: func() logr.Logger { return GinkgoLogr },
			NewRESTConfigRotator: func(cfg, bootstrapCfg *rest.Config, opts utilrest.ConfigRotatorOptions) (utilrest.ConfigRotator, error) {
				return restConfigRotator, nil
			},
		})
		Expect(err).NotTo(HaveOccurred())

		By("starting the controller")
		ctrlErr := make(chan error, 1)
		ctrlCtx, cancelCtrl := context.WithCancel(ctx)
		defer cancelCtrl()
		go func() {
			ctrlErr <- c.Start(ctrlCtx)
		}()

		By("waiting for the rest config rotator to be started")
		Eventually(ctx, restConfigRotator.Started).Should(BeTrue(), "rest config rotator was not started")

		By("asserting no client config becomes available and stored")
		Consistently(ctx, func(g Gomega) {
			g.Expect(c.ClientConfig()).To(BeNil())
			_, err := store.Get(ctx)
			g.Expect(err).To(MatchError(config.ErrConfigNotFound))
			g.Expect(c.Check(nil)).To(HaveOccurred())
		}).Should(Succeed())

		By("supplying a client config and enqueuing all listeners")
		authCfg := &rest.Config{
			Host:     "https://k8s.example.org",
			Username: "foo",
			Password: "bar",
		}
		restConfigRotator.SetClientConfig(authCfg)
		restConfigRotator.EnqueueAll()

		By("waiting for the client config to be available and stored")
		Eventually(ctx, func(g Gomega) {
			g.Expect(c.ClientConfig()).To(Equal(authCfg))
			g.Expect(store.Get(ctx)).To(Equal(authCfg))
		}).Should(Succeed())

		By("stopping the controller and waiting for it to return")
		cancelCtrl()
		Eventually(ctx, ctrlErr).Should(Receive(BeNil()))
	})
})
