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
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	"github.com/onmetal/onmetal-api/internal/api"
	"github.com/onmetal/onmetal-api/internal/apis/storage"
	bucketstorage "github.com/onmetal/onmetal-api/internal/registry/storage/bucket/storage"
	bucketclassstore "github.com/onmetal/onmetal-api/internal/registry/storage/bucketclass/storage"
	bucketpoolstorage "github.com/onmetal/onmetal-api/internal/registry/storage/bucketpool/storage"
	volumestorage "github.com/onmetal/onmetal-api/internal/registry/storage/volume/storage"
	volumeclassstore "github.com/onmetal/onmetal-api/internal/registry/storage/volumeclass/storage"
	volumepoolstorage "github.com/onmetal/onmetal-api/internal/registry/storage/volumepool/storage"
	onmetalapiserializer "github.com/onmetal/onmetal-api/internal/serializer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"

	serverstorage "k8s.io/apiserver/pkg/server/storage"
)

type StorageProvider struct {
	GroupsToShowPoolResources []string
}

func (p StorageProvider) GroupName() string {
	return storage.SchemeGroupVersion.Group
}

func (p StorageProvider) NewRESTStorage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) (genericapiserver.APIGroupInfo, bool, error) {
	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(p.GroupName(), api.Scheme, metav1.ParameterCodec, api.Codecs)
	apiGroupInfo.NegotiatedSerializer = onmetalapiserializer.DefaultSubsetNegotiatedSerializer(api.Codecs)

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

	volumePoolStorage, err := volumepoolstorage.NewStorage(restOptionsGetter, p.GroupsToShowPoolResources)
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

	bucketClassStorage, err := bucketclassstore.NewStorage(restOptionsGetter)
	if err != nil {
		return storageMap, err
	}

	storageMap["bucketclasses"] = bucketClassStorage.BucketClass

	bucketPoolStorage, err := bucketpoolstorage.NewStorage(restOptionsGetter)
	if err != nil {
		return storageMap, err
	}

	storageMap["bucketpools"] = bucketPoolStorage.BucketPool
	storageMap["bucketpools/status"] = bucketPoolStorage.Status

	bucketStorage, err := bucketstorage.NewStorage(restOptionsGetter)
	if err != nil {
		return storageMap, err
	}

	storageMap["buckets"] = bucketStorage.Bucket
	storageMap["buckets/status"] = bucketStorage.Status

	return storageMap, nil
}
