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
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	"github.com/onmetal/onmetal-api/internal/api"
	"github.com/onmetal/onmetal-api/internal/apis/networking"
	aliasprefixstorage "github.com/onmetal/onmetal-api/internal/registry/networking/aliasprefix/storage"
	aliasprefixroutingstoratge "github.com/onmetal/onmetal-api/internal/registry/networking/aliasprefixrouting/storage"
	loadbalancerstorage "github.com/onmetal/onmetal-api/internal/registry/networking/loadbalancer/storage"
	loadbalancerroutingstorage "github.com/onmetal/onmetal-api/internal/registry/networking/loadbalancerrouting/storage"
	natgatewaystorage "github.com/onmetal/onmetal-api/internal/registry/networking/natgateway/storage"
	natgatewayroutingstorage "github.com/onmetal/onmetal-api/internal/registry/networking/natgatewayrouting/storage"
	networkstorage "github.com/onmetal/onmetal-api/internal/registry/networking/network/storage"
	networkinterfacestorage "github.com/onmetal/onmetal-api/internal/registry/networking/networkinterface/storage"
	networkpolicystorage "github.com/onmetal/onmetal-api/internal/registry/networking/networkpolicy/storage"
	virtualipstorage "github.com/onmetal/onmetal-api/internal/registry/networking/virtualip/storage"
	onmetalapiserializer "github.com/onmetal/onmetal-api/internal/serializer"
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
	apiGroupInfo.NegotiatedSerializer = onmetalapiserializer.DefaultSubsetNegotiatedSerializer(api.Codecs)

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
	storageMap["networks/status"] = networkStorage.Status

	networkPolicyStorage, err := networkpolicystorage.NewStorage(restOptionsGetter)
	if err != nil {
		return storageMap, err
	}

	storageMap["networkpolicies"] = networkPolicyStorage.NetworkPolicy
	storageMap["networkpolicies/status"] = networkPolicyStorage.Status

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

	loadBalancerStorage, err := loadbalancerstorage.NewStorage(restOptionsGetter)
	if err != nil {
		return storageMap, err
	}

	storageMap["loadbalancers"] = loadBalancerStorage.LoadBalancer
	storageMap["loadbalancers/status"] = loadBalancerStorage.Status

	loadBalancerRoutingStorage, err := loadbalancerroutingstorage.NewStorage(restOptionsGetter)
	if err != nil {
		return storageMap, err
	}

	storageMap["loadbalancerroutings"] = loadBalancerRoutingStorage.LoadBalancerRouting

	natGatewayStorage, err := natgatewaystorage.NewStorage(restOptionsGetter)
	if err != nil {
		return storageMap, err
	}

	storageMap["natgateways"] = natGatewayStorage.NATGateway
	storageMap["natgateways/status"] = natGatewayStorage.Status

	natGatewayRoutingStorage, err := natgatewayroutingstorage.NewStorage(restOptionsGetter)
	if err != nil {
		return storageMap, err
	}

	storageMap["natgatewayroutings"] = natGatewayRoutingStorage.NATGatewayRouting

	return storageMap, nil
}
