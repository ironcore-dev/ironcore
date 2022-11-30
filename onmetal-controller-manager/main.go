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
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	onmetalapiclient "github.com/onmetal/onmetal-api/onmetal-controller-manager/client"
	"github.com/onmetal/onmetal-api/onmetal-controller-manager/controllers/compute"
	"github.com/onmetal/onmetal-api/onmetal-controller-manager/controllers/ipam"
	networkingcontrollers "github.com/onmetal/onmetal-api/onmetal-controller-manager/controllers/networking"
	"github.com/onmetal/onmetal-api/onmetal-controller-manager/controllers/storage"

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

	prefixController          = "prefix"
	prefixAllocationScheduler = "prefixallocationscheduler"

	networkProtectionController    = "networkprotection"
	networkInterfaceController     = "networkinterface"
	networkInterfaceBindController = "networkinterfacebind"
	virtualIPController            = "virtualip"
	aliasPrefixController          = "aliasprefix"
	loadBalancerController         = "loadbalancer"
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
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
	flag.DurationVar(&networkInterfaceBindTimeout, "network-interface-bind-timeout", 10*time.Second, "Timet to wait until considering a network interface bind to be failed.")

	controllers := switches.New(
		// Compute controllers
		machineClassController, machinePoolController, machineSchedulerController, machineController,

		// Storage controllers
		volumePoolController, volumeClassController, volumeController, volumeScheduler,

		// Networking controllers
		networkProtectionController, networkInterfaceController, networkInterfaceBindController, virtualIPController, aliasPrefixController, loadBalancerController,

		// IPAM controllers
		prefixController, prefixAllocationScheduler,
	)
	flag.Var(controllers, "controllers", fmt.Sprintf("Controllers to enable. All controllers: %v. Disabled-by-default controllers: %v", controllers.All(), controllers.DisabledByDefault()))

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	logger := zap.New(zap.UseFlagOptions(&opts))
	ctrl.SetLogger(logger)

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
		if err = (&compute.MachineClassReconciler{
			Client:    mgr.GetClient(),
			APIReader: mgr.GetAPIReader(),
			Scheme:    mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "MachineClassRef")
			os.Exit(1)
		}
	}

	if controllers.Enabled(machinePoolController) {
		if err = (&compute.MachinePoolReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "MachinePoolRef")
			os.Exit(1)
		}
	}

	if controllers.Enabled(machineSchedulerController) {
		if err := (&compute.MachineScheduler{
			Client:        mgr.GetClient(),
			EventRecorder: mgr.GetEventRecorderFor("machine-scheduler"),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "MachineScheduler")
			os.Exit(1)
		}
	}

	if controllers.Enabled(volumePoolController) {
		if err = (&storage.VolumePoolReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "VolumePoolRef")
			os.Exit(1)
		}
	}

	if controllers.Enabled(volumeClassController) {
		if err = (&storage.VolumeClassReconciler{
			Client:    mgr.GetClient(),
			APIReader: mgr.GetAPIReader(),
			Scheme:    mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "VolumeClass")
			os.Exit(1)
		}
	}

	if controllers.Enabled(volumeController) {
		if err = onmetalapiclient.SetupMachineSpecVolumeNamesFieldIndexer(context.TODO(), mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to index field", "field", onmetalapiclient.MachineSpecVolumeNamesField)
			os.Exit(1)
		}

		if err = (&storage.VolumeReconciler{
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

	if controllers.Enabled(volumeScheduler) {
		if err = (&storage.VolumeScheduler{
			EventRecorder: mgr.GetEventRecorderFor("volume-scheduler"),
			Client:        mgr.GetClient(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "VolumeScheduler")
		}
	}

	if controllers.Enabled(machineController) {
		if err = (&compute.MachineReconciler{
			EventRecorder: mgr.GetEventRecorderFor("machines"),
			Client:        mgr.GetClient(),
			Scheme:        mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "Machine")
			os.Exit(1)
		}
	}

	if controllers.Enabled(machineController) || controllers.Enabled(networkInterfaceBindController) {
		if err = onmetalapiclient.SetupMachineSpecNetworkInterfaceNamesFieldIndexer(context.TODO(), mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to setup field indexer", "field", onmetalapiclient.MachineSpecNetworkInterfaceNamesField)
			os.Exit(1)
		}
	}

	if controllers.Enabled(prefixController) || controllers.Enabled(prefixAllocationScheduler) {
		if err = onmetalapiclient.SetupPrefixSpecIPFamilyFieldIndexer(context.TODO(), mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to setup field indexer", "field", onmetalapiclient.PrefixSpecIPFamilyField)
			os.Exit(1)
		}
		if err = onmetalapiclient.SetupPrefixAllocationSpecIPFamilyFieldIndexer(context.TODO(), mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to setup field indexer", "field", onmetalapiclient.PrefixAllocationSpecIPFamilyField)
			os.Exit(1)
		}
	}

	if controllers.Enabled(networkInterfaceController) || controllers.Enabled(networkProtectionController) || controllers.Enabled(aliasPrefixController) || controllers.Enabled(loadBalancerController) {
		if err = onmetalapiclient.SetupNetworkInterfaceNetworkNameFieldIndexer(context.TODO(), mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to setup field indexer", "field", onmetalapiclient.NetworkInterfaceNetworkNameField)
			os.Exit(1)
		}
	}

	if controllers.Enabled(prefixController) {
		if err = (&ipam.PrefixReconciler{
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
		if err = (&ipam.PrefixAllocationScheduler{
			Client:        mgr.GetClient(),
			Scheme:        mgr.GetScheme(),
			EventRecorder: mgr.GetEventRecorderFor("prefix-allocation-scheduler"),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "PrefixAllocationScheduler")
			os.Exit(1)
		}
	}

	if controllers.Enabled(networkProtectionController) {
		if err = (&networkingcontrollers.NetworkProtectionReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "NetworkProtection")
			os.Exit(1)
		}
	}

	if controllers.Enabled(networkInterfaceController) || controllers.Enabled(virtualIPController) {
		if err = onmetalapiclient.SetupNetworkInterfaceVirtualIPNameFieldIndexer(context.TODO(), mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to setup field indexer", "field", "NetworkInterfaceVirtualIPName")
			os.Exit(1)
		}
	}

	if controllers.Enabled(networkInterfaceController) {
		if err = (&networkingcontrollers.NetworkInterfaceReconciler{
			EventRecorder: mgr.GetEventRecorderFor("networkinterfaces"),
			Client:        mgr.GetClient(),
			Scheme:        mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "NetworkInterface")
			os.Exit(1)
		}
	}

	if controllers.Enabled(networkInterfaceBindController) {
		if err = (&networkingcontrollers.NetworkInterfaceBindReconciler{
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
		if err = (&networkingcontrollers.VirtualIPReconciler{
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

	if controllers.Enabled(aliasPrefixController) || controllers.Enabled(networkProtectionController) {
		if err = onmetalapiclient.SetupAliasPrefixNetworkNameFieldIndexer(context.TODO(), mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to setup field indexer", "field", onmetalapiclient.AliasPrefixNetworkNameField)
			os.Exit(1)
		}
	}

	if controllers.Enabled(aliasPrefixController) {
		if err = (&networkingcontrollers.AliasPrefixReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "AliasPrefix")
			os.Exit(1)
		}
	}

	if controllers.Enabled(loadBalancerController) {
		if err = onmetalapiclient.SetupLoadBalancerNetworkNameFieldIndexer(context.TODO(), mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to setup field indexer", "field", onmetalapiclient.LoadBalancerNetworkNameField)
			os.Exit(1)
		}

		if err = (&networkingcontrollers.LoadBalancerReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "LoadBalancer")
			os.Exit(1)
		}
	}

	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
