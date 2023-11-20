// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package apiserver

import (
	"fmt"

	"github.com/ironcore-dev/ironcore/internal/machinepoollet/client"
	computerest "github.com/ironcore-dev/ironcore/internal/registry/compute/rest"
	corerest "github.com/ironcore-dev/ironcore/internal/registry/core/rest"
	ipamrest "github.com/ironcore-dev/ironcore/internal/registry/ipam/rest"
	networkingrest "github.com/ironcore-dev/ironcore/internal/registry/networking/rest"
	storagerest "github.com/ironcore-dev/ironcore/internal/registry/storage/rest"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/apiserver/pkg/registry/generic"
	genericapiserver "k8s.io/apiserver/pkg/server"
	serverstorage "k8s.io/apiserver/pkg/server/storage"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	logf = ctrl.Log.WithName("apiserver")
)

// ExtraConfig holds custom apiserver config
type ExtraConfig struct {
	APIResourceConfigSource serverstorage.APIResourceConfigSource
	MachinePoolletConfig    client.MachinePoolletClientConfig
}

// Config defines the config for the apiserver
type Config struct {
	GenericConfig *genericapiserver.RecommendedConfig
	ExtraConfig   ExtraConfig
}

// IronCoreAPIServer contains state for a Kubernetes cluster master/api server.
type IronCoreAPIServer struct {
	GenericAPIServer *genericapiserver.GenericAPIServer
}

type completedConfig struct {
	GenericConfig genericapiserver.CompletedConfig
	ExtraConfig   *ExtraConfig
}

// CompletedConfig embeds a private pointer that cannot be instantiated outside of this package.
type CompletedConfig struct {
	*completedConfig
}

// Complete fills in any fields not set that are required to have valid data. It's mutating the receiver.
func (cfg *Config) Complete() CompletedConfig {
	c := completedConfig{
		cfg.GenericConfig.Complete(),
		&cfg.ExtraConfig,
	}

	c.GenericConfig.Version = &version.Info{
		Major: "1",
		Minor: "0",
	}

	return CompletedConfig{&c}
}

type RESTStorageProvider interface {
	GroupName() string
	NewRESTStorage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) (genericapiserver.APIGroupInfo, bool, error)
}

// New returns a new instance of IronCoreAPIServer from the given config.
func (c completedConfig) New() (*IronCoreAPIServer, error) {
	genericServer, err := c.GenericConfig.New("sample-apiserver", genericapiserver.NewEmptyDelegate())
	if err != nil {
		return nil, err
	}

	s := &IronCoreAPIServer{
		GenericAPIServer: genericServer,
	}

	apiResourceConfigSource := c.ExtraConfig.APIResourceConfigSource
	restStorageProviders := []RESTStorageProvider{
		ipamrest.StorageProvider{},
		corerest.StorageProvider{},
		computerest.StorageProvider{
			MachinePoolletClientConfig: c.ExtraConfig.MachinePoolletConfig,
		},
		networkingrest.StorageProvider{},
		storagerest.StorageProvider{},
	}

	var apiGroupsInfos []*genericapiserver.APIGroupInfo
	for _, restStorageProvider := range restStorageProviders {
		groupName := restStorageProvider.GroupName()
		logf := logf.WithValues("GroupName", groupName)

		if !apiResourceConfigSource.AnyResourceForGroupEnabled(groupName) {
			logf.V(1).Info("Skipping disabled api group")
			continue
		}

		apiGroupInfo, enabled, err := restStorageProvider.NewRESTStorage(apiResourceConfigSource, c.GenericConfig.RESTOptionsGetter)
		if err != nil {
			return nil, fmt.Errorf("error initializing api group %s: %w", groupName, err)
		}
		if !enabled {
			logf.Info("API Group is not enabled, skipping")
			continue
		}

		if postHookProvider, ok := restStorageProvider.(genericapiserver.PostStartHookProvider); ok {
			name, hook, err := postHookProvider.PostStartHook()
			if err != nil {
				return nil, fmt.Errorf("error building post start hook: %w", err)
			}

			if err := s.GenericAPIServer.AddPostStartHook(name, hook); err != nil {
				return nil, err
			}
		}

		apiGroupsInfos = append(apiGroupsInfos, &apiGroupInfo)
	}

	if err := s.GenericAPIServer.InstallAPIGroups(apiGroupsInfos...); err != nil {
		return nil, err
	}

	return s, nil
}
