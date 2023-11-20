// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	"github.com/ironcore-dev/ironcore/internal/api"
	"github.com/ironcore-dev/ironcore/internal/apis/compute"
	machinepoolletclient "github.com/ironcore-dev/ironcore/internal/machinepoollet/client"
	machinestorage "github.com/ironcore-dev/ironcore/internal/registry/compute/machine/storage"
	machineclassstore "github.com/ironcore-dev/ironcore/internal/registry/compute/machineclass/storage"
	machinepoolstorage "github.com/ironcore-dev/ironcore/internal/registry/compute/machinepool/storage"
	ironcoreserializer "github.com/ironcore-dev/ironcore/internal/serializer"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"

	serverstorage "k8s.io/apiserver/pkg/server/storage"
)

type StorageProvider struct {
	MachinePoolletClientConfig machinepoolletclient.MachinePoolletClientConfig
}

func (p StorageProvider) GroupName() string {
	return compute.SchemeGroupVersion.Group
}

func (p StorageProvider) NewRESTStorage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) (genericapiserver.APIGroupInfo, bool, error) {
	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(p.GroupName(), api.Scheme, api.ParameterCodec, api.Codecs)
	apiGroupInfo.NegotiatedSerializer = ironcoreserializer.DefaultSubsetNegotiatedSerializer(api.Codecs)

	storageMap, err := p.v1alpha1Storage(restOptionsGetter)
	if err != nil {
		return genericapiserver.APIGroupInfo{}, false, err
	}

	apiGroupInfo.VersionedResourcesStorageMap[computev1alpha1.SchemeGroupVersion.Version] = storageMap

	return apiGroupInfo, true, nil
}

func (p StorageProvider) v1alpha1Storage(restOptionsGetter generic.RESTOptionsGetter) (map[string]rest.Storage, error) {
	storageMap := map[string]rest.Storage{}

	machineClassStorage, err := machineclassstore.NewStorage(restOptionsGetter)
	if err != nil {
		return storageMap, err
	}

	storageMap["machineclasses"] = machineClassStorage.MachineClass

	machinePoolStorage, err := machinepoolstorage.NewStorage(restOptionsGetter, p.MachinePoolletClientConfig)
	if err != nil {
		return storageMap, err
	}

	storageMap["machinepools"] = machinePoolStorage.MachinePool
	storageMap["machinepools/status"] = machinePoolStorage.Status

	machineStorage, err := machinestorage.NewStorage(restOptionsGetter, machinePoolStorage.MachinePoolletConnectionInfo)
	if err != nil {
		return storageMap, err
	}

	storageMap["machines"] = machineStorage.Machine
	storageMap["machines/status"] = machineStorage.Status
	storageMap["machines/exec"] = machineStorage.Exec

	return storageMap, nil
}
