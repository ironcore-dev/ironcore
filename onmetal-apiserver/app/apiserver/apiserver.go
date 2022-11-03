// Copyright 2022 OnMetal authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package apiserver

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/onmetal/onmetal-api/apis/compute"
	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	ipamv1alpha1 "github.com/onmetal/onmetal-api/apis/ipam/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/apis/networking/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/apis/storage/v1alpha1"
	clientset "github.com/onmetal/onmetal-api/generated/clientset/versioned"
	informers "github.com/onmetal/onmetal-api/generated/informers/externalversions"
	onmetalopenapi "github.com/onmetal/onmetal-api/generated/openapi"
	"github.com/onmetal/onmetal-api/onmetal-apiserver/admission/plugin/machinevolumedevices"
	"github.com/onmetal/onmetal-api/onmetal-apiserver/api"
	"github.com/onmetal/onmetal-api/onmetal-apiserver/apiserver"
	"github.com/onmetal/onmetal-api/onmetal-apiserver/machinepoollet/client"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/apiserver/pkg/endpoints/openapi"
	"k8s.io/apiserver/pkg/features"
	genericapiserver "k8s.io/apiserver/pkg/server"
	genericoptions "k8s.io/apiserver/pkg/server/options"
	serverstorage "k8s.io/apiserver/pkg/server/storage"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	netutils "k8s.io/utils/net"
)

const defaultEtcdPathPrefix = "/registry/onmetal.de"

func NewResourceConfig() *serverstorage.ResourceConfig {
	cfg := serverstorage.NewResourceConfig()
	cfg.EnableVersions(
		computev1alpha1.SchemeGroupVersion,
		storagev1alpha1.SchemeGroupVersion,
		networkingv1alpha1.SchemeGroupVersion,
		ipamv1alpha1.SchemeGroupVersion,
	)
	return cfg
}

type OnmetalAPIServerOptions struct {
	RecommendedOptions   *genericoptions.RecommendedOptions
	MachinePoolletConfig client.MachinePoolletClientConfig

	SharedInformerFactory informers.SharedInformerFactory
}

func (o *OnmetalAPIServerOptions) AddFlags(fs *pflag.FlagSet) {
	o.RecommendedOptions.AddFlags(fs)

	// machinepoollet related flags:
	fs.StringSliceVar(&o.MachinePoolletConfig.PreferredAddressTypes, "machinepoollet-preferred-address-types", o.MachinePoolletConfig.PreferredAddressTypes,
		"List of the preferred MachinePoolAddressTypes to use for machinepoollet connections.")

	fs.DurationVar(&o.MachinePoolletConfig.HTTPTimeout, "machinepoollet-timeout", o.MachinePoolletConfig.HTTPTimeout,
		"Timeout for machinepoollet operations.")

	fs.StringVar(&o.MachinePoolletConfig.CertFile, "machinepoollet-client-certificate", o.MachinePoolletConfig.CertFile,
		"Path to a client cert file for TLS.")

	fs.StringVar(&o.MachinePoolletConfig.KeyFile, "machinepoollet-client-key", o.MachinePoolletConfig.KeyFile,
		"Path to a client key file for TLS.")

	fs.StringVar(&o.MachinePoolletConfig.CAFile, "machinepoollet-certificate-authority", o.MachinePoolletConfig.CAFile,
		"Path to a cert file for the certificate authority.")
}

func NewOnmetalAPIServerOptions() *OnmetalAPIServerOptions {
	o := &OnmetalAPIServerOptions{
		RecommendedOptions: genericoptions.NewRecommendedOptions(
			defaultEtcdPathPrefix,
			api.Codecs.LegacyCodec(
				computev1alpha1.SchemeGroupVersion,
				storagev1alpha1.SchemeGroupVersion,
				networkingv1alpha1.SchemeGroupVersion,
				ipamv1alpha1.SchemeGroupVersion,
			),
		),
		MachinePoolletConfig: client.MachinePoolletClientConfig{
			Port:         12319,
			ReadOnlyPort: 12320,
			PreferredAddressTypes: []string{
				string(compute.MachinePoolHostName),

				// internal, preferring DNS if reported
				string(compute.MachinePoolInternalDNS),
				string(compute.MachinePoolInternalIP),

				// external, preferring DNS if reported
				string(compute.MachinePoolExternalDNS),
				string(compute.MachinePoolExternalIP),
			},
			HTTPTimeout: time.Duration(5) * time.Second,
		},
	}
	o.RecommendedOptions.Etcd.StorageConfig.EncodeVersioner = runtime.NewMultiGroupVersioner(
		computev1alpha1.SchemeGroupVersion,
		schema.GroupKind{Group: computev1alpha1.SchemeGroupVersion.Group},
		schema.GroupKind{Group: storagev1alpha1.SchemeGroupVersion.Group},
		schema.GroupKind{Group: networkingv1alpha1.SchemeGroupVersion.Group},
		schema.GroupKind{Group: ipamv1alpha1.SchemeGroupVersion.Group},
	)
	return o
}

func NewCommandStartOnmetalAPIServer(ctx context.Context, defaults *OnmetalAPIServerOptions) *cobra.Command {
	o := *defaults
	cmd := &cobra.Command{
		Short: "Launch an onmetal-api API server",
		Long:  "Launch an onmetal-api API server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Complete(); err != nil {
				return err
			}
			if err := o.Validate(args); err != nil {
				return err
			}
			if err := o.Run(ctx); err != nil {
				return err
			}
			return nil
		},
	}

	o.AddFlags(cmd.Flags())
	utilfeature.DefaultMutableFeatureGate.AddFlag(cmd.Flags())

	return cmd
}

func (o *OnmetalAPIServerOptions) Validate(args []string) error {
	var errors []error
	errors = append(errors, o.RecommendedOptions.Validate()...)
	return utilerrors.NewAggregate(errors)
}

func (o *OnmetalAPIServerOptions) Complete() error {
	machinevolumedevices.Register(o.RecommendedOptions.Admission.Plugins)

	o.RecommendedOptions.Admission.RecommendedPluginOrder = append(o.RecommendedOptions.Admission.RecommendedPluginOrder, machinevolumedevices.PluginName)

	return nil
}

func (o *OnmetalAPIServerOptions) Config() (*apiserver.Config, error) {
	if err := o.RecommendedOptions.SecureServing.MaybeDefaultWithSelfSignedCerts("localhost", nil, []net.IP{netutils.ParseIPSloppy("127.0.0.1")}); err != nil {
		return nil, fmt.Errorf("error creating self-signed certificates: %w", err)
	}

	o.RecommendedOptions.Etcd.StorageConfig.Paging = utilfeature.DefaultFeatureGate.Enabled(features.APIListChunking)

	o.RecommendedOptions.ExtraAdmissionInitializers = func(c *genericapiserver.RecommendedConfig) ([]admission.PluginInitializer, error) {
		client, err := clientset.NewForConfig(c.LoopbackClientConfig)
		if err != nil {
			return nil, err
		}

		informerFactory := informers.NewSharedInformerFactory(client, c.LoopbackClientConfig.Timeout)
		o.SharedInformerFactory = informerFactory
		return []admission.PluginInitializer{}, nil
	}

	serverConfig := genericapiserver.NewRecommendedConfig(api.Codecs)

	serverConfig.OpenAPIConfig = genericapiserver.DefaultOpenAPIConfig(onmetalopenapi.GetOpenAPIDefinitions, openapi.NewDefinitionNamer(api.Scheme))
	serverConfig.OpenAPIConfig.Info.Title = "onmetal-api"
	serverConfig.OpenAPIConfig.Info.Version = "0.1"

	if err := o.RecommendedOptions.ApplyTo(serverConfig); err != nil {
		return nil, err
	}

	apiResourceConfig := NewResourceConfig()

	config := &apiserver.Config{
		GenericConfig: serverConfig,
		ExtraConfig: apiserver.ExtraConfig{
			APIResourceConfigSource: apiResourceConfig,
			MachinePoolletConfig:    o.MachinePoolletConfig,
		},
	}

	if config.GenericConfig.EgressSelector != nil {
		// Use the config.GenericConfig.EgressSelector lookup to find the dialer to connect to the machinepoollet
		config.ExtraConfig.MachinePoolletConfig.Lookup = config.GenericConfig.EgressSelector.Lookup
	}

	return config, nil
}

func (o *OnmetalAPIServerOptions) Run(ctx context.Context) error {
	config, err := o.Config()
	if err != nil {
		return err
	}

	server, err := config.Complete().New()
	if err != nil {
		return err
	}

	server.GenericAPIServer.AddPostStartHookOrDie("start-onmetal-api-server-informers", func(context genericapiserver.PostStartHookContext) error {
		config.GenericConfig.SharedInformerFactory.Start(context.StopCh)
		o.SharedInformerFactory.Start(context.StopCh)
		return nil
	})

	return server.GenericAPIServer.PrepareRun().Run(ctx.Done())
}
