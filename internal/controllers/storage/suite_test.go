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
	computeclient "github.com/onmetal/onmetal-api/internal/client/compute"
	storageclient "github.com/onmetal/onmetal-api/internal/client/storage"
	"github.com/onmetal/onmetal-api/internal/controllers/storage/scheduler"
	utilsenvtest "github.com/onmetal/onmetal-api/utils/envtest"
	"github.com/onmetal/onmetal-api/utils/envtest/apiserver"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/lru"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/envtest/komega"
	metricserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"

	. "github.com/onmetal/onmetal-api/utils/testing"
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
	cfg            *rest.Config
	k8sClient      client.Client
	cacheK8sClient client.Client
	testEnv        *envtest.Environment
	testEnvExt     *utilsenvtest.EnvironmentExtensions
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
	ctx, cancel := context.WithCancel(context.Background())
	DeferCleanup(cancel)

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
		Metrics: metricserver.Options{
			BindAddress: "0",
		},
	})
	Expect(err).ToNot(HaveOccurred())

	cacheK8sClient = k8sManager.GetClient()

	// index fields here
	Expect(computeclient.SetupMachineSpecVolumeNamesFieldIndexer(ctx, k8sManager.GetFieldIndexer())).To(Succeed())

	Expect(storageclient.SetupVolumeSpecVolumeClassRefNameFieldIndexer(ctx, k8sManager.GetFieldIndexer())).To(Succeed())
	Expect(storageclient.SetupVolumeSpecVolumePoolRefNameFieldIndexer(ctx, k8sManager.GetFieldIndexer())).To(Succeed())
	Expect(storageclient.SetupVolumePoolAvailableVolumeClassesFieldIndexer(ctx, k8sManager.GetFieldIndexer())).To(Succeed())
	Expect(storageclient.SetupBucketSpecBucketClassRefNameFieldIndexer(ctx, k8sManager.GetFieldIndexer())).To(Succeed())
	Expect(storageclient.SetupBucketPoolAvailableBucketClassesFieldIndexer(ctx, k8sManager.GetFieldIndexer())).To(Succeed())
	Expect(storageclient.SetupBucketSpecBucketPoolRefNameFieldIndexer(ctx, k8sManager.GetFieldIndexer())).To(Succeed())

	schedulerCache := scheduler.NewCache(k8sManager.GetLogger(), scheduler.DefaultCacheStrategy)
	Expect(k8sManager.Add(schedulerCache)).To(Succeed())

	// register reconciler here
	Expect((&VolumeReleaseReconciler{
		Client:       k8sManager.GetClient(),
		APIReader:    k8sManager.GetAPIReader(),
		AbsenceCache: lru.New(100),
	}).SetupWithManager(k8sManager)).To(Succeed())

	Expect((&VolumeClassReconciler{
		Client:    k8sManager.GetClient(),
		APIReader: k8sManager.GetAPIReader(),
	}).SetupWithManager(k8sManager)).To(Succeed())

	Expect((&BucketClassReconciler{
		Client:    k8sManager.GetClient(),
		APIReader: k8sManager.GetAPIReader(),
	}).SetupWithManager(k8sManager)).To(Succeed())

	Expect((&VolumeScheduler{
		Client:        k8sManager.GetClient(),
		EventRecorder: &record.FakeRecorder{},
		Cache:         schedulerCache,
	}).SetupWithManager(k8sManager)).To(Succeed())

	Expect((&BucketClassReconciler{
		Client:    k8sManager.GetClient(),
		APIReader: k8sManager.GetAPIReader(),
	}).SetupWithManager(k8sManager)).To(Succeed())

	Expect((&BucketScheduler{
		Client:        k8sManager.GetClient(),
		EventRecorder: &record.FakeRecorder{},
	}).SetupWithManager(k8sManager)).To(Succeed())

	go func() {
		defer GinkgoRecover()
		Expect(k8sManager.Start(ctx)).To(Succeed(), "failed to start manager")
	}()
})

func SetupVolumeClass() *storagev1alpha1.VolumeClass {
	return SetupObjectStruct[*storagev1alpha1.VolumeClass](&k8sClient, func(volumeClass *storagev1alpha1.VolumeClass) {
		*volumeClass = storagev1alpha1.VolumeClass{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "volume-class-",
			},
			Capabilities: corev1alpha1.ResourceList{
				corev1alpha1.ResourceTPS:  resource.MustParse("250Mi"),
				corev1alpha1.ResourceIOPS: resource.MustParse("15000"),
			},
		}
	})
}
