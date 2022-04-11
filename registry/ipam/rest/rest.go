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
	"github.com/onmetal/onmetal-api/apis/ipam"
	ipamv1alpha1 "github.com/onmetal/onmetal-api/apis/ipam/v1alpha1"
	clusterprefixstorage "github.com/onmetal/onmetal-api/registry/ipam/clusterprefix/storage"
	clusterprefixallocationstorage "github.com/onmetal/onmetal-api/registry/ipam/clusterprefixallocation/storage"
	ipstorage "github.com/onmetal/onmetal-api/registry/ipam/ip/storage"
	prefixstorage "github.com/onmetal/onmetal-api/registry/ipam/prefix/storage"
	prefixallocationstorage "github.com/onmetal/onmetal-api/registry/ipam/prefixallocation/storage"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"

	serverstorage "k8s.io/apiserver/pkg/server/storage"
)

type StorageProvider struct{}

func (p StorageProvider) GroupName() string {
	return ipam.SchemeGroupVersion.Group
}

func (p StorageProvider) NewRESTStorage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) (genericapiserver.APIGroupInfo, bool, error) {
	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(p.GroupName(), api.Scheme, metav1.ParameterCodec, api.Codecs)

	if apiResourceConfigSource.VersionEnabled(ipamv1alpha1.SchemeGroupVersion) {
		storageMap, err := p.v1alpha1Storage(restOptionsGetter)
		if err != nil {
			return genericapiserver.APIGroupInfo{}, false, err
		}

		apiGroupInfo.VersionedResourcesStorageMap[ipamv1alpha1.SchemeGroupVersion.Version] = storageMap
	}

	return apiGroupInfo, true, nil
}

func (p StorageProvider) v1alpha1Storage(restOptionsGetter generic.RESTOptionsGetter) (map[string]rest.Storage, error) {
	storageMap := map[string]rest.Storage{}

	clusterPrefixStorage, err := clusterprefixstorage.NewStorage(restOptionsGetter)
	if err != nil {
		return storageMap, err
	}

	storageMap["clusterprefixes"] = clusterPrefixStorage.ClusterPrefix
	storageMap["clusterprefixes/status"] = clusterPrefixStorage.Status

	clusterPrefixAllocationStorage, err := clusterprefixallocationstorage.NewStorage(restOptionsGetter)
	if err != nil {
		return storageMap, err
	}

	storageMap["clusterprefixallocations"] = clusterPrefixAllocationStorage.ClusterPrefixAllocation
	storageMap["clusterprefixallocations/status"] = clusterPrefixAllocationStorage.Status

	ipStorage, err := ipstorage.NewStorage(restOptionsGetter)
	if err != nil {
		return storageMap, err
	}

	storageMap["ips"] = ipStorage.IP
	storageMap["ips/status"] = ipStorage.Status

	prefixStorage, err := prefixstorage.NewStorage(restOptionsGetter)
	if err != nil {
		return storageMap, err
	}

	storageMap["prefixes"] = prefixStorage.Prefix
	storageMap["prefixes/status"] = prefixStorage.Status

	prefixAllocationStorage, err := prefixallocationstorage.NewStorage(restOptionsGetter)
	if err != nil {
		return storageMap, err
	}

	storageMap["prefixallocations"] = prefixAllocationStorage.PrefixAllocation
	storageMap["prefixallocations/status"] = prefixAllocationStorage.Status

	return storageMap, nil
}
