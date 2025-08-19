// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	computeclient "github.com/ironcore-dev/ironcore/internal/client/compute"
	ipamclient "github.com/ironcore-dev/ironcore/internal/client/ipam"
	networkingclient "github.com/ironcore-dev/ironcore/internal/client/networking"
	storageclient "github.com/ironcore-dev/ironcore/internal/client/storage"
	computecontrollers "github.com/ironcore-dev/ironcore/internal/controllers/compute"
	computescheduler "github.com/ironcore-dev/ironcore/internal/controllers/compute/scheduler"
	corecontrollers "github.com/ironcore-dev/ironcore/internal/controllers/core"
	certificateironcore "github.com/ironcore-dev/ironcore/internal/controllers/core/certificate/ironcore"
	quotacontrollergeneric "github.com/ironcore-dev/ironcore/internal/controllers/core/quota/generic"
	quotacontrollerironcore "github.com/ironcore-dev/ironcore/internal/controllers/core/quota/ironcore"
	ipamcontrollers "github.com/ironcore-dev/ironcore/internal/controllers/ipam"
	networkingcontrollers "github.com/ironcore-dev/ironcore/internal/controllers/networking"
	storagecontrollers "github.com/ironcore-dev/ironcore/internal/controllers/storage"
	storagescheduler "github.com/ironcore-dev/ironcore/internal/controllers/storage/scheduler"
	quotaevaluatorironcore "github.com/ironcore-dev/ironcore/internal/quota/evaluator/ironcore"
	"github.com/ironcore-dev/ironcore/utils/quota"
	"k8s.io/utils/lru"
	"sigs.k8s.io/controller-runtime/pkg/certwatcher"
	"sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/ironcore-dev/controller-utils/cmdutils/switches"

	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	ipamv1alpha1 "github.com/ironcore-dev/ironcore/api/ipam/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

const (
	// compute controllers
	machineEphemeralNetworkInterfaceController = "machineephemeralnetworkinterface"
	machineEphemeralVolumeController           = "machineephemeralvolume"
	machineSchedulerController                 = "machinescheduler"
	machineClassController                     = "machineclass"

	// storage controllers
	bucketScheduler           = "bucketscheduler"
	bucketClassController     = "bucketclass"
	volumeReleaseController   = "volumerelease"
	volumeSchedulerController = "volumescheduler"
	volumeClassController     = "volumeclass"

	// ipam controllers
	prefixController          = "prefix"
	prefixAllocationScheduler = "prefixallocationscheduler"

	// networking controllers
	loadBalancerController                       = "loadbalancer"
	loadBalancerEphemeralPrefixController        = "loadbalancerephemeralprefix"
	networkProtectionController                  = "networkprotection"
	networkPeeringController                     = "networkpeering"
	networkReleaseController                     = "networkrelease"
	networkInterfaceEphemeralPrefixController    = "networkinterfaceephemeralprefix"
	networkInterfaceEphemeralVirtualIPController = "networkinterfaceephemeralvirtualip"
	networkInterfaceReleaseController            = "networkinterfacerelease"
	virtualIPReleaseController                   = "virtualiprelease"

	// core controllers
	resourceQuotaController       = "resourcequota"
	certificateApprovalController = "certificateapproval"
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(corev1alpha1.AddToScheme(scheme))
	utilruntime.Must(computev1alpha1.AddToScheme(scheme))
	utilruntime.Must(storagev1alpha1.AddToScheme(scheme))
	utilruntime.Must(networkingv1alpha1.AddToScheme(scheme))
	utilruntime.Must(ipamv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var secureMetrics bool
	var metricsCertPath, metricsCertName, metricsCertKey string
	var enableHTTP2 bool
	var enableLeaderElection bool
	var probeAddr string
	var pprofAddr string
	var prefixAllocationTimeout time.Duration
	var volumeBindTimeout time.Duration
	var virtualIPBindTimeout time.Duration
	var networkInterfaceBindTimeout time.Duration
	var tlsOpts []func(*tls.Config)
	flag.StringVar(&metricsAddr, "metrics-bind-address", "0", "The address the metrics endpoint binds to. "+
		"Use :8443 for HTTPS or :8080 for HTTP, or leave as 0 to disable the metrics service.")
	flag.BoolVar(&secureMetrics, "metrics-secure", true,
		"If set, the metrics endpoint is served securely via HTTPS. Use --metrics-secure=false to use HTTP instead.")
	flag.StringVar(&metricsCertPath, "metrics-cert-path", "",
		"The directory that contains the metrics server certificate.")
	flag.StringVar(&metricsCertName, "metrics-cert-name", "tls.crt", "The name of the metrics server certificate file.")
	flag.StringVar(&metricsCertKey, "metrics-cert-key", "tls.key", "The name of the metrics server key file.")
	flag.BoolVar(&enableHTTP2, "enable-http2", false,
		"If set, HTTP/2 will be enabled for the metrics server")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.StringVar(&pprofAddr, "pprof-bind-address", "", "The address the Pprof endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.DurationVar(&prefixAllocationTimeout, "prefix-allocation-timeout", 1*time.Second, "Time to wait until considering a pending allocation failed.")
	flag.DurationVar(&volumeBindTimeout, "volume-bind-timeout", 10*time.Second, "Time to wait until considering a volume bind to be failed.")
	flag.DurationVar(&virtualIPBindTimeout, "virtual-ip-bind-timeout", 10*time.Second, "Time to wait until considering a virtual ip bind to be failed.")
	flag.DurationVar(&networkInterfaceBindTimeout, "network-interface-bind-timeout", 10*time.Second, "Time to wait until considering a network interface bind to be failed.")

	controllers := switches.New(
		// compute controllers
		machineEphemeralNetworkInterfaceController,
		machineEphemeralVolumeController,
		machineSchedulerController,
		machineClassController,

		// storage controllers
		bucketScheduler,
		bucketClassController,
		volumeReleaseController,
		volumeSchedulerController,
		volumeClassController,

		// ipam controllers
		prefixController,
		prefixAllocationScheduler,

		// networking controllers
		loadBalancerController,
		loadBalancerEphemeralPrefixController,
		networkProtectionController,
		networkPeeringController,
		networkReleaseController,
		networkInterfaceEphemeralPrefixController,
		networkInterfaceEphemeralVirtualIPController,
		networkInterfaceReleaseController,
		virtualIPReleaseController,

		// core controllers
		resourceQuotaController,
		certificateApprovalController,
	)
	flag.Var(controllers, "controllers",
		fmt.Sprintf("Controllers to enable. All controllers: %v. Disabled-by-default controllers: %v",
			controllers.All(),
			controllers.DisabledByDefault(),
		),
	)

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	logger := zap.New(zap.UseFlagOptions(&opts))
	ctrl.SetLogger(logger)
	ctx := ctrl.SetupSignalHandler()

	// if the enable-http2 flag is false (the default), http/2 should be disabled
	// due to its vulnerabilities. More specifically, disabling http/2 will
	// prevent from being vulnerable to the HTTP/2 Stream Cancellation and
	// Rapid Reset CVEs. For more information see:
	// - https://github.com/advisories/GHSA-qppj-fm5r-hxr3
	// - https://github.com/advisories/GHSA-4374-p667-p6c8
	disableHTTP2 := func(c *tls.Config) {
		setupLog.Info("disabling http/2")
		c.NextProtos = []string{"http/1.1"}
	}

	if !enableHTTP2 {
		tlsOpts = append(tlsOpts, disableHTTP2)
	}

	// Metrics endpoint is enabled in 'config/controller/default/kustomization.yaml'. The Metrics options configure the server.
	// More info:
	// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.1/pkg/metrics/server
	// - https://book.kubebuilder.io/reference/metrics.html
	metricsServerOptions := metricsserver.Options{
		BindAddress:   metricsAddr,
		SecureServing: secureMetrics,
		TLSOpts:       tlsOpts,
	}

	if secureMetrics {
		// FilterProvider is used to protect the metrics endpoint with authn/authz.
		// These configurations ensure that only authorized users and service accounts
		// can access the metrics endpoint. The RBAC are configured in 'config/controller/rbac/kustomization.yaml'. More info:
		// https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.1/pkg/metrics/filters#WithAuthenticationAndAuthorization
		metricsServerOptions.FilterProvider = filters.WithAuthenticationAndAuthorization
	}

	// If the certificate is not specified, controller-runtime will automatically
	// generate self-signed certificates for the metrics server. While convenient for development and testing,
	// this setup is not recommended for production.
	//
	// TODO(user): If you enable certManager, uncomment the following lines:
	// - [METRICS-WITH-CERTS] at config/controller/default/kustomization.yaml to generate and use certificates
	// managed by cert-manager for the metrics server.
	// - [PROMETHEUS-WITH-CERTS] at config/controller/prometheus/kustomization.yaml for TLS certification.

	// Create watchers for metrics certificates
	var metricsCertWatcher *certwatcher.CertWatcher

	if len(metricsCertPath) > 0 {
		setupLog.Info("Initializing metrics certificate watcher using provided certificates",
			"metrics-cert-path", metricsCertPath, "metrics-cert-name", metricsCertName, "metrics-cert-key", metricsCertKey)

		var err error
		metricsCertWatcher, err = certwatcher.New(
			filepath.Join(metricsCertPath, metricsCertName),
			filepath.Join(metricsCertPath, metricsCertKey),
		)
		if err != nil {
			setupLog.Error(err, "to initialize metrics certificate watcher", "error", err)
			os.Exit(1)
		}

		metricsServerOptions.TLSOpts = append(metricsServerOptions.TLSOpts, func(config *tls.Config) {
			config.GetCertificate = metricsCertWatcher.GetCertificate
		})
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Logger:                 logger,
		Scheme:                 scheme,
		Metrics:                metricsServerOptions,
		HealthProbeBindAddress: probeAddr,
		PprofBindAddress:       pprofAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "d0ae00be.ironcore.dev",
	})
	if err != nil {
		setupLog.Error(err, "unable to create manager")
		os.Exit(1)
	}

	if metricsCertWatcher != nil {
		setupLog.Info("Adding metrics certificate watcher to manager")
		if err := mgr.Add(metricsCertWatcher); err != nil {
			setupLog.Error(err, "unable to add metrics certificate watcher to manager")
			os.Exit(1)
		}
	}

	// Register controllers

	// compute controllers

	if controllers.Enabled(machineEphemeralNetworkInterfaceController) {
		if err := (&computecontrollers.MachineEphemeralNetworkInterfaceReconciler{
			Client: mgr.GetClient(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "MachineEphemeralNetworkInterface")
			os.Exit(1)
		}
	}

	if controllers.Enabled(machineEphemeralVolumeController) {
		if err := (&computecontrollers.MachineEphemeralVolumeReconciler{
			EventRecorder: mgr.GetEventRecorderFor("machine-ephemeral-volume"),
			Client:        mgr.GetClient(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "MachineEphemeralVolume")
			os.Exit(1)
		}
	}

	if controllers.Enabled(machineSchedulerController) {
		schedulerCache := computescheduler.NewCache(mgr.GetLogger(), computescheduler.DefaultCacheStrategy)
		if err := mgr.Add(schedulerCache); err != nil {
			setupLog.Error(err, "unable to create cache", "controller", "MachineSchedulerCache")
			os.Exit(1)
		}

		if err := (&computecontrollers.MachineScheduler{
			Client:        mgr.GetClient(),
			EventRecorder: mgr.GetEventRecorderFor("machine-scheduler"),
			Cache:         schedulerCache,
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "MachineScheduler")
			os.Exit(1)
		}
	}

	if controllers.Enabled(machineClassController) {
		if err := (&computecontrollers.MachineClassReconciler{
			Client:    mgr.GetClient(),
			APIReader: mgr.GetAPIReader(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "MachineClass")
			os.Exit(1)
		}
	}

	// storage controllers

	if controllers.Enabled(bucketScheduler) {
		if err := (&storagecontrollers.BucketScheduler{
			EventRecorder: mgr.GetEventRecorderFor("bucket-scheduler"),
			Client:        mgr.GetClient(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "BucketScheduler")
			os.Exit(1)
		}
	}

	if controllers.Enabled(bucketClassController) {
		if err := (&storagecontrollers.BucketClassReconciler{
			Client:    mgr.GetClient(),
			APIReader: mgr.GetAPIReader(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "BucketClass")
			os.Exit(1)
		}
	}

	if controllers.Enabled(volumeReleaseController) {
		if err := (&storagecontrollers.VolumeReleaseReconciler{
			Client:       mgr.GetClient(),
			APIReader:    mgr.GetAPIReader(),
			AbsenceCache: lru.New(500),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "VolumeRelease")
			os.Exit(1)
		}
	}

	if controllers.Enabled(volumeSchedulerController) {
		schedulerCache := storagescheduler.NewCache(mgr.GetLogger(), storagescheduler.DefaultCacheStrategy)
		if err := mgr.Add(schedulerCache); err != nil {
			setupLog.Error(err, "unable to create cache", "controller", "VolumeSchedulerCache")
			os.Exit(1)
		}

		if err := (&storagecontrollers.VolumeScheduler{
			EventRecorder: mgr.GetEventRecorderFor("volume-scheduler"),
			Client:        mgr.GetClient(),
			Cache:         schedulerCache,
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "VolumeScheduler")
			os.Exit(1)
		}
	}

	if controllers.Enabled(volumeClassController) {
		if err := (&storagecontrollers.VolumeClassReconciler{
			Client:    mgr.GetClient(),
			APIReader: mgr.GetAPIReader(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "VolumeClass")
			os.Exit(1)
		}
	}

	// ipam controllers

	if controllers.Enabled(prefixController) {
		if err := (&ipamcontrollers.PrefixReconciler{
			Client:                  mgr.GetClient(),
			APIReader:               mgr.GetAPIReader(),
			Scheme:                  mgr.GetScheme(),
			PrefixAllocationTimeout: prefixAllocationTimeout,
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "Prefix")
			os.Exit(1)
		}
	}

	if controllers.Enabled(prefixAllocationScheduler) {
		if err := (&ipamcontrollers.PrefixAllocationScheduler{
			Client:        mgr.GetClient(),
			Scheme:        mgr.GetScheme(),
			EventRecorder: mgr.GetEventRecorderFor("prefix-allocation-scheduler"),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "PrefixAllocationScheduler")
			os.Exit(1)
		}
	}

	// networking controllers

	if controllers.Enabled(loadBalancerController) {
		if err := (&networkingcontrollers.LoadBalancerReconciler{
			Client: mgr.GetClient(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "LoadBalancer")
			os.Exit(1)
		}
	}

	if controllers.Enabled(loadBalancerEphemeralPrefixController) {
		if err := (&networkingcontrollers.LoadBalancerEphemeralPrefixReconciler{
			Client: mgr.GetClient(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "LoadBalancerEphemeralPrefix")
			os.Exit(1)
		}
	}

	if controllers.Enabled(networkProtectionController) {
		if err := (&networkingcontrollers.NetworkProtectionReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "NetworkProtection")
			os.Exit(1)
		}
	}

	if controllers.Enabled(networkPeeringController) {
		if err := (&networkingcontrollers.NetworkPeeringReconciler{
			Client: mgr.GetClient(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "NetworkPeering")
			os.Exit(1)
		}
	}

	if controllers.Enabled(networkReleaseController) {
		if err := (&networkingcontrollers.NetworkReleaseReconciler{
			Client:       mgr.GetClient(),
			APIReader:    mgr.GetAPIReader(),
			AbsenceCache: lru.New(500),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "NetworkRelease")
			os.Exit(1)
		}
	}

	if controllers.Enabled(networkInterfaceEphemeralPrefixController) {
		if err := (&networkingcontrollers.NetworkInterfaceEphemeralPrefixReconciler{
			Client: mgr.GetClient(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "NetworkInterfaceEphemeralPrefix")
			os.Exit(1)
		}
	}

	if controllers.Enabled(networkInterfaceEphemeralVirtualIPController) {
		if err := (&networkingcontrollers.NetworkInterfaceEphemeralVirtualIPReconciler{
			Client: mgr.GetClient(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "NetworkInterfaceEphemeralVirtualIP")
			os.Exit(1)
		}
	}

	if controllers.Enabled(networkInterfaceReleaseController) {
		if err := (&networkingcontrollers.NetworkInterfaceReleaseReconciler{
			Client:       mgr.GetClient(),
			APIReader:    mgr.GetAPIReader(),
			AbsenceCache: lru.New(500),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "NetworkInterfaceRelease")
			os.Exit(1)
		}
	}

	if controllers.Enabled(virtualIPReleaseController) {
		if err := (&networkingcontrollers.VirtualIPReleaseReconciler{
			Client:       mgr.GetClient(),
			APIReader:    mgr.GetAPIReader(),
			AbsenceCache: lru.New(500),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "VirtualIPRelease")
			os.Exit(1)
		}
	}

	// core controllers

	if controllers.Enabled(resourceQuotaController) {
		registry := quota.NewRegistry(mgr.GetScheme())
		if err := quota.AddAllToRegistry(registry, quotaevaluatorironcore.NewEvaluatorsForControllers(mgr.GetClient())); err != nil {
			setupLog.Error(err, "unable to add evaluators to registry")
			os.Exit(1)
		}

		if err := (&corecontrollers.ResourceQuotaReconciler{
			Client:    mgr.GetClient(),
			APIReader: mgr.GetAPIReader(),
			Scheme:    mgr.GetScheme(),
			Registry:  registry,
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "ResourceQuota")
			os.Exit(1)
		}

		replenishReconcilers, err := quotacontrollerironcore.NewReplenishReconcilers(mgr.GetClient(), registry)
		if err != nil {
			setupLog.Error(err, "unable to create quota replenish controllers")
			os.Exit(1)
		}

		if err := quotacontrollergeneric.SetupReplenishReconcilersWithManager(mgr, replenishReconcilers); err != nil {
			setupLog.Error(err, "unable to create replenish controllers")
			os.Exit(1)
		}
	}

	if controllers.Enabled(certificateApprovalController) {
		if err := (&corecontrollers.CertificateApprovalReconciler{
			Client:      mgr.GetClient(),
			Recognizers: certificateironcore.Recognizers,
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "CertificateApproval")
			os.Exit(1)
		}
	}

	// compute indexers

	if controllers.AnyEnabled(machineEphemeralNetworkInterfaceController) {
		if err := computeclient.SetupMachineSpecNetworkInterfaceNamesFieldIndexer(ctx, mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to index field", "field", computeclient.MachineSpecNetworkInterfaceNamesField)
			os.Exit(1)
		}
	}

	if controllers.AnyEnabled(machineEphemeralVolumeController) {
		if err := computeclient.SetupMachineSpecVolumeNamesFieldIndexer(ctx, mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to index field", "field", computeclient.MachineSpecVolumeNamesField)
			os.Exit(1)
		}
	}

	if controllers.AnyEnabled(machineSchedulerController) {
		if err := computeclient.SetupMachineSpecMachinePoolRefNameFieldIndexer(ctx, mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to index field", "field", computeclient.MachineSpecMachinePoolRefNameField)
			os.Exit(1)
		}
	}

	if controllers.AnyEnabled(machineClassController) {
		if err := computeclient.SetupMachineSpecMachineClassRefNameFieldIndexer(ctx, mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to setup field indexer", "field", computeclient.MachineSpecMachineClassRefNameField)
			os.Exit(1)
		}
	}

	if controllers.AnyEnabled(machineSchedulerController) {
		if err := computeclient.SetupMachinePoolAvailableMachineClassesFieldIndexer(ctx, mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to index field", "field", computeclient.MachinePoolAvailableMachineClassesField)
			os.Exit(1)
		}
	}

	// ipam indexers

	if controllers.AnyEnabled(prefixController, prefixAllocationScheduler) {
		if err := ipamclient.SetupPrefixSpecIPFamilyFieldIndexer(ctx, mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to setup field indexer", "field", ipamclient.PrefixSpecIPFamilyField)
			os.Exit(1)
		}
	}

	if controllers.AnyEnabled(prefixController) {
		if err := ipamclient.SetupPrefixSpecParentRefFieldIndexer(ctx, mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to setup field indexer", "field", ipamclient.PrefixSpecParentRefNameField)
			os.Exit(1)
		}
	}

	if controllers.AnyEnabled(prefixAllocationScheduler) {
		if err := ipamclient.SetupPrefixAllocationSpecIPFamilyFieldIndexer(ctx, mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to setup field indexer", "field", ipamclient.PrefixAllocationSpecIPFamilyField)
			os.Exit(1)
		}
	}

	if controllers.AnyEnabled(prefixController) {
		if err := ipamclient.SetupPrefixAllocationSpecPrefixRefNameField(ctx, mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to setup field indexer", "field", ipamclient.PrefixAllocationSpecPrefixRefNameField)
			os.Exit(1)
		}
	}

	// networking indexers

	if controllers.AnyEnabled(loadBalancerController, networkProtectionController) {
		if err := networkingclient.SetupLoadBalancerNetworkNameFieldIndexer(ctx, mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to setup field indexer", "field", networkingclient.LoadBalancerNetworkNameField)
			os.Exit(1)
		}
	}

	if controllers.AnyEnabled(loadBalancerEphemeralPrefixController) {
		if err := networkingclient.SetupLoadBalancerPrefixNamesFieldIndexer(ctx, mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to setup field indexer", "field", networkingclient.LoadBalancerPrefixNamesField)
			os.Exit(1)
		}
	}

	if controllers.AnyEnabled(networkProtectionController) {
		if err := networkingclient.SetupNATGatewayNetworkNameFieldIndexer(ctx, mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to setup field indexer", "field", networkingclient.NATGatewayNetworkNameField)
			os.Exit(1)
		}
	}

	if controllers.AnyEnabled(networkProtectionController) {
		if err := networkingclient.SetupNetworkSpecPeeringClaimRefNamesFieldIndexer(ctx, mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to setup field indexer", "field", networkingclient.NetworkSpecPeeringClaimRefNamesField)
			os.Exit(1)
		}
	}

	if controllers.AnyEnabled(networkInterfaceEphemeralPrefixController) {
		if err := networkingclient.SetupNetworkInterfacePrefixNamesFieldIndexer(ctx, mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to setup field indexer", "field", networkingclient.NetworkInterfacePrefixNamesField)
			os.Exit(1)
		}
	}

	if controllers.AnyEnabled(networkInterfaceEphemeralVirtualIPController) {
		if err := networkingclient.SetupNetworkInterfaceVirtualIPNameFieldIndexer(ctx, mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to setup field indexer", "field", networkingclient.NetworkInterfaceVirtualIPNamesField)
			os.Exit(1)
		}
	}

	if controllers.AnyEnabled(loadBalancerController, networkProtectionController, networkInterfaceReleaseController) {
		if err := networkingclient.SetupNetworkInterfaceNetworkNameFieldIndexer(ctx, mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to setup field indexer", "field", networkingclient.NetworkInterfaceSpecNetworkRefNameField)
			os.Exit(1)
		}
	}

	// storage indexers

	if controllers.AnyEnabled(bucketClassController) {
		if err := storageclient.SetupBucketSpecBucketClassRefNameFieldIndexer(ctx, mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to setup field indexer", "field", storageclient.BucketSpecBucketClassRefNameField)
			os.Exit(1)
		}
	}

	if controllers.AnyEnabled(bucketScheduler) {
		if err := storageclient.SetupBucketSpecBucketPoolRefNameFieldIndexer(ctx, mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to setup field indexer", "field", storageclient.BucketSpecBucketPoolRefNameField)
			os.Exit(1)
		}
	}

	if controllers.AnyEnabled(bucketScheduler) {
		if err := storageclient.SetupBucketPoolAvailableBucketClassesFieldIndexer(ctx, mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to setup field indexer", "field", storageclient.BucketPoolAvailableBucketClassesField)
			os.Exit(1)
		}
	}

	if controllers.AnyEnabled(volumeClassController) {
		if err := storageclient.SetupVolumeSpecVolumeClassRefNameFieldIndexer(ctx, mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to setup field indexer", "field", storageclient.VolumeSpecVolumeClassRefNameField)
			os.Exit(1)
		}
	}

	if controllers.AnyEnabled(volumeSchedulerController) {
		if err := storageclient.SetupVolumeSpecVolumePoolRefNameFieldIndexer(ctx, mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to setup field indexer", "field", storageclient.VolumeSpecVolumePoolRefNameField)
			os.Exit(1)
		}
	}

	if controllers.AnyEnabled(volumeSchedulerController) {
		if err := storageclient.SetupVolumePoolAvailableVolumeClassesFieldIndexer(ctx, mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to setup field indexer", "field", storageclient.VolumePoolAvailableVolumeClassesField)
			os.Exit(1)
		}
	}

	// healthz / readyz setup

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
