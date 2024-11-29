// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package reservation

import (
	"context"
	"fmt"

	"github.com/ironcore-dev/ironcore/internal/api"
	"github.com/ironcore-dev/ironcore/internal/apis/compute"
	"github.com/ironcore-dev/ironcore/internal/apis/compute/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/registry/generic"
	apisrvstorage "k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/names"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/structured-merge-diff/v4/fieldpath"
)

func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	reservation, ok := obj.(*compute.Reservation)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a Reservation")
	}
	return reservation.Labels, SelectableFields(reservation), nil
}

func MatchReservation(label labels.Selector, field fields.Selector) apisrvstorage.SelectionPredicate {
	return apisrvstorage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

func SelectableFields(reservation *compute.Reservation) fields.Set {
	fieldsSet := make(fields.Set)
	return generic.AddObjectMetaFieldsSet(fieldsSet, &reservation.ObjectMeta, true)
}

func Indexers() *cache.Indexers {
	return &cache.Indexers{}
}

type reservationStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

var Strategy = reservationStrategy{api.Scheme, names.SimpleNameGenerator}

func (reservationStrategy) NamespaceScoped() bool {
	return true
}

func (reservationStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
}

func (reservationStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

func (reservationStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	reservation := obj.(*compute.Reservation)
	return validation.ValidateReservation(reservation)
}

func (reservationStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (reservationStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (reservationStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (reservationStrategy) Canonicalize(obj runtime.Object) {
}

func (reservationStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	oldReservation := old.(*compute.Reservation)
	newReservation := obj.(*compute.Reservation)
	return validation.ValidateReservationUpdate(newReservation, oldReservation)
}

func (reservationStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

type reservationStatusStrategy struct {
	reservationStrategy
}

var StatusStrategy = reservationStatusStrategy{Strategy}

func (reservationStatusStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return map[fieldpath.APIVersion]*fieldpath.Set{
		"compute.ironcore.dev/v1alpha1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("spec"),
		),
	}
}

func (reservationStatusStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newReservation := obj.(*compute.Reservation)
	oldReservation := old.(*compute.Reservation)
	newReservation.Spec = oldReservation.Spec
}

func (reservationStatusStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newReservation := obj.(*compute.Reservation)
	oldReservation := old.(*compute.Reservation)
	return validation.ValidateReservationUpdate(newReservation, oldReservation)
}

func (reservationStatusStrategy) WarningsOnUpdate(cxt context.Context, obj, old runtime.Object) []string {
	return nil
}

type ResourceGetter interface {
	Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error)
}

func getReservation(ctx context.Context, getter ResourceGetter, name string) (*compute.Reservation, error) {
	obj, err := getter.Get(ctx, name, &metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	machine, ok := obj.(*compute.Reservation)
	if !ok {
		return nil, fmt.Errorf("unexpected object type %T", obj)
	}
	return machine, nil
}
