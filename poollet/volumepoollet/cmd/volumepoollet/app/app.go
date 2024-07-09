// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	goflag "flag"
	"fmt"
	"os"
	"time"

	"github.com/ironcore-dev/controller-utils/configutils"
	ipamv1alpha1 "github.com/ironcore-dev/ironcore/api/ipam/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	storageclient "github.com/ironcore-dev/ironcore/internal/client/storage"
	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
	iriremotevolume "github.com/ironcore-dev/ironcore/iri/remote/volume"
	"github.com/ironcore-dev/ironcore/poollet/irievent"
	volumepoolletconfig "github.com/ironcore-dev/ironcore/poollet/volumepoollet/client/config"
	"github.com/ironcore-dev/ironcore/poollet/volumepoollet/controllers"
	"github.com/ironcore-dev/ironcore/poollet/volumepoollet/vcm"
	"github.com/ironcore-dev/ironcore/utils/client/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(storagev1alpha1.AddToScheme(scheme))
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

	VolumePoolName                      string
	ProviderID                          string
	VolumeRuntimeEndpoint               string
	DialTimeout                         time.Duration
	VolumeRuntimeSocketDiscoveryTimeout time.Duration
	VolumeClassMapperSyncTimeout        time.Duration

	ChannelCapacity int

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

	fs.StringVar(&o.VolumePoolName, "volume-pool-name", o.VolumePoolName, "Name of the volume pool to announce / watch")
	fs.StringVar(&o.ProviderID, "provider-id", "", "Provider id to announce on the volume pool.")
	fs.StringVar(&o.VolumeRuntimeEndpoint, "volume-runtime-endpoint", o.VolumeRuntimeEndpoint, "Endpoint of the remote volume runtime service.")
	fs.DurationVar(&o.DialTimeout, "dial-timeout", 1*time.Second, "Timeout for dialing to the volume runtime endpoint.")
	fs.DurationVar(&o.VolumeRuntimeSocketDiscoveryTimeout, "volume-runtime-discovery-timeout", 20*time.Second, "Timeout for discovering the volume runtime socket.")
	fs.DurationVar(&o.VolumeClassMapperSyncTimeout, "vcm-sync-timeout", 10*time.Second, "Timeout waiting for the volume class mapper to sync.")

	fs.IntVar(&o.ChannelCapacity, "channel-capacity", 1024, "channel capacity for the volume event generator")

	fs.StringVar(&o.WatchFilterValue, "watch-filter", "", "Value to filter for while watching.")
}

func (o *Options) MarkFlagsRequired(cmd *cobra.Command) {
	_ = cmd.MarkFlagRequired("volume-pool-name")
	_ = cmd.MarkFlagRequired("provider-id")
}

func Command() *cobra.Command {
	var (
		zapOpts = zap.Options{Development: true}
		opts    Options
	)

	cmd := &cobra.Command{
		Use: "volumepoollet",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			logger := zap.New(zap.UseFlagOptions(&zapOpts))
			ctrl.SetLogger(logger)
			cmd.SetContext(ctrl.LoggerInto(cmd.Context(), ctrl.Log))
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return Run(ctx, opts)
		},
	}

	goFlags := goflag.NewFlagSet("", 0)
	zapOpts.BindFlags(goFlags)
	cmd.PersistentFlags().AddGoFlagSet(goFlags)

	opts.AddFlags(cmd.Flags())
	opts.MarkFlagsRequired(cmd)

	return cmd
}

func Run(ctx context.Context, opts Options) error {
	logger := ctrl.LoggerFrom(ctx)
	setupLog := ctrl.Log.WithName("setup")

	getter, err := volumepoolletconfig.NewGetter(opts.VolumePoolName)
	if err != nil {
		setupLog.Error(err, "Error creating new getter")
		os.Exit(1)
	}

	endpoint, err := iriremotevolume.GetAddressWithTimeout(opts.VolumeRuntimeSocketDiscoveryTimeout, opts.VolumeRuntimeEndpoint)
	if err != nil {
		return fmt.Errorf("error detecting volume runtime endpoint: %w", err)
	}

	volumeRuntime, err := iriremotevolume.NewRemoteRuntime(endpoint)
	if err != nil {
		return fmt.Errorf("error creating remote volume runtime: %w", err)
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
		LeaderElectionID:        "dfffbeaa.ironcore.dev",
		LeaderElectionNamespace: opts.LeaderElectionNamespace,
		LeaderElectionConfig:    leaderElectionCfg,
	})
	if err != nil {
		return fmt.Errorf("error creating manager: %w", err)
	}
	if err := config.SetupControllerWithManager(mgr, configCtrl); err != nil {
		return err
	}

	volumeClassMapper := vcm.NewGeneric(volumeRuntime, vcm.GenericOptions{})
	if err := mgr.Add(volumeClassMapper); err != nil {
		return fmt.Errorf("error adding volume class mapper: %w", err)
	}

	volumeEvents := irievent.NewGenerator(func(ctx context.Context) ([]*iri.Volume, error) {
		res, err := volumeRuntime.ListVolumes(ctx, &iri.ListVolumesRequest{})
		if err != nil {
			return nil, err
		}
		return res.Volumes, nil
	}, irievent.GeneratorOptions{
		ChannelCapacity: opts.ChannelCapacity,
	})
	if err := mgr.Add(volumeEvents); err != nil {
		return fmt.Errorf("error adding volume event generator: %w", err)
	}
	if err := mgr.AddHealthzCheck("volume-events", volumeEvents.Check); err != nil {
		return fmt.Errorf("error adding volume event generator healthz check: %w", err)
	}

	indexer := mgr.GetFieldIndexer()
	if err := storageclient.SetupVolumeSpecVolumePoolRefNameFieldIndexer(ctx, indexer); err != nil {
		return fmt.Errorf("error setting up %s indexer with manager: %w", storageclient.VolumeSpecVolumePoolRefNameField, err)
	}

	onInitialized := func(ctx context.Context) error {
		volumeClassMapperSyncCtx, cancel := context.WithTimeout(ctx, opts.VolumeClassMapperSyncTimeout)
		defer cancel()

		if err := volumeClassMapper.WaitForSync(volumeClassMapperSyncCtx); err != nil {
			return fmt.Errorf("error waiting for volume class mapper to sync: %w", err)
		}

		if err := (&controllers.VolumeReconciler{
			EventRecorder:     mgr.GetEventRecorderFor("volumes"),
			Client:            mgr.GetClient(),
			Scheme:            scheme,
			VolumeRuntime:     volumeRuntime,
			VolumeClassMapper: volumeClassMapper,
			VolumePoolName:    opts.VolumePoolName,
			WatchFilterValue:  opts.WatchFilterValue,
		}).SetupWithManager(mgr); err != nil {
			return fmt.Errorf("error setting up volume reconciler with manager: %w", err)
		}

		if err := (&controllers.VolumeAnnotatorReconciler{
			Client:       mgr.GetClient(),
			VolumeEvents: volumeEvents,
		}).SetupWithManager(mgr); err != nil {
			return fmt.Errorf("error setting up volume annotator reconciler with manager: %w", err)
		}

		if err := (&controllers.VolumePoolReconciler{
			Client:            mgr.GetClient(),
			VolumePoolName:    opts.VolumePoolName,
			VolumeClassMapper: volumeClassMapper,
			VolumeRuntime:     volumeRuntime,
		}).SetupWithManager(mgr); err != nil {
			return fmt.Errorf("error setting up volume pool reconciler with manager: %w", err)
		}

		if err := (&controllers.VolumePoolAnnotatorReconciler{
			Client:            mgr.GetClient(),
			VolumeClassMapper: volumeClassMapper,
			VolumePoolName:    opts.VolumePoolName,
		}).SetupWithManager(mgr); err != nil {
			return fmt.Errorf("error setting up volume pool annotator reconciler with manager: %w", err)
		}

		return nil
	}

	if err := (&controllers.VolumePoolInit{
		Client:         mgr.GetClient(),
		VolumePoolName: opts.VolumePoolName,
		ProviderID:     opts.ProviderID,
		OnInitialized:  onInitialized,
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("error setting up volume pool init with manager: %w", err)
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
	if err := mgr.Start(ctx); err != nil {
		return fmt.Errorf("error running manager: %w", err)
	}
	return nil
}
