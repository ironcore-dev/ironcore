// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"crypto/tls"
	goflag "flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

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
	"github.com/ironcore-dev/ironcore/poollet/volumepoollet/vem"
	"github.com/ironcore-dev/ironcore/utils/client/config"

	"github.com/ironcore-dev/controller-utils/configutils"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/certwatcher"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var scheme = runtime.NewScheme()

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
	SecureMetrics            bool
	MetricsCertPath          string
	MetricsCertName          string
	MetricsCertKey           string
	EnableHTTP2              bool
	EnableLeaderElection     bool
	LeaderElectionNamespace  string
	LeaderElectionKubeconfig string
	ProbeAddr                string

	VolumePoolName                      string
	VolumeDownwardAPILabels             map[string]string
	VolumeDownwardAPIAnnotations        map[string]string
	ProviderID                          string
	VolumeRuntimeEndpoint               string
	DialTimeout                         time.Duration
	VolumeRuntimeSocketDiscoveryTimeout time.Duration
	VolumeClassMapperSyncTimeout        time.Duration

	ChannelCapacity int
	RelistPeriod    time.Duration
	RelistThreshold time.Duration

	WatchFilterValue string

	MaxConcurrentReconciles int
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	o.GetConfigOptions.BindFlags(fs)
	fs.StringVar(&o.MetricsAddr, "metrics-bind-address", "0", "The address the metrics endpoint binds to. "+
		"Use :8443 for HTTPS or :8080 for HTTP, or leave as 0 to disable the metrics service.")
	fs.BoolVar(&o.SecureMetrics, "metrics-secure", true,
		"If set, the metrics endpoint is served securely via HTTPS. Use --metrics-secure=false to use HTTP instead.")
	fs.StringVar(&o.MetricsCertPath, "metrics-cert-path", "",
		"The directory that contains the metrics server certificate.")
	fs.StringVar(&o.MetricsCertName, "metrics-cert-name", "tls.crt", "The name of the metrics server certificate file.")
	fs.StringVar(&o.MetricsCertKey, "metrics-cert-key", "tls.key", "The name of the metrics server key file.")
	fs.BoolVar(&o.EnableHTTP2, "enable-http2", false,
		"If set, HTTP/2 will be enabled for the metrics server")
	fs.StringVar(&o.ProbeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	fs.BoolVar(&o.EnableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	fs.StringVar(&o.LeaderElectionNamespace, "leader-election-namespace", "", "Namespace to do leader election in.")
	fs.StringVar(&o.LeaderElectionKubeconfig, "leader-election-kubeconfig", "", "Path pointing to a kubeconfig to use for leader election.")

	fs.StringVar(&o.VolumePoolName, "volume-pool-name", o.VolumePoolName, "Name of the volume pool to announce / watch")
	fs.StringToStringVar(&o.VolumeDownwardAPILabels, "volume-downward-api-label", o.VolumeDownwardAPILabels, "Downward-API labels to set on IRI volume.")
	fs.StringToStringVar(&o.VolumeDownwardAPIAnnotations, "volume-downward-api-annotation", o.VolumeDownwardAPIAnnotations, "Downward-API annotations to set on the IRI volume.")
	fs.StringVar(&o.ProviderID, "provider-id", "", "Provider id to announce on the volume pool.")
	fs.StringVar(&o.VolumeRuntimeEndpoint, "volume-runtime-endpoint", o.VolumeRuntimeEndpoint, "Endpoint of the remote volume runtime service.")
	fs.DurationVar(&o.DialTimeout, "dial-timeout", 1*time.Second, "Timeout for dialing to the volume runtime endpoint.")
	fs.DurationVar(&o.VolumeRuntimeSocketDiscoveryTimeout, "volume-runtime-discovery-timeout", 20*time.Second, "Timeout for discovering the volume runtime socket.")
	fs.DurationVar(&o.VolumeClassMapperSyncTimeout, "vcm-sync-timeout", 10*time.Second, "Timeout waiting for the volume class mapper to sync.")

	fs.IntVar(&o.ChannelCapacity, "channel-capacity", 1024, "channel capacity for the bucket event generator.")
	fs.DurationVar(&o.RelistPeriod, "relist-period", 5*time.Second, "event channel relisting period.")
	fs.DurationVar(&o.RelistThreshold, "relist-threshold", 3*time.Minute, "event channel relisting threshold.")

	fs.StringVar(&o.WatchFilterValue, "watch-filter", "", "Value to filter for while watching.")

	fs.IntVar(&o.MaxConcurrentReconciles, "max-concurrent-reconciles", 1, "Maximum number of concurrent reconciles.")
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

	setupLog.Info("IRI Client configuration", "ChannelCapacity", opts.ChannelCapacity, "RelistPeriod", opts.RelistPeriod, "RelistThreshold", opts.RelistThreshold)
	setupLog.Info("Kubernetes Client configuration", "QPS", cfg.QPS, "Burst", cfg.Burst)

	leaderElectionCfg, err := configutils.GetConfig(
		configutils.Kubeconfig(opts.LeaderElectionKubeconfig),
	)
	if err != nil {
		return fmt.Errorf("error creating leader election kubeconfig: %w", err)
	}

	// if the enable-http2 flag is false (the default), http/2 should be disabled
	// due to its vulnerabilities. More specifically, disabling http/2 will
	// prevent from being vulnerable to the HTTP/2 Stream Cancellation and
	// Rapid Reset CVEs. For more information see:
	// - https://github.com/advisories/GHSA-qppj-fm5r-hxr3
	// - https://github.com/advisories/GHSA-4374-p667-p6c8
	var tlsOpts []func(*tls.Config)
	disableHTTP2 := func(c *tls.Config) {
		setupLog.Info("disabling http/2")
		c.NextProtos = []string{"http/1.1"}
	}
	if !opts.EnableHTTP2 {
		tlsOpts = append(tlsOpts, disableHTTP2)
	}
	// Metrics endpoint is enabled in 'config/volumepoollet-broker/default/kustomization.yaml'. The Metrics options configure the server.
	// More info:
	// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.1/pkg/metrics/server
	// - https://book.kubebuilder.io/reference/metrics.html
	metricsServerOptions := metricsserver.Options{
		BindAddress:   opts.MetricsAddr,
		SecureServing: opts.SecureMetrics,
		TLSOpts:       tlsOpts,
	}
	if opts.SecureMetrics {
		// FilterProvider is used to protect the metrics endpoint with authn/authz.
		// These configurations ensure that only authorized users and service accounts
		// can access the metrics endpoint. The RBAC are configured in 'config/volumepoollet-broker/broker-rbac/kustomization.yaml'. More info:
		// https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.1/pkg/metrics/filters#WithAuthenticationAndAuthorization
		metricsServerOptions.FilterProvider = filters.WithAuthenticationAndAuthorization
	}

	// If the certificate is not specified, controller-runtime will automatically
	// generate self-signed certificates for the metrics server. While convenient for development and testing,
	// this setup is not recommended for production.
	//
	// TODO(user): If you enable certManager, uncomment the following lines:
	// - [METRICS-WITH-CERTS] at config/volumepoollet-broker/default/kustomization.yaml to generate and use certificates
	// managed by cert-manager for the metrics server.
	// - [PROMETHEUS-WITH-CERTS] at config/volumepoollet-broker/prometheus/kustomization.yaml for TLS certification.

	// Create watchers for metrics certificates
	var metricsCertWatcher *certwatcher.CertWatcher

	if len(opts.MetricsCertPath) > 0 {
		setupLog.Info("Initializing metrics certificate watcher using provided certificates",
			"metrics-cert-path", opts.MetricsCertPath, "metrics-cert-name", opts.MetricsCertName, "metrics-cert-key", opts.MetricsCertKey)

		var err error
		metricsCertWatcher, err = certwatcher.New(
			filepath.Join(opts.MetricsCertPath, opts.MetricsCertName),
			filepath.Join(opts.MetricsCertPath, opts.MetricsCertKey),
		)
		if err != nil {
			setupLog.Error(err, "to initialize metrics certificate watcher", "error", err)
			os.Exit(1)
		}

		metricsServerOptions.TLSOpts = append(metricsServerOptions.TLSOpts, func(config *tls.Config) {
			config.GetCertificate = metricsCertWatcher.GetCertificate
		})
	}

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Logger:                  logger,
		Scheme:                  scheme,
		Metrics:                 metricsServerOptions,
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

	if metricsCertWatcher != nil {
		setupLog.Info("Adding metrics certificate watcher to manager")
		if err := mgr.Add(metricsCertWatcher); err != nil {
			setupLog.Error(err, "unable to add metrics certificate watcher to manager")
			os.Exit(1)
		}
	}

	version, err := volumeRuntime.Version(ctx, &iri.VersionRequest{})
	if err != nil {
		return fmt.Errorf("error getting volume runtime version: %w", err)
	}
	volumeClassMapper := vcm.NewGeneric(volumeRuntime, vcm.GenericOptions{})
	if err := mgr.Add(volumeClassMapper); err != nil {
		return fmt.Errorf("error adding volume class mapper: %w", err)
	}
	volumeEventMapper := vem.NewVolumeEventMapper(mgr.GetClient(), volumeRuntime, mgr.GetEventRecorderFor("volume-cluster-events"), vem.VolumeEventMapperOptions{})
	if err := mgr.Add(volumeEventMapper); err != nil {
		return fmt.Errorf("error adding volume event mapper: %w", err)
	}

	volumeEvents := irievent.NewGenerator(func(ctx context.Context) ([]*iri.Volume, error) {
		res, err := volumeRuntime.ListVolumes(ctx, &iri.ListVolumesRequest{})
		if err != nil {
			return nil, err
		}
		return res.Volumes, nil
	}, irievent.GeneratorOptions{
		ChannelCapacity: opts.ChannelCapacity,
		RelistPeriod:    opts.RelistPeriod,
		RelistThreshold: opts.RelistThreshold,
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
			EventRecorder:           mgr.GetEventRecorderFor("volumes"),
			Client:                  mgr.GetClient(),
			Scheme:                  scheme,
			VolumeRuntime:           volumeRuntime,
			VolumeRuntimeName:       version.RuntimeName,
			VolumeClassMapper:       volumeClassMapper,
			VolumePoolName:          opts.VolumePoolName,
			DownwardAPILabels:       opts.VolumeDownwardAPILabels,
			DownwardAPIAnnotations:  opts.VolumeDownwardAPIAnnotations,
			WatchFilterValue:        opts.WatchFilterValue,
			MaxConcurrentReconciles: opts.MaxConcurrentReconciles,
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
