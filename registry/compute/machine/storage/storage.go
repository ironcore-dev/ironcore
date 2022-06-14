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

package storage

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/onmetal/onmetal-api/apis/compute"
	"github.com/onmetal/onmetal-api/machinepoollet/client"
	"github.com/onmetal/onmetal-api/registry/compute/machine"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/proxy"
	"k8s.io/apiserver/pkg/registry/generic"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"
	"sigs.k8s.io/structured-merge-diff/v4/fieldpath"
)

type MachineStorage struct {
	Machine *REST
	Status  *StatusREST
	Exec    *ExecREST
}

type REST struct {
	*genericregistry.Store
}

func NewStorage(optsGetter generic.RESTOptionsGetter, k client.ConnectionInfoGetter) (MachineStorage, error) {
	store := &genericregistry.Store{
		NewFunc: func() runtime.Object {
			return &compute.Machine{}
		},
		NewListFunc: func() runtime.Object {
			return &compute.MachineList{}
		},
		PredicateFunc:            machine.MatchMachine,
		DefaultQualifiedResource: compute.Resource("machines"),

		CreateStrategy: machine.Strategy,
		UpdateStrategy: machine.Strategy,
		DeleteStrategy: machine.Strategy,

		TableConvertor: newTableConvertor(),
	}

	options := &generic.StoreOptions{RESTOptions: optsGetter, AttrFunc: machine.GetAttrs}
	if err := store.CompleteWithOptions(options); err != nil {
		return MachineStorage{}, err
	}

	statusStore := *store
	statusStore.UpdateStrategy = machine.StatusStrategy
	statusStore.ResetFieldsStrategy = machine.StatusStrategy

	return MachineStorage{
		Machine: &REST{store},
		Status:  &StatusREST{&statusStore},
		Exec:    &ExecREST{store, k},
	}, nil
}

type StatusREST struct {
	store *genericregistry.Store
}

func (r *StatusREST) New() runtime.Object {
	return &compute.Machine{}
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

// Support both GET and POST methods. We must support GET for browsers that want to use WebSockets.
var upgradeableMethods = []string{"GET", "POST"}

type ExecREST struct {
	Store       *genericregistry.Store
	MachineConn client.ConnectionInfoGetter
}

func (r *ExecREST) New() runtime.Object {
	return &compute.MachineExecOptions{}
}

func (r *ExecREST) Connect(ctx context.Context, name string, opts runtime.Object, responder rest.Responder) (http.Handler, error) {
	execOpts, ok := opts.(*compute.MachineExecOptions)
	if !ok {
		return nil, fmt.Errorf("invalid options objects: %#v", opts)
	}

	location, transport, err := machine.ExecLocation(ctx, r.Store, r.MachineConn, name, execOpts)
	if err != nil {
		return nil, err
	}

	return newThrottledUpgradeAwareProxyHandler(location, transport, false, false, responder), nil
}

func newThrottledUpgradeAwareProxyHandler(location *url.URL, transport http.RoundTripper, wrapTransport, upgradeRequired bool, responder rest.Responder) *proxy.UpgradeAwareHandler {
	handler := proxy.NewUpgradeAwareHandler(location, transport, wrapTransport, upgradeRequired, proxy.NewErrorResponder(responder))
	return handler
}

func (r *ExecREST) NewConnectOptions() (runtime.Object, bool, string) {
	return &compute.MachineExecOptions{}, false, ""
}

func (r *ExecREST) ConnectMethods() []string {
	return upgradeableMethods
}
