// Copyright 2022 OnMetal authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package apiutils

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/onmetal/controller-utils/metautils"
	machinebrokerv1alpha1 "github.com/onmetal/onmetal-api/broker/machinebroker/api/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetObjectMetadata(o metav1.Object) (*ori.ObjectMetadata, error) {
	annotations, err := GetAnnotationsAnnotation(o)
	if err != nil {
		return nil, err
	}

	labels, err := GetLabelsAnnotation(o)
	if err != nil {
		return nil, err
	}

	var deletedAt int64
	if !o.GetDeletionTimestamp().IsZero() {
		deletedAt = o.GetDeletionTimestamp().UnixNano()
	}

	return &ori.ObjectMetadata{
		Id:          o.GetName(),
		Annotations: annotations,
		Labels:      labels,
		Generation:  o.GetGeneration(),
		CreatedAt:   o.GetCreationTimestamp().UnixNano(),
		DeletedAt:   deletedAt,
	}, nil
}

func SetObjectMetadata(o metav1.Object, metadata *ori.ObjectMetadata) error {
	if err := SetAnnotationsAnnotation(o, metadata.Annotations); err != nil {
		return err
	}
	if err := SetLabelsAnnotation(o, metadata.Labels); err != nil {
		return err
	}
	return nil
}

func SetCreatedLabel(o metav1.Object) {
	metautils.SetLabel(o, machinebrokerv1alpha1.CreatedLabel, "true")
}

func IsCreated(o metav1.Object) bool {
	return metautils.HasLabel(o, machinebrokerv1alpha1.CreatedLabel)
}

func PatchControlledBy(ctx context.Context, c client.Client, owner, controlled client.Object) error {
	base := controlled.DeepCopyObject().(client.Object)
	if err := ctrl.SetControllerReference(owner, controlled, c.Scheme()); err != nil {
		return err
	}

	if err := c.Patch(ctx, controlled, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error patching object to be controlled: %w", err)
	}
	return nil
}

func PatchCreated(ctx context.Context, c client.Client, o client.Object) error {
	base := o.DeepCopyObject().(client.Object)
	SetCreatedLabel(o)
	if err := c.Patch(ctx, o, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error patching object to created: %w", err)
	}
	return nil
}

func SetLabelsAnnotation(o metav1.Object, labels map[string]string) error {
	data, err := json.Marshal(labels)
	if err != nil {
		return fmt.Errorf("error marshalling labels: %w", err)
	}
	metautils.SetAnnotation(o, machinebrokerv1alpha1.LabelsAnnotation, string(data))
	return nil
}

func GetLabelsAnnotation(o metav1.Object) (map[string]string, error) {
	data, ok := o.GetAnnotations()[machinebrokerv1alpha1.LabelsAnnotation]
	if !ok {
		return nil, fmt.Errorf("object has no labels at %s", machinebrokerv1alpha1.LabelsAnnotation)
	}

	var labels map[string]string
	if err := json.Unmarshal([]byte(data), &labels); err != nil {
		return nil, err
	}

	return labels, nil
}

func SetAnnotationsAnnotation(o metav1.Object, annotations map[string]string) error {
	data, err := json.Marshal(annotations)
	if err != nil {
		return fmt.Errorf("error marshalling annotations: %w", err)
	}
	metautils.SetAnnotation(o, machinebrokerv1alpha1.AnnotationsAnnotation, string(data))
	return nil
}

func GetAnnotationsAnnotation(o metav1.Object) (map[string]string, error) {
	data, ok := o.GetAnnotations()[machinebrokerv1alpha1.AnnotationsAnnotation]
	if !ok {
		return nil, fmt.Errorf("object has no annotations at %s", machinebrokerv1alpha1.AnnotationsAnnotation)
	}

	var annotations map[string]string
	if err := json.Unmarshal([]byte(data), &annotations); err != nil {
		return nil, err
	}

	return annotations, nil
}

func SetMachineIDLabel(o metav1.Object, id string) {
	metautils.SetLabel(o, machinebrokerv1alpha1.MachineIDLabel, id)
}

func SetVolumeNameLabel(o metav1.Object, name string) {
	metautils.SetLabel(o, machinebrokerv1alpha1.VolumeNameLabel, name)
}

func SetDeviceLabel(o metav1.Object, device string) {
	metautils.SetLabel(o, machinebrokerv1alpha1.DeviceLabel, device)
}

func SetNetworkInterfaceNameLabel(o metav1.Object, name string) {
	metautils.SetLabel(o, machinebrokerv1alpha1.NetworkInterfaceNameLabel, name)
}

func SetManagerLabel(o metav1.Object, manager string) {
	metautils.SetLabel(o, machinebrokerv1alpha1.ManagerLabel, manager)
}

func SetIPFamilyLabel(o metav1.Object, ipFamily corev1.IPFamily) {
	metautils.SetLabel(o, machinebrokerv1alpha1.IPFamilyLabel, string(ipFamily))
}

func SetPurpose(o metav1.Object, purpose string) {
	metautils.SetLabel(o, machinebrokerv1alpha1.PurposeLabel, purpose)
}

func GetPurpose(o metav1.Object) string {
	return o.GetLabels()[machinebrokerv1alpha1.PurposeLabel]
}

func IsManagedBy(o metav1.Object, manager string) bool {
	actual, ok := o.GetLabels()[machinebrokerv1alpha1.ManagerLabel]
	return ok && actual == manager
}
