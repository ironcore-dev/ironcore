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
var ipamrangelog = logf.Log.WithName("ipamrange-resource")

func (r *IPAMRange) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-network-onmetal-de-v1alpha1-ipamrange,mutating=true,failurePolicy=fail,sideEffects=None,groups=network.onmetal.de,resources=ipamranges,verbs=create;update,versions=v1alpha1,name=mipamrange.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &IPAMRange{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *IPAMRange) Default() {
	ipamrangelog.Info("default", "name", r.Name)
}

//+kubebuilder:webhook:path=/validate-network-onmetal-de-v1alpha1-ipamrange,mutating=false,failurePolicy=fail,sideEffects=None,groups=network.onmetal.de,resources=ipamranges,verbs=create;update;delete,versions=v1alpha1,name=vipamrange.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &IPAMRange{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *IPAMRange) ValidateCreate() error {
	ipamrangelog.Info("validate create", "name", r.Name)

	var allErrs field.ErrorList
	path := field.NewPath("spec")

	if r.Spec.Parent != nil {
		if r.Spec.Parent.Name == "" {
			allErrs = append(allErrs, field.Required(path.Child("parent").Child("name"), r.Spec.Parent.Name))
		}
	}

	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(IPAMRangeGK, r.Name, allErrs)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *IPAMRange) ValidateUpdate(old runtime.Object) error {
	ipamrangelog.Info("validate update", "name", r.Name)

	oldRange := old.(*IPAMRange)
	path := field.NewPath("spec")

	var allErrs field.ErrorList

	if !reflect.DeepEqual(r.Spec.Parent, oldRange.Spec.Parent) {
		allErrs = append(allErrs, field.Invalid(path.Child("parent"), r.Spec.Parent, fieldImmutable))
	}

	if !reflect.DeepEqual(r.Spec.Items, oldRange.Spec.Items) {
		allErrs = append(allErrs, field.Invalid(path.Child("elements"), r.Spec.Items, fieldImmutable))
	}

	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(IPAMRangeGK, r.Name, allErrs)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *IPAMRange) ValidateDelete() error {
	ipamrangelog.Info("validate delete", "name", r.Name)
	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
