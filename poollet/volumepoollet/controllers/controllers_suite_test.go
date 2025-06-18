// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package controllers_test

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/ironcore-dev/controller-utils/buildutils"
	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	storageclient "github.com/ironcore-dev/ironcore/internal/client/storage"
	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
	"github.com/ironcore-dev/ironcore/iri/testing/volume"
	"github.com/ironcore-dev/ironcore/poollet/volumepoollet/controllers"
	"github.com/ironcore-dev/ironcore/poollet/volumepoollet/vcm"
	utilsenvtest "github.com/ironcore-dev/ironcore/utils/envtest"
	"github.com/ironcore-dev/ironcore/utils/envtest/apiserver"
	"github.com/ironcore-dev/ironcore/utils/envtest/controllermanager"
	"github.com/ironcore-dev/ironcore/utils/envtest/process"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlconfig "sigs.k8s.io/controller-runtime/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
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
	testEnv = &envtest.Environment{
		// The BinaryAssetsDirectory is only required if you want to run the tests directly
		// without call the makefile target test. If not informed it will look for the
		// default path defined in controller-runtime which is /usr/local/kubebuilder/.
		// Note that you must have the required binaries setup under the bin directory to perform
		// the tests directly. When we run make test it will be setup and used automatically.
		BinaryAssetsDirectory: filepath.Join("..", "..", "..", "bin", "k8s",
			fmt.Sprintf("1.33.0-%s-%s", runtime.GOOS, runtime.GOARCH)),
	}
	testEnvExt = &utilsenvtest.EnvironmentExtensions{
		APIServiceDirectoryPaths:       []string{filepath.Join("..", "..", "..", "config", "apiserver", "apiservice", "bases")},
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

	ctrlMgr, err := controllermanager.New(cfg, controllermanager.Options{
		Args:         process.EmptyArgs().Set("controllers", "*"),
		MainPath:     "github.com/ironcore-dev/ironcore/cmd/ironcore-controller-manager",
		BuildOptions: []buildutils.BuildOption{buildutils.ModModeMod},
		Host:         testEnvExt.GetAdditionalServiceHost(controllerManagerService),
		Port:         testEnvExt.GetAdditionalServicePort(controllerManagerService),
	})
	Expect(err).NotTo(HaveOccurred())

	Expect(ctrlMgr.Start()).To(Succeed())
	DeferCleanup(ctrlMgr.Stop)
})

func SetupTest() (*corev1.Namespace, *storagev1alpha1.VolumePool, *storagev1alpha1.VolumeClass, *storagev1alpha1.VolumeClass, *volume.FakeRuntimeService) {
	var (
		ns           = &corev1.Namespace{}
		vp           = &storagev1alpha1.VolumePool{}
		vc           = &storagev1alpha1.VolumeClass{}
		expandableVc = &storagev1alpha1.VolumeClass{}
		srv          = &volume.FakeRuntimeService{}
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

		*expandableVc = storagev1alpha1.VolumeClass{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-vc-expandable-",
			},
			ResizePolicy: storagev1alpha1.ResizePolicyExpandOnly,
			Capabilities: corev1alpha1.ResourceList{
				corev1alpha1.ResourceTPS:  resource.MustParse("250Mi"),
				corev1alpha1.ResourceIOPS: resource.MustParse("1000"),
			},
		}
		Expect(k8sClient.Create(ctx, expandableVc)).To(Succeed(), "failed to create test volume class")
		DeferCleanup(k8sClient.Delete, expandableVc)

		*srv = *volume.NewFakeRuntimeService()
		srv.SetVolumeClasses([]*volume.FakeVolumeClassStatus{
			{
				VolumeClassStatus: iri.VolumeClassStatus{
					VolumeClass: &iri.VolumeClass{
						Name: vc.Name,
						Capabilities: &iri.VolumeClassCapabilities{
							Tps:  262144000,
							Iops: 15000,
						},
					},
				},
			},
			{
				VolumeClassStatus: iri.VolumeClassStatus{
					VolumeClass: &iri.VolumeClass{
						Name: expandableVc.Name,
						Capabilities: &iri.VolumeClassCapabilities{
							Tps:  262144000,
							Iops: 1000,
						},
					},
				},
			},
		})
		DeferCleanup(srv.SetVolumeClasses, []*volume.FakeVolumeClassStatus{})

		k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
			Scheme: scheme.Scheme,
			Metrics: metricserver.Options{
				BindAddress: "0",
			},
			Controller: ctrlconfig.Controller{SkipNameValidation: ptr.To(true)},
		})
		Expect(err).ToNot(HaveOccurred())

		indexer := k8sManager.GetFieldIndexer()
		Expect(storageclient.SetupVolumeSpecVolumePoolRefNameFieldIndexer(ctx, indexer)).To(Succeed())

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
			VolumeRuntimeName: volume.FakeRuntimeName,
			VolumeClassMapper: volumeClassMapper,
			VolumePoolName:    vp.Name,
		}).SetupWithManager(k8sManager)).To(Succeed())

		Expect((&controllers.VolumePoolReconciler{
			Client:            k8sManager.GetClient(),
			VolumeRuntime:     srv,
			VolumeClassMapper: volumeClassMapper,
			VolumePoolName:    vp.Name,
		}).SetupWithManager(k8sManager)).To(Succeed())

		Expect((&controllers.VolumePoolAnnotatorReconciler{
			Client:            k8sManager.GetClient(),
			VolumeClassMapper: volumeClassMapper,
			VolumePoolName:    vp.Name,
		}).SetupWithManager(k8sManager)).To(Succeed())

		go func() {
			defer GinkgoRecover()
			Expect(k8sManager.Start(mgrCtx)).To(Succeed(), "failed to start manager")
		}()
	})

	return ns, vp, vc, expandableVc, srv
}

func expectVolumeDeleted(ctx context.Context, volume *storagev1alpha1.Volume) {
	Expect(k8sClient.Delete(ctx, volume)).Should(Succeed())
	Eventually(Get(volume)).Should(Satisfy(errors.IsNotFound))
}
