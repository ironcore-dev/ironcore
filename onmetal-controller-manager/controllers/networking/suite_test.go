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

	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	onmetalapiclient "github.com/onmetal/onmetal-api/onmetal-controller-manager/client"
	"github.com/onmetal/onmetal-api/onmetal-controller-manager/controllers/ipam"
	utilsenvtest "github.com/onmetal/onmetal-api/utils/envtest"
	"github.com/onmetal/onmetal-api/utils/envtest/apiserver"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	ipamv1alpha1 "github.com/onmetal/onmetal-api/api/ipam/v1alpha1"
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
	testEnvExt      *utilsenvtest.EnvironmentExtensions
)

func TestAPIs(t *testing.T) {
	SetDefaultConsistentlyPollingInterval(pollingInterval)
	SetDefaultEventuallyPollingInterval(pollingInterval)
	SetDefaultEventuallyTimeout(eventuallyTimeout)
	SetDefaultConsistentlyDuration(consistentlyDuration)
	RegisterFailHandler(Fail)

	RunSpecs(t, "Networking Controller Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	path, err := gexec.Build(filepath.Join("..", "..", "..", "onmetal-apiserver", "cmd", "apiserver"))
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(gexec.CleanupBuildArtifacts)
	return []byte(path)
}, func(path []byte) {
	apiServerBinary = string(path)

	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{}
	testEnvExt = &utilsenvtest.EnvironmentExtensions{
		APIServiceDirectoryPaths:       []string{filepath.Join("..", "..", "..", "config", "apiserver", "apiservice", "bases")},
		ErrorIfAPIServicePathIsMissing: true,
	}

	cfg, err := utilsenvtest.StartWithExtensions(testEnv, testEnvExt)
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	DeferCleanup(utilsenvtest.StopWithExtensions, testEnv, testEnvExt)

	Expect(ipamv1alpha1.AddToScheme(scheme.Scheme)).Should(Succeed())
	Expect(networkingv1alpha1.AddToScheme(scheme.Scheme)).Should(Succeed())
	Expect(computev1alpha1.AddToScheme(scheme.Scheme)).Should(Succeed())

	// Init package-level k8sClient
	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())
	SetClient(k8sClient)

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

	Expect(utilsenvtest.WaitUntilAPIServicesReadyWithTimeout(apiServiceTimeout, testEnvExt, k8sClient, scheme.Scheme)).To(Succeed())

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme.Scheme,
		MetricsBindAddress: "0",
	})
	Expect(err).ToNot(HaveOccurred())

	ctx, cancel := context.WithCancel(context.Background())
	DeferCleanup(cancel)
	Expect(onmetalapiclient.SetupNetworkInterfaceNetworkNameFieldIndexer(ctx, k8sManager.GetFieldIndexer())).To(Succeed())
	Expect(onmetalapiclient.SetupNetworkInterfaceVirtualIPNameFieldIndexer(ctx, k8sManager.GetFieldIndexer())).To(Succeed())
	Expect(onmetalapiclient.SetupMachineSpecNetworkInterfaceNamesFieldIndexer(ctx, k8sManager.GetFieldIndexer())).To(Succeed())
	Expect(onmetalapiclient.SetupPrefixSpecIPFamilyFieldIndexer(ctx, k8sManager.GetFieldIndexer())).To(Succeed())
	Expect(onmetalapiclient.SetupAliasPrefixNetworkNameFieldIndexer(ctx, k8sManager.GetFieldIndexer())).To(Succeed())
	Expect(onmetalapiclient.SetupLoadBalancerNetworkNameFieldIndexer(ctx, k8sManager.GetFieldIndexer())).To(Succeed())
	Expect(onmetalapiclient.SetupNATGatewayNetworkNameFieldIndexer(ctx, k8sManager.GetFieldIndexer())).To(Succeed())

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
		EventRecorder: &record.FakeRecorder{},
		Client:        k8sManager.GetClient(),
		Scheme:        k8sManager.GetScheme(),
	}).SetupWithManager(k8sManager)
	Expect(err).NotTo(HaveOccurred())

	err = (&LoadBalancerReconciler{
		Client: k8sManager.GetClient(),
		Scheme: k8sManager.GetScheme(),
	}).SetupWithManager(k8sManager)
	Expect(err).NotTo(HaveOccurred())

	err = (&NatGatewayReconciler{
		EventRecorder: &record.FakeRecorder{},
		Client:        k8sManager.GetClient(),
		Scheme:        k8sManager.GetScheme(),
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

func SetupTest(ctx context.Context) (*corev1.Namespace, *computev1alpha1.MachineClass) {
	var (
		ns           = &corev1.Namespace{}
		machineClass = &computev1alpha1.MachineClass{}
	)

	BeforeEach(func() {
		*ns = corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "testns-"},
		}
		Expect(k8sClient.Create(ctx, ns)).NotTo(HaveOccurred(), "failed to create test namespace")
		DeferCleanup(k8sClient.Delete, ctx, ns)

		*machineClass = computev1alpha1.MachineClass{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "machine-class-",
			},
			Capabilities: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("1"),
				corev1.ResourceMemory: resource.MustParse("1Gi"),
			},
		}
		Expect(k8sClient.Create(ctx, machineClass)).To(Succeed(), "failed to create test machine class")
		DeferCleanup(k8sClient.Delete, ctx, machineClass)
	})

	return ns, machineClass
}
