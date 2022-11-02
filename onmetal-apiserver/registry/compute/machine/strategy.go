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

package machine

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/onmetal/onmetal-api/apis/compute"
	"github.com/onmetal/onmetal-api/apis/compute/validation"
	"github.com/onmetal/onmetal-api/onmetal-apiserver/api"
	"github.com/onmetal/onmetal-api/onmetal-apiserver/machinepoollet/client"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/registry/generic"
	apisrvstorage "k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/names"
	"sigs.k8s.io/structured-merge-diff/v4/fieldpath"
)

func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	machine, ok := obj.(*compute.Machine)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a Machine")
	}
	return machine.Labels, SelectableFields(machine), nil
}

func MatchMachine(label labels.Selector, field fields.Selector) apisrvstorage.SelectionPredicate {
	return apisrvstorage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

func SelectableFields(machine *compute.Machine) fields.Set {
	return generic.ObjectMetaFieldsSet(&machine.ObjectMeta, true)
}

type machineStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

var Strategy = machineStrategy{api.Scheme, names.SimpleNameGenerator}

func (machineStrategy) NamespaceScoped() bool {
	return true
}

func (machineStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
}

func (machineStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

func (machineStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	machine := obj.(*compute.Machine)
	return validation.ValidateMachine(machine)
}

func (machineStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (machineStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (machineStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (machineStrategy) Canonicalize(obj runtime.Object) {
}

func (machineStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	oldMachine := old.(*compute.Machine)
	newMachine := obj.(*compute.Machine)
	return validation.ValidateMachineUpdate(newMachine, oldMachine)
}

func (machineStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

type machineStatusStrategy struct {
	machineStrategy
}

var StatusStrategy = machineStatusStrategy{Strategy}

func (machineStatusStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return map[fieldpath.APIVersion]*fieldpath.Set{
		"compute.api.onmetal.de/v1alpha1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("spec"),
		),
	}
}

func (machineStatusStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newMachine := obj.(*compute.Machine)
	oldMachine := old.(*compute.Machine)
	newMachine.Spec = oldMachine.Spec
}

func (machineStatusStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newMachine := obj.(*compute.Machine)
	oldMachine := old.(*compute.Machine)
	return validation.ValidateMachineUpdate(newMachine, oldMachine)
}

func (machineStatusStrategy) WarningsOnUpdate(cxt context.Context, obj, old runtime.Object) []string {
	return nil
}

type ResourceGetter interface {
	Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error)
}

func ExecLocation(
	ctx context.Context,
	getter ResourceGetter,
	connInfo client.ConnectionInfoGetter,
	name string,
	opts *compute.MachineExecOptions,
) (*url.URL, http.RoundTripper, error) {
	machine, err := getMachine(ctx, getter, name)
	if err != nil {
		return nil, nil, err
	}

	machinePoolRef := machine.Spec.MachinePoolRef
	if machinePoolRef == nil {
		return nil, nil, apierrors.NewBadRequest(fmt.Sprintf("machine %s has no machine pool assigned", name))
	}

	machinePoolName := machinePoolRef.Name
	machinePoolInfo, err := connInfo.GetConnectionInfo(ctx, machinePoolName)
	if err != nil {
		return nil, nil, err
	}

	loc := &url.URL{
		Scheme: machinePoolInfo.Scheme,
		Host:   net.JoinHostPort(machinePoolInfo.Hostname, machinePoolInfo.Port),
		Path:   fmt.Sprintf("/apis/compute.api.onmetal.de/namespaces/%s/machines/%s/exec", machine.Namespace, machine.Name),
	}
	transport := machinePoolInfo.Transport
	if opts.InsecureSkipTLSVerifyBackend {
		transport = machinePoolInfo.InsecureSkipTLSVerifyTransport
	}

	return loc, transport, nil
}

func getMachine(ctx context.Context, getter ResourceGetter, name string) (*compute.Machine, error) {
	obj, err := getter.Get(ctx, name, &metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	machine, ok := obj.(*compute.Machine)
	if !ok {
		return nil, fmt.Errorf("unexpected object type %T", obj)
	}
	return machine, nil
}
