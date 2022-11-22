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
	"encoding/json"
	"fmt"

	"github.com/onmetal/controller-utils/metautils"
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/volume/v1alpha1"
	volumebrokerv1alpha1 "github.com/onmetal/onmetal-api/volumebroker/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func SetMetadataAnnotation(o metav1.Object, metadata *ori.VolumeMetadata) error {
	data, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("error marshalling metadata: %w", err)
	}
	metautils.SetAnnotation(o, volumebrokerv1alpha1.MetadataAnnotation, string(data))
	return nil
}

func GetMetadataAnnotation(o metav1.Object) (*ori.VolumeMetadata, error) {
	data, ok := o.GetAnnotations()[volumebrokerv1alpha1.MetadataAnnotation]
	if !ok {
		return nil, fmt.Errorf("object has no metadata at %s", volumebrokerv1alpha1.MetadataAnnotation)
	}

	metadata := &ori.VolumeMetadata{}
	if err := json.Unmarshal([]byte(data), metadata); err != nil {
		return nil, err
	}
	return metadata, nil
}

func SetLabelsAnnotation(o metav1.Object, labels map[string]string) error {
	data, err := json.Marshal(labels)
	if err != nil {
		return fmt.Errorf("error marshalling labels: %w", err)
	}
	metautils.SetAnnotation(o, volumebrokerv1alpha1.LabelsAnnotation, string(data))
	return nil
}

func GetLabelsAnnotation(o metav1.Object) (map[string]string, error) {
	data, ok := o.GetAnnotations()[volumebrokerv1alpha1.LabelsAnnotation]
	if !ok {
		return nil, fmt.Errorf("object has no labels at %s", volumebrokerv1alpha1.LabelsAnnotation)
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
	metautils.SetAnnotation(o, volumebrokerv1alpha1.AnnotationsAnnotation, string(data))
	return nil
}

func GetAnnotationsAnnotation(o metav1.Object) (map[string]string, error) {
	data, ok := o.GetAnnotations()[volumebrokerv1alpha1.AnnotationsAnnotation]
	if !ok {
		return nil, fmt.Errorf("object has no annotations at %s", volumebrokerv1alpha1.AnnotationsAnnotation)
	}

	var annotations map[string]string
	if err := json.Unmarshal([]byte(data), &annotations); err != nil {
		return nil, err
	}

	return annotations, nil
}

func SetVolumeIDLabel(o metav1.Object, id string) {
	metautils.SetLabel(o, volumebrokerv1alpha1.VolumeIDLabel, id)
}

func SetVolumeManagerLabel(volume *storagev1alpha1.Volume, manager string) {
	metautils.SetLabel(volume, volumebrokerv1alpha1.VolumeManagerLabel, manager)
}
