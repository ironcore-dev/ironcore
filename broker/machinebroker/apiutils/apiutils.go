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
	"strings"

	"github.com/onmetal/controller-utils/metautils"
	corev1alpha1 "github.com/onmetal/onmetal-api/api/core/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	machinebrokerv1alpha1 "github.com/onmetal/onmetal-api/broker/machinebroker/api/v1alpha1"
	orimeta "github.com/onmetal/onmetal-api/ori/apis/meta/v1alpha1"
	"golang.org/x/exp/slices"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func GetObjectMetadata(o metav1.Object) (*orimeta.ObjectMetadata, error) {
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

	return &orimeta.ObjectMetadata{
		Id:          o.GetName(),
		Annotations: annotations,
		Labels:      labels,
		Generation:  o.GetGeneration(),
		CreatedAt:   o.GetCreationTimestamp().UnixNano(),
		DeletedAt:   deletedAt,
	}, nil
}

func SetObjectMetadata(o metav1.Object, metadata *orimeta.ObjectMetadata) error {
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

func PatchOwnedBy(ctx context.Context, c client.Client, owner, obj client.Object) error {
	base := obj.DeepCopyObject().(client.Object)
	if err := controllerutil.SetOwnerReference(owner, obj, c.Scheme()); err != nil {
		return err
	}

	if err := c.Patch(ctx, obj, client.MergeFrom(base)); err != nil {
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

var (
	ipAndPrefixReplacer        = strings.NewReplacer("/", "-", ":", "_")
	reverseIPAndPrefixReplacer = strings.NewReplacer("-", "/", "_", ":")
)

func SetNetworkHandleLabel(o metav1.Object, handle string) {
	metautils.SetLabel(o, machinebrokerv1alpha1.NetworkHandleLabel, handle)
}

func EscapePrefix(prefix corev1alpha1.IPPrefix) string {
	return ipAndPrefixReplacer.Replace(prefix.String())
}

func UnescapePrefix(escapedPrefix string) (corev1alpha1.IPPrefix, error) {
	unescaped := reverseIPAndPrefixReplacer.Replace(escapedPrefix)
	return corev1alpha1.ParseIPPrefix(unescaped)
}

func SetPrefixLabel(o metav1.Object, prefix corev1alpha1.IPPrefix) {
	metautils.SetLabel(o, machinebrokerv1alpha1.PrefixLabel, EscapePrefix(prefix))
}

func GetPrefixLabel(o metav1.Object) (corev1alpha1.IPPrefix, error) {
	escapedPrefix := o.GetLabels()[machinebrokerv1alpha1.PrefixLabel]
	return UnescapePrefix(escapedPrefix)
}

func SetLoadBalancerTypeLabel(o metav1.Object, typ networkingv1alpha1.LoadBalancerType) {
	metautils.SetLabel(o, machinebrokerv1alpha1.LoadBalancerTypeLabel, string(typ))
}

func EscapeIP(ip corev1alpha1.IP) string {
	return ipAndPrefixReplacer.Replace(ip.String())
}

func UnescapeIP(escaped string) (corev1alpha1.IP, error) {
	unescaped := reverseIPAndPrefixReplacer.Replace(escaped)
	return corev1alpha1.ParseIP(unescaped)
}

func SetIPLabel(o metav1.Object, ip corev1alpha1.IP) {
	metautils.SetLabel(o, machinebrokerv1alpha1.IPLabel, EscapeIP(ip))
}

func GetIPLabel(o metav1.Object) (corev1alpha1.IP, error) {
	escaped := o.GetLabels()[machinebrokerv1alpha1.IPLabel]
	return UnescapeIP(escaped)
}

func SetManagerLabel(o metav1.Object, manager string) {
	metautils.SetLabel(o, machinebrokerv1alpha1.ManagerLabel, manager)
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

func GetDependents(o metav1.Object) ([]string, error) {
	dependentsData, ok := o.GetAnnotations()[machinebrokerv1alpha1.DependentsAnnotation]
	if !ok {
		return nil, nil
	}
	var dependents []string
	if err := json.Unmarshal([]byte(dependentsData), &dependents); err != nil {
		return nil, fmt.Errorf("error unmarshalling dependents: %w", err)
	}
	return dependents, nil
}

func SetDependents(o metav1.Object, dependents []string) error {
	dependentsData, err := json.Marshal(dependents)
	if err != nil {
		return fmt.Errorf("error marshalling dependents: %w", err)
	}

	metautils.SetAnnotation(o, machinebrokerv1alpha1.DependentsAnnotation, string(dependentsData))
	return nil
}

func AddDependent(o metav1.Object, newDependent string) (bool, error) {
	dependents, err := GetDependents(o)
	if err != nil {
		return false, fmt.Errorf("error getting dependents: %w", err)
	}

	for _, dependent := range dependents {
		if dependent == newDependent {
			return false, nil
		}
	}

	if err := SetDependents(o, append(dependents, newDependent)); err != nil {
		return false, fmt.Errorf("error setting dependents: %w", err)
	}
	return true, nil
}

func RemoveDependent(o metav1.Object, removeDependent string) (bool, error) {
	dependents, err := GetDependents(o)
	if err != nil {
		return false, fmt.Errorf("error getting dependents: %w", err)
	}

	var updated bool
	for i := 0; i < len(dependents); i++ {
		if dependents[i] == removeDependent {
			dependents = append(dependents[:i], dependents[i+1:]...)
			i--
			updated = true
		}
	}
	if !updated {
		return false, nil
	}

	if err := SetDependents(o, dependents); err != nil {
		return false, fmt.Errorf("error setting dependents: %w", err)
	}
	return true, nil
}

func PatchCreatedWithDependent(ctx context.Context, c client.Client, o client.Object, dependent string) error {
	base := o.DeepCopyObject().(client.Object)
	SetCreatedLabel(o)
	if _, err := AddDependent(o, dependent); err != nil {
		return fmt.Errorf("error adding dependent: %w", err)
	}
	if err := c.Patch(ctx, o, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error patching created / adding dependent: %w", err)
	}
	return nil
}

func DeleteAndGarbageCollect(ctx context.Context, c client.Client, o client.Object, dependent string) error {
	dependents, err := GetDependents(o)
	if err != nil {
		return fmt.Errorf("error getting dependents: %w", err)
	}

	var updated bool
	for i := 0; i < len(dependents); i++ {
		if dependents[i] == dependent {
			dependents = append(dependents[:i], dependents[i+1:]...)
			i--
			updated = true
		}
	}
	if !updated {
		return nil
	}

	if len(dependents) == 0 {
		if err := c.Delete(ctx, o); client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("error deleting object: %w", err)
		}
		return nil
	}

	if !updated {
		return nil
	}
	base := o.DeepCopyObject().(client.Object)
	if err := SetDependents(o, dependents); err != nil {
		return fmt.Errorf("error setting dependents: %w", err)
	}
	if err := c.Patch(ctx, o, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error updating dependents: %w", err)
	}
	return nil
}

func HasDependent(obj client.Object, dependent string) (bool, error) {
	dependents, err := GetDependents(obj)
	if err != nil {
		return false, fmt.Errorf("error getting dependents: %w", err)
	}

	return slices.Contains(dependents, dependent), nil
}

func FilterObjectListByDependent[S ~[]Obj, ObjPtr interface {
	client.Object
	*Obj
}, Obj any](objs S, dependent string) ([]Obj, error) {
	var filtered []Obj
	for _, obj := range objs {
		objPtr := ObjPtr(&obj)

		ok, err := HasDependent(objPtr, dependent)
		if err != nil {
			return nil, fmt.Errorf("[object %s] %w", client.ObjectKeyFromObject(objPtr), err)
		}
		if !ok {
			continue
		}

		filtered = append(filtered, obj)
	}
	return filtered, nil
}
