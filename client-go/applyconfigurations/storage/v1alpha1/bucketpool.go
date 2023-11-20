// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1alpha1

import (
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	internal "github.com/ironcore-dev/ironcore/client-go/applyconfigurations/internal"
	v1 "github.com/ironcore-dev/ironcore/client-go/applyconfigurations/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	managedfields "k8s.io/apimachinery/pkg/util/managedfields"
)

// BucketPoolApplyConfiguration represents an declarative configuration of the BucketPool type for use
// with apply.
type BucketPoolApplyConfiguration struct {
	v1.TypeMetaApplyConfiguration    `json:",inline"`
	*v1.ObjectMetaApplyConfiguration `json:"metadata,omitempty"`
	Spec                             *BucketPoolSpecApplyConfiguration   `json:"spec,omitempty"`
	Status                           *BucketPoolStatusApplyConfiguration `json:"status,omitempty"`
}

// BucketPool constructs an declarative configuration of the BucketPool type for use with
// apply.
func BucketPool(name string) *BucketPoolApplyConfiguration {
	b := &BucketPoolApplyConfiguration{}
	b.WithName(name)
	b.WithKind("BucketPool")
	b.WithAPIVersion("storage.ironcore.dev/v1alpha1")
	return b
}

// ExtractBucketPool extracts the applied configuration owned by fieldManager from
// bucketPool. If no managedFields are found in bucketPool for fieldManager, a
// BucketPoolApplyConfiguration is returned with only the Name, Namespace (if applicable),
// APIVersion and Kind populated. It is possible that no managed fields were found for because other
// field managers have taken ownership of all the fields previously owned by fieldManager, or because
// the fieldManager never owned fields any fields.
// bucketPool must be a unmodified BucketPool API object that was retrieved from the Kubernetes API.
// ExtractBucketPool provides a way to perform a extract/modify-in-place/apply workflow.
// Note that an extracted apply configuration will contain fewer fields than what the fieldManager previously
// applied if another fieldManager has updated or force applied any of the previously applied fields.
// Experimental!
func ExtractBucketPool(bucketPool *storagev1alpha1.BucketPool, fieldManager string) (*BucketPoolApplyConfiguration, error) {
	return extractBucketPool(bucketPool, fieldManager, "")
}

// ExtractBucketPoolStatus is the same as ExtractBucketPool except
// that it extracts the status subresource applied configuration.
// Experimental!
func ExtractBucketPoolStatus(bucketPool *storagev1alpha1.BucketPool, fieldManager string) (*BucketPoolApplyConfiguration, error) {
	return extractBucketPool(bucketPool, fieldManager, "status")
}

func extractBucketPool(bucketPool *storagev1alpha1.BucketPool, fieldManager string, subresource string) (*BucketPoolApplyConfiguration, error) {
	b := &BucketPoolApplyConfiguration{}
	err := managedfields.ExtractInto(bucketPool, internal.Parser().Type("com.github.ironcore-dev.ironcore.api.storage.v1alpha1.BucketPool"), fieldManager, b, subresource)
	if err != nil {
		return nil, err
	}
	b.WithName(bucketPool.Name)

	b.WithKind("BucketPool")
	b.WithAPIVersion("storage.ironcore.dev/v1alpha1")
	return b, nil
}

// WithKind sets the Kind field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Kind field is set to the value of the last call.
func (b *BucketPoolApplyConfiguration) WithKind(value string) *BucketPoolApplyConfiguration {
	b.Kind = &value
	return b
}

// WithAPIVersion sets the APIVersion field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the APIVersion field is set to the value of the last call.
func (b *BucketPoolApplyConfiguration) WithAPIVersion(value string) *BucketPoolApplyConfiguration {
	b.APIVersion = &value
	return b
}

// WithName sets the Name field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Name field is set to the value of the last call.
func (b *BucketPoolApplyConfiguration) WithName(value string) *BucketPoolApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	b.Name = &value
	return b
}

// WithGenerateName sets the GenerateName field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the GenerateName field is set to the value of the last call.
func (b *BucketPoolApplyConfiguration) WithGenerateName(value string) *BucketPoolApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	b.GenerateName = &value
	return b
}

// WithNamespace sets the Namespace field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Namespace field is set to the value of the last call.
func (b *BucketPoolApplyConfiguration) WithNamespace(value string) *BucketPoolApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	b.Namespace = &value
	return b
}

// WithUID sets the UID field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the UID field is set to the value of the last call.
func (b *BucketPoolApplyConfiguration) WithUID(value types.UID) *BucketPoolApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	b.UID = &value
	return b
}

// WithResourceVersion sets the ResourceVersion field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ResourceVersion field is set to the value of the last call.
func (b *BucketPoolApplyConfiguration) WithResourceVersion(value string) *BucketPoolApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	b.ResourceVersion = &value
	return b
}

// WithGeneration sets the Generation field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Generation field is set to the value of the last call.
func (b *BucketPoolApplyConfiguration) WithGeneration(value int64) *BucketPoolApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	b.Generation = &value
	return b
}

// WithCreationTimestamp sets the CreationTimestamp field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the CreationTimestamp field is set to the value of the last call.
func (b *BucketPoolApplyConfiguration) WithCreationTimestamp(value metav1.Time) *BucketPoolApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	b.CreationTimestamp = &value
	return b
}

// WithDeletionTimestamp sets the DeletionTimestamp field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the DeletionTimestamp field is set to the value of the last call.
func (b *BucketPoolApplyConfiguration) WithDeletionTimestamp(value metav1.Time) *BucketPoolApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	b.DeletionTimestamp = &value
	return b
}

// WithDeletionGracePeriodSeconds sets the DeletionGracePeriodSeconds field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the DeletionGracePeriodSeconds field is set to the value of the last call.
func (b *BucketPoolApplyConfiguration) WithDeletionGracePeriodSeconds(value int64) *BucketPoolApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	b.DeletionGracePeriodSeconds = &value
	return b
}

// WithLabels puts the entries into the Labels field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, the entries provided by each call will be put on the Labels field,
// overwriting an existing map entries in Labels field with the same key.
func (b *BucketPoolApplyConfiguration) WithLabels(entries map[string]string) *BucketPoolApplyConfiguration {
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
func (b *BucketPoolApplyConfiguration) WithAnnotations(entries map[string]string) *BucketPoolApplyConfiguration {
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
func (b *BucketPoolApplyConfiguration) WithOwnerReferences(values ...*v1.OwnerReferenceApplyConfiguration) *BucketPoolApplyConfiguration {
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
func (b *BucketPoolApplyConfiguration) WithFinalizers(values ...string) *BucketPoolApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	for i := range values {
		b.Finalizers = append(b.Finalizers, values[i])
	}
	return b
}

func (b *BucketPoolApplyConfiguration) ensureObjectMetaApplyConfigurationExists() {
	if b.ObjectMetaApplyConfiguration == nil {
		b.ObjectMetaApplyConfiguration = &v1.ObjectMetaApplyConfiguration{}
	}
}

// WithSpec sets the Spec field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Spec field is set to the value of the last call.
func (b *BucketPoolApplyConfiguration) WithSpec(value *BucketPoolSpecApplyConfiguration) *BucketPoolApplyConfiguration {
	b.Spec = value
	return b
}

// WithStatus sets the Status field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Status field is set to the value of the last call.
func (b *BucketPoolApplyConfiguration) WithStatus(value *BucketPoolStatusApplyConfiguration) *BucketPoolApplyConfiguration {
	b.Status = value
	return b
}
