// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0
package networking

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/ironcore-dev/controller-utils/buildutils"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	computeclient "github.com/ironcore-dev/ironcore/internal/client/compute"
	ipamclient "github.com/ironcore-dev/ironcore/internal/client/ipam"
	networkingclient "github.com/ironcore-dev/ironcore/internal/client/networking"
	utilsenvtest "github.com/ironcore-dev/ironcore/utils/envtest"
	"github.com/ironcore-dev/ironcore/utils/envtest/apiserver"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/lru"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlconfig "sigs.k8s.io/controller-runtime/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	ipamv1alpha1 "github.com/ironcore-dev/ironcore/api/ipam/v1alpha1"
	//+kubebuilder:scaffold:imports
)

const (
	pollingInterval      = 50 * time.Millisecond
	eventuallyTimeout    = 3 * time.Second
	consistentlyDuration = 1 * time.Second
	apiServiceTimeout    = 5 * time.Minute
)

var (
	k8sClient  client.Client
	testEnv    *envtest.Environment
	testEnvExt *utilsenvtest.EnvironmentExtensions
)

func TestAPIs(t *testing.T) {
	SetDefaultConsistentlyPollingInterval(pollingInterval)
	SetDefaultEventuallyPollingInterval(pollingInterval)
	SetDefaultEventuallyTimeout(eventuallyTimeout)
	SetDefaultConsistentlyDuration(consistentlyDuration)
	RegisterFailHandler(Fail)

	RunSpecs(t, "Networking Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		// The BinaryAssetsDirectory is only required if you want to run the tests directly
		// without call the makefile target test. If not informed it will look for the
		// default path defined in controller-runtime which is /usr/local/kubebuilder/.
		// Note that you must have the required binaries setup under the bin directory to perform
		// the tests directly. When we run make test it will be setup and used automatically.
		BinaryAssetsDirectory: filepath.Join("..", "..", "bin", "k8s",
			fmt.Sprintf("1.33.0-%s-%s", runtime.GOOS, runtime.GOARCH)),
	}
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
		MainPath:     "github.com/ironcore-dev/ironcore/cmd/ironcore-apiserver",
		BuildOptions: []buildutils.BuildOption{buildutils.ModModeMod},
		ETCDServers:  []string{testEnv.ControlPlane.Etcd.URL.String()},
		Host:         testEnvExt.APIServiceInstallOptions.LocalServingHost,
		Port:         testEnvExt.APIServiceInstallOptions.LocalServingPort,
		CertDir:      testEnvExt.APIServiceInstallOptions.LocalServingCertDir,
	})
	Expect(err).NotTo(HaveOccurred())

	Expect(apiSrv.Start()).To(Succeed())
	DeferCleanup(apiSrv.Stop)

	Expect(utilsenvtest.WaitUntilAPIServicesReadyWithTimeout(apiServiceTimeout, testEnvExt, cfg, k8sClient, scheme.Scheme)).To(Succeed())

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
		Metrics: metricserver.Options{
			BindAddress: "0",
		},
		Controller: ctrlconfig.Controller{SkipNameValidation: ptr.To(true)},
	})
	Expect(err).ToNot(HaveOccurred())

	ctx, cancel := context.WithCancel(context.Background())
	DeferCleanup(cancel)

	Expect(computeclient.SetupMachineSpecNetworkInterfaceNamesFieldIndexer(ctx, k8sManager.GetFieldIndexer())).To(Succeed())
	Expect(computeclient.SetupMachineSpecMachineClassRefNameFieldIndexer(ctx, k8sManager.GetFieldIndexer())).To(Succeed())

	Expect(ipamclient.SetupPrefixSpecIPFamilyFieldIndexer(ctx, k8sManager.GetFieldIndexer())).To(Succeed())
	Expect(ipamclient.SetupPrefixSpecParentRefFieldIndexer(ctx, k8sManager.GetFieldIndexer())).To(Succeed())
	Expect(ipamclient.SetupPrefixAllocationSpecIPFamilyFieldIndexer(ctx, k8sManager.GetFieldIndexer())).To(Succeed())
	Expect(ipamclient.SetupPrefixAllocationSpecPrefixRefNameField(ctx, k8sManager.GetFieldIndexer())).To(Succeed())

	Expect(networkingclient.SetupNetworkInterfaceNetworkNameFieldIndexer(ctx, k8sManager.GetFieldIndexer())).To(Succeed())
	Expect(networkingclient.SetupNetworkInterfaceVirtualIPNameFieldIndexer(ctx, k8sManager.GetFieldIndexer())).To(Succeed())
	Expect(networkingclient.SetupLoadBalancerNetworkNameFieldIndexer(ctx, k8sManager.GetFieldIndexer())).To(Succeed())
	Expect(networkingclient.SetupNATGatewayNetworkNameFieldIndexer(ctx, k8sManager.GetFieldIndexer())).To(Succeed())
	Expect(networkingclient.SetupNetworkSpecPeeringClaimRefNamesFieldIndexer(ctx, k8sManager.GetFieldIndexer())).To(Succeed())
	Expect(networkingclient.SetupNetworkInterfacePrefixNamesFieldIndexer(ctx, k8sManager.GetFieldIndexer())).To(Succeed())
	Expect(networkingclient.SetupLoadBalancerPrefixNamesFieldIndexer(ctx, k8sManager.GetFieldIndexer())).To(Succeed())

	// Register reconcilers
	Expect((&VirtualIPReleaseReconciler{
		Client:       k8sManager.GetClient(),
		APIReader:    k8sManager.GetAPIReader(),
		AbsenceCache: lru.New(100),
	}).SetupWithManager(k8sManager)).To(Succeed())

	Expect((&LoadBalancerReconciler{
		Client: k8sManager.GetClient(),
	}).SetupWithManager(k8sManager)).To(Succeed())

	err = (&NetworkProtectionReconciler{
		Client: k8sManager.GetClient(),
		Scheme: k8sManager.GetScheme(),
	}).SetupWithManager(k8sManager)
	Expect(err).NotTo(HaveOccurred())

	err = (&NetworkPeeringReconciler{
		Client: k8sManager.GetClient(),
	}).SetupWithManager(k8sManager)
	Expect(err).NotTo(HaveOccurred())

	Expect((&NetworkReleaseReconciler{
		Client:       k8sManager.GetClient(),
		APIReader:    k8sManager.GetAPIReader(),
		AbsenceCache: lru.New(100),
	}).SetupWithManager(k8sManager)).To(Succeed())

	Expect((&NetworkInterfaceReleaseReconciler{
		Client:       k8sManager.GetClient(),
		APIReader:    k8sManager.GetAPIReader(),
		AbsenceCache: lru.New(100),
	}).SetupWithManager(k8sManager)).To(Succeed())

	Expect((&LoadBalancerEphemeralPrefixReconciler{
		Client: k8sManager.GetClient(),
	}).SetupWithManager(k8sManager)).To(Succeed())

	Expect((&NetworkInterfaceEphemeralPrefixReconciler{
		Client: k8sManager.GetClient(),
	}).SetupWithManager(k8sManager)).To(Succeed())

	Expect((&NetworkInterfaceEphemeralVirtualIPReconciler{
		Client: k8sManager.GetClient(),
	}).SetupWithManager(k8sManager)).To(Succeed())

	go func() {
		defer GinkgoRecover()
		Expect(k8sManager.Start(ctx)).To(Succeed())
	}()
})
