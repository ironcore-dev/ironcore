// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"crypto/tls"
	goflag "flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
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
	"github.com/ironcore-dev/ironcore/poollet/machinepoollet/mem"
	"github.com/ironcore-dev/ironcore/poollet/machinepoollet/server"
	"github.com/ironcore-dev/ironcore/utils/client/config"

	"github.com/ironcore-dev/controller-utils/configutils"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/certwatcher"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var scheme = runtime.NewScheme()

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
	SecureMetrics            bool
	MetricsCertPath          string
	MetricsCertName          string
	MetricsCertKey           string
	EnableHTTP2              bool
	EnableLeaderElection     bool
	LeaderElectionNamespace  string
	LeaderElectionKubeconfig string
	ProbeAddr                string
	PprofAddr                string

	MachinePoolName               string
	MachineDownwardAPILabels      map[string]string
	MachineDownwardAPIAnnotations map[string]string

	VolumeDownwardAPILabels      map[string]string
	VolumeDownwardAPIAnnotations map[string]string

	NicDownwardAPILabels      map[string]string
	NicDownwardAPIAnnotations map[string]string

	NetworkDownwardAPILabels      map[string]string
	NetworkDownwardAPIAnnotations map[string]string

	TopologyRegionLabel string
	TopologyZoneLabel   string

	ProviderID                           string
	MachineRuntimeEndpoint               string
	MachineRuntimeSocketDiscoveryTimeout time.Duration
	DialTimeout                          time.Duration
	MachineClassMapperSyncTimeout        time.Duration

	ChannelCapacity int
	RelistPeriod    time.Duration
	RelistThreshold time.Duration

	ServerFlags server.Flags

	AddressesOptions addresses.GetOptions

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
	fs.StringVar(&o.PprofAddr, "pprof-bind-address", "", "The address the Pprof endpoint binds to.")
	fs.BoolVar(&o.EnableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	fs.StringVar(&o.LeaderElectionNamespace, "leader-election-namespace", "", "Namespace to do leader election in.")
	fs.StringVar(&o.LeaderElectionKubeconfig, "leader-election-kubeconfig", "", "Path pointing to a kubeconfig to use for leader election.")

	fs.StringVar(&o.MachinePoolName, "machine-pool-name", o.MachinePoolName, "Name of the machine pool to announce / watch")
	fs.StringToStringVar(&o.MachineDownwardAPILabels, "machine-downward-api-label", o.MachineDownwardAPILabels, "Downward-API labels to set on the iri machine.")
	fs.StringToStringVar(&o.MachineDownwardAPIAnnotations, "machine-downward-api-annotation", o.MachineDownwardAPIAnnotations, "Downward-API annotations to set on the iri machine.")
	fs.StringToStringVar(&o.VolumeDownwardAPILabels, "volume-downward-api-label", o.VolumeDownwardAPILabels, "Downward-API labels to set on the iri volume.")
	fs.StringToStringVar(&o.VolumeDownwardAPIAnnotations, "volume-downward-api-annotation", o.VolumeDownwardAPIAnnotations, "Downward-API annotations to set on the iri volume.")
	fs.StringToStringVar(&o.NicDownwardAPILabels, "nic-downward-api-label", o.NicDownwardAPILabels, "Downward-API labels to set on the iri nic.")
	fs.StringToStringVar(&o.NicDownwardAPIAnnotations, "nic-downward-api-annotation", o.NicDownwardAPIAnnotations, "Downward-API annotations to set on the iri nic.")
	fs.StringToStringVar(&o.NetworkDownwardAPILabels, "network-downward-api-label", o.NetworkDownwardAPILabels, "Downward-API labels to set on the iri network.")
	fs.StringToStringVar(&o.NetworkDownwardAPIAnnotations, "network-downward-api-annotation", o.NetworkDownwardAPIAnnotations, "Downward-API annotations to set on the iri network.")

	fs.StringVar(&o.TopologyRegionLabel, "topology-region-label", "", "Label to use for the region topology information.")
	fs.StringVar(&o.TopologyZoneLabel, "topology-zone-label", "", "Label to use for the zone topology information.")

	fs.StringVar(&o.ProviderID, "provider-id", "", "Provider id to announce on the machine pool.")
	fs.StringVar(&o.MachineRuntimeEndpoint, "machine-runtime-endpoint", o.MachineRuntimeEndpoint, "Endpoint of the remote machine runtime service.")
	fs.DurationVar(&o.MachineRuntimeSocketDiscoveryTimeout, "machine-runtime-socket-discovery-timeout", 20*time.Second, "Timeout for discovering the machine runtime socket.")
	fs.DurationVar(&o.DialTimeout, "dial-timeout", 1*time.Second, "Timeout for dialing to the machine runtime endpoint.")
	fs.DurationVar(&o.MachineClassMapperSyncTimeout, "mcm-sync-timeout", 10*time.Second, "Timeout waiting for the machine class mapper to sync.")

	fs.IntVar(&o.ChannelCapacity, "channel-capacity", 1024, "channel capacity for the machine event generator.")
	fs.DurationVar(&o.RelistPeriod, "relist-period", 5*time.Second, "event channel relisting period.")
	fs.DurationVar(&o.RelistThreshold, "relist-threshold", 3*time.Minute, "event channel relisting threshold.")

	o.ServerFlags.BindFlags(fs)

	o.AddressesOptions.BindFlags(fs)

	fs.StringVar(&o.WatchFilterValue, "watch-filter", "", "Value to filter for while watching.")

	fs.IntVar(&o.MaxConcurrentReconciles, "max-concurrent-reconciles", 1, "Maximum number of concurrent reconciles.")
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
	// Metrics endpoint is enabled in 'config/machinepoollet-broker/default/kustomization.yaml'. The Metrics options configure the server.
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
		// can access the metrics endpoint. The RBAC are configured in 'config/machinepoollet-broker/broker-rbac/kustomization.yaml'. More info:
		// https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.1/pkg/metrics/filters#WithAuthenticationAndAuthorization
		metricsServerOptions.FilterProvider = filters.WithAuthenticationAndAuthorization
	}

	// If the certificate is not specified, controller-runtime will automatically
	// generate self-signed certificates for the metrics server. While convenient for development and testing,
	// this setup is not recommended for production.
	//
	// TODO(user): If you enable certManager, uncomment the following lines:
	// - [METRICS-WITH-CERTS] at config/machinepoollet-broker/default/kustomization.yaml to generate and use certificates
	// managed by cert-manager for the metrics server.
	// - [PROMETHEUS-WITH-CERTS] at config/machinepoollet-broker/prometheus/kustomization.yaml for TLS certification.

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

	topologyLabels := map[commonv1alpha1.TopologyLabel]string{}
	if opts.TopologyRegionLabel != "" {
		topologyLabels[commonv1alpha1.TopologyLabelRegion] = opts.TopologyRegionLabel
	}
	if opts.TopologyZoneLabel != "" {
		topologyLabels[commonv1alpha1.TopologyLabelZone] = opts.TopologyZoneLabel
	}

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Logger:                  logger,
		Scheme:                  scheme,
		Metrics:                 metricsServerOptions,
		HealthProbeBindAddress:  opts.ProbeAddr,
		PprofBindAddress:        opts.PprofAddr,
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

	if metricsCertWatcher != nil {
		setupLog.Info("Adding metrics certificate watcher to manager")
		if err := mgr.Add(metricsCertWatcher); err != nil {
			setupLog.Error(err, "unable to add metrics certificate watcher to manager")
			os.Exit(1)
		}
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

	machineEventMapper := mem.NewMachineEventMapper(mgr.GetClient(), machineRuntime, mgr.GetEventRecorderFor("machine-cluster-events"), mem.MachineEventMapperOptions{})
	if err := mgr.Add(machineEventMapper); err != nil {
		return fmt.Errorf("error adding machine event mapper: %w", err)
	}

	machineEvents := irievent.NewGenerator(func(ctx context.Context) ([]*iri.Machine, error) {
		res, err := machineRuntime.ListMachines(ctx, &iri.ListMachinesRequest{})
		if err != nil {
			return nil, err
		}
		return res.Machines, nil
	}, irievent.GeneratorOptions{
		ChannelCapacity: opts.ChannelCapacity,
		RelistPeriod:    opts.RelistPeriod,
		RelistThreshold: opts.RelistThreshold,
	})
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
			EventRecorder:                 mgr.GetEventRecorderFor("machines"),
			Client:                        mgr.GetClient(),
			MachineRuntime:                machineRuntime,
			MachineRuntimeName:            version.RuntimeName,
			MachineRuntimeVersion:         version.RuntimeVersion,
			MachineClassMapper:            machineClassMapper,
			MachinePoolName:               opts.MachinePoolName,
			MachineDownwardAPILabels:      opts.MachineDownwardAPILabels,
			MachineDownwardAPIAnnotations: opts.MachineDownwardAPIAnnotations,
			VolumeDownwardAPILabels:       opts.VolumeDownwardAPILabels,
			VolumeDownwardAPIAnnotations:  opts.VolumeDownwardAPIAnnotations,
			NicDownwardAPILabels:          opts.NicDownwardAPILabels,
			NicDownwardAPIAnnotations:     opts.NicDownwardAPIAnnotations,
			NetworkDownwardAPILabels:      opts.NetworkDownwardAPILabels,
			NetworkDownwardAPIAnnotations: opts.NetworkDownwardAPIAnnotations,
			WatchFilterValue:              opts.WatchFilterValue,
			MaxConcurrentReconciles:       opts.MaxConcurrentReconciles,
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
			TopologyLabels:     topologyLabels,
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
		Client:              mgr.GetClient(),
		MachinePoolName:     opts.MachinePoolName,
		ProviderID:          opts.ProviderID,
		TopologyRegionLabel: opts.TopologyRegionLabel,
		TopologyZoneLabel:   opts.TopologyZoneLabel,
		TopologyLabels:      topologyLabels,
		OnInitialized:       onInitialized,
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
