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
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	ipamv1alpha1 "github.com/onmetal/onmetal-api/api/ipam/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"github.com/onmetal/onmetal-api/ori/testing/machine"
	machinepoolletclient "github.com/onmetal/onmetal-api/poollet/machinepoollet/client"
	"github.com/onmetal/onmetal-api/poollet/machinepoollet/controllers"
	"github.com/onmetal/onmetal-api/poollet/machinepoollet/mcm"
	"github.com/onmetal/onmetal-api/testutils/envtestutils"
	"github.com/onmetal/onmetal-api/testutils/envtestutils/apiserver"
	"github.com/onmetal/onmetal-api/testutils/envtestutils/controllermanager"
	"github.com/onmetal/onmetal-api/testutils/envtestutils/process"
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
	testEnvExt *envtestutils.EnvironmentExtensions
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
	testEnvExt = &envtestutils.EnvironmentExtensions{
		APIServiceDirectoryPaths: []string{
			modutils.Dir("github.com/onmetal/onmetal-api", "config", "apiserver", "apiservice", "bases"),
		},
		ErrorIfAPIServicePathIsMissing: true,
		AdditionalServices: []envtestutils.AdditionalService{
			{
				Name: controllerManagerService,
			},
		},
	}

	cfg, err = envtestutils.StartWithExtensions(testEnv, testEnvExt)
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	DeferCleanup(envtestutils.StopWithExtensions, testEnv, testEnvExt)

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
		MainPath:     "github.com/onmetal/onmetal-api/onmetal-apiserver/cmd/apiserver",
		BuildOptions: []buildutils.BuildOption{buildutils.ModModeMod},
		ETCDServers:  []string{testEnv.ControlPlane.Etcd.URL.String()},
		Host:         testEnvExt.APIServiceInstallOptions.LocalServingHost,
		Port:         testEnvExt.APIServiceInstallOptions.LocalServingPort,
		CertDir:      testEnvExt.APIServiceInstallOptions.LocalServingCertDir,
	})
	Expect(err).NotTo(HaveOccurred())

	Expect(apiSrv.Start()).To(Succeed())
	DeferCleanup(apiSrv.Stop)

	Expect(envtestutils.WaitUntilAPIServicesReadyWithTimeout(apiServiceTimeout, testEnvExt, k8sClient, scheme.Scheme)).To(Succeed())

	ctrlMgr, err := controllermanager.New(cfg, controllermanager.Options{
		Args:         process.EmptyArgs().Set("controllers", "*"),
		MainPath:     "github.com/onmetal/onmetal-api/onmetal-controller-manager",
		BuildOptions: []buildutils.BuildOption{buildutils.ModModeMod},
		Host:         testEnvExt.GetAdditionalServiceHost(controllerManagerService),
		Port:         testEnvExt.GetAdditionalServicePort(controllerManagerService),
	})
	Expect(err).NotTo(HaveOccurred())

	Expect(ctrlMgr.Start()).To(Succeed())
	DeferCleanup(ctrlMgr.Stop)
})

func SetupTest(ctx context.Context) (*corev1.Namespace, *computev1alpha1.MachinePool, *computev1alpha1.MachineClass, *machine.FakeRuntimeService) {
	var (
		ns  = &corev1.Namespace{}
		mp  = &computev1alpha1.MachinePool{}
		mc  = &computev1alpha1.MachineClass{}
		srv = &machine.FakeRuntimeService{}
	)

	BeforeEach(func() {
		*ns = corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-ns-",
			},
		}
		Expect(k8sClient.Create(ctx, ns)).To(Succeed(), "failed to create test namespace")
		DeferCleanup(k8sClient.Delete, ctx, ns)

		*mp = computev1alpha1.MachinePool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-mp-",
			},
		}
		Expect(k8sClient.Create(ctx, mp)).To(Succeed(), "failed to create test machine pool")
		DeferCleanup(k8sClient.Delete, ctx, mp)

		*mc = computev1alpha1.MachineClass{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-mc-",
			},
			Capabilities: map[corev1.ResourceName]resource.Quantity{
				corev1.ResourceCPU:    resource.MustParse("1"),
				corev1.ResourceMemory: resource.MustParse("1Gi"),
			},
		}
		Expect(k8sClient.Create(ctx, mc)).To(Succeed(), "failed to create test machine class")
		DeferCleanup(k8sClient.Delete, ctx, mc)

		*srv = *machine.NewFakeRuntimeService()
		srv.SetMachineClasses([]*machine.FakeMachineClass{
			{
				MachineClass: ori.MachineClass{
					Name: mc.Name,
					Capabilities: &ori.MachineClassCapabilities{
						CpuMillis:   mc.Capabilities.Cpu().MilliValue(),
						MemoryBytes: mc.Capabilities.Memory().AsDec().UnscaledBig().Uint64(),
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

		indexer := k8sManager.GetFieldIndexer()
		Expect(machinepoolletclient.SetupMachineSpecNetworkInterfaceNamesField(ctx, indexer, mp.Name)).To(Succeed())
		Expect(machinepoolletclient.SetupMachineSpecVolumeNamesField(ctx, indexer, mp.Name)).To(Succeed())
		Expect(machinepoolletclient.SetupMachineSpecSecretNamesField(ctx, indexer, mp.Name)).To(Succeed())
		Expect(machinepoolletclient.SetupAliasPrefixRoutingNetworkRefNameField(ctx, indexer)).To(Succeed())
		Expect(machinepoolletclient.SetupLoadBalancerRoutingNetworkRefNameField(ctx, indexer)).To(Succeed())
		Expect(machinepoolletclient.SetupNATGatewayRoutingNetworkRefNameField(ctx, indexer)).To(Succeed())
		Expect(machinepoolletclient.SetupNetworkInterfaceNetworkMachinePoolID(ctx, indexer, mp.Name)).To(Succeed())

		machineClassMapper := mcm.NewGeneric(srv, mcm.GenericOptions{})
		Expect(k8sManager.Add(machineClassMapper)).To(Succeed())

		mgrCtx, cancel := context.WithCancel(ctx)
		DeferCleanup(cancel)

		Expect((&controllers.MachineReconciler{
			EventRecorder:      &record.FakeRecorder{},
			Client:             k8sManager.GetClient(),
			MachineRuntime:     srv,
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
