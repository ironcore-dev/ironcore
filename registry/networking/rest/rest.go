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
	"github.com/onmetal/onmetal-api/apis/networking"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/apis/networking/v1alpha1"
	aliasprefixstorage "github.com/onmetal/onmetal-api/registry/networking/aliasprefix/storage"
	aliasprefixroutingstoratge "github.com/onmetal/onmetal-api/registry/networking/aliasprefixrouting/storage"
	networkstorage "github.com/onmetal/onmetal-api/registry/networking/network/storage"
	networkinterfacestorage "github.com/onmetal/onmetal-api/registry/networking/networkinterface/storage"
	virtualipstorage "github.com/onmetal/onmetal-api/registry/networking/virtualip/storage"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"

	"k8s.io/apiserver/pkg/server/storage"
)

type StorageProvider struct{}

func (p StorageProvider) GroupName() string {
	return networking.SchemeGroupVersion.Group
}

func (p StorageProvider) NewRESTStorage(apiResourceConfigSource storage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) (genericapiserver.APIGroupInfo, bool, error) {
	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(p.GroupName(), api.Scheme, metav1.ParameterCodec, api.Codecs)
	apiGroupInfo.PrioritizedVersions = []schema.GroupVersion{networkingv1alpha1.SchemeGroupVersion}

	storageMap, err := p.v1alpha1Storage(restOptionsGetter)
	if err != nil {
		return genericapiserver.APIGroupInfo{}, false, err
	}

	apiGroupInfo.VersionedResourcesStorageMap[networkingv1alpha1.SchemeGroupVersion.Version] = storageMap

	return apiGroupInfo, true, nil
}

func (p StorageProvider) v1alpha1Storage(restOptionsGetter generic.RESTOptionsGetter) (map[string]rest.Storage, error) {
	storageMap := map[string]rest.Storage{}

	networkInterfaceStorage, err := networkinterfacestorage.NewStorage(restOptionsGetter)
	if err != nil {
		return storageMap, err
	}

	storageMap["networkinterfaces"] = networkInterfaceStorage.NetworkInterface
	storageMap["networkinterfaces/status"] = networkInterfaceStorage.Status

	networkStorage, err := networkstorage.NewStorage(restOptionsGetter)
	if err != nil {
		return storageMap, err
	}

	storageMap["networks"] = networkStorage.Network

	virtualIPStorage, err := virtualipstorage.NewStorage(restOptionsGetter)
	if err != nil {
		return storageMap, err
	}

	storageMap["virtualips"] = virtualIPStorage.VirtualIP
	storageMap["virtualips/status"] = virtualIPStorage.Status

	aliasPrefixStorage, err := aliasprefixstorage.NewStorage(restOptionsGetter)
	if err != nil {
		return storageMap, err
	}

	storageMap["aliasprefixes"] = aliasPrefixStorage.AliasPrefix
	storageMap["aliasprefixes/status"] = aliasPrefixStorage.Status

	aliasPrefixRoutingStorage, err := aliasprefixroutingstoratge.NewStorage(restOptionsGetter)
	if err != nil {
		return storageMap, err
	}

	storageMap["aliasprefixroutings"] = aliasPrefixRoutingStorage.AliasPrefixRouting

	return storageMap, nil
}
