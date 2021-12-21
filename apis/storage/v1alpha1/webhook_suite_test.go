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

package v1alpha1

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	cfg           *rest.Config
	ctx           = ctrl.SetupSignalHandler()
	k8sClient     client.Client
	webhookScheme *runtime.Scheme
	testEnv       *envtest.Environment
)

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: false,
		WebhookInstallOptions: envtest.WebhookInstallOptions{
			Paths: []string{filepath.Join("..", "..", "..", "config", "webhook")},
		},
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	webhookScheme = runtime.NewScheme()
	err = AddToScheme(webhookScheme)
	Expect(err).NotTo(HaveOccurred())

	err = admissionv1beta1.AddToScheme(webhookScheme)
	Expect(err).NotTo(HaveOccurred())

	err = corev1.AddToScheme(webhookScheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: webhookScheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

func SetupTest() *corev1.Namespace {
	var (
		cancel      context.CancelFunc
		ctxForSetup context.Context
	)

	ns := &corev1.Namespace{}

	BeforeEach(func() {
		// start webhook server using Manager
		ctxForSetup, cancel = context.WithCancel(ctx)
		webhookInstallOptions := &testEnv.WebhookInstallOptions
		mgr, err := ctrl.NewManager(cfg, ctrl.Options{
			Scheme:             webhookScheme,
			Host:               webhookInstallOptions.LocalServingHost,
			Port:               webhookInstallOptions.LocalServingPort,
			CertDir:            webhookInstallOptions.LocalServingCertDir,
			LeaderElection:     false,
			MetricsBindAddress: "0",
		})
		Expect(err).NotTo(HaveOccurred())

		err = (&Volume{}).SetupWebhookWithManager(mgr)
		Expect(err).NotTo(HaveOccurred())

		//+kubebuilder:scaffold:webhook

		go func() {
			defer GinkgoRecover()

			err = mgr.Start(ctxForSetup)
			if err != nil {
				Expect(err).NotTo(HaveOccurred())
			}
		}()

		// wait for the webhook server to get ready
		dialer := &net.Dialer{Timeout: time.Second}
		addrPort := fmt.Sprintf("%s:%d", webhookInstallOptions.LocalServingHost, webhookInstallOptions.LocalServingPort)
		Eventually(func() error {
			conn, err := tls.DialWithDialer(dialer, "tcp", addrPort, &tls.Config{InsecureSkipVerify: true})
			if err != nil {
				return err
			}
			conn.Close()
			return nil
		}).Should(Succeed())

		*ns = corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "ns-"},
		}
		Expect(k8sClient.Create(ctxForSetup, ns)).NotTo(HaveOccurred(), "failed to create test namespace")
	})

	AfterEach(func() {
		Expect(k8sClient.Delete(ctxForSetup, ns)).NotTo(HaveOccurred(), "failed to delete test namespace")

		defer cancel()
	})

	return ns
}

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Webhook Suite")
}
