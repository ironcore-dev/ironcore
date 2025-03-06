// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	"context"
	"testing"
	"time"

	"github.com/ironcore-dev/controller-utils/buildutils"
	"github.com/ironcore-dev/controller-utils/modutils"
	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	"github.com/ironcore-dev/ironcore/broker/bucketbroker/server"
	utilsenvtest "github.com/ironcore-dev/ironcore/utils/envtest"
	"github.com/ironcore-dev/ironcore/utils/envtest/apiserver"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	cfg        *rest.Config
	testEnv    *envtest.Environment
	testEnvExt *utilsenvtest.EnvironmentExtensions
	k8sClient  client.Client
)

const (
	eventuallyTimeout    = 3 * time.Second
	pollingInterval      = 50 * time.Millisecond
	consistentlyDuration = 1 * time.Second
	apiServiceTimeout    = 10 * time.Minute
)

func TestServer(t *testing.T) {
	SetDefaultConsistentlyPollingInterval(pollingInterval)
	SetDefaultEventuallyPollingInterval(pollingInterval)
	SetDefaultEventuallyTimeout(eventuallyTimeout)
	SetDefaultConsistentlyDuration(consistentlyDuration)

	RegisterFailHandler(Fail)
	RunSpecs(t, "Server Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	var err error
	By("bootstrapping test environment")
	testEnv = &envtest.Environment{}
	testEnvExt = &utilsenvtest.EnvironmentExtensions{
		APIServiceDirectoryPaths: []string{
			modutils.Dir("github.com/ironcore-dev/ironcore", "config", "apiserver", "apiservice", "bases"),
		},
		ErrorIfAPIServicePathIsMissing: true,
	}

	cfg, err = utilsenvtest.StartWithExtensions(testEnv, testEnvExt)
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	DeferCleanup(utilsenvtest.StopWithExtensions, testEnv, testEnvExt)

	Expect(storagev1alpha1.AddToScheme(scheme.Scheme)).To(Succeed())

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

	Expect(utilsenvtest.WaitUntilAPIServicesReadyWithTimeout(apiServiceTimeout, testEnvExt, k8sClient, scheme.Scheme)).To(Succeed())
})

func SetupTest() (*corev1.Namespace, *storagev1alpha1.BucketPool, *server.Server) {
	var (
		ns         = &corev1.Namespace{}
		srv        = &server.Server{}
		bucketPool = &storagev1alpha1.BucketPool{}
	)

	BeforeEach(func(ctx SpecContext) {
		*ns = corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-ns-",
			},
		}
		Expect(k8sClient.Create(ctx, ns)).To(Succeed(), "failed to create test namespace")
		DeferCleanup(k8sClient.Delete, ns)

		*bucketPool = storagev1alpha1.BucketPool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-bp-",
				Labels:       map[string]string{"pool": "test-pool"},
			},
		}
		Expect(k8sClient.Create(ctx, bucketPool)).To(Succeed())
		DeferCleanup(k8sClient.Delete, bucketPool)

		newSrv, err := server.New(cfg, server.Options{
			Namespace:      ns.Name,
			BucketPoolName: bucketPool.Name,
			BucketPoolSelector: map[string]string{
				"pool": "test-pool",
			},
		})
		Expect(err).NotTo(HaveOccurred())
		*srv = *newSrv
	})

	return ns, bucketPool, srv
}

func SetupBucketClass(tps, iops string) *storagev1alpha1.BucketClass {
	bucketClass := &storagev1alpha1.BucketClass{}

	BeforeEach(func(ctx SpecContext) {
		*bucketClass = storagev1alpha1.BucketClass{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "bucket-class-",
			},
			Capabilities: corev1alpha1.ResourceList{
				corev1alpha1.ResourceTPS:  resource.MustParse(tps),
				corev1alpha1.ResourceIOPS: resource.MustParse(iops),
			},
		}
		Expect(k8sClient.Create(ctx, bucketClass)).To(Succeed())
		DeferCleanup(func(ctx context.Context) error {
			return client.IgnoreNotFound(k8sClient.Delete(ctx, bucketClass))
		})
	})

	return bucketClass
}
