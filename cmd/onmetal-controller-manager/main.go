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

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	corev1alpha1 "github.com/onmetal/onmetal-api/api/core/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	computeclient "github.com/onmetal/onmetal-api/internal/client/compute"
	ipamclient "github.com/onmetal/onmetal-api/internal/client/ipam"
	networkingclient "github.com/onmetal/onmetal-api/internal/client/networking"
	storageclient "github.com/onmetal/onmetal-api/internal/client/storage"
	computecontrollers "github.com/onmetal/onmetal-api/internal/controllers/compute"
	computescheduler "github.com/onmetal/onmetal-api/internal/controllers/compute/scheduler"
	corecontrollers "github.com/onmetal/onmetal-api/internal/controllers/core"
	certificateonmetal "github.com/onmetal/onmetal-api/internal/controllers/core/certificate/onmetal"
	quotacontrollergeneric "github.com/onmetal/onmetal-api/internal/controllers/core/quota/generic"
	quotacontrolleronmetal "github.com/onmetal/onmetal-api/internal/controllers/core/quota/onmetal"
	ipamcontrollers "github.com/onmetal/onmetal-api/internal/controllers/ipam"
	networkingcontrollers "github.com/onmetal/onmetal-api/internal/controllers/networking"
	storagecontrollers "github.com/onmetal/onmetal-api/internal/controllers/storage"
	storagescheduler "github.com/onmetal/onmetal-api/internal/controllers/storage/scheduler"
	quotaevaluatoronmetal "github.com/onmetal/onmetal-api/internal/quota/evaluator/onmetal"
	"github.com/onmetal/onmetal-api/utils/quota"
	"k8s.io/utils/lru"
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

	"github.com/onmetal/controller-utils/cmdutils/switches"

	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	ipamv1alpha1 "github.com/onmetal/onmetal-api/api/ipam/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
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
	var enableLeaderElection bool
	var probeAddr string
	var prefixAllocationTimeout time.Duration
	var volumeBindTimeout time.Duration
	var virtualIPBindTimeout time.Duration
	var networkInterfaceBindTimeout time.Duration
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
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

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Logger:                 logger,
		Scheme:                 scheme,
		Metrics:                metricsserver.Options{BindAddress: metricsAddr},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "d0ae00be.onmetal.de",
	})
	if err != nil {
		setupLog.Error(err, "unable to create manager")
		os.Exit(1)
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
			Client: mgr.GetClient(),
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
		if err := quota.AddAllToRegistry(registry, quotaevaluatoronmetal.NewEvaluatorsForControllers(mgr.GetClient())); err != nil {
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

		replenishReconcilers, err := quotacontrolleronmetal.NewReplenishReconcilers(mgr.GetClient(), registry)
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
			Recognizers: certificateonmetal.Recognizers,
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
