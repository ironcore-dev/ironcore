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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EgressSecurityGroupRule) DeepCopyInto(out *EgressSecurityGroupRule) {
	*out = *in
	in.SecurityGroupRule.DeepCopyInto(&out.SecurityGroupRule)
	in.Destination.DeepCopyInto(&out.Destination)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EgressSecurityGroupRule.
func (in *EgressSecurityGroupRule) DeepCopy() *EgressSecurityGroupRule {
	if in == nil {
		return nil
	}
	out := new(EgressSecurityGroupRule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FilterRule) DeepCopyInto(out *FilterRule) {
	*out = *in
	out.SecurityGroup = in.SecurityGroup
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FilterRule.
func (in *FilterRule) DeepCopy() *FilterRule {
	if in == nil {
		return nil
	}
	out := new(FilterRule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Gateway) DeepCopyInto(out *Gateway) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Gateway.
func (in *Gateway) DeepCopy() *Gateway {
	if in == nil {
		return nil
	}
	out := new(Gateway)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Gateway) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GatewayCondition) DeepCopyInto(out *GatewayCondition) {
	*out = *in
	in.LastUpdateTime.DeepCopyInto(&out.LastUpdateTime)
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GatewayCondition.
func (in *GatewayCondition) DeepCopy() *GatewayCondition {
	if in == nil {
		return nil
	}
	out := new(GatewayCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GatewayList) DeepCopyInto(out *GatewayList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Gateway, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GatewayList.
func (in *GatewayList) DeepCopy() *GatewayList {
	if in == nil {
		return nil
	}
	out := new(GatewayList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GatewayList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GatewaySpec) DeepCopyInto(out *GatewaySpec) {
	*out = *in
	if in.FilterRules != nil {
		in, out := &in.FilterRules, &out.FilterRules
		*out = make([]FilterRule, len(*in))
		copy(*out, *in)
	}
	in.Uplink.DeepCopyInto(&out.Uplink)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GatewaySpec.
func (in *GatewaySpec) DeepCopy() *GatewaySpec {
	if in == nil {
		return nil
	}
	out := new(GatewaySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GatewayStatus) DeepCopyInto(out *GatewayStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]GatewayCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.IPs != nil {
		in, out := &in.IPs, &out.IPs
		*out = make([]commonv1alpha1.IPAddr, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GatewayStatus.
func (in *GatewayStatus) DeepCopy() *GatewayStatus {
	if in == nil {
		return nil
	}
	out := new(GatewayStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IPAMRange) DeepCopyInto(out *IPAMRange) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IPAMRange.
func (in *IPAMRange) DeepCopy() *IPAMRange {
	if in == nil {
		return nil
	}
	out := new(IPAMRange)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *IPAMRange) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IPAMRangeAllocationStatus) DeepCopyInto(out *IPAMRangeAllocationStatus) {
	*out = *in
	if in.CIDR != nil {
		in, out := &in.CIDR, &out.CIDR
		*out = (*in).DeepCopy()
	}
	if in.IPs != nil {
		in, out := &in.IPs, &out.IPs
		*out = new(commonv1alpha1.IPRange)
		(*in).DeepCopyInto(*out)
	}
	if in.Item != nil {
		in, out := &in.Item, &out.Item
		*out = new(IPAMRangeItem)
		(*in).DeepCopyInto(*out)
	}
	if in.User != nil {
		in, out := &in.User, &out.User
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IPAMRangeAllocationStatus.
func (in *IPAMRangeAllocationStatus) DeepCopy() *IPAMRangeAllocationStatus {
	if in == nil {
		return nil
	}
	out := new(IPAMRangeAllocationStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IPAMRangeItem) DeepCopyInto(out *IPAMRangeItem) {
	*out = *in
	if in.CIDR != nil {
		in, out := &in.CIDR, &out.CIDR
		*out = (*in).DeepCopy()
	}
	if in.IPs != nil {
		in, out := &in.IPs, &out.IPs
		*out = new(commonv1alpha1.IPRange)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IPAMRangeItem.
func (in *IPAMRangeItem) DeepCopy() *IPAMRangeItem {
	if in == nil {
		return nil
	}
	out := new(IPAMRangeItem)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IPAMRangeList) DeepCopyInto(out *IPAMRangeList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]IPAMRange, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IPAMRangeList.
func (in *IPAMRangeList) DeepCopy() *IPAMRangeList {
	if in == nil {
		return nil
	}
	out := new(IPAMRangeList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *IPAMRangeList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IPAMRangeSpec) DeepCopyInto(out *IPAMRangeSpec) {
	*out = *in
	if in.Parent != nil {
		in, out := &in.Parent, &out.Parent
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]IPAMRangeItem, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IPAMRangeSpec.
func (in *IPAMRangeSpec) DeepCopy() *IPAMRangeSpec {
	if in == nil {
		return nil
	}
	out := new(IPAMRangeSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IPAMRangeStatus) DeepCopyInto(out *IPAMRangeStatus) {
	*out = *in
	if in.Allocations != nil {
		in, out := &in.Allocations, &out.Allocations
		*out = make([]IPAMRangeAllocationStatus, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IPAMRangeStatus.
func (in *IPAMRangeStatus) DeepCopy() *IPAMRangeStatus {
	if in == nil {
		return nil
	}
	out := new(IPAMRangeStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IPSetSpec) DeepCopyInto(out *IPSetSpec) {
	*out = *in
	in.CIDR.DeepCopyInto(&out.CIDR)
	out.SecurityGroupRef = in.SecurityGroupRef
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IPSetSpec.
func (in *IPSetSpec) DeepCopy() *IPSetSpec {
	if in == nil {
		return nil
	}
	out := new(IPSetSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IngressSecurityGroupRule) DeepCopyInto(out *IngressSecurityGroupRule) {
	*out = *in
	in.SecurityGroupRule.DeepCopyInto(&out.SecurityGroupRule)
	in.Source.DeepCopyInto(&out.Source)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IngressSecurityGroupRule.
func (in *IngressSecurityGroupRule) DeepCopy() *IngressSecurityGroupRule {
	if in == nil {
		return nil
	}
	out := new(IngressSecurityGroupRule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MachineRouteTarget) DeepCopyInto(out *MachineRouteTarget) {
	*out = *in
	out.LocalObjectReference = in.LocalObjectReference
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MachineRouteTarget.
func (in *MachineRouteTarget) DeepCopy() *MachineRouteTarget {
	if in == nil {
		return nil
	}
	out := new(MachineRouteTarget)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PortRange) DeepCopyInto(out *PortRange) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PortRange.
func (in *PortRange) DeepCopy() *PortRange {
	if in == nil {
		return nil
	}
	out := new(PortRange)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RangeType) DeepCopyInto(out *RangeType) {
	*out = *in
	if in.CIDR != nil {
		in, out := &in.CIDR, &out.CIDR
		*out = (*in).DeepCopy()
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RangeType.
func (in *RangeType) DeepCopy() *RangeType {
	if in == nil {
		return nil
	}
	out := new(RangeType)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReservedIP) DeepCopyInto(out *ReservedIP) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReservedIP.
func (in *ReservedIP) DeepCopy() *ReservedIP {
	if in == nil {
		return nil
	}
	out := new(ReservedIP)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ReservedIP) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReservedIPAssignment) DeepCopyInto(out *ReservedIPAssignment) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReservedIPAssignment.
func (in *ReservedIPAssignment) DeepCopy() *ReservedIPAssignment {
	if in == nil {
		return nil
	}
	out := new(ReservedIPAssignment)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReservedIPBound) DeepCopyInto(out *ReservedIPBound) {
	*out = *in
	out.Assignment = in.Assignment
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReservedIPBound.
func (in *ReservedIPBound) DeepCopy() *ReservedIPBound {
	if in == nil {
		return nil
	}
	out := new(ReservedIPBound)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReservedIPCondition) DeepCopyInto(out *ReservedIPCondition) {
	*out = *in
	in.LastUpdateTime.DeepCopyInto(&out.LastUpdateTime)
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReservedIPCondition.
func (in *ReservedIPCondition) DeepCopy() *ReservedIPCondition {
	if in == nil {
		return nil
	}
	out := new(ReservedIPCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReservedIPList) DeepCopyInto(out *ReservedIPList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ReservedIP, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReservedIPList.
func (in *ReservedIPList) DeepCopy() *ReservedIPList {
	if in == nil {
		return nil
	}
	out := new(ReservedIPList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ReservedIPList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReservedIPSpec) DeepCopyInto(out *ReservedIPSpec) {
	*out = *in
	out.Subnet = in.Subnet
	if in.IP != nil {
		in, out := &in.IP, &out.IP
		*out = (*in).DeepCopy()
	}
	out.Assignment = in.Assignment
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReservedIPSpec.
func (in *ReservedIPSpec) DeepCopy() *ReservedIPSpec {
	if in == nil {
		return nil
	}
	out := new(ReservedIPSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReservedIPStatus) DeepCopyInto(out *ReservedIPStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]ReservedIPCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.IP != nil {
		in, out := &in.IP, &out.IP
		*out = (*in).DeepCopy()
	}
	if in.Bound != nil {
		in, out := &in.Bound, &out.Bound
		*out = new(ReservedIPBound)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReservedIPStatus.
func (in *ReservedIPStatus) DeepCopy() *ReservedIPStatus {
	if in == nil {
		return nil
	}
	out := new(ReservedIPStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Route) DeepCopyInto(out *Route) {
	*out = *in
	out.SubnetRef = in.SubnetRef
	in.Target.DeepCopyInto(&out.Target)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Route.
func (in *Route) DeepCopy() *Route {
	if in == nil {
		return nil
	}
	out := new(Route)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RoutingDomain) DeepCopyInto(out *RoutingDomain) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RoutingDomain.
func (in *RoutingDomain) DeepCopy() *RoutingDomain {
	if in == nil {
		return nil
	}
	out := new(RoutingDomain)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RoutingDomain) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RoutingDomainCondition) DeepCopyInto(out *RoutingDomainCondition) {
	*out = *in
	in.LastUpdateTime.DeepCopyInto(&out.LastUpdateTime)
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RoutingDomainCondition.
func (in *RoutingDomainCondition) DeepCopy() *RoutingDomainCondition {
	if in == nil {
		return nil
	}
	out := new(RoutingDomainCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RoutingDomainList) DeepCopyInto(out *RoutingDomainList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]RoutingDomain, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RoutingDomainList.
func (in *RoutingDomainList) DeepCopy() *RoutingDomainList {
	if in == nil {
		return nil
	}
	out := new(RoutingDomainList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RoutingDomainList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RoutingDomainSpec) DeepCopyInto(out *RoutingDomainSpec) {
	*out = *in
	if in.Routes != nil {
		in, out := &in.Routes, &out.Routes
		*out = make([]Route, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RoutingDomainSpec.
func (in *RoutingDomainSpec) DeepCopy() *RoutingDomainSpec {
	if in == nil {
		return nil
	}
	out := new(RoutingDomainSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RoutingDomainStatus) DeepCopyInto(out *RoutingDomainStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]RoutingDomainCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RoutingDomainStatus.
func (in *RoutingDomainStatus) DeepCopy() *RoutingDomainStatus {
	if in == nil {
		return nil
	}
	out := new(RoutingDomainStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecurityGroup) DeepCopyInto(out *SecurityGroup) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecurityGroup.
func (in *SecurityGroup) DeepCopy() *SecurityGroup {
	if in == nil {
		return nil
	}
	out := new(SecurityGroup)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SecurityGroup) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecurityGroupCondition) DeepCopyInto(out *SecurityGroupCondition) {
	*out = *in
	in.LastUpdateTime.DeepCopyInto(&out.LastUpdateTime)
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecurityGroupCondition.
func (in *SecurityGroupCondition) DeepCopy() *SecurityGroupCondition {
	if in == nil {
		return nil
	}
	out := new(SecurityGroupCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecurityGroupList) DeepCopyInto(out *SecurityGroupList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]SecurityGroup, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecurityGroupList.
func (in *SecurityGroupList) DeepCopy() *SecurityGroupList {
	if in == nil {
		return nil
	}
	out := new(SecurityGroupList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SecurityGroupList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecurityGroupRule) DeepCopyInto(out *SecurityGroupRule) {
	*out = *in
	out.SecurityGroupRef = in.SecurityGroupRef
	if in.PortRange != nil {
		in, out := &in.PortRange, &out.PortRange
		*out = new(PortRange)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecurityGroupRule.
func (in *SecurityGroupRule) DeepCopy() *SecurityGroupRule {
	if in == nil {
		return nil
	}
	out := new(SecurityGroupRule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecurityGroupSpec) DeepCopyInto(out *SecurityGroupSpec) {
	*out = *in
	if in.Ingress != nil {
		in, out := &in.Ingress, &out.Ingress
		*out = make([]IngressSecurityGroupRule, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Egress != nil {
		in, out := &in.Egress, &out.Egress
		*out = make([]EgressSecurityGroupRule, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecurityGroupSpec.
func (in *SecurityGroupSpec) DeepCopy() *SecurityGroupSpec {
	if in == nil {
		return nil
	}
	out := new(SecurityGroupSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecurityGroupStatus) DeepCopyInto(out *SecurityGroupStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]SecurityGroupCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecurityGroupStatus.
func (in *SecurityGroupStatus) DeepCopy() *SecurityGroupStatus {
	if in == nil {
		return nil
	}
	out := new(SecurityGroupStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Subnet) DeepCopyInto(out *Subnet) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Subnet.
func (in *Subnet) DeepCopy() *Subnet {
	if in == nil {
		return nil
	}
	out := new(Subnet)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Subnet) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SubnetCondition) DeepCopyInto(out *SubnetCondition) {
	*out = *in
	in.LastUpdateTime.DeepCopyInto(&out.LastUpdateTime)
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SubnetCondition.
func (in *SubnetCondition) DeepCopy() *SubnetCondition {
	if in == nil {
		return nil
	}
	out := new(SubnetCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SubnetList) DeepCopyInto(out *SubnetList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Subnet, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SubnetList.
func (in *SubnetList) DeepCopy() *SubnetList {
	if in == nil {
		return nil
	}
	out := new(SubnetList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SubnetList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SubnetSpec) DeepCopyInto(out *SubnetSpec) {
	*out = *in
	if in.Parent != nil {
		in, out := &in.Parent, &out.Parent
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	if in.MachinePools != nil {
		in, out := &in.MachinePools, &out.MachinePools
		*out = make([]v1.LocalObjectReference, len(*in))
		copy(*out, *in)
	}
	out.RoutingDomain = in.RoutingDomain
	if in.Ranges != nil {
		in, out := &in.Ranges, &out.Ranges
		*out = make([]RangeType, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SubnetSpec.
func (in *SubnetSpec) DeepCopy() *SubnetSpec {
	if in == nil {
		return nil
	}
	out := new(SubnetSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SubnetStatus) DeepCopyInto(out *SubnetStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]SubnetCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.CIDRs != nil {
		in, out := &in.CIDRs, &out.CIDRs
		*out = make([]commonv1alpha1.CIDR, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SubnetStatus.
func (in *SubnetStatus) DeepCopy() *SubnetStatus {
	if in == nil {
		return nil
	}
	out := new(SubnetStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Target) DeepCopyInto(out *Target) {
	*out = *in
	if in.Machine != nil {
		in, out := &in.Machine, &out.Machine
		*out = new(MachineRouteTarget)
		**out = **in
	}
	if in.Gateway != nil {
		in, out := &in.Gateway, &out.Gateway
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	if in.ReservedIP != nil {
		in, out := &in.ReservedIP, &out.ReservedIP
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Target.
func (in *Target) DeepCopy() *Target {
	if in == nil {
		return nil
	}
	out := new(Target)
	in.DeepCopyInto(out)
	return out
}
