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

	networkingv1alpha1 "github.com/onmetal/onmetal-api/apis/networking/v1alpha1"
	"github.com/onmetal/onmetal-api/controllers/networking"
	"github.com/onmetal/onmetal-api/controllers/shared"

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

	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	ipamv1alpha1 "github.com/onmetal/onmetal-api/apis/ipam/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/apis/storage/v1alpha1"
	computecontrollers "github.com/onmetal/onmetal-api/controllers/compute"
	ipamcontrollers "github.com/onmetal/onmetal-api/controllers/ipam"
	storagecontrollers "github.com/onmetal/onmetal-api/controllers/storage"
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

	networkInterfaceController     = "networkinterface"
	networkInterfaceBindController = "networkinterfacebind"
	virtualIPController            = "virtualip"
	aliasPrefixController          = "aliasprefix"
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
		networkInterfaceController, networkInterfaceBindController, virtualIPController, aliasPrefixController,

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
		if err = (&computecontrollers.MachineClassReconciler{
			Client:    mgr.GetClient(),
			APIReader: mgr.GetAPIReader(),
			Scheme:    mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "MachineClassRef")
			os.Exit(1)
		}
	}

	if controllers.Enabled(machinePoolController) {
		if err = (&computecontrollers.MachinePoolReconciler{
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
		if err = (&storagecontrollers.VolumePoolReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "VolumePoolRef")
			os.Exit(1)
		}
	}

	if controllers.Enabled(volumeClassController) {
		if err = (&storagecontrollers.VolumeClassReconciler{
			Client:    mgr.GetClient(),
			APIReader: mgr.GetAPIReader(),
			Scheme:    mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "VolumeClass")
			os.Exit(1)
		}
	}

	if controllers.Enabled(volumeController) {
		if err = shared.SetupMachineSpecVolumeNamesFieldIndexer(context.TODO(), mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to index field", "field", shared.MachineSpecVolumeNamesField)
			os.Exit(1)
		}

		if err = (&storagecontrollers.VolumeReconciler{
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
		if err = (&storagecontrollers.VolumeScheduler{
			EventRecorder: mgr.GetEventRecorderFor("volume-scheduler"),
			Client:        mgr.GetClient(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "VolumeScheduler")
		}
	}

	if controllers.Enabled(machineController) {
		if err = (&computecontrollers.MachineReconciler{
			EventRecorder: mgr.GetEventRecorderFor("machines"),
			Client:        mgr.GetClient(),
			Scheme:        mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "Machine")
			os.Exit(1)
		}
	}

	if controllers.Enabled(machineController) || controllers.Enabled(networkInterfaceBindController) {
		if err = shared.SetupMachineSpecNetworkInterfaceNamesFieldIndexer(context.TODO(), mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to setup field indexer", "field", shared.MachineSpecNetworkInterfaceNamesField)
			os.Exit(1)
		}
	}

	if controllers.Enabled(prefixController) || controllers.Enabled(prefixAllocationScheduler) {
		if err = ipamcontrollers.SetupPrefixSpecIPFamilyFieldIndexer(context.TODO(), mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to setup field indexer", "field", ipamcontrollers.PrefixSpecIPFamilyField)
			os.Exit(1)
		}
	}

	if controllers.Enabled(prefixController) {
		if err = (&ipamcontrollers.PrefixReconciler{
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
		if err = (&ipamcontrollers.PrefixAllocationScheduler{
			Client:        mgr.GetClient(),
			Scheme:        mgr.GetScheme(),
			EventRecorder: mgr.GetEventRecorderFor("prefix-allocation-scheduler"),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "PrefixAllocationScheduler")
			os.Exit(1)
		}
	}

	if controllers.Enabled(networkInterfaceController) || controllers.Enabled(virtualIPController) {
		if err = shared.SetupNetworkInterfaceVirtualIPNameFieldIndexer(context.TODO(), mgr.GetFieldIndexer()); err != nil {
			setupLog.Error(err, "unable to setup field indexer", "field", "NetworkInterfaceVirtualIPName")
			os.Exit(1)
		}
	}

	if controllers.Enabled(networkInterfaceController) {
		if err = (&networking.NetworkInterfaceReconciler{
			EventRecorder: mgr.GetEventRecorderFor("networkinterfaces"),
			Client:        mgr.GetClient(),
			Scheme:        mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "NetworkInterface")
			os.Exit(1)
		}
	}

	if controllers.Enabled(networkInterfaceBindController) {
		if err = (&networking.NetworkInterfaceBindReconciler{
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
		if err = (&networking.VirtualIPReconciler{
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
		if err = (&networking.AliasPrefixReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "AliasPrefix")
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
