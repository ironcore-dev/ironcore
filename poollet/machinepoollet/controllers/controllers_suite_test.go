// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package controllers_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ironcore-dev/controller-utils/buildutils"
	"github.com/ironcore-dev/controller-utils/modutils"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	ipamv1alpha1 "github.com/ironcore-dev/ironcore/api/ipam/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	computeclient "github.com/ironcore-dev/ironcore/internal/client/compute"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	"github.com/ironcore-dev/ironcore/iri/testing/machine"
	"github.com/ironcore-dev/ironcore/poollet/irievent"
	machinepoolletclient "github.com/ironcore-dev/ironcore/poollet/machinepoollet/client"
	"github.com/ironcore-dev/ironcore/poollet/machinepoollet/controllers"
	"github.com/ironcore-dev/ironcore/poollet/machinepoollet/mcm"
	utilsenvtest "github.com/ironcore-dev/ironcore/utils/envtest"
	"github.com/ironcore-dev/ironcore/utils/envtest/apiserver"
	"github.com/ironcore-dev/ironcore/utils/envtest/controllermanager"
	"github.com/ironcore-dev/ironcore/utils/envtest/process"
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

	fooDownwardAPILabel = "custom-downward-api-label"
	fooAnnotation       = "foo"
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

	Expect(computev1alpha1.AddToScheme(scheme.Scheme)).To(Succeed())
	Expect(networkingv1alpha1.AddToScheme(scheme.Scheme)).To(Succeed())
	Expect(ipamv1alpha1.AddToScheme(scheme.Scheme)).To(Succeed())
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

func SetupTest() (*corev1.Namespace, *computev1alpha1.MachinePool, *computev1alpha1.MachineClass, *machine.FakeRuntimeService) {
	var (
		ns  = &corev1.Namespace{}
		mp  = &computev1alpha1.MachinePool{}
		mc  = &computev1alpha1.MachineClass{}
		srv = &machine.FakeRuntimeService{}
	)

	BeforeEach(func(ctx SpecContext) {
		*ns = corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-ns-",
			},
		}
		Expect(k8sClient.Create(ctx, ns)).To(Succeed(), "failed to create test namespace")
		DeferCleanup(k8sClient.Delete, ns)

		*mp = computev1alpha1.MachinePool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-mp-",
			},
		}
		Expect(k8sClient.Create(ctx, mp)).To(Succeed(), "failed to create test machine pool")
		DeferCleanup(k8sClient.Delete, mp)

		*mc = computev1alpha1.MachineClass{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-mc-",
			},
			Capabilities: corev1alpha1.ResourceList{
				corev1alpha1.ResourceCPU:    resource.MustParse("1"),
				corev1alpha1.ResourceMemory: resource.MustParse("1Gi"),
			},
		}
		Expect(k8sClient.Create(ctx, mc)).To(Succeed(), "failed to create test machine class")
		DeferCleanup(k8sClient.Delete, mc)

		*srv = *machine.NewFakeRuntimeService()
		srv.SetMachineClasses([]*machine.FakeMachineClassStatus{
			{
				MachineClassStatus: iri.MachineClassStatus{
					MachineClass: &iri.MachineClass{
						Name: mc.Name,
						Capabilities: &iri.MachineClassCapabilities{
							CpuMillis:   mc.Capabilities.CPU().MilliValue(),
							MemoryBytes: mc.Capabilities.Memory().Value(),
						},
					},
				},
			},
		})

		k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
			Scheme: scheme.Scheme,
			Metrics: metricserver.Options{
				BindAddress: "0",
			},
		})
		Expect(err).ToNot(HaveOccurred())

		indexer := k8sManager.GetFieldIndexer()
		Expect(machinepoolletclient.SetupMachineSpecNetworkInterfaceNamesField(ctx, indexer, mp.Name)).To(Succeed())
		Expect(machinepoolletclient.SetupMachineSpecVolumeNamesField(ctx, indexer, mp.Name)).To(Succeed())
		Expect(machinepoolletclient.SetupMachineSpecSecretNamesField(ctx, indexer, mp.Name)).To(Succeed())
		Expect(computeclient.SetupMachineSpecMachinePoolRefNameFieldIndexer(ctx, indexer)).To(Succeed())

		machineClassMapper := mcm.NewGeneric(srv, mcm.GenericOptions{
			RelistPeriod: 2 * time.Second,
		})
		Expect(k8sManager.Add(machineClassMapper)).To(Succeed())

		mgrCtx, cancel := context.WithCancel(context.Background())
		DeferCleanup(cancel)

		Expect((&controllers.MachineReconciler{
			EventRecorder:         &record.FakeRecorder{},
			Client:                k8sManager.GetClient(),
			MachineRuntime:        srv,
			MachineRuntimeName:    machine.FakeRuntimeName,
			MachineRuntimeVersion: machine.FakeVersion,
			MachineClassMapper:    machineClassMapper,
			MachinePoolName:       mp.Name,
			DownwardAPILabels: map[string]string{
				fooDownwardAPILabel: fmt.Sprintf("metadata.annotations['%s']", fooAnnotation),
			},
		}).SetupWithManager(k8sManager)).To(Succeed())

		machineEvents := irievent.NewGenerator(func(ctx context.Context) ([]*iri.Machine, error) {
			res, err := srv.ListMachines(ctx, &iri.ListMachinesRequest{})
			if err != nil {
				return nil, err
			}
			return res.Machines, nil
		}, irievent.GeneratorOptions{})

		Expect(k8sManager.Add(machineEvents)).To(Succeed())

		Expect((&controllers.MachineAnnotatorReconciler{
			Client:        k8sManager.GetClient(),
			MachineEvents: machineEvents,
		}).SetupWithManager(k8sManager)).To(Succeed())

		Expect((&controllers.MachinePoolReconciler{
			Client:             k8sManager.GetClient(),
			MachineRuntime:     srv,
			MachineClassMapper: machineClassMapper,
			MachinePoolName:    mp.Name,
		}).SetupWithManager(k8sManager)).To(Succeed())

		Expect((&controllers.MachinePoolAnnotatorReconciler{
			Client:             k8sManager.GetClient(),
			MachineClassMapper: machineClassMapper,
			MachinePoolName:    mp.Name,
		}).SetupWithManager(k8sManager)).To(Succeed())

		go func() {
			defer GinkgoRecover()
			Expect(k8sManager.Start(mgrCtx)).To(Succeed(), "failed to start manager")
		}()
	})

	return ns, mp, mc, srv
}
