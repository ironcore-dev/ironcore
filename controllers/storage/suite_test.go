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

package storage

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/onmetal/onmetal-api/envtestutils"
	"github.com/onmetal/onmetal-api/envtestutils/apiserver"
	"github.com/onmetal/onmetal-api/testutils/apiserverbin"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"

	storagev1alpha1 "github.com/onmetal/onmetal-api/apis/storage/v1alpha1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

const (
	interval = 50 * time.Millisecond
	timeout  = 3 * time.Second
)

var (
	ctx, cancel = context.WithCancel(context.Background())

	cfg        *rest.Config
	k8sClient  client.Client
	testEnv    *envtest.Environment
	testEnvExt *envtestutils.EnvironmentExtensions
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	var err error

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{}
	testEnvExt = &envtestutils.EnvironmentExtensions{
		APIServiceDirectoryPaths:       []string{filepath.Join("..", "..", "config", "apiserver", "apiservice", "bases")},
		ErrorIfAPIServicePathIsMissing: true,
	}

	cfg, err = envtestutils.StartWithExtensions(testEnv, testEnvExt)
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = storagev1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	apiSrv, err := apiserver.New(cfg, apiserver.Options{
		Command:     []string{apiserverbin.Path},
		ETCDServers: []string{testEnv.ControlPlane.Etcd.URL.String()},
		Host:        testEnvExt.APIServiceInstallOptions.LocalServingHost,
		Port:        testEnvExt.APIServiceInstallOptions.LocalServingPort,
		CertDir:     testEnvExt.APIServiceInstallOptions.LocalServingCertDir,
	})
	Expect(err).NotTo(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err := apiSrv.Start(ctx)
		Expect(err).NotTo(HaveOccurred())
	}()

	err = envtestutils.WaitUntilAPIServicesReadyWithTimeout(30*time.Second, testEnvExt, k8sClient, scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
}, 60)

func SetupTest(ctx context.Context) *corev1.Namespace {
	var (
		cancel context.CancelFunc
	)
	ns := &corev1.Namespace{}
	BeforeEach(func() {
		var mgrCtx context.Context
		mgrCtx, cancel = context.WithCancel(ctx)
		*ns = corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "testns-",
			},
		}
		Expect(k8sClient.Create(ctx, ns)).To(Succeed(), "failed to create test namespace")

		k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
			Scheme:             scheme.Scheme,
			Host:               "127.0.0.1",
			MetricsBindAddress: "0",
		})
		Expect(err).ToNot(HaveOccurred())

		// index fields here
		fieldIndexer := NewSharedIndexer(k8sManager)
		Expect(fieldIndexer.IndexField(ctx, &storagev1alpha1.Volume{}, VolumeSpecVolumeClaimNameRefField)).ToNot(HaveOccurred())

		// register reconciler here
		Expect((&VolumeClaimScheduler{
			Client:        k8sManager.GetClient(),
			EventRecorder: &record.FakeRecorder{},
		}).SetupWithManager(k8sManager)).To(Succeed())

		Expect((&VolumeReconciler{
			Client:             k8sManager.GetClient(),
			APIReader:          k8sManager.GetAPIReader(),
			Scheme:             k8sManager.GetScheme(),
			SharedFieldIndexer: fieldIndexer,
			BoundTimeout:       1 * time.Second,
		}).SetupWithManager(k8sManager)).To(Succeed())

		Expect((&VolumeClaimReconciler{
			Client:             k8sManager.GetClient(),
			APIReader:          k8sManager.GetAPIReader(),
			Scheme:             k8sManager.GetScheme(),
			SharedFieldIndexer: fieldIndexer,
		}).SetupWithManager(k8sManager)).To(Succeed())

		Expect((&VolumeScheduler{
			Client: k8sManager.GetClient(),
			Events: &record.FakeRecorder{},
		}).SetupWithManager(k8sManager)).To(Succeed())

		Expect((&VolumeClassReconciler{
			Client:    k8sManager.GetClient(),
			APIReader: k8sManager.GetAPIReader(),
			Scheme:    k8sManager.GetScheme(),
		}).SetupWithManager(k8sManager)).To(Succeed())

		go func() {
			Expect(k8sManager.Start(mgrCtx)).To(Succeed(), "failed to start manager")
		}()
	})

	AfterEach(func() {
		cancel()
		Expect(k8sClient.Delete(ctx, ns)).To(Succeed(), "failed to delete test namespace")
		Expect(k8sClient.DeleteAllOf(ctx, &storagev1alpha1.VolumePool{})).To(Succeed())
	})

	return ns
}

var _ = AfterSuite(func() {
	cancel()

	By("tearing down the test environment")
	err := envtestutils.StopWithExtensions(testEnv, testEnvExt)
	Expect(err).NotTo(HaveOccurred())
})
