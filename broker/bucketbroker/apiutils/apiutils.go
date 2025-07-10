// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package apiutils

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ironcore-dev/controller-utils/metautils"
	bucketbrokerv1alpha1 "github.com/ironcore-dev/ironcore/broker/bucketbroker/api/v1alpha1"
	irimeta "github.com/ironcore-dev/ironcore/iri/apis/meta/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetObjectMetadata(o metav1.Object) (*irimeta.ObjectMetadata, error) {
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

	return &irimeta.ObjectMetadata{
		Id:          o.GetName(),
		Annotations: annotations,
		Labels:      labels,
		Generation:  o.GetGeneration(),
		CreatedAt:   o.GetCreationTimestamp().UnixNano(),
		DeletedAt:   deletedAt,
	}, nil
}

func SetObjectMetadata(o metav1.Object, metadata *irimeta.ObjectMetadata) error {
	if err := SetAnnotationsAnnotation(o, metadata.Annotations); err != nil {
		return err
	}
	if err := SetLabelsAnnotation(o, metadata.Labels); err != nil {
		return err
	}
	return nil
}

func SetCreatedLabel(o metav1.Object) {
	metautils.SetLabel(o, bucketbrokerv1alpha1.CreatedLabel, "true")
}

func IsCreated(o metav1.Object) bool {
	return metautils.HasLabel(o, bucketbrokerv1alpha1.CreatedLabel)
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
	metautils.SetAnnotation(o, bucketbrokerv1alpha1.LabelsAnnotation, string(data))
	return nil
}

func GetLabelsAnnotation(o metav1.Object) (map[string]string, error) {
	data, ok := o.GetAnnotations()[bucketbrokerv1alpha1.LabelsAnnotation]
	if !ok {
		return nil, fmt.Errorf("object has no labels at %s", bucketbrokerv1alpha1.LabelsAnnotation)
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
	metautils.SetAnnotation(o, bucketbrokerv1alpha1.AnnotationsAnnotation, string(data))
	return nil
}

func GetAnnotationsAnnotation(o metav1.Object) (map[string]string, error) {
	data, ok := o.GetAnnotations()[bucketbrokerv1alpha1.AnnotationsAnnotation]
	if !ok {
		return nil, fmt.Errorf("object has no annotations at %s", bucketbrokerv1alpha1.AnnotationsAnnotation)
	}

	var annotations map[string]string
	if err := json.Unmarshal([]byte(data), &annotations); err != nil {
		return nil, err
	}

	return annotations, nil
}

func IsManagedBy(o metav1.Object, manager string) bool {
	actual, ok := o.GetLabels()[bucketbrokerv1alpha1.ManagerLabel]
	return ok && actual == manager
}
