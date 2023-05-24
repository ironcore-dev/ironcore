// Copyright 2022 OnMetal authors
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

package controllers_test

import (
	"context"
	"testing"
	"time"

	"github.com/onmetal/controller-utils/buildutils"
	"github.com/onmetal/controller-utils/modutils"
	corev1alpha1 "github.com/onmetal/onmetal-api/api/core/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/volume/v1alpha1"
	"github.com/onmetal/onmetal-api/ori/testing/volume"
	"github.com/onmetal/onmetal-api/poollet/volumepoollet/controllers"
	"github.com/onmetal/onmetal-api/poollet/volumepoollet/vcm"
	utilsenvtest "github.com/onmetal/onmetal-api/utils/envtest"
	"github.com/onmetal/onmetal-api/utils/envtest/apiserver"
	"github.com/onmetal/onmetal-api/utils/envtest/controllermanager"
	"github.com/onmetal/onmetal-api/utils/envtest/process"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
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
	apiServiceTimeout    = 5 * time.Minute

	controllerManagerService = "controller-manager"
)

func TestControllers(t *testing.T) {
	SetDefaultConsistentlyPollingInterval(pollingInterval)
	SetDefaultEventuallyPollingInterval(pollingInterval)
	SetDefaultEventuallyTimeout(eventuallyTimeout)
	SetDefaultConsistentlyDuration(consistentlyDuration)

	RegisterFailHandler(Fail)
	RunSpecs(t, "Controllers Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	var err error
	By("bootstrapping test environment")
	testEnv = &envtest.Environment{}
	testEnvExt = &utilsenvtest.EnvironmentExtensions{
		APIServiceDirectoryPaths: []string{
			modutils.Dir("github.com/onmetal/onmetal-api", "config", "apiserver", "apiservice", "bases"),
		},
		ErrorIfAPIServicePathIsMissing: true,
		AdditionalServices: []utilsenvtest.AdditionalService{
			{
				Name: controllerManagerService,
			},
		},
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

	ctrlMgr, err := controllermanager.New(cfg, controllermanager.Options{
		Args:         process.EmptyArgs().Set("controllers", "*"),
		MainPath:     "github.com/onmetal/onmetal-api/cmd/onmetal-controller-manager",
		BuildOptions: []buildutils.BuildOption{buildutils.ModModeMod},
		Host:         testEnvExt.GetAdditionalServiceHost(controllerManagerService),
		Port:         testEnvExt.GetAdditionalServicePort(controllerManagerService),
	})
	Expect(err).NotTo(HaveOccurred())

	Expect(ctrlMgr.Start()).To(Succeed())
	DeferCleanup(ctrlMgr.Stop)
})

func SetupTest() (*corev1.Namespace, *storagev1alpha1.VolumePool, *storagev1alpha1.VolumeClass, *volume.FakeRuntimeService) {
	var (
		ns  = &corev1.Namespace{}
		vp  = &storagev1alpha1.VolumePool{}
		vc  = &storagev1alpha1.VolumeClass{}
		srv = &volume.FakeRuntimeService{}
	)

	BeforeEach(func(ctx SpecContext) {
		*ns = corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-ns-",
			},
		}
		Expect(k8sClient.Create(ctx, ns)).To(Succeed(), "failed to create test namespace")
		DeferCleanup(k8sClient.Delete, ns)

		*vp = storagev1alpha1.VolumePool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-vp-",
			},
		}
		Expect(k8sClient.Create(ctx, vp)).To(Succeed(), "failed to create test volume pool")
		DeferCleanup(k8sClient.Delete, vp)

		*vc = storagev1alpha1.VolumeClass{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-vc-",
			},
			Capabilities: corev1alpha1.ResourceList{
				corev1alpha1.ResourceTPS:  resource.MustParse("250Mi"),
				corev1alpha1.ResourceIOPS: resource.MustParse("15000"),
			},
		}
		Expect(k8sClient.Create(ctx, vc)).To(Succeed(), "failed to create test volume class")
		DeferCleanup(k8sClient.Delete, vc)

		*srv = *volume.NewFakeRuntimeService()
		srv.SetVolumeClasses([]*volume.FakeVolumeClass{
			{
				VolumeClass: ori.VolumeClass{
					Name: vc.Name,
					Capabilities: &ori.VolumeClassCapabilities{
						Tps:  262144000,
						Iops: 15000,
					},
				},
			},
		})

		k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
			Scheme:             scheme.Scheme,
			Host:               "127.0.0.1",
			MetricsBindAddress: "0",
		})
		Expect(err).ToNot(HaveOccurred())

		volumeClassMapper := vcm.NewGeneric(srv, vcm.GenericOptions{
			RelistPeriod: 2 * time.Second,
		})
		Expect(k8sManager.Add(volumeClassMapper)).To(Succeed())

		mgrCtx, cancel := context.WithCancel(context.Background())
		DeferCleanup(cancel)

		Expect((&controllers.VolumeReconciler{
			EventRecorder:     &record.FakeRecorder{},
			Client:            k8sManager.GetClient(),
			Scheme:            scheme.Scheme,
			VolumeRuntime:     srv,
			VolumeClassMapper: volumeClassMapper,
			VolumePoolName:    vp.Name,
		}).SetupWithManager(k8sManager)).To(Succeed())

		Expect((&controllers.VolumePoolReconciler{
			Client:            k8sManager.GetClient(),
			VolumeRuntime:     srv,
			VolumeClassMapper: volumeClassMapper,
			VolumePoolName:    TestVolumePool,
		}).SetupWithManager(k8sManager)).To(Succeed())

		Expect((&controllers.VolumePoolAnnotatorReconciler{
			Client:            k8sManager.GetClient(),
			VolumeClassMapper: volumeClassMapper,
			VolumePoolName:    TestVolumePool,
		}).SetupWithManager(k8sManager)).To(Succeed())

		go func() {
			defer GinkgoRecover()
			Expect(k8sManager.Start(mgrCtx)).To(Succeed(), "failed to start manager")
		}()
	})

	return ns, vp, vc, srv
}
