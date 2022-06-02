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
// Code generated by conversion-gen. DO NOT EDIT.

package v1alpha1

import (
	unsafe "unsafe"

	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	storage "github.com/onmetal/onmetal-api/apis/storage"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	conversion "k8s.io/apimachinery/pkg/conversion"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

func init() {
	localSchemeBuilder.Register(RegisterConversions)
}

// RegisterConversions adds conversion functions to the given scheme.
// Public to allow building arbitrary schemes.
func RegisterConversions(s *runtime.Scheme) error {
	if err := s.AddGeneratedConversionFunc((*Volume)(nil), (*storage.Volume)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_Volume_To_storage_Volume(a.(*Volume), b.(*storage.Volume), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*storage.Volume)(nil), (*Volume)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_storage_Volume_To_v1alpha1_Volume(a.(*storage.Volume), b.(*Volume), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*VolumeAccess)(nil), (*storage.VolumeAccess)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_VolumeAccess_To_storage_VolumeAccess(a.(*VolumeAccess), b.(*storage.VolumeAccess), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*storage.VolumeAccess)(nil), (*VolumeAccess)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_storage_VolumeAccess_To_v1alpha1_VolumeAccess(a.(*storage.VolumeAccess), b.(*VolumeAccess), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*VolumeClass)(nil), (*storage.VolumeClass)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_VolumeClass_To_storage_VolumeClass(a.(*VolumeClass), b.(*storage.VolumeClass), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*storage.VolumeClass)(nil), (*VolumeClass)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_storage_VolumeClass_To_v1alpha1_VolumeClass(a.(*storage.VolumeClass), b.(*VolumeClass), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*VolumeClassList)(nil), (*storage.VolumeClassList)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_VolumeClassList_To_storage_VolumeClassList(a.(*VolumeClassList), b.(*storage.VolumeClassList), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*storage.VolumeClassList)(nil), (*VolumeClassList)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_storage_VolumeClassList_To_v1alpha1_VolumeClassList(a.(*storage.VolumeClassList), b.(*VolumeClassList), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*VolumeCondition)(nil), (*storage.VolumeCondition)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_VolumeCondition_To_storage_VolumeCondition(a.(*VolumeCondition), b.(*storage.VolumeCondition), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*storage.VolumeCondition)(nil), (*VolumeCondition)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_storage_VolumeCondition_To_v1alpha1_VolumeCondition(a.(*storage.VolumeCondition), b.(*VolumeCondition), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*VolumeList)(nil), (*storage.VolumeList)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_VolumeList_To_storage_VolumeList(a.(*VolumeList), b.(*storage.VolumeList), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*storage.VolumeList)(nil), (*VolumeList)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_storage_VolumeList_To_v1alpha1_VolumeList(a.(*storage.VolumeList), b.(*VolumeList), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*VolumePool)(nil), (*storage.VolumePool)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_VolumePool_To_storage_VolumePool(a.(*VolumePool), b.(*storage.VolumePool), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*storage.VolumePool)(nil), (*VolumePool)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_storage_VolumePool_To_v1alpha1_VolumePool(a.(*storage.VolumePool), b.(*VolumePool), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*VolumePoolCondition)(nil), (*storage.VolumePoolCondition)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_VolumePoolCondition_To_storage_VolumePoolCondition(a.(*VolumePoolCondition), b.(*storage.VolumePoolCondition), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*storage.VolumePoolCondition)(nil), (*VolumePoolCondition)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_storage_VolumePoolCondition_To_v1alpha1_VolumePoolCondition(a.(*storage.VolumePoolCondition), b.(*VolumePoolCondition), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*VolumePoolList)(nil), (*storage.VolumePoolList)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_VolumePoolList_To_storage_VolumePoolList(a.(*VolumePoolList), b.(*storage.VolumePoolList), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*storage.VolumePoolList)(nil), (*VolumePoolList)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_storage_VolumePoolList_To_v1alpha1_VolumePoolList(a.(*storage.VolumePoolList), b.(*VolumePoolList), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*VolumePoolSpec)(nil), (*storage.VolumePoolSpec)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_VolumePoolSpec_To_storage_VolumePoolSpec(a.(*VolumePoolSpec), b.(*storage.VolumePoolSpec), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*storage.VolumePoolSpec)(nil), (*VolumePoolSpec)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_storage_VolumePoolSpec_To_v1alpha1_VolumePoolSpec(a.(*storage.VolumePoolSpec), b.(*VolumePoolSpec), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*VolumePoolStatus)(nil), (*storage.VolumePoolStatus)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_VolumePoolStatus_To_storage_VolumePoolStatus(a.(*VolumePoolStatus), b.(*storage.VolumePoolStatus), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*storage.VolumePoolStatus)(nil), (*VolumePoolStatus)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_storage_VolumePoolStatus_To_v1alpha1_VolumePoolStatus(a.(*storage.VolumePoolStatus), b.(*VolumePoolStatus), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*VolumeSpec)(nil), (*storage.VolumeSpec)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_VolumeSpec_To_storage_VolumeSpec(a.(*VolumeSpec), b.(*storage.VolumeSpec), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*storage.VolumeSpec)(nil), (*VolumeSpec)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_storage_VolumeSpec_To_v1alpha1_VolumeSpec(a.(*storage.VolumeSpec), b.(*VolumeSpec), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*VolumeStatus)(nil), (*storage.VolumeStatus)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_VolumeStatus_To_storage_VolumeStatus(a.(*VolumeStatus), b.(*storage.VolumeStatus), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*storage.VolumeStatus)(nil), (*VolumeStatus)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_storage_VolumeStatus_To_v1alpha1_VolumeStatus(a.(*storage.VolumeStatus), b.(*VolumeStatus), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*VolumeTemplateSpec)(nil), (*storage.VolumeTemplateSpec)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_VolumeTemplateSpec_To_storage_VolumeTemplateSpec(a.(*VolumeTemplateSpec), b.(*storage.VolumeTemplateSpec), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*storage.VolumeTemplateSpec)(nil), (*VolumeTemplateSpec)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_storage_VolumeTemplateSpec_To_v1alpha1_VolumeTemplateSpec(a.(*storage.VolumeTemplateSpec), b.(*VolumeTemplateSpec), scope)
	}); err != nil {
		return err
	}
	return nil
}

func autoConvert_v1alpha1_Volume_To_storage_Volume(in *Volume, out *storage.Volume, s conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	if err := Convert_v1alpha1_VolumeSpec_To_storage_VolumeSpec(&in.Spec, &out.Spec, s); err != nil {
		return err
	}
	if err := Convert_v1alpha1_VolumeStatus_To_storage_VolumeStatus(&in.Status, &out.Status, s); err != nil {
		return err
	}
	return nil
}

// Convert_v1alpha1_Volume_To_storage_Volume is an autogenerated conversion function.
func Convert_v1alpha1_Volume_To_storage_Volume(in *Volume, out *storage.Volume, s conversion.Scope) error {
	return autoConvert_v1alpha1_Volume_To_storage_Volume(in, out, s)
}

func autoConvert_storage_Volume_To_v1alpha1_Volume(in *storage.Volume, out *Volume, s conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	if err := Convert_storage_VolumeSpec_To_v1alpha1_VolumeSpec(&in.Spec, &out.Spec, s); err != nil {
		return err
	}
	if err := Convert_storage_VolumeStatus_To_v1alpha1_VolumeStatus(&in.Status, &out.Status, s); err != nil {
		return err
	}
	return nil
}

// Convert_storage_Volume_To_v1alpha1_Volume is an autogenerated conversion function.
func Convert_storage_Volume_To_v1alpha1_Volume(in *storage.Volume, out *Volume, s conversion.Scope) error {
	return autoConvert_storage_Volume_To_v1alpha1_Volume(in, out, s)
}

func autoConvert_v1alpha1_VolumeAccess_To_storage_VolumeAccess(in *VolumeAccess, out *storage.VolumeAccess, s conversion.Scope) error {
	out.SecretRef = (*v1.LocalObjectReference)(unsafe.Pointer(in.SecretRef))
	out.Driver = in.Driver
	out.VolumeAttributes = *(*map[string]string)(unsafe.Pointer(&in.VolumeAttributes))
	return nil
}

// Convert_v1alpha1_VolumeAccess_To_storage_VolumeAccess is an autogenerated conversion function.
func Convert_v1alpha1_VolumeAccess_To_storage_VolumeAccess(in *VolumeAccess, out *storage.VolumeAccess, s conversion.Scope) error {
	return autoConvert_v1alpha1_VolumeAccess_To_storage_VolumeAccess(in, out, s)
}

func autoConvert_storage_VolumeAccess_To_v1alpha1_VolumeAccess(in *storage.VolumeAccess, out *VolumeAccess, s conversion.Scope) error {
	out.SecretRef = (*v1.LocalObjectReference)(unsafe.Pointer(in.SecretRef))
	out.Driver = in.Driver
	out.VolumeAttributes = *(*map[string]string)(unsafe.Pointer(&in.VolumeAttributes))
	return nil
}

// Convert_storage_VolumeAccess_To_v1alpha1_VolumeAccess is an autogenerated conversion function.
func Convert_storage_VolumeAccess_To_v1alpha1_VolumeAccess(in *storage.VolumeAccess, out *VolumeAccess, s conversion.Scope) error {
	return autoConvert_storage_VolumeAccess_To_v1alpha1_VolumeAccess(in, out, s)
}

func autoConvert_v1alpha1_VolumeClass_To_storage_VolumeClass(in *VolumeClass, out *storage.VolumeClass, s conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	out.Capabilities = *(*v1.ResourceList)(unsafe.Pointer(&in.Capabilities))
	return nil
}

// Convert_v1alpha1_VolumeClass_To_storage_VolumeClass is an autogenerated conversion function.
func Convert_v1alpha1_VolumeClass_To_storage_VolumeClass(in *VolumeClass, out *storage.VolumeClass, s conversion.Scope) error {
	return autoConvert_v1alpha1_VolumeClass_To_storage_VolumeClass(in, out, s)
}

func autoConvert_storage_VolumeClass_To_v1alpha1_VolumeClass(in *storage.VolumeClass, out *VolumeClass, s conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	out.Capabilities = *(*v1.ResourceList)(unsafe.Pointer(&in.Capabilities))
	return nil
}

// Convert_storage_VolumeClass_To_v1alpha1_VolumeClass is an autogenerated conversion function.
func Convert_storage_VolumeClass_To_v1alpha1_VolumeClass(in *storage.VolumeClass, out *VolumeClass, s conversion.Scope) error {
	return autoConvert_storage_VolumeClass_To_v1alpha1_VolumeClass(in, out, s)
}

func autoConvert_v1alpha1_VolumeClassList_To_storage_VolumeClassList(in *VolumeClassList, out *storage.VolumeClassList, s conversion.Scope) error {
	out.ListMeta = in.ListMeta
	out.Items = *(*[]storage.VolumeClass)(unsafe.Pointer(&in.Items))
	return nil
}

// Convert_v1alpha1_VolumeClassList_To_storage_VolumeClassList is an autogenerated conversion function.
func Convert_v1alpha1_VolumeClassList_To_storage_VolumeClassList(in *VolumeClassList, out *storage.VolumeClassList, s conversion.Scope) error {
	return autoConvert_v1alpha1_VolumeClassList_To_storage_VolumeClassList(in, out, s)
}

func autoConvert_storage_VolumeClassList_To_v1alpha1_VolumeClassList(in *storage.VolumeClassList, out *VolumeClassList, s conversion.Scope) error {
	out.ListMeta = in.ListMeta
	out.Items = *(*[]VolumeClass)(unsafe.Pointer(&in.Items))
	return nil
}

// Convert_storage_VolumeClassList_To_v1alpha1_VolumeClassList is an autogenerated conversion function.
func Convert_storage_VolumeClassList_To_v1alpha1_VolumeClassList(in *storage.VolumeClassList, out *VolumeClassList, s conversion.Scope) error {
	return autoConvert_storage_VolumeClassList_To_v1alpha1_VolumeClassList(in, out, s)
}

func autoConvert_v1alpha1_VolumeCondition_To_storage_VolumeCondition(in *VolumeCondition, out *storage.VolumeCondition, s conversion.Scope) error {
	out.Type = storage.VolumeConditionType(in.Type)
	out.Status = v1.ConditionStatus(in.Status)
	out.Reason = in.Reason
	out.Message = in.Message
	out.ObservedGeneration = in.ObservedGeneration
	out.LastTransitionTime = in.LastTransitionTime
	return nil
}

// Convert_v1alpha1_VolumeCondition_To_storage_VolumeCondition is an autogenerated conversion function.
func Convert_v1alpha1_VolumeCondition_To_storage_VolumeCondition(in *VolumeCondition, out *storage.VolumeCondition, s conversion.Scope) error {
	return autoConvert_v1alpha1_VolumeCondition_To_storage_VolumeCondition(in, out, s)
}

func autoConvert_storage_VolumeCondition_To_v1alpha1_VolumeCondition(in *storage.VolumeCondition, out *VolumeCondition, s conversion.Scope) error {
	out.Type = VolumeConditionType(in.Type)
	out.Status = v1.ConditionStatus(in.Status)
	out.Reason = in.Reason
	out.Message = in.Message
	out.ObservedGeneration = in.ObservedGeneration
	out.LastTransitionTime = in.LastTransitionTime
	return nil
}

// Convert_storage_VolumeCondition_To_v1alpha1_VolumeCondition is an autogenerated conversion function.
func Convert_storage_VolumeCondition_To_v1alpha1_VolumeCondition(in *storage.VolumeCondition, out *VolumeCondition, s conversion.Scope) error {
	return autoConvert_storage_VolumeCondition_To_v1alpha1_VolumeCondition(in, out, s)
}

func autoConvert_v1alpha1_VolumeList_To_storage_VolumeList(in *VolumeList, out *storage.VolumeList, s conversion.Scope) error {
	out.ListMeta = in.ListMeta
	out.Items = *(*[]storage.Volume)(unsafe.Pointer(&in.Items))
	return nil
}

// Convert_v1alpha1_VolumeList_To_storage_VolumeList is an autogenerated conversion function.
func Convert_v1alpha1_VolumeList_To_storage_VolumeList(in *VolumeList, out *storage.VolumeList, s conversion.Scope) error {
	return autoConvert_v1alpha1_VolumeList_To_storage_VolumeList(in, out, s)
}

func autoConvert_storage_VolumeList_To_v1alpha1_VolumeList(in *storage.VolumeList, out *VolumeList, s conversion.Scope) error {
	out.ListMeta = in.ListMeta
	out.Items = *(*[]Volume)(unsafe.Pointer(&in.Items))
	return nil
}

// Convert_storage_VolumeList_To_v1alpha1_VolumeList is an autogenerated conversion function.
func Convert_storage_VolumeList_To_v1alpha1_VolumeList(in *storage.VolumeList, out *VolumeList, s conversion.Scope) error {
	return autoConvert_storage_VolumeList_To_v1alpha1_VolumeList(in, out, s)
}

func autoConvert_v1alpha1_VolumePool_To_storage_VolumePool(in *VolumePool, out *storage.VolumePool, s conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	if err := Convert_v1alpha1_VolumePoolSpec_To_storage_VolumePoolSpec(&in.Spec, &out.Spec, s); err != nil {
		return err
	}
	if err := Convert_v1alpha1_VolumePoolStatus_To_storage_VolumePoolStatus(&in.Status, &out.Status, s); err != nil {
		return err
	}
	return nil
}

// Convert_v1alpha1_VolumePool_To_storage_VolumePool is an autogenerated conversion function.
func Convert_v1alpha1_VolumePool_To_storage_VolumePool(in *VolumePool, out *storage.VolumePool, s conversion.Scope) error {
	return autoConvert_v1alpha1_VolumePool_To_storage_VolumePool(in, out, s)
}

func autoConvert_storage_VolumePool_To_v1alpha1_VolumePool(in *storage.VolumePool, out *VolumePool, s conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	if err := Convert_storage_VolumePoolSpec_To_v1alpha1_VolumePoolSpec(&in.Spec, &out.Spec, s); err != nil {
		return err
	}
	if err := Convert_storage_VolumePoolStatus_To_v1alpha1_VolumePoolStatus(&in.Status, &out.Status, s); err != nil {
		return err
	}
	return nil
}

// Convert_storage_VolumePool_To_v1alpha1_VolumePool is an autogenerated conversion function.
func Convert_storage_VolumePool_To_v1alpha1_VolumePool(in *storage.VolumePool, out *VolumePool, s conversion.Scope) error {
	return autoConvert_storage_VolumePool_To_v1alpha1_VolumePool(in, out, s)
}

func autoConvert_v1alpha1_VolumePoolCondition_To_storage_VolumePoolCondition(in *VolumePoolCondition, out *storage.VolumePoolCondition, s conversion.Scope) error {
	out.Type = storage.VolumePoolConditionType(in.Type)
	out.Status = v1.ConditionStatus(in.Status)
	out.Reason = in.Reason
	out.Message = in.Message
	out.ObservedGeneration = in.ObservedGeneration
	out.LastTransitionTime = in.LastTransitionTime
	return nil
}

// Convert_v1alpha1_VolumePoolCondition_To_storage_VolumePoolCondition is an autogenerated conversion function.
func Convert_v1alpha1_VolumePoolCondition_To_storage_VolumePoolCondition(in *VolumePoolCondition, out *storage.VolumePoolCondition, s conversion.Scope) error {
	return autoConvert_v1alpha1_VolumePoolCondition_To_storage_VolumePoolCondition(in, out, s)
}

func autoConvert_storage_VolumePoolCondition_To_v1alpha1_VolumePoolCondition(in *storage.VolumePoolCondition, out *VolumePoolCondition, s conversion.Scope) error {
	out.Type = VolumePoolConditionType(in.Type)
	out.Status = v1.ConditionStatus(in.Status)
	out.Reason = in.Reason
	out.Message = in.Message
	out.ObservedGeneration = in.ObservedGeneration
	out.LastTransitionTime = in.LastTransitionTime
	return nil
}

// Convert_storage_VolumePoolCondition_To_v1alpha1_VolumePoolCondition is an autogenerated conversion function.
func Convert_storage_VolumePoolCondition_To_v1alpha1_VolumePoolCondition(in *storage.VolumePoolCondition, out *VolumePoolCondition, s conversion.Scope) error {
	return autoConvert_storage_VolumePoolCondition_To_v1alpha1_VolumePoolCondition(in, out, s)
}

func autoConvert_v1alpha1_VolumePoolList_To_storage_VolumePoolList(in *VolumePoolList, out *storage.VolumePoolList, s conversion.Scope) error {
	out.ListMeta = in.ListMeta
	out.Items = *(*[]storage.VolumePool)(unsafe.Pointer(&in.Items))
	return nil
}

// Convert_v1alpha1_VolumePoolList_To_storage_VolumePoolList is an autogenerated conversion function.
func Convert_v1alpha1_VolumePoolList_To_storage_VolumePoolList(in *VolumePoolList, out *storage.VolumePoolList, s conversion.Scope) error {
	return autoConvert_v1alpha1_VolumePoolList_To_storage_VolumePoolList(in, out, s)
}

func autoConvert_storage_VolumePoolList_To_v1alpha1_VolumePoolList(in *storage.VolumePoolList, out *VolumePoolList, s conversion.Scope) error {
	out.ListMeta = in.ListMeta
	out.Items = *(*[]VolumePool)(unsafe.Pointer(&in.Items))
	return nil
}

// Convert_storage_VolumePoolList_To_v1alpha1_VolumePoolList is an autogenerated conversion function.
func Convert_storage_VolumePoolList_To_v1alpha1_VolumePoolList(in *storage.VolumePoolList, out *VolumePoolList, s conversion.Scope) error {
	return autoConvert_storage_VolumePoolList_To_v1alpha1_VolumePoolList(in, out, s)
}

func autoConvert_v1alpha1_VolumePoolSpec_To_storage_VolumePoolSpec(in *VolumePoolSpec, out *storage.VolumePoolSpec, s conversion.Scope) error {
	out.ProviderID = in.ProviderID
	out.Taints = *(*[]commonv1alpha1.Taint)(unsafe.Pointer(&in.Taints))
	return nil
}

// Convert_v1alpha1_VolumePoolSpec_To_storage_VolumePoolSpec is an autogenerated conversion function.
func Convert_v1alpha1_VolumePoolSpec_To_storage_VolumePoolSpec(in *VolumePoolSpec, out *storage.VolumePoolSpec, s conversion.Scope) error {
	return autoConvert_v1alpha1_VolumePoolSpec_To_storage_VolumePoolSpec(in, out, s)
}

func autoConvert_storage_VolumePoolSpec_To_v1alpha1_VolumePoolSpec(in *storage.VolumePoolSpec, out *VolumePoolSpec, s conversion.Scope) error {
	out.ProviderID = in.ProviderID
	out.Taints = *(*[]commonv1alpha1.Taint)(unsafe.Pointer(&in.Taints))
	return nil
}

// Convert_storage_VolumePoolSpec_To_v1alpha1_VolumePoolSpec is an autogenerated conversion function.
func Convert_storage_VolumePoolSpec_To_v1alpha1_VolumePoolSpec(in *storage.VolumePoolSpec, out *VolumePoolSpec, s conversion.Scope) error {
	return autoConvert_storage_VolumePoolSpec_To_v1alpha1_VolumePoolSpec(in, out, s)
}

func autoConvert_v1alpha1_VolumePoolStatus_To_storage_VolumePoolStatus(in *VolumePoolStatus, out *storage.VolumePoolStatus, s conversion.Scope) error {
	out.State = storage.VolumePoolState(in.State)
	out.Conditions = *(*[]storage.VolumePoolCondition)(unsafe.Pointer(&in.Conditions))
	out.AvailableVolumeClasses = *(*[]v1.LocalObjectReference)(unsafe.Pointer(&in.AvailableVolumeClasses))
	out.Available = *(*v1.ResourceList)(unsafe.Pointer(&in.Available))
	out.Used = *(*v1.ResourceList)(unsafe.Pointer(&in.Used))
	return nil
}

// Convert_v1alpha1_VolumePoolStatus_To_storage_VolumePoolStatus is an autogenerated conversion function.
func Convert_v1alpha1_VolumePoolStatus_To_storage_VolumePoolStatus(in *VolumePoolStatus, out *storage.VolumePoolStatus, s conversion.Scope) error {
	return autoConvert_v1alpha1_VolumePoolStatus_To_storage_VolumePoolStatus(in, out, s)
}

func autoConvert_storage_VolumePoolStatus_To_v1alpha1_VolumePoolStatus(in *storage.VolumePoolStatus, out *VolumePoolStatus, s conversion.Scope) error {
	out.State = VolumePoolState(in.State)
	out.Conditions = *(*[]VolumePoolCondition)(unsafe.Pointer(&in.Conditions))
	out.AvailableVolumeClasses = *(*[]v1.LocalObjectReference)(unsafe.Pointer(&in.AvailableVolumeClasses))
	out.Available = *(*v1.ResourceList)(unsafe.Pointer(&in.Available))
	out.Used = *(*v1.ResourceList)(unsafe.Pointer(&in.Used))
	return nil
}

// Convert_storage_VolumePoolStatus_To_v1alpha1_VolumePoolStatus is an autogenerated conversion function.
func Convert_storage_VolumePoolStatus_To_v1alpha1_VolumePoolStatus(in *storage.VolumePoolStatus, out *VolumePoolStatus, s conversion.Scope) error {
	return autoConvert_storage_VolumePoolStatus_To_v1alpha1_VolumePoolStatus(in, out, s)
}

func autoConvert_v1alpha1_VolumeSpec_To_storage_VolumeSpec(in *VolumeSpec, out *storage.VolumeSpec, s conversion.Scope) error {
	out.VolumeClassRef = in.VolumeClassRef
	out.VolumePoolSelector = *(*map[string]string)(unsafe.Pointer(&in.VolumePoolSelector))
	out.VolumePoolRef = (*v1.LocalObjectReference)(unsafe.Pointer(in.VolumePoolRef))
	out.ClaimRef = (*commonv1alpha1.LocalUIDReference)(unsafe.Pointer(in.ClaimRef))
	out.Resources = *(*v1.ResourceList)(unsafe.Pointer(&in.Resources))
	out.Image = in.Image
	out.ImagePullSecretRef = (*v1.LocalObjectReference)(unsafe.Pointer(in.ImagePullSecretRef))
	out.Unclaimable = in.Unclaimable
	out.Tolerations = *(*[]commonv1alpha1.Toleration)(unsafe.Pointer(&in.Tolerations))
	return nil
}

// Convert_v1alpha1_VolumeSpec_To_storage_VolumeSpec is an autogenerated conversion function.
func Convert_v1alpha1_VolumeSpec_To_storage_VolumeSpec(in *VolumeSpec, out *storage.VolumeSpec, s conversion.Scope) error {
	return autoConvert_v1alpha1_VolumeSpec_To_storage_VolumeSpec(in, out, s)
}

func autoConvert_storage_VolumeSpec_To_v1alpha1_VolumeSpec(in *storage.VolumeSpec, out *VolumeSpec, s conversion.Scope) error {
	out.VolumeClassRef = in.VolumeClassRef
	out.VolumePoolSelector = *(*map[string]string)(unsafe.Pointer(&in.VolumePoolSelector))
	out.VolumePoolRef = (*v1.LocalObjectReference)(unsafe.Pointer(in.VolumePoolRef))
	out.ClaimRef = (*commonv1alpha1.LocalUIDReference)(unsafe.Pointer(in.ClaimRef))
	out.Resources = *(*v1.ResourceList)(unsafe.Pointer(&in.Resources))
	out.Image = in.Image
	out.ImagePullSecretRef = (*v1.LocalObjectReference)(unsafe.Pointer(in.ImagePullSecretRef))
	out.Unclaimable = in.Unclaimable
	out.Tolerations = *(*[]commonv1alpha1.Toleration)(unsafe.Pointer(&in.Tolerations))
	return nil
}

// Convert_storage_VolumeSpec_To_v1alpha1_VolumeSpec is an autogenerated conversion function.
func Convert_storage_VolumeSpec_To_v1alpha1_VolumeSpec(in *storage.VolumeSpec, out *VolumeSpec, s conversion.Scope) error {
	return autoConvert_storage_VolumeSpec_To_v1alpha1_VolumeSpec(in, out, s)
}

func autoConvert_v1alpha1_VolumeStatus_To_storage_VolumeStatus(in *VolumeStatus, out *storage.VolumeStatus, s conversion.Scope) error {
	out.State = storage.VolumeState(in.State)
	out.LastStateTransitionTime = (*metav1.Time)(unsafe.Pointer(in.LastStateTransitionTime))
	out.Phase = storage.VolumePhase(in.Phase)
	out.LastPhaseTransitionTime = (*metav1.Time)(unsafe.Pointer(in.LastPhaseTransitionTime))
	out.Access = (*storage.VolumeAccess)(unsafe.Pointer(in.Access))
	out.Conditions = *(*[]storage.VolumeCondition)(unsafe.Pointer(&in.Conditions))
	return nil
}

// Convert_v1alpha1_VolumeStatus_To_storage_VolumeStatus is an autogenerated conversion function.
func Convert_v1alpha1_VolumeStatus_To_storage_VolumeStatus(in *VolumeStatus, out *storage.VolumeStatus, s conversion.Scope) error {
	return autoConvert_v1alpha1_VolumeStatus_To_storage_VolumeStatus(in, out, s)
}

func autoConvert_storage_VolumeStatus_To_v1alpha1_VolumeStatus(in *storage.VolumeStatus, out *VolumeStatus, s conversion.Scope) error {
	out.State = VolumeState(in.State)
	out.LastStateTransitionTime = (*metav1.Time)(unsafe.Pointer(in.LastStateTransitionTime))
	out.Phase = VolumePhase(in.Phase)
	out.LastPhaseTransitionTime = (*metav1.Time)(unsafe.Pointer(in.LastPhaseTransitionTime))
	out.Access = (*VolumeAccess)(unsafe.Pointer(in.Access))
	out.Conditions = *(*[]VolumeCondition)(unsafe.Pointer(&in.Conditions))
	return nil
}

// Convert_storage_VolumeStatus_To_v1alpha1_VolumeStatus is an autogenerated conversion function.
func Convert_storage_VolumeStatus_To_v1alpha1_VolumeStatus(in *storage.VolumeStatus, out *VolumeStatus, s conversion.Scope) error {
	return autoConvert_storage_VolumeStatus_To_v1alpha1_VolumeStatus(in, out, s)
}

func autoConvert_v1alpha1_VolumeTemplateSpec_To_storage_VolumeTemplateSpec(in *VolumeTemplateSpec, out *storage.VolumeTemplateSpec, s conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	if err := Convert_v1alpha1_VolumeSpec_To_storage_VolumeSpec(&in.Spec, &out.Spec, s); err != nil {
		return err
	}
	return nil
}

// Convert_v1alpha1_VolumeTemplateSpec_To_storage_VolumeTemplateSpec is an autogenerated conversion function.
func Convert_v1alpha1_VolumeTemplateSpec_To_storage_VolumeTemplateSpec(in *VolumeTemplateSpec, out *storage.VolumeTemplateSpec, s conversion.Scope) error {
	return autoConvert_v1alpha1_VolumeTemplateSpec_To_storage_VolumeTemplateSpec(in, out, s)
}

func autoConvert_storage_VolumeTemplateSpec_To_v1alpha1_VolumeTemplateSpec(in *storage.VolumeTemplateSpec, out *VolumeTemplateSpec, s conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	if err := Convert_storage_VolumeSpec_To_v1alpha1_VolumeSpec(&in.Spec, &out.Spec, s); err != nil {
		return err
	}
	return nil
}

// Convert_storage_VolumeTemplateSpec_To_v1alpha1_VolumeTemplateSpec is an autogenerated conversion function.
func Convert_storage_VolumeTemplateSpec_To_v1alpha1_VolumeTemplateSpec(in *storage.VolumeTemplateSpec, out *VolumeTemplateSpec, s conversion.Scope) error {
	return autoConvert_storage_VolumeTemplateSpec_To_v1alpha1_VolumeTemplateSpec(in, out, s)
}
