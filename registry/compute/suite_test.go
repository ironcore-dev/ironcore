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
package compute

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	networkingv1alpha1 "github.com/onmetal/onmetal-api/apis/networking/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/apis/storage/v1alpha1"
	"github.com/onmetal/onmetal-api/envtestutils"
	"github.com/onmetal/onmetal-api/envtestutils/apiserver"
	onmetalclientset "github.com/onmetal/onmetal-api/generated/clientset/versioned"
	"github.com/onmetal/onmetal-api/internal/testing/apiserverbin"
	"github.com/onmetal/onmetal-api/internal/testing/certs"
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

	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	ipamv1alpha1 "github.com/onmetal/onmetal-api/apis/ipam/v1alpha1"
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
	cfg              *rest.Config
	k8sClient        client.Client
	onmetalClientSet onmetalclientset.Interface
	testEnv          *envtest.Environment
	testEnvExt       *envtestutils.EnvironmentExtensions

	machinePoolletCA             *certs.TinyCA
	machinePoolletCertDir        string
	machinePoolletCAFile         string
	machinePoolletClientCertFile string
	machinePoolletClientKeyFile  string
)

func TestAPIs(t *testing.T) {
	_, reporterConfig := GinkgoConfiguration()
	reporterConfig.SlowSpecThreshold = 10 * time.Second
	SetDefaultConsistentlyPollingInterval(pollingInterval)
	SetDefaultEventuallyPollingInterval(pollingInterval)
	SetDefaultEventuallyTimeout(eventuallyTimeout)
	SetDefaultConsistentlyDuration(consistentlyDuration)

	RegisterFailHandler(Fail)

	RunSpecs(t, "Compute Registry Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	var err error

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{}
	testEnvExt = &envtestutils.EnvironmentExtensions{
		APIServiceDirectoryPaths:       []string{filepath.Join("..", "..", "config", "apiserver", "apiservice", "bases")},
		ErrorIfAPIServicePathIsMissing: true,
	}

	cfg, err = envtestutils.StartWithExtensions(testEnv, testEnvExt)
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())
	DeferCleanup(envtestutils.StopWithExtensions, testEnv, testEnvExt)

	Expect(computev1alpha1.AddToScheme(scheme.Scheme)).To(Succeed())
	Expect(storagev1alpha1.AddToScheme(scheme.Scheme)).To(Succeed())
	Expect(ipamv1alpha1.AddToScheme(scheme.Scheme)).To(Succeed())
	Expect(networkingv1alpha1.AddToScheme(scheme.Scheme)).To(Succeed())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	onmetalClientSet, err = onmetalclientset.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())
	Expect(onmetalClientSet).NotTo(BeNil())

	komega.SetClient(k8sClient)

	By("initializing a ca for machinepoollet testing")
	machinePoolletCA, err = certs.NewTinyCA()
	Expect(err).NotTo(HaveOccurred())

	machinePoolletCertDir, err = os.MkdirTemp("", "machinepoollet-cert")
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(os.RemoveAll, machinePoolletCertDir)

	machinePoolletCAFile = filepath.Join(machinePoolletCertDir, "ca.crt")
	Expect(os.WriteFile(machinePoolletCAFile, machinePoolletCA.CA.CertBytes(), 0640)).To(Succeed())

	machinePoolletClientCert, err := machinePoolletCA.NewClientCert(certs.ClientInfo{
		Name:   "admin",
		Groups: []string{"system:masters"},
	})
	Expect(err).NotTo(HaveOccurred())

	machinePoolletClientCertData, machinePoolletClientKey, err := machinePoolletClientCert.AsBytes()
	Expect(err).NotTo(HaveOccurred())

	machinePoolletClientCertFile = filepath.Join(machinePoolletCertDir, "client.crt")
	machinePoolletClientKeyFile = filepath.Join(machinePoolletCertDir, "client.key")
	Expect(os.WriteFile(machinePoolletClientCertFile, machinePoolletClientCertData, 0640)).To(Succeed())
	Expect(os.WriteFile(machinePoolletClientKeyFile, machinePoolletClientKey, 0640)).To(Succeed())

	machinePoolletServerCert, err := machinePoolletCA.NewServingCert(nil, nil)
	Expect(err).NotTo(HaveOccurred())

	machinePoolletServerCertData, machinePoolletServerKey, err := machinePoolletServerCert.AsBytes()
	Expect(err).NotTo(HaveOccurred())

	Expect(os.WriteFile(filepath.Join(machinePoolletCertDir, "tls.crt"), machinePoolletServerCertData, 0640)).To(Succeed())
	Expect(os.WriteFile(filepath.Join(machinePoolletCertDir, "tls.key"), machinePoolletServerKey, 0640)).To(Succeed())

	apiSrv, err := apiserver.New(cfg, apiserver.Options{
		Command:     []string{apiserverbin.Path},
		ETCDServers: []string{testEnv.ControlPlane.Etcd.URL.String()},
		Host:        testEnvExt.APIServiceInstallOptions.LocalServingHost,
		Port:        testEnvExt.APIServiceInstallOptions.LocalServingPort,
		CertDir:     testEnvExt.APIServiceInstallOptions.LocalServingCertDir,
		Args: apiserver.EmptyProcessArgs().
			Set("machinepoollet-certificate-authority", machinePoolletCAFile).
			Set("machinepoollet-client-certificate", machinePoolletClientCertFile).
			Set("machinepoollet-client-key", machinePoolletClientKeyFile),
	})
	Expect(err).NotTo(HaveOccurred())

	ctx, cancel := context.WithCancel(context.Background())
	DeferCleanup(cancel)
	go func() {
		defer GinkgoRecover()
		err := apiSrv.Start(ctx)
		Expect(err).NotTo(HaveOccurred())
	}()

	err = envtestutils.WaitUntilAPIServicesReadyWithTimeout(apiServiceTimeout, testEnvExt, k8sClient, scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
})

// SetupTest returns a namespace which will be created before each ginkgo `It` block and deleted at the end of `It`
// so that each test case can run in an independent way
func SetupTest(ctx context.Context) *corev1.Namespace {
	ns := &corev1.Namespace{}

	BeforeEach(func() {
		*ns = corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "testns-",
			},
		}
		Expect(k8sClient.Create(ctx, ns)).To(Succeed(), "failed to create test namespace")
	})

	AfterEach(func() {
		Expect(k8sClient.Delete(ctx, ns)).To(Succeed(), "failed to delete test namespace")
		Expect(k8sClient.DeleteAllOf(ctx, &computev1alpha1.MachinePool{})).To(Succeed())
	})

	return ns
}
