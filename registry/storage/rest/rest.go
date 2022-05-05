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
	"github.com/onmetal/onmetal-api/apis/storage"
	storagev1alpha1 "github.com/onmetal/onmetal-api/apis/storage/v1alpha1"
	volumestorage "github.com/onmetal/onmetal-api/registry/storage/volume/storage"
	volumeclaimstorage "github.com/onmetal/onmetal-api/registry/storage/volumeclaim/storage"
	volumepoolstorage "github.com/onmetal/onmetal-api/registry/storage/volumepool/storage"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"

	volumeclassstore "github.com/onmetal/onmetal-api/registry/storage/volumeclass/storage"

	serverstorage "k8s.io/apiserver/pkg/server/storage"
)

type StorageProvider struct{}

func (p StorageProvider) GroupName() string {
	return storage.SchemeGroupVersion.Group
}

func (p StorageProvider) NewRESTStorage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) (genericapiserver.APIGroupInfo, bool, error) {
	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(p.GroupName(), api.Scheme, metav1.ParameterCodec, api.Codecs)

	storageMap, err := p.v1alpha1Storage(restOptionsGetter)
	if err != nil {
		return genericapiserver.APIGroupInfo{}, false, err
	}

	apiGroupInfo.VersionedResourcesStorageMap[storagev1alpha1.SchemeGroupVersion.Version] = storageMap

	return apiGroupInfo, true, nil
}

func (p StorageProvider) v1alpha1Storage(restOptionsGetter generic.RESTOptionsGetter) (map[string]rest.Storage, error) {
	storageMap := map[string]rest.Storage{}

	volumeClassStorage, err := volumeclassstore.NewStorage(restOptionsGetter)
	if err != nil {
		return storageMap, err
	}

	storageMap["volumeclasses"] = volumeClassStorage.VolumeClass

	volumePoolStorage, err := volumepoolstorage.NewStorage(restOptionsGetter)
	if err != nil {
		return storageMap, err
	}

	storageMap["volumepools"] = volumePoolStorage.VolumePool
	storageMap["volumepools/status"] = volumePoolStorage.Status

	volumeStorage, err := volumestorage.NewStorage(restOptionsGetter)
	if err != nil {
		return storageMap, err
	}

	storageMap["volumes"] = volumeStorage.Volume
	storageMap["volumes/status"] = volumeStorage.Status

	volumeClaimStorage, err := volumeclaimstorage.NewStorage(restOptionsGetter)
	if err != nil {
		return storageMap, err
	}

	storageMap["volumeclaims"] = volumeClaimStorage.VolumeClaim
	storageMap["volumeclaims/status"] = volumeClaimStorage.Status

	return storageMap, nil
}
