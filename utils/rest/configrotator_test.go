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

package rest_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"time"

	"github.com/go-logr/logr"
	"github.com/onmetal/onmetal-api/utils/certificate"
	certificatetesting "github.com/onmetal/onmetal-api/utils/certificate/testing"
	. "github.com/onmetal/onmetal-api/utils/rest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/authorization/v1"
	certificatesv1 "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var _ = Describe("ConfigRotator", func() {
	It("should rotate the certificates", func(ctx SpecContext) {
		By("initializing a bootstrap kubeconfig and a kubeconfig store filepath")
		bootstrapUser, err := testEnv.AddUser(envtest.User{
			Name:   "Bootstrap",
			Groups: []string{"system:authenticated"},
		}, cfg)
		Expect(err).NotTo(HaveOccurred())

		By("creating an authenticated user")
		authenticatedUser, err := testEnv.AddUser(envtest.User{
			Name:   "Authenticated",
			Groups: []string{"system:authenticated", "system:masters"},
		}, cfg)
		Expect(err).NotTo(HaveOccurred())
		authenticatedUserCertData, authenticatedUserKeyData := authenticatedUser.Config().CertData, authenticatedUser.Config().KeyData
		authenticatedUserCert, err := tls.X509KeyPair(authenticatedUserCertData, authenticatedUserKeyData)
		Expect(err).NotTo(HaveOccurred())

		certRotator := certificatetesting.NewFakeRotator()

		By("creating a rotator")
		template := &x509.CertificateRequest{}
		signerName := "rotator-signer.api.onmetal.de"
		requestedDuration := pointer.Duration(1 * time.Hour)
		rotatorName := "rotator"
		r, err := NewConfigRotator(nil, bootstrapUser.Config(), ConfigRotatorOptions{
			Name:              rotatorName,
			SignerName:        signerName,
			Template:          template,
			RequestedDuration: requestedDuration,
			GetUsages: func(privateKey any) []certificatesv1.KeyUsage {
				return []certificatesv1.KeyUsage{certificatesv1.UsageKeyEncipherment}
			},
			LogConstructor: func() logr.Logger {
				return GinkgoLogr
			},
			NewCertificateRotator: func(opts certificate.RotatorOptions) (certificate.Rotator, error) {
				Expect(opts.Name).To(Equal(rotatorName))
				Expect(opts.SignerName).To(Equal(signerName))
				Expect(opts.Template).To(Equal(template))
				Expect(opts.RequestedDuration).To(Equal(requestedDuration))
				Expect(opts.ForceInitial).To(BeFalse())
				Expect(opts.InitCertificate).To(BeNil())
				return certRotator, nil
			},
		})
		Expect(err).NotTo(HaveOccurred())

		By("running the rotator")
		rotatorDone := make(chan struct{})
		rotatorCtx, cancelRotator := context.WithCancel(ctx)
		defer cancelRotator()

		go func() {
			defer GinkgoRecover()
			defer close(rotatorDone)
			Expect(r.Start(rotatorCtx)).To(Succeed())
		}()

		By("waiting for the certificate rotator to be started")
		Eventually(ctx, certRotator.Started).Should(BeTrue(), "cert rotator was not started")

		By("asserting there is no client config available and the rotator marks itself as unhealthy")
		Consistently(ctx, func(g Gomega) {
			g.Expect(r.ClientConfig()).To(BeNil())
			g.Expect(r.Check(nil)).To(HaveOccurred())
		}).Should(Succeed())

		By("creating a client")
		c, err := kubernetes.NewForConfig(r.TransportConfig())
		Expect(err).NotTo(HaveOccurred())

		By("asserting we are not authenticated")
		_, err = c.AuthorizationV1().SelfSubjectRulesReviews().Create(ctx, &v1.SelfSubjectRulesReview{
			Spec: v1.SelfSubjectRulesReviewSpec{
				Namespace: corev1.NamespaceDefault,
			},
		}, metav1.CreateOptions{})
		Expect(err).To(Satisfy(apierrors.IsForbidden))

		By("setting the certificate to the one of the authenticated user")
		certRotator.SetCertificate(&authenticatedUserCert)
		certRotator.EnqueueAll()

		By("waiting for the client config to be available and the rotator to report as healthy")
		var clientConfig *rest.Config
		Eventually(ctx, func(g Gomega) {
			clientConfig = r.ClientConfig()
			g.Expect(clientConfig).NotTo(BeNil())
			g.Expect(clientConfig.CertData).To(Equal(authenticatedUserCertData))
			g.Expect(clientConfig.KeyData).To(Equal(authenticatedUserKeyData))
			g.Expect(r.Check(nil)).NotTo(HaveOccurred())
		}).Should(Succeed())

		By("asserting we are now authenticated")
		_, err = c.AuthorizationV1().SelfSubjectRulesReviews().Create(ctx, &v1.SelfSubjectRulesReview{
			Spec: v1.SelfSubjectRulesReviewSpec{
				Namespace: corev1.NamespaceDefault,
			},
		}, metav1.CreateOptions{})
		Expect(err).NotTo(HaveOccurred())

		By("stopping the rotator")
		cancelRotator()

		By("waiting for the rotator to stop")
		Eventually(rotatorDone).Should(BeClosed())
	})
})
