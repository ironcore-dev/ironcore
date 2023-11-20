// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	internal "github.com/ironcore-dev/ironcore/client-go/applyconfigurations/internal"
	v1 "github.com/ironcore-dev/ironcore/client-go/applyconfigurations/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	managedfields "k8s.io/apimachinery/pkg/util/managedfields"
)

// VolumeClassApplyConfiguration represents an declarative configuration of the VolumeClass type for use
// with apply.
type VolumeClassApplyConfiguration struct {
	v1.TypeMetaApplyConfiguration    `json:",inline"`
	*v1.ObjectMetaApplyConfiguration `json:"metadata,omitempty"`
	Capabilities                     *v1alpha1.ResourceList        `json:"capabilities,omitempty"`
	ResizePolicy                     *storagev1alpha1.ResizePolicy `json:"resizePolicy,omitempty"`
}

// VolumeClass constructs an declarative configuration of the VolumeClass type for use with
// apply.
func VolumeClass(name string) *VolumeClassApplyConfiguration {
	b := &VolumeClassApplyConfiguration{}
	b.WithName(name)
	b.WithKind("VolumeClass")
	b.WithAPIVersion("storage.ironcore.dev/v1alpha1")
	return b
}

// ExtractVolumeClass extracts the applied configuration owned by fieldManager from
// volumeClass. If no managedFields are found in volumeClass for fieldManager, a
// VolumeClassApplyConfiguration is returned with only the Name, Namespace (if applicable),
// APIVersion and Kind populated. It is possible that no managed fields were found for because other
// field managers have taken ownership of all the fields previously owned by fieldManager, or because
// the fieldManager never owned fields any fields.
// volumeClass must be a unmodified VolumeClass API object that was retrieved from the Kubernetes API.
// ExtractVolumeClass provides a way to perform a extract/modify-in-place/apply workflow.
// Note that an extracted apply configuration will contain fewer fields than what the fieldManager previously
// applied if another fieldManager has updated or force applied any of the previously applied fields.
// Experimental!
func ExtractVolumeClass(volumeClass *storagev1alpha1.VolumeClass, fieldManager string) (*VolumeClassApplyConfiguration, error) {
	return extractVolumeClass(volumeClass, fieldManager, "")
}

// ExtractVolumeClassStatus is the same as ExtractVolumeClass except
// that it extracts the status subresource applied configuration.
// Experimental!
func ExtractVolumeClassStatus(volumeClass *storagev1alpha1.VolumeClass, fieldManager string) (*VolumeClassApplyConfiguration, error) {
	return extractVolumeClass(volumeClass, fieldManager, "status")
}

func extractVolumeClass(volumeClass *storagev1alpha1.VolumeClass, fieldManager string, subresource string) (*VolumeClassApplyConfiguration, error) {
	b := &VolumeClassApplyConfiguration{}
	err := managedfields.ExtractInto(volumeClass, internal.Parser().Type("com.github.ironcore-dev.ironcore.api.storage.v1alpha1.VolumeClass"), fieldManager, b, subresource)
	if err != nil {
		return nil, err
	}
	b.WithName(volumeClass.Name)

	b.WithKind("VolumeClass")
	b.WithAPIVersion("storage.ironcore.dev/v1alpha1")
	return b, nil
}

// WithKind sets the Kind field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Kind field is set to the value of the last call.
func (b *VolumeClassApplyConfiguration) WithKind(value string) *VolumeClassApplyConfiguration {
	b.Kind = &value
	return b
}

// WithAPIVersion sets the APIVersion field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the APIVersion field is set to the value of the last call.
func (b *VolumeClassApplyConfiguration) WithAPIVersion(value string) *VolumeClassApplyConfiguration {
	b.APIVersion = &value
	return b
}

// WithName sets the Name field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Name field is set to the value of the last call.
func (b *VolumeClassApplyConfiguration) WithName(value string) *VolumeClassApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	b.Name = &value
	return b
}

// WithGenerateName sets the GenerateName field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the GenerateName field is set to the value of the last call.
func (b *VolumeClassApplyConfiguration) WithGenerateName(value string) *VolumeClassApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	b.GenerateName = &value
	return b
}

// WithNamespace sets the Namespace field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Namespace field is set to the value of the last call.
func (b *VolumeClassApplyConfiguration) WithNamespace(value string) *VolumeClassApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	b.Namespace = &value
	return b
}

// WithUID sets the UID field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the UID field is set to the value of the last call.
func (b *VolumeClassApplyConfiguration) WithUID(value types.UID) *VolumeClassApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	b.UID = &value
	return b
}

// WithResourceVersion sets the ResourceVersion field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ResourceVersion field is set to the value of the last call.
func (b *VolumeClassApplyConfiguration) WithResourceVersion(value string) *VolumeClassApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	b.ResourceVersion = &value
	return b
}

// WithGeneration sets the Generation field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Generation field is set to the value of the last call.
func (b *VolumeClassApplyConfiguration) WithGeneration(value int64) *VolumeClassApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	b.Generation = &value
	return b
}

// WithCreationTimestamp sets the CreationTimestamp field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the CreationTimestamp field is set to the value of the last call.
func (b *VolumeClassApplyConfiguration) WithCreationTimestamp(value metav1.Time) *VolumeClassApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	b.CreationTimestamp = &value
	return b
}

// WithDeletionTimestamp sets the DeletionTimestamp field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the DeletionTimestamp field is set to the value of the last call.
func (b *VolumeClassApplyConfiguration) WithDeletionTimestamp(value metav1.Time) *VolumeClassApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	b.DeletionTimestamp = &value
	return b
}

// WithDeletionGracePeriodSeconds sets the DeletionGracePeriodSeconds field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the DeletionGracePeriodSeconds field is set to the value of the last call.
func (b *VolumeClassApplyConfiguration) WithDeletionGracePeriodSeconds(value int64) *VolumeClassApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	b.DeletionGracePeriodSeconds = &value
	return b
}

// WithLabels puts the entries into the Labels field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, the entries provided by each call will be put on the Labels field,
// overwriting an existing map entries in Labels field with the same key.
func (b *VolumeClassApplyConfiguration) WithLabels(entries map[string]string) *VolumeClassApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	if b.Labels == nil && len(entries) > 0 {
		b.Labels = make(map[string]string, len(entries))
	}
	for k, v := range entries {
		b.Labels[k] = v
	}
	return b
}

// WithAnnotations puts the entries into the Annotations field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, the entries provided by each call will be put on the Annotations field,
// overwriting an existing map entries in Annotations field with the same key.
func (b *VolumeClassApplyConfiguration) WithAnnotations(entries map[string]string) *VolumeClassApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	if b.Annotations == nil && len(entries) > 0 {
		b.Annotations = make(map[string]string, len(entries))
	}
	for k, v := range entries {
		b.Annotations[k] = v
	}
	return b
}

// WithOwnerReferences adds the given value to the OwnerReferences field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the OwnerReferences field.
func (b *VolumeClassApplyConfiguration) WithOwnerReferences(values ...*v1.OwnerReferenceApplyConfiguration) *VolumeClassApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithOwnerReferences")
		}
		b.OwnerReferences = append(b.OwnerReferences, *values[i])
	}
	return b
}

// WithFinalizers adds the given value to the Finalizers field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the Finalizers field.
func (b *VolumeClassApplyConfiguration) WithFinalizers(values ...string) *VolumeClassApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	for i := range values {
		b.Finalizers = append(b.Finalizers, values[i])
	}
	return b
}

func (b *VolumeClassApplyConfiguration) ensureObjectMetaApplyConfigurationExists() {
	if b.ObjectMetaApplyConfiguration == nil {
		b.ObjectMetaApplyConfiguration = &v1.ObjectMetaApplyConfiguration{}
	}
}

// WithCapabilities sets the Capabilities field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Capabilities field is set to the value of the last call.
func (b *VolumeClassApplyConfiguration) WithCapabilities(value v1alpha1.ResourceList) *VolumeClassApplyConfiguration {
	b.Capabilities = &value
	return b
}

// WithResizePolicy sets the ResizePolicy field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ResizePolicy field is set to the value of the last call.
func (b *VolumeClassApplyConfiguration) WithResizePolicy(value storagev1alpha1.ResizePolicy) *VolumeClassApplyConfiguration {
	b.ResizePolicy = &value
	return b
}
