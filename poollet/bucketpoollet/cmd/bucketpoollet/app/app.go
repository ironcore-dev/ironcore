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
	iri "github.com/ironcore-dev/ironcore/iri/apis/bucket/v1alpha1"
	iriremotebucket "github.com/ironcore-dev/ironcore/iri/remote/bucket"
	"github.com/ironcore-dev/ironcore/poollet/bucketpoollet/bcm"
	bucketpoolletconfig "github.com/ironcore-dev/ironcore/poollet/bucketpoollet/client/config"
	"github.com/ironcore-dev/ironcore/poollet/bucketpoollet/controllers"
	"github.com/ironcore-dev/ironcore/poollet/irievent"
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

	BucketPoolName                      string
	ProviderID                          string
	BucketRuntimeEndpoint               string
	DialTimeout                         time.Duration
	BucketRuntimeSocketDiscoveryTimeout time.Duration
	BucketClassMapperSyncTimeout        time.Duration

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

	fs.StringVar(&o.BucketPoolName, "bucket-pool-name", o.BucketPoolName, "Name of the bucket pool to announce / watch")
	fs.StringVar(&o.ProviderID, "provider-id", "", "Provider id to announce on the bucket pool.")
	fs.StringVar(&o.BucketRuntimeEndpoint, "bucket-runtime-endpoint", o.BucketRuntimeEndpoint, "Endpoint of the remote bucket runtime service.")
	fs.DurationVar(&o.DialTimeout, "dial-timeout", 1*time.Second, "Timeout for dialing to the bucket runtime endpoint.")
	fs.DurationVar(&o.BucketRuntimeSocketDiscoveryTimeout, "bucket-runtime-discovery-timeout", 20*time.Second, "Timeout for discovering the bucket runtime socket.")
	fs.DurationVar(&o.BucketClassMapperSyncTimeout, "bcm-sync-timeout", 10*time.Second, "Timeout waiting for the bucket class mapper to sync.")

	fs.IntVar(&o.ChannelCapacity, "channel-capacity", 1024, "channel capacity for the bucket event generator")

	fs.StringVar(&o.WatchFilterValue, "watch-filter", "", "Value to filter for while watching.")
}

func (o *Options) MarkFlagsRequired(cmd *cobra.Command) {
	_ = cmd.MarkFlagRequired("bucket-pool-name")
	_ = cmd.MarkFlagRequired("provider-id")
}

func Command() *cobra.Command {
	var (
		zapOpts = zap.Options{Development: true}
		opts    Options
	)

	cmd := &cobra.Command{
		Use: "bucketpoollet",
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

	getter, err := bucketpoolletconfig.NewGetter(opts.BucketPoolName)
	if err != nil {
		setupLog.Error(err, "Error creating new getter")
		os.Exit(1)
	}

	endpoint, err := iriremotebucket.GetAddressWithTimeout(opts.BucketRuntimeSocketDiscoveryTimeout, opts.BucketRuntimeEndpoint)
	if err != nil {
		return fmt.Errorf("error detecting bucket runtime endpoint: %w", err)
	}

	bucketRuntime, err := iriremotebucket.NewRemoteRuntime(endpoint)
	if err != nil {
		return fmt.Errorf("error creating remote bucket runtime: %w", err)
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
		LeaderElectionID:        "dwfepysc.ironcore.dev",
		LeaderElectionNamespace: opts.LeaderElectionNamespace,
		LeaderElectionConfig:    leaderElectionCfg,
	})
	if err != nil {
		return fmt.Errorf("error creating manager: %w", err)
	}
	if err := config.SetupControllerWithManager(mgr, configCtrl); err != nil {
		return err
	}

	bucketClassMapper := bcm.NewGeneric(bucketRuntime, bcm.GenericOptions{})
	if err := mgr.Add(bucketClassMapper); err != nil {
		return fmt.Errorf("error adding bucket class mapper: %w", err)
	}

	bucketEvents := irievent.NewGenerator(func(ctx context.Context) ([]*iri.Bucket, error) {
		res, err := bucketRuntime.ListBuckets(ctx, &iri.ListBucketsRequest{})
		if err != nil {
			return nil, err
		}
		return res.Buckets, nil
	}, irievent.GeneratorOptions{
		ChannelCapacity: opts.ChannelCapacity,
	})
	if err := mgr.Add(bucketEvents); err != nil {
		return fmt.Errorf("error adding bucket event generator: %w", err)
	}
	if err := mgr.AddHealthzCheck("bucket-events", bucketEvents.Check); err != nil {
		return fmt.Errorf("error adding bucket event generator healthz check: %w", err)
	}

	onInitialized := func(ctx context.Context) error {
		bucketClassMapperSyncCtx, cancel := context.WithTimeout(ctx, opts.BucketClassMapperSyncTimeout)
		defer cancel()

		if err := bucketClassMapper.WaitForSync(bucketClassMapperSyncCtx); err != nil {
			return fmt.Errorf("error waiting for bucket class mapper to sync: %w", err)
		}

		if err := (&controllers.BucketReconciler{
			EventRecorder:     mgr.GetEventRecorderFor("buckets"),
			Client:            mgr.GetClient(),
			Scheme:            scheme,
			BucketRuntime:     bucketRuntime,
			BucketClassMapper: bucketClassMapper,
			BucketPoolName:    opts.BucketPoolName,
			WatchFilterValue:  opts.WatchFilterValue,
		}).SetupWithManager(mgr); err != nil {
			return fmt.Errorf("error setting up bucket reconciler with manager: %w", err)
		}

		if err := (&controllers.BucketAnnotatorReconciler{
			Client:       mgr.GetClient(),
			BucketEvents: bucketEvents,
		}).SetupWithManager(mgr); err != nil {
			return fmt.Errorf("error setting up bucket annotator reconciler with manager: %w", err)
		}

		if err := (&controllers.BucketPoolReconciler{
			Client:            mgr.GetClient(),
			BucketPoolName:    opts.BucketPoolName,
			BucketClassMapper: bucketClassMapper,
			BucketRuntime:     bucketRuntime,
		}).SetupWithManager(mgr); err != nil {
			return fmt.Errorf("error setting up bucket pool reconciler with manager: %w", err)
		}

		return nil
	}

	if err := (&controllers.BucketPoolInit{
		Client:         mgr.GetClient(),
		BucketPoolName: opts.BucketPoolName,
		ProviderID:     opts.ProviderID,
		OnInitialized:  onInitialized,
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("error setting up bucket pool init with manager: %w", err)
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
