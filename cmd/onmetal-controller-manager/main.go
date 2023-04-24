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
	corecontrollers "github.com/onmetal/onmetal-api/internal/controllers/core"
	certificateonmetal "github.com/onmetal/onmetal-api/internal/controllers/core/certificate/onmetal"
	quotacontrollergeneric "github.com/onmetal/onmetal-api/internal/controllers/core/quota/generic"
	quotacontrolleronmetal "github.com/onmetal/onmetal-api/internal/controllers/core/quota/onmetal"
	ipamcontrollers "github.com/onmetal/onmetal-api/internal/controllers/ipam"
	networkingcontrollers "github.com/onmetal/onmetal-api/internal/controllers/networking"
	storagecontrollers "github.com/onmetal/onmetal-api/internal/controllers/storage"
	quotaevaluatoronmetal "github.com/onmetal/onmetal-api/internal/quota/evaluator/onmetal"
	"github.com/onmetal/onmetal-api/utils/quota"

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
	machineClassController     = "machineclass"
	machinePoolController      = "machinepool"
	machineSchedulerController = "machinescheduler"
	machineController          = "machine"

	volumePoolController  = "volumepool"
	volumeClassController = "volumeclass"
	volumeController      = "volume"
	volumeScheduler       = "volumescheduler"

	bucketClassController = "bucketclass"
	bucketScheduler       = "bucketscheduler"

	prefixController          = "prefix"
	prefixAllocationScheduler = "prefixallocationscheduler"

	networkBindController          = "networkbind"
	networkProtectionController    = "networkprotection"
	networkInterfaceController     = "networkinterface"
	networkInterfaceBindController = "networkinterfacebind"
	virtualIPController            = "virtualip"
	aliasPrefixController          = "aliasprefix"
	loadBalancerController         = "loadbalancer"
	natGatewayController           = "natgateway"

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
		// Compute controllers
		machineClassController, machinePoolController, machineSchedulerController, machineController,

		// Storage controllers
		volumePoolController, volumeClassController, volumeController, volumeScheduler,
		bucketClassController, bucketScheduler,

		// Networking controllers
		networkBindController, networkProtectionController,
		networkInterfaceController, networkInterfaceBindController, virtualIPController, aliasPrefixController, loadBalancerController, natGatewayController,

		// IPAM controllers
		prefixController, prefixAllocationScheduler,

		// Core controllers
		resourceQuotaController, certificateApprovalController,
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
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "d0ae00be.onmetal.de",
	})
	if err != nil {
		setupLog.Error(err, "unable to create manager")
		os.Exit(1)
	}

	// Register controllers
	if controllers.Enabled(machineClassController) {
		if err := (&computecontrollers.MachineClassReconciler{
			Client:    mgr.GetClient(),
			APIReader: mgr.GetAPIReader(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "MachineClassRef")
			os.Exit(1)
		}
	}

	if controllers.Enabled(machinePoolController) {
		if err := (&computecontrollers.MachinePoolReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "MachinePoolRef")
			os.Exit(1)
		}
	}

	if controllers.Enabled(machineSchedulerController) {
		if err := (&computecontrollers.MachineScheduler{
			Client:        mgr.GetClient(),
			EventRecorder: mgr.GetEventRecorderFor("machine-scheduler"),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "MachineScheduler")
			os.Exit(1)
		}
	}

	if controllers.Enabled(volumePoolController) {
		if err := (&storagecontrollers.VolumePoolReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "VolumePoolRef")
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

	if controllers.Enabled(volumeController) {
		if err := (&storagecontrollers.VolumeReconciler{
			EventRecorder: mgr.GetEventRecorderFor("volumes"),
			Client:        mgr.GetClient(),
			APIReader:     mgr.GetAPIReader(),
			Scheme:        mgr.GetScheme(),
			BindTimeout:   volumeBindTimeout,
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "Volume")
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

	if controllers.Enabled(bucketScheduler) {
		if err := (&storagecontrollers.BucketScheduler{
			EventRecorder: mgr.GetEventRecorderFor("bucket-scheduler"),
			Client:        mgr.GetClient(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "BucketScheduler")
		}
	}

	if controllers.Enabled(volumeScheduler) {
		if err := (&storagecontrollers.VolumeScheduler{
			EventRecorder: mgr.GetEventRecorderFor("volume-scheduler"),
			Client:        mgr.GetClient(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "VolumeScheduler")
		}
	}

	if controllers.Enabled(machineController) {
		if err := (&computecontrollers.MachineReconciler{
			EventRecorder: mgr.GetEventRecorderFor("machines"),
			Client:        mgr.GetClient(),
			Scheme:        mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "Machine")
			os.Exit(1)
		}
	}

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

	if controllers.Enabled(networkBindController) {
		if err := (&networkingcontrollers.NetworkBindReconciler{
			EventRecorder: mgr.GetEventRecorderFor("networkbind"),
			Client:        mgr.GetClient(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "NetworkBind")
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

	if controllers.Enabled(networkInterfaceController) {
		if err := (&networkingcontrollers.NetworkInterfaceReconciler{
			EventRecorder: mgr.GetEventRecorderFor("networkinterfaces"),
			Client:        mgr.GetClient(),
			Scheme:        mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "NetworkInterface")
			os.Exit(1)
		}
	}

	if controllers.Enabled(networkInterfaceBindController) {
		if err := (&networkingcontrollers.NetworkInterfaceBindReconciler{
			EventRecorder: mgr.GetEventRecorderFor("networkinterfaces"),
			Client:        mgr.GetClient(),
			APIReader:     mgr.GetAPIReader(),
			Scheme:        mgr.GetScheme(),
			BindTimeout:   networkInterfaceBindTimeout,
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "NetworkInterfaceBind")
			os.Exit(1)
		}
	}

	if controllers.Enabled(virtualIPController) {
		if err := (&networkingcontrollers.VirtualIPReconciler{
			EventRecorder: mgr.GetEventRecorderFor("virtualips"),
			Client:        mgr.GetClient(),
			APIReader:     mgr.GetAPIReader(),
			Scheme:        mgr.GetScheme(),
			BindTimeout:   virtualIPBindTimeout,
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "VirtualIP")
			os.Exit(1)
		}
	}

	if controllers.Enabled(aliasPrefixController) {
		if err := (&networkingcontrollers.AliasPrefixReconciler{
			EventRecorder: mgr.GetEventRecorderFor("aliasprefixes"),
			Client:        mgr.GetClient(),
			Scheme:        mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "AliasPrefix")
			os.Exit(1)
		}
	}

	if controllers.Enabled(loadBalancerController) {
		if err := (&networkingcontrollers.LoadBalancerReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "LoadBalancer")
			os.Exit(1)
		}
	}

	if controllers.Enabled(natGatewayController) {
		if err := (&networkingcontrollers.NATGatewayReconciler{
			Client:        mgr.GetClient(),
			Scheme:        mgr.GetScheme(),
			EventRecorder: mgr.GetEventRecorderFor("natgateways"),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "NATGateway")
			os.Exit(1)
		}
	}

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

	if controllers.AnyEnabled(machineController, networkInterfaceBindController) {
		if err := computeclient.SetupMachineSpecNetworkInterfaceNamesFieldIndexer(ctx, mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to setup field indexer", "field", computeclient.MachineSpecNetworkInterfaceNamesField)
			os.Exit(1)
		}
	}

	if controllers.AnyEnabled(machineController, volumeController) {
		if err := computeclient.SetupMachineSpecVolumeNamesFieldIndexer(ctx, mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to index field", "field", computeclient.MachineSpecVolumeNamesField)
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

	if controllers.AnyEnabled(aliasPrefixController, networkProtectionController) {
		if err := networkingclient.SetupAliasPrefixNetworkNameFieldIndexer(ctx, mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to setup field indexer", "field", networkingclient.AliasPrefixNetworkNameField)
			os.Exit(1)
		}
	}

	if controllers.AnyEnabled(loadBalancerController, networkProtectionController) {
		if err := networkingclient.SetupLoadBalancerNetworkNameFieldIndexer(ctx, mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to setup field indexer", "field", networkingclient.LoadBalancerNetworkNameField)
			os.Exit(1)
		}
	}

	if controllers.AnyEnabled(natGatewayController, networkProtectionController) {
		if err := networkingclient.SetupNATGatewayNetworkNameFieldIndexer(ctx, mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to setup field indexer", "field", networkingclient.NATGatewayNetworkNameField)
			os.Exit(1)
		}
	}

	if controllers.AnyEnabled(networkBindController) {
		if err := networkingclient.SetupNetworkPeeringKeysFieldIndexer(ctx, mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to setup field indexer", "field", networkingclient.NetworkPeeringKeysField)
			os.Exit(1)
		}
	}

	if controllers.AnyEnabled(networkInterfaceController, virtualIPController) {
		if err := networkingclient.SetupNetworkInterfaceVirtualIPNameFieldIndexer(ctx, mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to setup field indexer", "field", networkingclient.NetworkInterfaceVirtualIPNamesField)
			os.Exit(1)
		}
	}

	if controllers.AnyEnabled(aliasPrefixController, loadBalancerController, natGatewayController, networkProtectionController, networkInterfaceController) {
		if err := networkingclient.SetupNetworkInterfaceNetworkNameFieldIndexer(ctx, mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to setup field indexer", "field", networkingclient.NetworkInterfaceSpecNetworkRefNameField)
			os.Exit(1)
		}
	}

	if controllers.AnyEnabled(networkInterfaceBindController) {
		if err := networkingclient.SetupNetworkInterfaceSpecMachineRefNameField(ctx, mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to setup field indexer", "field", networkingclient.NetworkInterfaceSpecMachineRefNameField)
			os.Exit(1)
		}
	}

	if controllers.AnyEnabled(virtualIPController) {
		if err := networkingclient.SetupVirtualIPSpecTargetRefNameFieldIndexer(ctx, mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to setup field indexer", "field", networkingclient.VirtualIPSpecTargetRefNameField)
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

	if controllers.AnyEnabled(volumeScheduler) {
		if err := storageclient.SetupVolumeSpecVolumePoolRefNameFieldIndexer(ctx, mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to setup field indexer", "field", storageclient.VolumeSpecVolumePoolRefNameField)
			os.Exit(1)
		}
	}

	if controllers.AnyEnabled(volumeScheduler) {
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
