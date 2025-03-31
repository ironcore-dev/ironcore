// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package app_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/ironcore-dev/controller-utils/buildutils"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	ipamv1alpha1 "github.com/ironcore-dev/ironcore/api/ipam/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	utilsenvtest "github.com/ironcore-dev/ironcore/utils/envtest"
	"github.com/ironcore-dev/ironcore/utils/envtest/apiserver"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/komega"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	pollingInterval      = 50 * time.Millisecond
	eventuallyTimeout    = 3 * time.Second
	consistentlyDuration = 1 * time.Second
	apiServiceTimeout    = 5 * time.Minute
)

var (
	cfg        *rest.Config
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

	RunSpecs(t, "Compute Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	var err error

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{}
	testEnvExt = &utilsenvtest.EnvironmentExtensions{
		APIServiceDirectoryPaths:       []string{filepath.Join("..", "..", "config", "apiserver", "apiservice", "bases")},
		ErrorIfAPIServicePathIsMissing: true,
	}

	cfg, err = utilsenvtest.StartWithExtensions(testEnv, testEnvExt)
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())
	DeferCleanup(utilsenvtest.StopWithExtensions, testEnv, testEnvExt)

	Expect(corev1alpha1.AddToScheme(scheme.Scheme)).To(Succeed())
	Expect(computev1alpha1.AddToScheme(scheme.Scheme)).To(Succeed())
	Expect(storagev1alpha1.AddToScheme(scheme.Scheme)).To(Succeed())
	Expect(ipamv1alpha1.AddToScheme(scheme.Scheme)).To(Succeed())
	Expect(networkingv1alpha1.AddToScheme(scheme.Scheme)).To(Succeed())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	komega.SetClient(k8sClient)

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

	err = utilsenvtest.WaitUntilAPIServicesReadyWithTimeout(apiServiceTimeout, testEnvExt, cfg, k8sClient, scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
})

func SetupTest(ctx context.Context) *corev1.Namespace {
	var (
		ns = &corev1.Namespace{}
	)

	BeforeEach(func() {
		*ns = corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "testns-",
			},
		}
		Expect(k8sClient.Create(ctx, ns)).To(Succeed(), "failed to create test namespace")
		DeferCleanup(k8sClient.Delete, ctx, ns)
	})

	return ns
}
