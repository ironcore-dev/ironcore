// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package controllers_test

import (
	"context"
	"testing"
	"time"

	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	storageclient "github.com/ironcore-dev/ironcore/internal/client/storage"
	iri "github.com/ironcore-dev/ironcore/iri/apis/bucket/v1alpha1"
	"github.com/ironcore-dev/ironcore/iri/testing/bucket"
	"github.com/ironcore-dev/ironcore/poollet/bucketpoollet/bcm"
	"github.com/ironcore-dev/ironcore/poollet/bucketpoollet/controllers"
	"github.com/ironcore-dev/ironcore/poollet/irievent"
	utilsenvtest "github.com/ironcore-dev/ironcore/utils/envtest"
	"github.com/ironcore-dev/ironcore/utils/envtest/apiserver"
	"github.com/ironcore-dev/ironcore/utils/envtest/controllermanager"
	"github.com/ironcore-dev/ironcore/utils/envtest/process"

	"github.com/ironcore-dev/controller-utils/buildutils"
	"github.com/ironcore-dev/controller-utils/modutils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
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
			modutils.Dir("github.com/ironcore-dev/ironcore", "config", "apiserver", "apiservice", "bases"),
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

func SetupTest() (*corev1.Namespace, *storagev1alpha1.BucketPool, *storagev1alpha1.BucketClass, *bucket.FakeRuntimeService) {
	var (
		ns  = &corev1.Namespace{}
		bp  = &storagev1alpha1.BucketPool{}
		bc  = &storagev1alpha1.BucketClass{}
		srv = &bucket.FakeRuntimeService{}
	)

	BeforeEach(func(ctx SpecContext) {
		*ns = corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-ns-",
			},
		}
		Expect(k8sClient.Create(ctx, ns)).To(Succeed(), "failed to create test namespace")
		DeferCleanup(k8sClient.Delete, ns)

		*bp = storagev1alpha1.BucketPool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-bp-",
			},
		}
		Expect(k8sClient.Create(ctx, bp)).To(Succeed(), "failed to create test bucket pool")
		DeferCleanup(k8sClient.Delete, bp)

		*bc = storagev1alpha1.BucketClass{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-bc-",
			},
			Capabilities: corev1alpha1.ResourceList{
				corev1alpha1.ResourceTPS:  resource.MustParse("250Mi"),
				corev1alpha1.ResourceIOPS: resource.MustParse("15000"),
			},
		}
		Expect(k8sClient.Create(ctx, bc)).To(Succeed(), "failed to create test bucket class")
		DeferCleanup(k8sClient.Delete, bc)

		*srv = *bucket.NewFakeRuntimeService()
		srv.SetBucketClasses([]*bucket.FakeBucketClass{
			{
				BucketClass: iri.BucketClass{
					Name: bc.Name,
					Capabilities: &iri.BucketClassCapabilities{
						Tps:  262144000,
						Iops: 15000,
					},
				},
			},
		})
		DeferCleanup(srv.SetBucketClasses, []*bucket.FakeBucketClass{})

		k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
			Scheme: scheme.Scheme,
			Metrics: metricserver.Options{
				BindAddress: "0",
			},
		})
		Expect(err).ToNot(HaveOccurred())

		indexer := k8sManager.GetFieldIndexer()
		Expect(storageclient.SetupBucketSpecBucketPoolRefNameFieldIndexer(ctx, indexer)).To(Succeed())

		bucketClassMapper := bcm.NewGeneric(srv, bcm.GenericOptions{
			RelistPeriod: 2 * time.Second,
		})
		Expect(k8sManager.Add(bucketClassMapper)).To(Succeed())

		mgrCtx, cancel := context.WithCancel(context.Background())
		DeferCleanup(cancel)

		Expect((&controllers.BucketReconciler{
			EventRecorder:     &record.FakeRecorder{},
			Client:            k8sManager.GetClient(),
			Scheme:            scheme.Scheme,
			BucketRuntime:     srv,
			BucketClassMapper: bucketClassMapper,
			BucketPoolName:    bp.Name,
		}).SetupWithManager(k8sManager)).To(Succeed())

		bucketEvents := irievent.NewGenerator(func(ctx context.Context) ([]*iri.Bucket, error) {
			res, err := srv.ListBuckets(ctx, &iri.ListBucketsRequest{})
			if err != nil {
				return nil, err
			}
			return res.Buckets, nil
		}, irievent.GeneratorOptions{})

		Expect(k8sManager.Add(bucketEvents)).To(Succeed())

		Expect((&controllers.BucketAnnotatorReconciler{
			Client:       k8sManager.GetClient(),
			BucketEvents: bucketEvents,
		}).SetupWithManager(k8sManager)).To(Succeed())

		Expect((&controllers.BucketPoolReconciler{
			Client:            k8sManager.GetClient(),
			BucketRuntime:     srv,
			BucketClassMapper: bucketClassMapper,
			BucketPoolName:    bp.Name,
		}).SetupWithManager(k8sManager)).To(Succeed())

		go func() {
			defer GinkgoRecover()
			Expect(k8sManager.Start(mgrCtx)).To(Succeed(), "failed to start manager")
		}()
	})

	return ns, bp, bc, srv
}

func expectBucketDeleted(ctx context.Context, bucket *storagev1alpha1.Bucket) {
	Expect(k8sClient.Delete(ctx, bucket)).Should(Succeed())
	Eventually(Get(bucket)).Should(Satisfy(errors.IsNotFound))
}
