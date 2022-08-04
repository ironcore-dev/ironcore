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
package networking

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/apis/networking/v1alpha1"
	"github.com/onmetal/onmetal-api/controllers/ipam"
	"github.com/onmetal/onmetal-api/controllers/shared"
	"github.com/onmetal/onmetal-api/envtestutils"
	"github.com/onmetal/onmetal-api/envtestutils/apiserver"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	ipamv1alpha1 "github.com/onmetal/onmetal-api/apis/ipam/v1alpha1"
	//+kubebuilder:scaffold:imports
)

const (
	pollingInterval      = 50 * time.Millisecond
	eventuallyTimeout    = 3 * time.Second
	consistentlyDuration = 1 * time.Second
	apiServiceTimeout    = 5 * time.Minute
)

var (
	apiServerBinary string
	k8sClient       client.Client
	testEnv         *envtest.Environment
	testEnvExt      *envtestutils.EnvironmentExtensions
)

func TestAPIs(t *testing.T) {
	_, reporterConfig := GinkgoConfiguration()
	reporterConfig.SlowSpecThreshold = 10 * time.Second
	SetDefaultConsistentlyPollingInterval(pollingInterval)
	SetDefaultEventuallyPollingInterval(pollingInterval)
	SetDefaultEventuallyTimeout(eventuallyTimeout)
	SetDefaultConsistentlyDuration(consistentlyDuration)
	RegisterFailHandler(Fail)

	RunSpecs(t, "Networking Controller Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	path, err := gexec.Build(filepath.Join("..", "..", "cmd", "apiserver"))
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(gexec.CleanupBuildArtifacts)
	return []byte(path)
}, func(path []byte) {
	apiServerBinary = string(path)

	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{}
	testEnvExt = &envtestutils.EnvironmentExtensions{
		APIServiceDirectoryPaths:       []string{filepath.Join("..", "..", "config", "apiserver", "apiservice", "bases")},
		ErrorIfAPIServicePathIsMissing: true,
	}

	cfg, err := envtestutils.StartWithExtensions(testEnv, testEnvExt)
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	DeferCleanup(envtestutils.StopWithExtensions, testEnv, testEnvExt)

	Expect(ipamv1alpha1.AddToScheme(scheme.Scheme)).Should(Succeed())
	Expect(networkingv1alpha1.AddToScheme(scheme.Scheme)).Should(Succeed())
	Expect(computev1alpha1.AddToScheme(scheme.Scheme)).Should(Succeed())

	// Init package-level k8sClient
	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	apiSrv, err := apiserver.New(cfg, apiserver.Options{
		Command:     []string{apiServerBinary},
		ETCDServers: []string{testEnv.ControlPlane.Etcd.URL.String()},
		Host:        testEnvExt.APIServiceInstallOptions.LocalServingHost,
		Port:        testEnvExt.APIServiceInstallOptions.LocalServingPort,
		CertDir:     testEnvExt.APIServiceInstallOptions.LocalServingCertDir,
	})
	Expect(err).NotTo(HaveOccurred())

	Expect(apiSrv.Start()).To(Succeed())
	DeferCleanup(apiSrv.Stop)

	Expect(envtestutils.WaitUntilAPIServicesReadyWithTimeout(apiServiceTimeout, testEnvExt, k8sClient, scheme.Scheme)).To(Succeed())

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme.Scheme,
		MetricsBindAddress: "0",
	})
	Expect(err).ToNot(HaveOccurred())

	ctx, cancel := context.WithCancel(context.Background())
	DeferCleanup(cancel)
	Expect(shared.SetupNetworkInterfaceVirtualIPNameFieldIndexer(ctx, k8sManager.GetFieldIndexer())).To(Succeed())
	Expect(shared.SetupMachineSpecNetworkInterfaceNamesFieldIndexer(ctx, k8sManager.GetFieldIndexer())).To(Succeed())

	// Register reconcilers
	err = (&NetworkInterfaceReconciler{
		EventRecorder: &record.FakeRecorder{},
		Client:        k8sManager.GetClient(),
		Scheme:        k8sManager.GetScheme(),
	}).SetupWithManager(k8sManager)
	Expect(err).NotTo(HaveOccurred())

	err = (&NetworkInterfaceBindReconciler{
		EventRecorder: &record.FakeRecorder{},
		Client:        k8sManager.GetClient(),
		APIReader:     k8sManager.GetAPIReader(),
		Scheme:        k8sManager.GetScheme(),
		BindTimeout:   1 * time.Second,
	}).SetupWithManager(k8sManager)
	Expect(err).NotTo(HaveOccurred())

	err = (&VirtualIPReconciler{
		EventRecorder: &record.FakeRecorder{},
		Client:        k8sManager.GetClient(),
		APIReader:     k8sManager.GetAPIReader(),
		Scheme:        k8sManager.GetScheme(),
		BindTimeout:   1 * time.Second,
	}).SetupWithManager(k8sManager)
	Expect(err).NotTo(HaveOccurred())

	err = (&AliasPrefixReconciler{
		Client: k8sManager.GetClient(),
		Scheme: k8sManager.GetScheme(),
	}).SetupWithManager(k8sManager)
	Expect(err).NotTo(HaveOccurred())

	err = (&ipam.PrefixReconciler{
		Client:                  k8sManager.GetClient(),
		APIReader:               k8sManager.GetAPIReader(),
		Scheme:                  k8sManager.GetScheme(),
		PrefixAllocationTimeout: 1 * time.Second,
	}).SetupWithManager(k8sManager)
	Expect(err).NotTo(HaveOccurred())

	err = (&NetworkProtectionReconciler{
		Client: k8sManager.GetClient(),
		Scheme: k8sManager.GetScheme(),
	}).SetupWithManager(k8sManager)
	Expect(err).NotTo(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err = k8sManager.Start(ctx)
		Expect(err).ToNot(HaveOccurred())
	}()
})

func SetupTest(ctx context.Context) *corev1.Namespace {
	ns := &corev1.Namespace{}

	BeforeEach(func() {
		*ns = corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "testns-"},
		}
		Expect(k8sClient.Create(ctx, ns)).NotTo(HaveOccurred(), "failed to create test namespace")

		DeferCleanup(k8sClient.Delete, ctx, ns)
	})

	return ns
}
