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
	"fmt"
	"reflect"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
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
		if len(r.Spec.CIDRs) > 0 {
			allErrs = append(allErrs, field.Forbidden(path.Child("cidrs"), "CIDRs should be empty for child IPAMRange"))
		}
		allErrs = append(allErrs, r.overlappingRequestExist()...)
	} else {
		allErrs = append(allErrs, r.overlappingParentExists()...)
	}

	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(IPAMRangeGK, r.Name, allErrs)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *IPAMRange) ValidateUpdate(oldObj runtime.Object) error {
	ipamrangelog.Info("validate update", "name", r.Name)

	oldRange := oldObj.(*IPAMRange)
	path := field.NewPath("spec")

	var allErrs field.ErrorList
	if r.Spec.Parent != nil {
		if !reflect.DeepEqual(r.Spec.Parent, oldRange.Spec.Parent) {
			allErrs = append(allErrs, field.Invalid(path.Child("parent"), r.Spec.Parent, fieldImmutable))
		}
		allErrs = append(allErrs, r.overlappingRequestExist()...)
	} else {
		allErrs = append(allErrs, r.overlappingParentExists()...)
		allErrs = append(allErrs, r.deletedCIDRsUsed(oldRange)...)
	}

	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(IPAMRangeGK, r.Name, allErrs)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *IPAMRange) ValidateDelete() error {
	ipamrangelog.Info("validate delete", "name", r.Name)

	for _, alloc := range r.Status.Allocations {
		if alloc.State == IPAMRangeAllocationUsed {
			return apierrors.NewForbidden(schema.GroupResource{
				Group:    IPAMRangeGK.Group,
				Resource: IPAMRangeGK.Kind,
			}, r.Name, fmt.Errorf("there's still children that depend on this IPAMRange"))
		}
	}

	return nil
}

// overlappingParentExists checks if any other parent IPAM range is already overlapping passed CIDRs
func (r *IPAMRange) overlappingParentExists() (errs field.ErrorList) {
	path := field.NewPath("spec").Child("cidrs")

	for i, cidr1 := range r.Spec.CIDRs {
	loop:
		for _, cidr2 := range r.Spec.CIDRs[i+1:] {
			if cidr1.Overlaps(cidr2.IPPrefix) {
				errs = append(errs, field.Duplicate(path, cidr2.String()))
				continue loop
			}
		}
	}

	return
}

// deletedCIDRsUsed checks if some of deleted CIDRs are used by children requests
func (r *IPAMRange) deletedCIDRsUsed(oldReq *IPAMRange) (errs field.ErrorList) {
	deletedCIDRs := []commonv1alpha1.CIDR{}
loop:
	for _, cidr := range oldReq.Spec.CIDRs {
		for _, newCidr := range r.Spec.CIDRs {
			if cidr == newCidr {
				continue loop
			}
		}

		deletedCIDRs = append(deletedCIDRs, cidr)
	}
	if len(deletedCIDRs) == 0 {
		return
	}

	path := field.NewPath("spec").Child("cidrs")
	for _, alloc := range oldReq.Status.Allocations {
		if alloc.State == IPAMRangeAllocationUsed {
			for _, cidr := range deletedCIDRs {
				if alloc.IPs != nil && alloc.IPs.Range().Overlaps(cidr.Range()) {
					errs = append(errs, field.Forbidden(path, fmt.Sprintf("CIDR %s is used by child request", cidr.String())))
				} else if alloc.CIDR != nil && alloc.CIDR.Overlaps(cidr.IPPrefix) {
					errs = append(errs, field.Forbidden(path, fmt.Sprintf("CIDR %s is used by child request", cidr.String())))
				}
			}
		}
	}

	return
}

// overlappingRequestExist checks if any other IPAM range request overlaps CIDR or IP range of passed request
func (r *IPAMRange) overlappingRequestExist() (errs field.ErrorList) {
	for i, req1 := range r.Spec.Requests {
	loop:
		for _, req2 := range r.Spec.Requests[i+1:] {
			if err := checkRequestOverlap(req1, req2); err != nil {
				errs = append(errs, err)
				continue loop
			}
		}
	}

	return
}

// checkRequestOverlap checks if requests overlap by CIDR or IP range
func checkRequestOverlap(req1, req2 IPAMRangeRequest) *field.Error {
	path := field.NewPath("spec").Child("requests")

	if req1.CIDR != nil {
		subpath := path.Child("cidr")
		if req2.CIDR != nil && req1.CIDR.Overlaps(req2.CIDR.IPPrefix) {
			return field.Duplicate(subpath, req2.CIDR.String())
		} else if req2.IPs != nil && req1.CIDR.Range().Overlaps(req2.IPs.Range()) {
			return field.Duplicate(subpath, req1.CIDR.String())
		}
	} else if req1.IPs != nil {
		subpath := path.Child("ips")
		if req2.CIDR != nil && req2.CIDR.Range().Overlaps(req1.IPs.Range()) {
			return field.Duplicate(subpath, req2.IPs.Range().String())
		}
		if req2.IPs != nil && req2.IPs.Range().Overlaps(req1.IPs.Range()) {
			return field.Duplicate(subpath, req2.IPs.Range().String())
		}
	}

	return nil
}
