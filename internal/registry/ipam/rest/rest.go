// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	ipamv1alpha1 "github.com/ironcore-dev/ironcore/api/ipam/v1alpha1"
	"github.com/ironcore-dev/ironcore/internal/api"
	"github.com/ironcore-dev/ironcore/internal/apis/ipam"
	prefixstorage "github.com/ironcore-dev/ironcore/internal/registry/ipam/prefix/storage"
	prefixallocationstorage "github.com/ironcore-dev/ironcore/internal/registry/ipam/prefixallocation/storage"
	ironcoreserializer "github.com/ironcore-dev/ironcore/internal/serializer"
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
	apiGroupInfo.NegotiatedSerializer = ironcoreserializer.DefaultSubsetNegotiatedSerializer(api.Codecs)

	storageMap, err := p.v1alpha1Storage(restOptionsGetter)
	if err != nil {
		return genericapiserver.APIGroupInfo{}, false, err
	}

	apiGroupInfo.VersionedResourcesStorageMap[ipamv1alpha1.SchemeGroupVersion.Version] = storageMap

	return apiGroupInfo, true, nil
}

func (p StorageProvider) v1alpha1Storage(restOptionsGetter generic.RESTOptionsGetter) (map[string]rest.Storage, error) {
	storageMap := map[string]rest.Storage{}

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
