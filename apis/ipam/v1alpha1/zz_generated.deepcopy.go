//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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
// Code generated by deepcopy-gen. DO NOT EDIT.

package v1alpha1

import (
	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterPrefix) DeepCopyInto(out *ClusterPrefix) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterPrefix.
func (in *ClusterPrefix) DeepCopy() *ClusterPrefix {
	if in == nil {
		return nil
	}
	out := new(ClusterPrefix)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ClusterPrefix) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterPrefixAllocation) DeepCopyInto(out *ClusterPrefixAllocation) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterPrefixAllocation.
func (in *ClusterPrefixAllocation) DeepCopy() *ClusterPrefixAllocation {
	if in == nil {
		return nil
	}
	out := new(ClusterPrefixAllocation)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ClusterPrefixAllocation) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterPrefixAllocationCondition) DeepCopyInto(out *ClusterPrefixAllocationCondition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterPrefixAllocationCondition.
func (in *ClusterPrefixAllocationCondition) DeepCopy() *ClusterPrefixAllocationCondition {
	if in == nil {
		return nil
	}
	out := new(ClusterPrefixAllocationCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterPrefixAllocationList) DeepCopyInto(out *ClusterPrefixAllocationList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ClusterPrefixAllocation, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterPrefixAllocationList.
func (in *ClusterPrefixAllocationList) DeepCopy() *ClusterPrefixAllocationList {
	if in == nil {
		return nil
	}
	out := new(ClusterPrefixAllocationList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ClusterPrefixAllocationList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterPrefixAllocationRequest) DeepCopyInto(out *ClusterPrefixAllocationRequest) {
	*out = *in
	in.Prefix.DeepCopyInto(&out.Prefix)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterPrefixAllocationRequest.
func (in *ClusterPrefixAllocationRequest) DeepCopy() *ClusterPrefixAllocationRequest {
	if in == nil {
		return nil
	}
	out := new(ClusterPrefixAllocationRequest)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterPrefixAllocationResult) DeepCopyInto(out *ClusterPrefixAllocationResult) {
	*out = *in
	in.Prefix.DeepCopyInto(&out.Prefix)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterPrefixAllocationResult.
func (in *ClusterPrefixAllocationResult) DeepCopy() *ClusterPrefixAllocationResult {
	if in == nil {
		return nil
	}
	out := new(ClusterPrefixAllocationResult)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterPrefixAllocationSpec) DeepCopyInto(out *ClusterPrefixAllocationSpec) {
	*out = *in
	if in.PrefixRef != nil {
		in, out := &in.PrefixRef, &out.PrefixRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	if in.PrefixSelector != nil {
		in, out := &in.PrefixSelector, &out.PrefixSelector
		*out = new(metav1.LabelSelector)
		(*in).DeepCopyInto(*out)
	}
	in.ClusterPrefixAllocationRequest.DeepCopyInto(&out.ClusterPrefixAllocationRequest)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterPrefixAllocationSpec.
func (in *ClusterPrefixAllocationSpec) DeepCopy() *ClusterPrefixAllocationSpec {
	if in == nil {
		return nil
	}
	out := new(ClusterPrefixAllocationSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterPrefixAllocationStatus) DeepCopyInto(out *ClusterPrefixAllocationStatus) {
	*out = *in
	in.ClusterPrefixAllocationResult.DeepCopyInto(&out.ClusterPrefixAllocationResult)
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]ClusterPrefixAllocationCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterPrefixAllocationStatus.
func (in *ClusterPrefixAllocationStatus) DeepCopy() *ClusterPrefixAllocationStatus {
	if in == nil {
		return nil
	}
	out := new(ClusterPrefixAllocationStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterPrefixCondition) DeepCopyInto(out *ClusterPrefixCondition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterPrefixCondition.
func (in *ClusterPrefixCondition) DeepCopy() *ClusterPrefixCondition {
	if in == nil {
		return nil
	}
	out := new(ClusterPrefixCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterPrefixList) DeepCopyInto(out *ClusterPrefixList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ClusterPrefix, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterPrefixList.
func (in *ClusterPrefixList) DeepCopy() *ClusterPrefixList {
	if in == nil {
		return nil
	}
	out := new(ClusterPrefixList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ClusterPrefixList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterPrefixSpec) DeepCopyInto(out *ClusterPrefixSpec) {
	*out = *in
	if in.ParentRef != nil {
		in, out := &in.ParentRef, &out.ParentRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	if in.ParentSelector != nil {
		in, out := &in.ParentSelector, &out.ParentSelector
		*out = new(metav1.LabelSelector)
		(*in).DeepCopyInto(*out)
	}
	in.PrefixSpace.DeepCopyInto(&out.PrefixSpace)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterPrefixSpec.
func (in *ClusterPrefixSpec) DeepCopy() *ClusterPrefixSpec {
	if in == nil {
		return nil
	}
	out := new(ClusterPrefixSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterPrefixStatus) DeepCopyInto(out *ClusterPrefixStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]ClusterPrefixCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Available != nil {
		in, out := &in.Available, &out.Available
		*out = make([]commonv1alpha1.IPPrefix, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Reserved != nil {
		in, out := &in.Reserved, &out.Reserved
		*out = make([]commonv1alpha1.IPPrefix, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterPrefixStatus.
func (in *ClusterPrefixStatus) DeepCopy() *ClusterPrefixStatus {
	if in == nil {
		return nil
	}
	out := new(ClusterPrefixStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IP) DeepCopyInto(out *IP) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IP.
func (in *IP) DeepCopy() *IP {
	if in == nil {
		return nil
	}
	out := new(IP)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *IP) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IPCondition) DeepCopyInto(out *IPCondition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IPCondition.
func (in *IPCondition) DeepCopy() *IPCondition {
	if in == nil {
		return nil
	}
	out := new(IPCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IPList) DeepCopyInto(out *IPList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]IP, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IPList.
func (in *IPList) DeepCopy() *IPList {
	if in == nil {
		return nil
	}
	out := new(IPList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *IPList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IPSpec) DeepCopyInto(out *IPSpec) {
	*out = *in
	if in.PrefixRef != nil {
		in, out := &in.PrefixRef, &out.PrefixRef
		*out = new(PrefixReference)
		**out = **in
	}
	if in.PrefixSelector != nil {
		in, out := &in.PrefixSelector, &out.PrefixSelector
		*out = new(PrefixSelector)
		(*in).DeepCopyInto(*out)
	}
	in.IP.DeepCopyInto(&out.IP)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IPSpec.
func (in *IPSpec) DeepCopy() *IPSpec {
	if in == nil {
		return nil
	}
	out := new(IPSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IPStatus) DeepCopyInto(out *IPStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]IPCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IPStatus.
func (in *IPStatus) DeepCopy() *IPStatus {
	if in == nil {
		return nil
	}
	out := new(IPStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Prefix) DeepCopyInto(out *Prefix) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Prefix.
func (in *Prefix) DeepCopy() *Prefix {
	if in == nil {
		return nil
	}
	out := new(Prefix)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Prefix) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PrefixAllocation) DeepCopyInto(out *PrefixAllocation) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PrefixAllocation.
func (in *PrefixAllocation) DeepCopy() *PrefixAllocation {
	if in == nil {
		return nil
	}
	out := new(PrefixAllocation)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PrefixAllocation) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PrefixAllocationCondition) DeepCopyInto(out *PrefixAllocationCondition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PrefixAllocationCondition.
func (in *PrefixAllocationCondition) DeepCopy() *PrefixAllocationCondition {
	if in == nil {
		return nil
	}
	out := new(PrefixAllocationCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PrefixAllocationList) DeepCopyInto(out *PrefixAllocationList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]PrefixAllocation, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PrefixAllocationList.
func (in *PrefixAllocationList) DeepCopy() *PrefixAllocationList {
	if in == nil {
		return nil
	}
	out := new(PrefixAllocationList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PrefixAllocationList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PrefixAllocationRequest) DeepCopyInto(out *PrefixAllocationRequest) {
	*out = *in
	in.Prefix.DeepCopyInto(&out.Prefix)
	if in.Range != nil {
		in, out := &in.Range, &out.Range
		*out = new(commonv1alpha1.IPRange)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PrefixAllocationRequest.
func (in *PrefixAllocationRequest) DeepCopy() *PrefixAllocationRequest {
	if in == nil {
		return nil
	}
	out := new(PrefixAllocationRequest)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PrefixAllocationResult) DeepCopyInto(out *PrefixAllocationResult) {
	*out = *in
	in.Prefix.DeepCopyInto(&out.Prefix)
	if in.Range != nil {
		in, out := &in.Range, &out.Range
		*out = new(commonv1alpha1.IPRange)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PrefixAllocationResult.
func (in *PrefixAllocationResult) DeepCopy() *PrefixAllocationResult {
	if in == nil {
		return nil
	}
	out := new(PrefixAllocationResult)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PrefixAllocationSpec) DeepCopyInto(out *PrefixAllocationSpec) {
	*out = *in
	if in.PrefixRef != nil {
		in, out := &in.PrefixRef, &out.PrefixRef
		*out = new(PrefixReference)
		**out = **in
	}
	if in.PrefixSelector != nil {
		in, out := &in.PrefixSelector, &out.PrefixSelector
		*out = new(PrefixSelector)
		(*in).DeepCopyInto(*out)
	}
	in.PrefixAllocationRequest.DeepCopyInto(&out.PrefixAllocationRequest)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PrefixAllocationSpec.
func (in *PrefixAllocationSpec) DeepCopy() *PrefixAllocationSpec {
	if in == nil {
		return nil
	}
	out := new(PrefixAllocationSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PrefixAllocationStatus) DeepCopyInto(out *PrefixAllocationStatus) {
	*out = *in
	in.PrefixAllocationResult.DeepCopyInto(&out.PrefixAllocationResult)
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]PrefixAllocationCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PrefixAllocationStatus.
func (in *PrefixAllocationStatus) DeepCopy() *PrefixAllocationStatus {
	if in == nil {
		return nil
	}
	out := new(PrefixAllocationStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PrefixCondition) DeepCopyInto(out *PrefixCondition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PrefixCondition.
func (in *PrefixCondition) DeepCopy() *PrefixCondition {
	if in == nil {
		return nil
	}
	out := new(PrefixCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PrefixList) DeepCopyInto(out *PrefixList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Prefix, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PrefixList.
func (in *PrefixList) DeepCopy() *PrefixList {
	if in == nil {
		return nil
	}
	out := new(PrefixList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PrefixList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PrefixReference) DeepCopyInto(out *PrefixReference) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PrefixReference.
func (in *PrefixReference) DeepCopy() *PrefixReference {
	if in == nil {
		return nil
	}
	out := new(PrefixReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PrefixSelector) DeepCopyInto(out *PrefixSelector) {
	*out = *in
	in.LabelSelector.DeepCopyInto(&out.LabelSelector)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PrefixSelector.
func (in *PrefixSelector) DeepCopy() *PrefixSelector {
	if in == nil {
		return nil
	}
	out := new(PrefixSelector)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PrefixSpace) DeepCopyInto(out *PrefixSpace) {
	*out = *in
	in.Prefix.DeepCopyInto(&out.Prefix)
	if in.Reservations != nil {
		in, out := &in.Reservations, &out.Reservations
		*out = make([]commonv1alpha1.IPPrefix, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.ReservationLengths != nil {
		in, out := &in.ReservationLengths, &out.ReservationLengths
		*out = make([]int32, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PrefixSpace.
func (in *PrefixSpace) DeepCopy() *PrefixSpace {
	if in == nil {
		return nil
	}
	out := new(PrefixSpace)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PrefixSpec) DeepCopyInto(out *PrefixSpec) {
	*out = *in
	if in.ParentRef != nil {
		in, out := &in.ParentRef, &out.ParentRef
		*out = new(PrefixReference)
		**out = **in
	}
	if in.ParentSelector != nil {
		in, out := &in.ParentSelector, &out.ParentSelector
		*out = new(PrefixSelector)
		(*in).DeepCopyInto(*out)
	}
	in.PrefixSpace.DeepCopyInto(&out.PrefixSpace)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PrefixSpec.
func (in *PrefixSpec) DeepCopy() *PrefixSpec {
	if in == nil {
		return nil
	}
	out := new(PrefixSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PrefixStatus) DeepCopyInto(out *PrefixStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]PrefixCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Available != nil {
		in, out := &in.Available, &out.Available
		*out = make([]commonv1alpha1.IPPrefix, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Reserved != nil {
		in, out := &in.Reserved, &out.Reserved
		*out = make([]commonv1alpha1.IPPrefix, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PrefixStatus.
func (in *PrefixStatus) DeepCopy() *PrefixStatus {
	if in == nil {
		return nil
	}
	out := new(PrefixStatus)
	in.DeepCopyInto(out)
	return out
}
