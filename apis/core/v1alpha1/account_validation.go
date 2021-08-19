/*
 * Copyright (c) 2021 by the OnMetal authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package v1alpha1

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// validateAccount validates the Account spec for errors
func (a *Account) validateAccount() error {
	var allErrs field.ErrorList

	// TODO: validation code goes here
	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(AccountGK, a.Name, allErrs)
}

// validationAccountUpdate validates an update of an Account
func (a *Account) validateAccountUpdate(old runtime.Object) error {
	var allErrs field.ErrorList

	// TODO: validation code goes here
	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(AccountGK, a.Name, allErrs)
}

// validateAccountDelete validates the deletion of an Account
func (a *Account) validateAccountDelete() error {
	var allErrs field.ErrorList

	// TODO: validation code goes here
	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(AccountGK, a.Name, allErrs)
}
