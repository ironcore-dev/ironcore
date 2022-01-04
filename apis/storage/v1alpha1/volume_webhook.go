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
	"reflect"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

const (
	fieldImmutable = "field is immutable"
)

// log is for logging in this package.
var volumelog = logf.Log.WithName("volume-resource")

func (r *Volume) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/validate-storage-onmetal-de-v1alpha1-volume,mutating=false,failurePolicy=fail,sideEffects=None,groups=storage.onmetal.de,resources=volumes,verbs=create;update,versions=v1alpha1,name=vvolume.storage.onmetal.de,admissionReviewVersions=v1

var _ webhook.Validator = &Volume{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Volume) ValidateCreate() error {
	volumelog.Info("validate create", "name", r.Name)
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Volume) ValidateUpdate(old runtime.Object) error {
	volumelog.Info("validate update", "name", r.Name)
	oldRange := old.(*Volume)
	path := field.NewPath("spec")

	var allErrs field.ErrorList
	if !reflect.DeepEqual(r.Spec.StorageClass, oldRange.Spec.StorageClass) {
		allErrs = append(allErrs, field.Invalid(path.Child("storageClass"), r.Spec.StorageClass, fieldImmutable))
	}

	if oldRange.Spec.StoragePool.Name != "" && !reflect.DeepEqual(r.Spec.StoragePool, oldRange.Spec.StoragePool) {
		allErrs = append(allErrs, field.Invalid(path.Child("storagePool"), r.Spec.StoragePool, fieldImmutable))
	}

	if !reflect.DeepEqual(r.Spec.StoragePoolSelector, oldRange.Spec.StoragePoolSelector) {
		allErrs = append(allErrs, field.Invalid(path.Child("storagePoolSelector"), r.Spec.StoragePoolSelector, fieldImmutable))
	}

	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(VolumeGK, r.Name, allErrs)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Volume) ValidateDelete() error {
	volumelog.Info("validate delete", "name", r.Name)
	return nil
}
