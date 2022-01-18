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
	networkv1alpha1 "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/apis/storage/v1alpha1"
	computecontrollers "github.com/onmetal/onmetal-api/controllers/compute"
	networkcontrollers "github.com/onmetal/onmetal-api/controllers/network"
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
	storagePoolController      = "storagepool"
	storageClassController     = "storageclass"
	volumeController           = "volume"
	volumeAttachmentController = "volumeattachment"
	reservedIPController       = "reservedip"
	securityGroupController    = "securitygroup"
	subnetController           = "subnet"
	machineController          = "machine"
	routingDomainController    = "routingdomain"
	ipamRangeController        = "ipamrange"
	gatewayController          = "gateway"

	ipamRangeWebhook = "ipamrange"
	machineWebhook   = "machine"
	volumeWebhook    = "volume"
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(computev1alpha1.AddToScheme(scheme))
	utilruntime.Must(storagev1alpha1.AddToScheme(scheme))
	utilruntime.Must(networkv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var enableWebhooks bool
	flag.BoolVar(&enableWebhooks, "enable-webhooks", true, "Enable webhooks.")
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")

	controllers := switches.New(
		machineClassController, machinePoolController, machineSchedulerController, storagePoolController,
		storageClassController, volumeController, volumeAttachmentController, reservedIPController, securityGroupController,
		subnetController, machineController, routingDomainController, ipamRangeController, gatewayController,
	)
	flag.Var(controllers, "controllers", fmt.Sprintf("Controllers to enable. All controllers: %v. Disabled-by-default controllers: %v", controllers.All(), controllers.DisabledByDefault()))

	webhooks := switches.New(ipamRangeWebhook, machineWebhook, volumeWebhook)
	flag.Var(webhooks, "webhooks", fmt.Sprintf("Webhooks to enable. All webhooks: %v. Disabled-by-default webhooks: %v", webhooks.All(), webhooks.DisabledByDefault()))

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()
	enableWebhooks = enableWebhooks && os.Getenv("ENABLE_WEBHOOKS") != "false"

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

	if controllers.Enabled(machineClassController) {
		if err = (&computecontrollers.MachineClassReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "MachineClass")
			os.Exit(1)
		}
	}
	if controllers.Enabled(machinePoolController) {
		if err = (&computecontrollers.MachinePoolReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "MachinePool")
			os.Exit(1)
		}
	}
	if controllers.Enabled(machineSchedulerController) {
		if err := (&computecontrollers.MachineScheduler{
			Client: mgr.GetClient(),
			Events: mgr.GetEventRecorderFor("machine-scheduler"),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "MachineScheduler")
			os.Exit(1)
		}
	}
	if controllers.Enabled(storagePoolController) {
		if err = (&storagecontrollers.StoragePoolReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "StoragePool")
			os.Exit(1)
		}
	}
	if controllers.Enabled(storageClassController) {
		if err = (&storagecontrollers.StorageClassReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "StorageClass")
			os.Exit(1)
		}
	}
	if controllers.Enabled(volumeController) {
		if err = (&storagecontrollers.VolumeReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "Volume")
			os.Exit(1)
		}
	}
	if controllers.Enabled(volumeAttachmentController) {
		if err = (&storagecontrollers.VolumeAttachmentReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "VolumeAttachment")
			os.Exit(1)
		}
	}
	if controllers.Enabled(reservedIPController) {
		if err = (&networkcontrollers.ReservedIPReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "ReservedIP")
			os.Exit(1)
		}
	}
	if controllers.Enabled(securityGroupController) {
		if err = (&networkcontrollers.SecurityGroupReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "SecurityGroup")
			os.Exit(1)
		}
	}
	if controllers.Enabled(subnetController) {
		if err = (&networkcontrollers.SubnetReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "Subnet")
			os.Exit(1)
		}
	}
	if controllers.Enabled(machineController) {
		if err = (&computecontrollers.MachineReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "Machine")
			os.Exit(1)
		}
	}
	if controllers.Enabled(routingDomainController) {
		if err = (&networkcontrollers.RoutingDomainReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "RoutingDomain")
			os.Exit(1)
		}
	}
	if controllers.Enabled(ipamRangeController) {
		if err = (&networkcontrollers.IPAMRangeReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "IPAMRange")
			os.Exit(1)
		}
	}
	if controllers.Enabled(gatewayController) {
		if err = (&networkcontrollers.GatewayReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "Gateway")
			os.Exit(1)
		}
	}

	// webhook
	if enableWebhooks {
		if webhooks.Enabled(ipamRangeWebhook) {
			if err = (&networkv1alpha1.IPAMRange{}).SetupWebhookWithManager(mgr); err != nil {
				setupLog.Error(err, "unable to create webhook", "webhook", "IPAMRange")
				os.Exit(1)
			}
		}

		if webhooks.Enabled(machineWebhook) {
			if err = (&computev1alpha1.Machine{}).SetupWebhookWithManager(mgr); err != nil {
				setupLog.Error(err, "unable to create webhook", "webhook", "Machine")
				os.Exit(1)
			}
		}

		if webhooks.Enabled(volumeWebhook) {
			if err = (&storagev1alpha1.Volume{}).SetupWebhookWithManager(mgr); err != nil {
				setupLog.Error(err, "unable to create webhook", "webhook", "Volume")
				os.Exit(1)
			}
		}
	}

	if err = (&networkcontrollers.PrefixReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Prefix")
		os.Exit(1)
	}
	if err = (&networkcontrollers.IPReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "IP")
		os.Exit(1)
	}
	if err = (&networkcontrollers.ClusterPrefixReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ClusterPrefix")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err = (&networkcontrollers.ClusterPrefixAllocationSchedulerReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ClusterPrefixAllocationScheduler")
		os.Exit(1)
	}
	if err = (&networkcontrollers.PrefixAllocationSchedulerReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "PrefixAllocationScheduler")
		os.Exit(1)
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
