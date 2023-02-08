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

	"github.com/onmetal/controller-utils/buildutils"
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	corev1alpha1 "github.com/onmetal/onmetal-api/api/core/v1alpha1"
	onmetalapiclient "github.com/onmetal/onmetal-api/internal/client"
	utilsenvtest "github.com/onmetal/onmetal-api/utils/envtest"
	"github.com/onmetal/onmetal-api/utils/envtest/apiserver"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/envtest/komega"

	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"

	. "github.com/onsi/ginkgo/v2"
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

	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	var err error

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{}
	testEnvExt = &utilsenvtest.EnvironmentExtensions{
		APIServiceDirectoryPaths:       []string{filepath.Join("..", "..", "..", "config", "apiserver", "apiservice", "bases")},
		ErrorIfAPIServicePathIsMissing: true,
	}

	cfg, err = utilsenvtest.StartWithExtensions(testEnv, testEnvExt)
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	DeferCleanup(utilsenvtest.StopWithExtensions, testEnv, testEnvExt)

	Expect(storagev1alpha1.AddToScheme(scheme.Scheme)).To(Succeed())
	Expect(computev1alpha1.AddToScheme(scheme.Scheme)).To(Succeed())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	komega.SetClient(k8sClient)

	apiSrv, err := apiserver.New(cfg, apiserver.Options{
		MainPath:     "github.com/onmetal/onmetal-api/cmd/onmetal-apiserver",
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

func SetupTest(ctx context.Context) (*corev1.Namespace, *computev1alpha1.MachineClass) {
	var (
		cancel       context.CancelFunc
		ns           = &corev1.Namespace{}
		machineClass = &computev1alpha1.MachineClass{}
	)
	BeforeEach(func() {
		var mgrCtx context.Context
		mgrCtx, cancel = context.WithCancel(ctx)
		*ns = corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "testns-",
			},
		}
		Expect(k8sClient.Create(ctx, ns)).To(Succeed(), "failed to create test namespace")

		*machineClass = computev1alpha1.MachineClass{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "machine-class-",
			},
			Capabilities: corev1alpha1.ResourceList{
				corev1alpha1.ResourceCPU:    resource.MustParse("1"),
				corev1alpha1.ResourceMemory: resource.MustParse("1Gi"),
			},
		}
		Expect(k8sClient.Create(ctx, machineClass)).To(Succeed(), "failed to create test machine class")

		k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
			Scheme:             scheme.Scheme,
			Host:               "127.0.0.1",
			MetricsBindAddress: "0",
		})
		Expect(err).ToNot(HaveOccurred())

		// index fields here
		Expect(onmetalapiclient.SetupMachineSpecVolumeNamesFieldIndexer(ctx, k8sManager.GetFieldIndexer())).To(Succeed())

		// register reconciler here
		Expect((&VolumeReconciler{
			EventRecorder: &record.FakeRecorder{},
			Client:        k8sManager.GetClient(),
			APIReader:     k8sManager.GetAPIReader(),
			Scheme:        k8sManager.GetScheme(),
			BindTimeout:   1 * time.Second,
		}).SetupWithManager(k8sManager)).To(Succeed())

		Expect((&VolumeClassReconciler{
			Client:    k8sManager.GetClient(),
			APIReader: k8sManager.GetAPIReader(),
			Scheme:    k8sManager.GetScheme(),
		}).SetupWithManager(k8sManager)).To(Succeed())

		Expect((&VolumeScheduler{
			Client:        k8sManager.GetClient(),
			EventRecorder: &record.FakeRecorder{},
		}).SetupWithManager(k8sManager)).To(Succeed())

		Expect((&BucketClassReconciler{
			Client:    k8sManager.GetClient(),
			APIReader: k8sManager.GetAPIReader(),
			Scheme:    k8sManager.GetScheme(),
		}).SetupWithManager(k8sManager)).To(Succeed())

		Expect((&BucketScheduler{
			Client:        k8sManager.GetClient(),
			EventRecorder: &record.FakeRecorder{},
		}).SetupWithManager(k8sManager)).To(Succeed())

		go func() {
			defer GinkgoRecover()
			Expect(k8sManager.Start(mgrCtx)).To(Succeed(), "failed to start manager")
		}()
	})

	AfterEach(func() {
		cancel()
		Expect(k8sClient.Delete(ctx, ns)).To(Succeed(), "failed to delete test namespace")
		Expect(k8sClient.DeleteAllOf(ctx, &storagev1alpha1.VolumePool{})).To(Succeed())
		Expect(k8sClient.DeleteAllOf(ctx, &storagev1alpha1.BucketPool{})).To(Succeed())
		Expect(k8sClient.DeleteAllOf(ctx, &computev1alpha1.MachineClass{})).To(Succeed())
	})

	return ns, machineClass
}
