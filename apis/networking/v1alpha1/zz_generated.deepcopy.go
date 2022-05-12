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
	ipamv1alpha1 "github.com/onmetal/onmetal-api/apis/ipam/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AliasPrefix) DeepCopyInto(out *AliasPrefix) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AliasPrefix.
func (in *AliasPrefix) DeepCopy() *AliasPrefix {
	if in == nil {
		return nil
	}
	out := new(AliasPrefix)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AliasPrefix) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AliasPrefixList) DeepCopyInto(out *AliasPrefixList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AliasPrefix, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AliasPrefixList.
func (in *AliasPrefixList) DeepCopy() *AliasPrefixList {
	if in == nil {
		return nil
	}
	out := new(AliasPrefixList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AliasPrefixList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AliasPrefixRouting) DeepCopyInto(out *AliasPrefixRouting) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	if in.Destinations != nil {
		in, out := &in.Destinations, &out.Destinations
		*out = make([]commonv1alpha1.LocalUIDReference, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AliasPrefixRouting.
func (in *AliasPrefixRouting) DeepCopy() *AliasPrefixRouting {
	if in == nil {
		return nil
	}
	out := new(AliasPrefixRouting)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AliasPrefixRouting) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AliasPrefixRoutingList) DeepCopyInto(out *AliasPrefixRoutingList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AliasPrefixRouting, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AliasPrefixRoutingList.
func (in *AliasPrefixRoutingList) DeepCopy() *AliasPrefixRoutingList {
	if in == nil {
		return nil
	}
	out := new(AliasPrefixRoutingList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AliasPrefixRoutingList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AliasPrefixSpec) DeepCopyInto(out *AliasPrefixSpec) {
	*out = *in
	out.NetworkRef = in.NetworkRef
	if in.NetworkInterfaceSelector != nil {
		in, out := &in.NetworkInterfaceSelector, &out.NetworkInterfaceSelector
		*out = new(v1.LabelSelector)
		(*in).DeepCopyInto(*out)
	}
	in.Prefix.DeepCopyInto(&out.Prefix)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AliasPrefixSpec.
func (in *AliasPrefixSpec) DeepCopy() *AliasPrefixSpec {
	if in == nil {
		return nil
	}
	out := new(AliasPrefixSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AliasPrefixStatus) DeepCopyInto(out *AliasPrefixStatus) {
	*out = *in
	if in.Prefix != nil {
		in, out := &in.Prefix, &out.Prefix
		*out = (*in).DeepCopy()
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AliasPrefixStatus.
func (in *AliasPrefixStatus) DeepCopy() *AliasPrefixStatus {
	if in == nil {
		return nil
	}
	out := new(AliasPrefixStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EphemeralPrefixSource) DeepCopyInto(out *EphemeralPrefixSource) {
	*out = *in
	if in.PrefixTemplate != nil {
		in, out := &in.PrefixTemplate, &out.PrefixTemplate
		*out = new(ipamv1alpha1.PrefixTemplateSpec)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EphemeralPrefixSource.
func (in *EphemeralPrefixSource) DeepCopy() *EphemeralPrefixSource {
	if in == nil {
		return nil
	}
	out := new(EphemeralPrefixSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EphemeralVirtualIPSource) DeepCopyInto(out *EphemeralVirtualIPSource) {
	*out = *in
	if in.VirtualIPTemplate != nil {
		in, out := &in.VirtualIPTemplate, &out.VirtualIPTemplate
		*out = new(VirtualIPTemplateSpec)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EphemeralVirtualIPSource.
func (in *EphemeralVirtualIPSource) DeepCopy() *EphemeralVirtualIPSource {
	if in == nil {
		return nil
	}
	out := new(EphemeralVirtualIPSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IPSource) DeepCopyInto(out *IPSource) {
	*out = *in
	if in.Value != nil {
		in, out := &in.Value, &out.Value
		*out = (*in).DeepCopy()
	}
	if in.Ephemeral != nil {
		in, out := &in.Ephemeral, &out.Ephemeral
		*out = new(EphemeralPrefixSource)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IPSource.
func (in *IPSource) DeepCopy() *IPSource {
	if in == nil {
		return nil
	}
	out := new(IPSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Network) DeepCopyInto(out *Network) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Network.
func (in *Network) DeepCopy() *Network {
	if in == nil {
		return nil
	}
	out := new(Network)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Network) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NetworkInterface) DeepCopyInto(out *NetworkInterface) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NetworkInterface.
func (in *NetworkInterface) DeepCopy() *NetworkInterface {
	if in == nil {
		return nil
	}
	out := new(NetworkInterface)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NetworkInterface) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NetworkInterfaceList) DeepCopyInto(out *NetworkInterfaceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]NetworkInterface, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NetworkInterfaceList.
func (in *NetworkInterfaceList) DeepCopy() *NetworkInterfaceList {
	if in == nil {
		return nil
	}
	out := new(NetworkInterfaceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NetworkInterfaceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NetworkInterfaceSpec) DeepCopyInto(out *NetworkInterfaceSpec) {
	*out = *in
	out.NetworkRef = in.NetworkRef
	if in.MachineRef != nil {
		in, out := &in.MachineRef, &out.MachineRef
		*out = new(commonv1alpha1.LocalUIDReference)
		**out = **in
	}
	if in.IPFamilies != nil {
		in, out := &in.IPFamilies, &out.IPFamilies
		*out = make([]corev1.IPFamily, len(*in))
		copy(*out, *in)
	}
	if in.IPs != nil {
		in, out := &in.IPs, &out.IPs
		*out = make([]IPSource, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.VirtualIP != nil {
		in, out := &in.VirtualIP, &out.VirtualIP
		*out = new(VirtualIPSource)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NetworkInterfaceSpec.
func (in *NetworkInterfaceSpec) DeepCopy() *NetworkInterfaceSpec {
	if in == nil {
		return nil
	}
	out := new(NetworkInterfaceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NetworkInterfaceStatus) DeepCopyInto(out *NetworkInterfaceStatus) {
	*out = *in
	if in.IPs != nil {
		in, out := &in.IPs, &out.IPs
		*out = make([]commonv1alpha1.IP, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.VirtualIP != nil {
		in, out := &in.VirtualIP, &out.VirtualIP
		*out = (*in).DeepCopy()
	}
	if in.LastPhaseTransitionTime != nil {
		in, out := &in.LastPhaseTransitionTime, &out.LastPhaseTransitionTime
		*out = (*in).DeepCopy()
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NetworkInterfaceStatus.
func (in *NetworkInterfaceStatus) DeepCopy() *NetworkInterfaceStatus {
	if in == nil {
		return nil
	}
	out := new(NetworkInterfaceStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NetworkInterfaceTemplateSpec) DeepCopyInto(out *NetworkInterfaceTemplateSpec) {
	*out = *in
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NetworkInterfaceTemplateSpec.
func (in *NetworkInterfaceTemplateSpec) DeepCopy() *NetworkInterfaceTemplateSpec {
	if in == nil {
		return nil
	}
	out := new(NetworkInterfaceTemplateSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NetworkList) DeepCopyInto(out *NetworkList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Network, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NetworkList.
func (in *NetworkList) DeepCopy() *NetworkList {
	if in == nil {
		return nil
	}
	out := new(NetworkList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NetworkList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PrefixSource) DeepCopyInto(out *PrefixSource) {
	*out = *in
	if in.Value != nil {
		in, out := &in.Value, &out.Value
		*out = (*in).DeepCopy()
	}
	if in.Ephemeral != nil {
		in, out := &in.Ephemeral, &out.Ephemeral
		*out = new(EphemeralPrefixSource)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PrefixSource.
func (in *PrefixSource) DeepCopy() *PrefixSource {
	if in == nil {
		return nil
	}
	out := new(PrefixSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VirtualIP) DeepCopyInto(out *VirtualIP) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VirtualIP.
func (in *VirtualIP) DeepCopy() *VirtualIP {
	if in == nil {
		return nil
	}
	out := new(VirtualIP)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VirtualIP) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VirtualIPList) DeepCopyInto(out *VirtualIPList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]VirtualIP, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VirtualIPList.
func (in *VirtualIPList) DeepCopy() *VirtualIPList {
	if in == nil {
		return nil
	}
	out := new(VirtualIPList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VirtualIPList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VirtualIPSource) DeepCopyInto(out *VirtualIPSource) {
	*out = *in
	if in.VirtualIPRef != nil {
		in, out := &in.VirtualIPRef, &out.VirtualIPRef
		*out = new(corev1.LocalObjectReference)
		**out = **in
	}
	if in.Ephemeral != nil {
		in, out := &in.Ephemeral, &out.Ephemeral
		*out = new(EphemeralVirtualIPSource)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VirtualIPSource.
func (in *VirtualIPSource) DeepCopy() *VirtualIPSource {
	if in == nil {
		return nil
	}
	out := new(VirtualIPSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VirtualIPSpec) DeepCopyInto(out *VirtualIPSpec) {
	*out = *in
	if in.TargetRef != nil {
		in, out := &in.TargetRef, &out.TargetRef
		*out = new(commonv1alpha1.LocalUIDReference)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VirtualIPSpec.
func (in *VirtualIPSpec) DeepCopy() *VirtualIPSpec {
	if in == nil {
		return nil
	}
	out := new(VirtualIPSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VirtualIPStatus) DeepCopyInto(out *VirtualIPStatus) {
	*out = *in
	if in.IP != nil {
		in, out := &in.IP, &out.IP
		*out = (*in).DeepCopy()
	}
	if in.LastPhaseTransitionTime != nil {
		in, out := &in.LastPhaseTransitionTime, &out.LastPhaseTransitionTime
		*out = (*in).DeepCopy()
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VirtualIPStatus.
func (in *VirtualIPStatus) DeepCopy() *VirtualIPStatus {
	if in == nil {
		return nil
	}
	out := new(VirtualIPStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VirtualIPTemplateSpec) DeepCopyInto(out *VirtualIPTemplateSpec) {
	*out = *in
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VirtualIPTemplateSpec.
func (in *VirtualIPTemplateSpec) DeepCopy() *VirtualIPTemplateSpec {
	if in == nil {
		return nil
	}
	out := new(VirtualIPTemplateSpec)
	in.DeepCopyInto(out)
	return out
}
