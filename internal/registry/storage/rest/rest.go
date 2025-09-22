// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	"github.com/ironcore-dev/ironcore/internal/api"
	"github.com/ironcore-dev/ironcore/internal/apis/storage"
	bucketstorage "github.com/ironcore-dev/ironcore/internal/registry/storage/bucket/storage"
	bucketclassstore "github.com/ironcore-dev/ironcore/internal/registry/storage/bucketclass/storage"
	bucketpoolstorage "github.com/ironcore-dev/ironcore/internal/registry/storage/bucketpool/storage"
	volumestorage "github.com/ironcore-dev/ironcore/internal/registry/storage/volume/storage"
	volumeclassstore "github.com/ironcore-dev/ironcore/internal/registry/storage/volumeclass/storage"
	volumepoolstorage "github.com/ironcore-dev/ironcore/internal/registry/storage/volumepool/storage"
	volumesnapshotstorage "github.com/ironcore-dev/ironcore/internal/registry/storage/volumesnapshot/storage"
	ironcoreserializer "github.com/ironcore-dev/ironcore/internal/serializer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"

	serverstorage "k8s.io/apiserver/pkg/server/storage"
)

type StorageProvider struct{}

func (p StorageProvider) GroupName() string {
	return storage.SchemeGroupVersion.Group
}

func (p StorageProvider) NewRESTStorage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) (genericapiserver.APIGroupInfo, bool, error) {
	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(p.GroupName(), api.Scheme, metav1.ParameterCodec, api.Codecs)
	apiGroupInfo.NegotiatedSerializer = ironcoreserializer.DefaultSubsetNegotiatedSerializer(api.Codecs)

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

	volumeSnapshotStorage, err := volumesnapshotstorage.NewStorage(restOptionsGetter)
	if err != nil {
		return storageMap, err
	}

	storageMap["volumesnapshots"] = volumeSnapshotStorage.VolumeSnapshot
	storageMap["volumesnapshots/status"] = volumeSnapshotStorage.Status

	return storageMap, nil
}
