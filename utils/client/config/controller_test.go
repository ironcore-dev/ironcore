// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"context"
	"crypto/x509"

	"github.com/go-logr/logr"
	utilcertificate "github.com/ironcore-dev/ironcore/utils/certificate"
	"github.com/ironcore-dev/ironcore/utils/client/config"
	utilrest "github.com/ironcore-dev/ironcore/utils/rest"
	utilresttesting "github.com/ironcore-dev/ironcore/utils/rest/testing"
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
