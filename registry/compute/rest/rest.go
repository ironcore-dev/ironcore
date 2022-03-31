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

package rest

import (
	"github.com/onmetal/onmetal-api/api"
	"github.com/onmetal/onmetal-api/apis/compute"
	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	machinestorage "github.com/onmetal/onmetal-api/registry/compute/machine/storage"
	machinepoolstorage "github.com/onmetal/onmetal-api/registry/compute/machinepool/storage"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"

	machineclassstore "github.com/onmetal/onmetal-api/registry/compute/machineclass/storage"

	serverstorage "k8s.io/apiserver/pkg/server/storage"
)

type StorageProvider struct{}

func (p StorageProvider) GroupName() string {
	return compute.SchemeGroupVersion.Group
}

func (p StorageProvider) NewRESTStorage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) (genericapiserver.APIGroupInfo, bool, error) {
	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(p.GroupName(), api.Scheme, metav1.ParameterCodec, api.Codecs)

	if apiResourceConfigSource.VersionEnabled(computev1alpha1.SchemeGroupVersion) {
		storageMap, err := p.v1alpha1Storage(restOptionsGetter)
		if err != nil {
			return genericapiserver.APIGroupInfo{}, false, err
		}

		apiGroupInfo.VersionedResourcesStorageMap[computev1alpha1.SchemeGroupVersion.Version] = storageMap
	}

	return apiGroupInfo, true, nil
}

func (p StorageProvider) v1alpha1Storage(restOptionsGetter generic.RESTOptionsGetter) (map[string]rest.Storage, error) {
	storageMap := map[string]rest.Storage{}

	machineClassStorage, err := machineclassstore.NewStorage(restOptionsGetter)
	if err != nil {
		return storageMap, err
	}

	storageMap["machineclasses"] = machineClassStorage.MachineClass

	machinePoolStorage, err := machinepoolstorage.NewStorage(restOptionsGetter)
	if err != nil {
		return storageMap, err
	}

	storageMap["machinepools"] = machinePoolStorage.MachinePool
	storageMap["machinepools/status"] = machinePoolStorage.Status

	machineStorage, err := machinestorage.NewStorage(restOptionsGetter)
	if err != nil {
		return storageMap, err
	}

	storageMap["machines"] = machineStorage.Machine
	storageMap["machines/status"] = machineStorage.Status

	return storageMap, nil
}
