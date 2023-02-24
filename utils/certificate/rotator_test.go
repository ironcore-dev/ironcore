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

package certificate_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"sync/atomic"
	"time"

	. "github.com/onmetal/onmetal-api/utils/certificate"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	certificatesv1 "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Rotator", func() {
	It("should rotate the certificates", func(ctx SpecContext) {
		By("creating a new rotator")
		signerName := "api.onmetal.de/test-signer"
		r, err := NewRotator(RotatorOptions{
			Name: "my-signer",
			NewClient: func(cert *tls.Certificate) (client.WithWatch, error) {
				return client.NewWithWatch(cfg, client.Options{})
			},
			SignerName:        signerName,
			Template:          &x509.CertificateRequest{},
			RequestedDuration: pointer.Duration(3 * time.Hour),
			GetUsages: func(privateKey interface{}) []certificatesv1.KeyUsage {
				return []certificatesv1.KeyUsage{
					certificatesv1.UsageKeyEncipherment,
				}
			},
			LogConstructor: ConstantLogger(GinkgoLogr.WithName("my-signer")),
		})
		Expect(err).NotTo(HaveOccurred())

		var enqueueCt atomic.Int32
		r.AddListener(RotatorListenerFunc(func() {
			enqueueCt.Add(1)
		}))

		By("starting the rotator")
		rotatorCtx, rotatorCancel := context.WithCancel(ctx)
		defer rotatorCancel()
		rotatorDone := make(chan struct{})

		go func() {
			defer GinkgoRecover()
			defer close(rotatorDone)
			Expect(r.Start(rotatorCtx)).To(Succeed())
		}()

		By("waiting for the csr to be there")
		csr := &certificatesv1.CertificateSigningRequest{}
		Eventually(ctx, func(g Gomega) error {
			csrList := &certificatesv1.CertificateSigningRequestList{}
			Expect(k8sClient.List(ctx, csrList)).Should(Succeed())

			for _, csrItem := range csrList.Items {
				if csrItem.Spec.SignerName == signerName {
					*csr = csrItem
					return nil
				}
			}
			return fmt.Errorf("no csr available for signer")
		}).Should(Succeed())

		By("parsing the certificate request")
		block, _ := pem.Decode(csr.Spec.Request)
		Expect(block).NotTo(BeNil())
		Expect(block.Type).To(Equal("CERTIFICATE REQUEST"))
		req, err := x509.ParseCertificateRequest(block.Bytes)
		Expect(err).NotTo(HaveOccurred())

		By("inspecting the certificate request")
		Expect(req.CheckSignature()).To(Succeed())

		By("manually approving the csr")
		now := metav1.Now()
		csr.Status.Conditions = append(csr.Status.Conditions, certificatesv1.CertificateSigningRequestCondition{
			Type:               certificatesv1.CertificateApproved,
			Status:             corev1.ConditionTrue,
			Reason:             "AutoApproved",
			Message:            "Approved for testing",
			LastUpdateTime:     now,
			LastTransitionTime: now,
		})
		Expect(k8sClient.SubResource("approval").Update(ctx, csr)).To(Succeed())

		By("creating a self-signed ca")
		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		Expect(err).NotTo(HaveOccurred())
		caCert, err := certutil.NewSelfSignedCACert(certutil.Config{
			CommonName:   "certificate-test",
			Organization: []string{"certificate-test"},
		}, privateKey)
		Expect(err).NotTo(HaveOccurred())

		By("creating a certificate from the request")
		tmpl := &x509.Certificate{
			SerialNumber: big.NewInt(1),
		}
		der, err := x509.CreateCertificate(rand.Reader, tmpl, caCert, req.PublicKey, privateKey)
		Expect(err).NotTo(HaveOccurred())
		cert, err := x509.ParseCertificate(der)
		Expect(err).NotTo(HaveOccurred())
		certData := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})

		By("updating the csr status to have the certificate")
		csr.Status.Certificate = certData
		Expect(k8sClient.Status().Update(ctx, csr)).To(Succeed())

		By("waiting for the rotator to report the certificate")
		Eventually(ctx, r.Certificate).Should(HaveField("Leaf", Satisfy(cert.Equal)))

		By("inspecting that the listener has been called")
		Expect(enqueueCt.Load()).To(Equal(int32(1)))

		By("stopping the rotator and waiting for it to shut down")
		rotatorCancel()
		Eventually(rotatorDone).Should(BeClosed())
	})
})
