// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"

	"github.com/ironcore-dev/ironcore/internal/apis/compute"
	"github.com/ironcore-dev/ironcore/internal/registry/compute/reservation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/generic"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/apiserver/pkg/storage"
	"sigs.k8s.io/structured-merge-diff/v4/fieldpath"
)

type ReservationStorage struct {
	Reservation *REST
	Status      *StatusREST
}

type REST struct {
	*genericregistry.Store
}

func NewStorage(optsGetter generic.RESTOptionsGetter) (ReservationStorage, error) {
	store := &genericregistry.Store{
		NewFunc: func() runtime.Object {
			return &compute.Reservation{}
		},
		NewListFunc: func() runtime.Object {
			return &compute.ReservationList{}
		},
		PredicateFunc:             reservation.MatchReservation,
		DefaultQualifiedResource:  compute.Resource("reservations"),
		SingularQualifiedResource: compute.Resource("reservation"),

		CreateStrategy: reservation.Strategy,
		UpdateStrategy: reservation.Strategy,
		DeleteStrategy: reservation.Strategy,

		TableConvertor: newTableConvertor(),
	}

	options := &generic.StoreOptions{
		RESTOptions: optsGetter,
		AttrFunc:    reservation.GetAttrs,
		TriggerFunc: map[string]storage.IndexerFunc{},
		Indexers:    reservation.Indexers(),
	}
	if err := store.CompleteWithOptions(options); err != nil {
		return ReservationStorage{}, err
	}

	statusStore := *store
	statusStore.UpdateStrategy = reservation.StatusStrategy
	statusStore.ResetFieldsStrategy = reservation.StatusStrategy

	return ReservationStorage{
		Reservation: &REST{store},
		Status:      &StatusREST{&statusStore},
	}, nil
}

type StatusREST struct {
	store *genericregistry.Store
}

func (r *StatusREST) New() runtime.Object {
	return &compute.Reservation{}
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
