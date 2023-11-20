// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	goflag "flag"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/ironcore-dev/controller-utils/configutils"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	ipamv1alpha1 "github.com/ironcore-dev/ironcore/api/ipam/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	computeclient "github.com/ironcore-dev/ironcore/internal/client/compute"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	iriremotemachine "github.com/ironcore-dev/ironcore/iri/remote/machine"
	"github.com/ironcore-dev/ironcore/poollet/irievent"
	"github.com/ironcore-dev/ironcore/poollet/machinepoollet/addresses"
	machinepoolletclient "github.com/ironcore-dev/ironcore/poollet/machinepoollet/client"
	machinepoolletconfig "github.com/ironcore-dev/ironcore/poollet/machinepoollet/client/config"
	"github.com/ironcore-dev/ironcore/poollet/machinepoollet/controllers"
	"github.com/ironcore-dev/ironcore/poollet/machinepoollet/mcm"
	"github.com/ironcore-dev/ironcore/poollet/machinepoollet/server"
	"github.com/ironcore-dev/ironcore/utils/client/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(computev1alpha1.AddToScheme(scheme))
	utilruntime.Must(storagev1alpha1.AddToScheme(scheme))
	utilruntime.Must(networkingv1alpha1.AddToScheme(scheme))
	utilruntime.Must(ipamv1alpha1.AddToScheme(scheme))
}

type Options struct {
	GetConfigOptions         config.GetConfigOptions
	MetricsAddr              string
	EnableLeaderElection     bool
	LeaderElectionNamespace  string
	LeaderElectionKubeconfig string
	ProbeAddr                string

	MachinePoolName                      string
	MachineDownwardAPILabels             map[string]string
	MachineDownwardAPIAnnotations        map[string]string
	ProviderID                           string
	MachineRuntimeEndpoint               string
	MachineRuntimeSocketDiscoveryTimeout time.Duration
	DialTimeout                          time.Duration
	MachineClassMapperSyncTimeout        time.Duration

	ServerFlags server.Flags

	AddressesOptions addresses.GetOptions

	WatchFilterValue string
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	o.GetConfigOptions.BindFlags(fs)
	fs.StringVar(&o.MetricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	fs.StringVar(&o.ProbeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	fs.BoolVar(&o.EnableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	fs.StringVar(&o.LeaderElectionNamespace, "leader-election-namespace", "", "Namespace to do leader election in.")
	fs.StringVar(&o.LeaderElectionKubeconfig, "leader-election-kubeconfig", "", "Path pointing to a kubeconfig to use for leader election.")

	fs.StringVar(&o.MachinePoolName, "machine-pool-name", o.MachinePoolName, "Name of the machine pool to announce / watch")
	fs.StringToStringVar(&o.MachineDownwardAPILabels, "machine-downward-api-label", o.MachineDownwardAPILabels, "Downward-API labels to set on the iri machine.")
	fs.StringToStringVar(&o.MachineDownwardAPIAnnotations, "machine-downward-api-annotation", o.MachineDownwardAPIAnnotations, "Downward-API annotations to set on the iri machine.")
	fs.StringVar(&o.ProviderID, "provider-id", "", "Provider id to announce on the machine pool.")
	fs.StringVar(&o.MachineRuntimeEndpoint, "machine-runtime-endpoint", o.MachineRuntimeEndpoint, "Endpoint of the remote machine runtime service.")
	fs.DurationVar(&o.MachineRuntimeSocketDiscoveryTimeout, "machine-runtime-socket-discovery-timeout", 20*time.Second, "Timeout for discovering the machine runtime socket.")
	fs.DurationVar(&o.DialTimeout, "dial-timeout", 1*time.Second, "Timeout for dialing to the machine runtime endpoint.")
	fs.DurationVar(&o.MachineClassMapperSyncTimeout, "mcm-sync-timeout", 10*time.Second, "Timeout waiting for the machine class mapper to sync.")

	o.ServerFlags.BindFlags(fs)

	o.AddressesOptions.BindFlags(fs)

	fs.StringVar(&o.WatchFilterValue, "watch-filter", "", "Value to filter for while watching.")
}

func (o *Options) MarkFlagsRequired(cmd *cobra.Command) {
	_ = cmd.MarkFlagRequired("machine-pool-name")
	_ = cmd.MarkFlagRequired("provider-id")
}

func NewOptions() *Options {
	return &Options{
		ServerFlags: *server.NewServerFlags(),
	}
}

func Command() *cobra.Command {
	var (
		zapOpts = zap.Options{Development: true}
		opts    = NewOptions()
	)

	cmd := &cobra.Command{
		Use: "machinepoollet",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			logger := zap.New(zap.UseFlagOptions(&zapOpts))
			ctrl.SetLogger(logger)
			cmd.SetContext(ctrl.LoggerInto(cmd.Context(), ctrl.Log))
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return Run(ctx, *opts)
		},
	}

	goFlags := goflag.NewFlagSet("", 0)
	zapOpts.BindFlags(goFlags)
	cmd.PersistentFlags().AddGoFlagSet(goFlags)

	opts.AddFlags(cmd.Flags())
	opts.MarkFlagsRequired(cmd)

	return cmd
}

func getPort(address string) (int32, error) {
	_, portString, err := net.SplitHostPort(address)
	if err != nil {
		return 0, fmt.Errorf("error splitting serving address into host / port: %w", err)
	}

	portInt64, err := strconv.ParseInt(portString, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("error parsing port %q: %w", portString, err)
	}

	if portInt64 == 0 {
		return 0, fmt.Errorf("cannot specify dynamic port")
	}
	return int32(portInt64), nil
}

func Run(ctx context.Context, opts Options) error {
	logger := ctrl.LoggerFrom(ctx)
	setupLog := ctrl.Log.WithName("setup")

	port, err := getPort(opts.ServerFlags.Serving.Address)
	if err != nil {
		return fmt.Errorf("error getting port from address: %w", err)
	}

	getter, err := machinepoolletconfig.NewGetter(opts.MachinePoolName)
	if err != nil {
		return fmt.Errorf("error creating new getter: %w", err)
	}

	endpoint, err := iriremotemachine.GetAddressWithTimeout(opts.MachineRuntimeSocketDiscoveryTimeout, opts.MachineRuntimeEndpoint)
	if err != nil {
		return fmt.Errorf("error detecting machine runtime endpoint: %w", err)
	}

	machinePoolAddresses, err := addresses.Get(&opts.AddressesOptions)
	if err != nil {
		return fmt.Errorf("error getting machine pool endpoints: %w", err)
	}

	setupLog.V(1).Info("Discovered addresses to report", "MachinePoolAddresses", machinePoolAddresses)

	machineRuntime, err := iriremotemachine.NewRemoteRuntime(endpoint)
	if err != nil {
		return fmt.Errorf("error creating remote machine runtime: %w", err)
	}

	cfg, configCtrl, err := getter.GetConfig(ctx, &opts.GetConfigOptions)
	if err != nil {
		return fmt.Errorf("error getting config: %w", err)
	}

	leaderElectionCfg, err := configutils.GetConfig(
		configutils.Kubeconfig(opts.LeaderElectionKubeconfig),
	)
	if err != nil {
		return fmt.Errorf("error creating leader election kubeconfig: %w", err)
	}

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Logger:                  logger,
		Scheme:                  scheme,
		Metrics:                 metricsserver.Options{BindAddress: opts.MetricsAddr},
		HealthProbeBindAddress:  opts.ProbeAddr,
		LeaderElection:          opts.EnableLeaderElection,
		LeaderElectionID:        "bfafcebe.ironcore.dev",
		LeaderElectionNamespace: opts.LeaderElectionNamespace,
		LeaderElectionConfig:    leaderElectionCfg,
		Cache:                   cache.Options{ByObject: map[client.Object]cache.ByObject{}},
		NewCache: func(config *rest.Config, cacheOpts cache.Options) (cache.Cache, error) {
			cacheOpts.ByObject[&computev1alpha1.Machine{}] = cache.ByObject{
				Field: fields.OneTermEqualSelector(
					computev1alpha1.MachineMachinePoolRefNameField,
					opts.MachinePoolName,
				),
			}
			return cache.New(config, cacheOpts)
		},
	})
	if err != nil {
		return fmt.Errorf("error creating manager: %w", err)
	}
	if err := config.SetupControllerWithManager(mgr, configCtrl); err != nil {
		return err
	}

	version, err := machineRuntime.Version(ctx, &iri.VersionRequest{})
	if err != nil {
		return fmt.Errorf("error getting machine runtime version: %w", err)
	}

	srvOpts := opts.ServerFlags.ServerOptions(
		opts.MachinePoolName,
		machineRuntime,
		logger.WithName("server"),
	)
	srv, err := server.New(cfg, srvOpts)
	if err != nil {
		return fmt.Errorf("error creating machinepoollet server: %w", err)
	}

	if err := mgr.Add(srv); err != nil {
		return fmt.Errorf("error adding machinepoollet server to manager: %w", err)
	}

	machineClassMapper := mcm.NewGeneric(machineRuntime, mcm.GenericOptions{})
	if err := mgr.Add(machineClassMapper); err != nil {
		return fmt.Errorf("error adding machine class mapper: %w", err)
	}

	machineEvents := irievent.NewGenerator(func(ctx context.Context) ([]*iri.Machine, error) {
		res, err := machineRuntime.ListMachines(ctx, &iri.ListMachinesRequest{})
		if err != nil {
			return nil, err
		}
		return res.Machines, nil
	}, irievent.GeneratorOptions{})
	if err := mgr.Add(machineEvents); err != nil {
		return fmt.Errorf("error adding machine event generator: %w", err)
	}
	if err := mgr.AddHealthzCheck("machine-events", machineEvents.Check); err != nil {
		return fmt.Errorf("error adding machine event generator healthz check: %w", err)
	}

	indexer := mgr.GetFieldIndexer()
	if err := machinepoolletclient.SetupMachineSpecNetworkInterfaceNamesField(ctx, indexer, opts.MachinePoolName); err != nil {
		return fmt.Errorf("error setting up %s indexer with manager: %w", machinepoolletclient.MachineSpecNetworkInterfaceNamesField, err)
	}
	if err := machinepoolletclient.SetupMachineSpecSecretNamesField(ctx, indexer, opts.MachinePoolName); err != nil {
		return fmt.Errorf("error setting up %s indexer with manager: %w", machinepoolletclient.MachineSpecSecretNamesField, err)
	}
	if err := machinepoolletclient.SetupMachineSpecVolumeNamesField(ctx, indexer, opts.MachinePoolName); err != nil {
		return fmt.Errorf("error setting up %s indexer with manager: %w", machinepoolletclient.MachineSpecVolumeNamesField, err)
	}

	if err := computeclient.SetupMachineSpecMachinePoolRefNameFieldIndexer(ctx, indexer); err != nil {
		return fmt.Errorf("error setting up %s indexer with manager: %w", computeclient.MachineSpecMachinePoolRefNameField, err)
	}

	onInitialized := func(ctx context.Context) error {
		machineClassMapperSyncCtx, cancel := context.WithTimeout(ctx, opts.MachineClassMapperSyncTimeout)
		defer cancel()

		if err := machineClassMapper.WaitForSync(machineClassMapperSyncCtx); err != nil {
			return fmt.Errorf("error waiting for machine class mapper to sync: %w", err)
		}

		if err := (&controllers.MachineReconciler{
			EventRecorder:          mgr.GetEventRecorderFor("machines"),
			Client:                 mgr.GetClient(),
			MachineRuntime:         machineRuntime,
			MachineRuntimeName:     version.RuntimeName,
			MachineRuntimeVersion:  version.RuntimeVersion,
			MachineClassMapper:     machineClassMapper,
			MachinePoolName:        opts.MachinePoolName,
			DownwardAPILabels:      opts.MachineDownwardAPILabels,
			DownwardAPIAnnotations: opts.MachineDownwardAPIAnnotations,
			WatchFilterValue:       opts.WatchFilterValue,
		}).SetupWithManager(mgr); err != nil {
			return fmt.Errorf("error setting up machine reconciler with manager: %w", err)
		}

		if err := (&controllers.MachineAnnotatorReconciler{
			Client:        mgr.GetClient(),
			MachineEvents: machineEvents,
		}).SetupWithManager(mgr); err != nil {
			return fmt.Errorf("error setting up machine annotator reconciler with manager: %w", err)
		}

		if err := (&controllers.MachinePoolReconciler{
			Client:             mgr.GetClient(),
			MachinePoolName:    opts.MachinePoolName,
			Addresses:          machinePoolAddresses,
			Port:               port,
			MachineRuntime:     machineRuntime,
			MachineClassMapper: machineClassMapper,
		}).SetupWithManager(mgr); err != nil {
			return fmt.Errorf("error setting up machine pool reconciler with manager: %w", err)
		}

		if err := (&controllers.MachinePoolAnnotatorReconciler{
			Client:             mgr.GetClient(),
			MachinePoolName:    opts.MachinePoolName,
			MachineClassMapper: machineClassMapper,
		}).SetupWithManager(mgr); err != nil {
			return fmt.Errorf("error setting up machine pool annotator reconciler with manager: %w", err)
		}

		return nil
	}

	if err := (&controllers.MachinePoolInit{
		Client:          mgr.GetClient(),
		MachinePoolName: opts.MachinePoolName,
		ProviderID:      opts.ProviderID,
		OnInitialized:   onInitialized,
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("error setting up machine pool init with manager: %w", err)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return fmt.Errorf("error adding healthz check: %w", err)
	}

	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return fmt.Errorf("error adding readyz check: %w", err)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		return fmt.Errorf("error running manager: %w", err)
	}
	return nil
}
