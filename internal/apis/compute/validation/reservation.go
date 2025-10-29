// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"github.com/ironcore-dev/ironcore/internal/apis/core"

	ironcorevalidation "github.com/ironcore-dev/ironcore/internal/api/validation"
	"github.com/ironcore-dev/ironcore/internal/apis/compute"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateReservation validates a Reservation object.
func ValidateReservation(reservation *compute.Reservation) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessor(reservation, true, apivalidation.NameIsDNSLabel, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateReservationSpec(&reservation.Spec, field.NewPath("spec"))...)

	return allErrs
}

// ValidateReservationUpdate validates a Reservation object before an update.
func ValidateReservationUpdate(newReservation, oldReservation *compute.Reservation) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, apivalidation.ValidateObjectMetaAccessorUpdate(newReservation, oldReservation, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateReservationSpecUpdate(&newReservation.Spec, &oldReservation.Spec, newReservation.DeletionTimestamp != nil, field.NewPath("spec"))...)
	allErrs = append(allErrs, ValidateReservation(newReservation)...)

	return allErrs
}

// validateReservationSpec validates the spec of a Reservation object.
func validateReservationSpec(reservationSpec *compute.ReservationSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, validateReservationResources(reservationSpec.Resources, fldPath.Child("resources"))...)

	return allErrs
}

func validateReservationResources(resources core.ResourceList, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	for name, resource := range resources {
		allErrs = append(allErrs, ironcorevalidation.ValidatePositiveQuantity(resource, fldPath.Key(string(name)))...)
	}

	return allErrs
}

// validateReservationSpecUpdate validates the spec of a Reservation object before an update.
func validateReservationSpecUpdate(new, old *compute.ReservationSpec, deletionTimestampSet bool, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, ironcorevalidation.ValidateImmutableField(new.Resources, old.Resources, fldPath.Child("resources"))...)

	return allErrs
}
