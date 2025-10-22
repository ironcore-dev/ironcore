// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"
	"fmt"

	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	"github.com/ironcore-dev/ironcore/internal/apis/compute"
	"github.com/ironcore-dev/ironcore/internal/apis/compute/v1alpha1"
	"github.com/ironcore-dev/ironcore/internal/machinepoollet/client"
	"github.com/ironcore-dev/ironcore/internal/registry/compute/machinepool"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/generic"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"
	"sigs.k8s.io/structured-merge-diff/v6/fieldpath"
)

type REST struct {
	*genericregistry.Store
}

func (r *REST) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	return r.Store.Get(ctx, name, options)
}

type MachinePoolStorage struct {
	MachinePool                  *REST
	Status                       *StatusREST
	MachinePoolletConnectionInfo client.ConnectionInfoGetter
}

func NewStorage(optsGetter generic.RESTOptionsGetter, machinePoolletClientConfig client.MachinePoolletClientConfig) (MachinePoolStorage, error) {
	store := &genericregistry.Store{
		NewFunc: func() runtime.Object {
			return &compute.MachinePool{}
		},
		NewListFunc: func() runtime.Object {
			return &compute.MachinePoolList{}
		},
		PredicateFunc:             machinepool.MatchMachinePool,
		DefaultQualifiedResource:  compute.Resource("machinepools"),
		SingularQualifiedResource: compute.Resource("machinepool"),

		CreateStrategy: machinepool.Strategy,
		UpdateStrategy: machinepool.Strategy,
		DeleteStrategy: machinepool.Strategy,

		TableConvertor: newTableConvertor(),
	}

	options := &generic.StoreOptions{RESTOptions: optsGetter, AttrFunc: machinepool.GetAttrs}
	if err := store.CompleteWithOptions(options); err != nil {
		return MachinePoolStorage{}, err
	}

	statusStore := *store
	statusStore.UpdateStrategy = machinepool.StatusStrategy
	statusStore.ResetFieldsStrategy = machinepool.StatusStrategy

	machinePoolRest := &REST{store}

	// Build a MachinePoolGetter that looks up nodes using the REST handler
	machinePoolGetter := client.MachinePoolGetterFunc(func(ctx context.Context, machinePoolName string, options metav1.GetOptions) (*computev1alpha1.MachinePool, error) {
		obj, err := machinePoolRest.Get(ctx, machinePoolName, &options)
		if err != nil {
			return nil, err
		}
		machinePool, ok := obj.(*compute.MachinePool)
		if !ok {
			return nil, fmt.Errorf("unexpected type %T", obj)
		}
		// TODO: Remove the conversion. Consider only return the MachinePoolAddresses
		externalMachinePool := &computev1alpha1.MachinePool{}
		if err := v1alpha1.Convert_compute_MachinePool_To_v1alpha1_MachinePool(machinePool, externalMachinePool, nil); err != nil {
			return nil, fmt.Errorf("failed to convert to v1alpha1.MachinePool: %v", err)
		}
		return externalMachinePool, nil
	})
	connectionInfoGetter, err := client.NewMachinePoolConnectionInfoGetter(machinePoolGetter, machinePoolletClientConfig)
	if err != nil {
		return MachinePoolStorage{}, err
	}

	return MachinePoolStorage{
		MachinePool:                  machinePoolRest,
		Status:                       &StatusREST{&statusStore},
		MachinePoolletConnectionInfo: connectionInfoGetter,
	}, nil
}

type StatusREST struct {
	store *genericregistry.Store
}

func (r *StatusREST) New() runtime.Object {
	return &compute.MachinePool{}
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
