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

package network

import (
	"context"
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	networkv1alpha1 "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
)

const (
	parentField    = ".spec.parent.name"
	fieldImmutable = "field is immutable"
)

// log is for logging in this package.
var ipamrangelog = logf.Log.WithName("ipamrange-resource")

func SetupIPAMRangeWebhookWithManager(mgr ctrl.Manager) error {
	defaulter := &IPAMRangeDefaulter{}
	validator := &IPAMRangeValidator{mgr.GetClient()}

	return ctrl.NewWebhookManagedBy(mgr).
		WithDefaulter(defaulter).
		WithValidator(validator).
		For(&networkv1alpha1.IPAMRange{}).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-network-onmetal-de-v1alpha1-ipamrange,mutating=true,failurePolicy=fail,sideEffects=None,groups=network.onmetal.de,resources=ipamranges,verbs=create;update,versions=v1alpha1,name=mipamrange.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomDefaulter = &IPAMRangeDefaulter{}

type IPAMRangeDefaulter struct{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (d *IPAMRangeDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	r := obj.(*networkv1alpha1.IPAMRange)
	ipamrangelog.Info("default", "name", r.Name)
	return nil
}

//+kubebuilder:webhook:path=/validate-network-onmetal-de-v1alpha1-ipamrange,mutating=false,failurePolicy=fail,sideEffects=None,groups=network.onmetal.de,resources=ipamranges,verbs=create;update;delete,versions=v1alpha1,name=vipamrange.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomValidator = &IPAMRangeValidator{}

type IPAMRangeValidator struct {
	client.Client
}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type
func (v *IPAMRangeValidator) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	r := obj.(*networkv1alpha1.IPAMRange)
	ipamrangelog.Info("validate create", "name", r.Name)

	var allErrs field.ErrorList
	path := field.NewPath("spec")

	if r.Spec.Parent != nil {
		if r.Spec.Parent.Name == "" {
			allErrs = append(allErrs, field.Required(path.Child("parent").Child("name"), r.Spec.Parent.Name))
		}

		if errs := v.overlappingRequestExist(ctx, r); len(errs) > 0 {
			allErrs = append(allErrs, errs...)
		}
	} else {
		if errs := v.overlappingParentExists(ctx, r); len(errs) > 0 {
			allErrs = append(allErrs, errs...)
		}
	}

	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(networkv1alpha1.IPAMRangeGK, r.Name, allErrs)
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type
func (v *IPAMRangeValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) error {
	r := newObj.(*networkv1alpha1.IPAMRange)
	ipamrangelog.Info("validate update", "name", r.Name)

	oldRange := oldObj.(*networkv1alpha1.IPAMRange)
	path := field.NewPath("spec")

	var allErrs field.ErrorList
	if r.Spec.Parent != nil {
		if !reflect.DeepEqual(r.Spec.Parent, oldRange.Spec.Parent) {
			allErrs = append(allErrs, field.Invalid(path.Child("parent"), r.Spec.Parent, fieldImmutable))
		}
		if errs := v.overlappingRequestExist(ctx, r); len(errs) > 0 {
			allErrs = append(allErrs, errs...)
		}
	} else {
		if errs := v.overlappingParentExists(ctx, r); len(errs) > 0 {
			allErrs = append(allErrs, errs...)
		}
	}

	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(networkv1alpha1.IPAMRangeGK, r.Name, allErrs)
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type
func (v *IPAMRangeValidator) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	r := obj.(*networkv1alpha1.IPAMRange)
	ipamrangelog.Info("validate delete", "name", r.Name)

	list := &networkv1alpha1.IPAMRangeList{}
	if err := v.List(
		ctx, list, client.InNamespace(r.Namespace),
		client.MatchingFields{parentField: r.Name},
	); err != nil {
		return apierrors.NewInternalError(fmt.Errorf("failed to list children: %w", err))
	}

	if len(list.Items) > 0 {
		return apierrors.NewForbidden(schema.GroupResource{
			Group:    networkv1alpha1.IPAMRangeGK.Group,
			Resource: networkv1alpha1.IPAMRangeGK.Kind,
		}, r.Name, fmt.Errorf("there's still children that depend on this IPAMRange"))
	}

	return nil
}

// overlappingParentExists checks if any other parent IPAM range is already overlapping passed CIDRs
func (v *IPAMRangeValidator) overlappingParentExists(ctx context.Context, req *networkv1alpha1.IPAMRange) (errs field.ErrorList) {
	list := &networkv1alpha1.IPAMRangeList{}
	path := field.NewPath("spec").Child("cidrs")

	if err := v.List(ctx, list, client.InNamespace(req.Namespace)); err != nil {
		errs = append(errs, field.InternalError(nil, fmt.Errorf("failed to list existing IPAMRanges: %w", err)))
		return
	}

	for _, i := range list.Items {
		if i.Spec.Parent == nil && i.Name != req.Name {
		loop:
			for _, cidr := range req.Spec.CIDRs {
				for _, otherCidr := range i.Spec.CIDRs {
					if cidr.Overlaps(otherCidr.IPPrefix) {
						errs = append(errs, field.Duplicate(path, cidr.String()))
						continue loop
					}
				}
			}
		}
	}

	return
}

// overlappingRequestExist checks if any other IPAM range request overlaps CIDR or IP range of passed request
func (v *IPAMRangeValidator) overlappingRequestExist(ctx context.Context, req *networkv1alpha1.IPAMRange) (errs field.ErrorList) {
	list := &networkv1alpha1.IPAMRangeList{}

	if err := v.List(
		ctx, list, client.InNamespace(req.Namespace),
		client.MatchingFields{parentField: req.Spec.Parent.Name},
	); err != nil {
		errs = append(errs, field.InternalError(nil, fmt.Errorf("failed to list existing IPAMRanges: %w", err)))
		return
	}

	for _, i := range list.Items {
		if i.Name != req.Name {
		loop:
			for _, req := range req.Spec.Requests {
				for _, otherReq := range i.Spec.Requests {
					if err := checkRequestOverlap(req, otherReq); err != nil {
						errs = append(errs, err)
						continue loop
					}
				}
			}
		}
	}

	return
}

// checkRequestOverlap checks if requests overlap by CIDR or IP range
func checkRequestOverlap(req, otherReq networkv1alpha1.IPAMRangeRequest) *field.Error {
	path := field.NewPath("spec").Child("requests")

	if req.CIDR != nil {
		subpath := path.Child("cidr")
		if otherReq.CIDR != nil && req.CIDR.Overlaps(otherReq.CIDR.IPPrefix) {
			return field.Duplicate(subpath, req.CIDR.String())
		} else if otherReq.IPs != nil && req.CIDR.Range().Overlaps(otherReq.IPs.Range()) {
			return field.Duplicate(subpath, req.CIDR.String())
		}
	} else if req.IPs != nil {
		subpath := path.Child("ips")
		if otherReq.CIDR != nil && otherReq.CIDR.Range().Overlaps(req.IPs.Range()) {
			return field.Duplicate(subpath, req.IPs.Range().String())
		}
		if otherReq.IPs != nil && otherReq.IPs.Range().Overlaps(req.IPs.Range()) {
			return field.Duplicate(subpath, req.IPs.Range().String())
		}
	}

	return nil
}
