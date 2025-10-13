// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"

	"github.com/ironcore-dev/ironcore/internal/apis/storage"
	"github.com/ironcore-dev/ironcore/internal/registry/storage/bucket"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/generic"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"
	apisrvstorage "k8s.io/apiserver/pkg/storage"
	"sigs.k8s.io/structured-merge-diff/v6/fieldpath"
)

type BucketStorage struct {
	Bucket *REST
	Status *StatusREST
}

type REST struct {
	*genericregistry.Store
}

func NewStorage(optsGetter generic.RESTOptionsGetter) (BucketStorage, error) {
	store := &genericregistry.Store{
		NewFunc: func() runtime.Object {
			return &storage.Bucket{}
		},
		NewListFunc: func() runtime.Object {
			return &storage.BucketList{}
		},
		PredicateFunc:             bucket.MatchBucket,
		DefaultQualifiedResource:  storage.Resource("buckets"),
		SingularQualifiedResource: storage.Resource("bucket"),

		CreateStrategy: bucket.Strategy,
		UpdateStrategy: bucket.Strategy,
		DeleteStrategy: bucket.Strategy,

		TableConvertor: newTableConvertor(),
	}

	options := &generic.StoreOptions{
		RESTOptions: optsGetter,
		AttrFunc:    bucket.GetAttrs,
		TriggerFunc: map[string]apisrvstorage.IndexerFunc{
			storage.BucketBucketPoolRefNameField: bucket.BucketPoolRefNameTriggerFunc,
		},
		Indexers: bucket.Indexers(),
	}
	if err := store.CompleteWithOptions(options); err != nil {
		return BucketStorage{}, err
	}

	statusStore := *store
	statusStore.UpdateStrategy = bucket.StatusStrategy
	statusStore.ResetFieldsStrategy = bucket.StatusStrategy

	return BucketStorage{
		Bucket: &REST{store},
		Status: &StatusREST{&statusStore},
	}, nil
}

type StatusREST struct {
	store *genericregistry.Store
}

func (r *StatusREST) New() runtime.Object {
	return &storage.Bucket{}
}

func (r *StatusREST) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	return r.store.Get(ctx, name, options)
}

func (r *StatusREST) Update(ctx context.Context, name string, objInfo rest.UpdatedObjectInfo, createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc, forceAllowCreate bool, options *metav1.UpdateOptions) (runtime.Object, bool, error) {
	return r.store.Update(ctx, name, objInfo, createValidation, updateValidation, false, options)
}

func (r *StatusREST) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return r.store.GetResetFields()
}

func (r *StatusREST) Destroy() {}
