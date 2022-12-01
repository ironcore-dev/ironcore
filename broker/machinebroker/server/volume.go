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

package server

import (
	"fmt"

	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	"github.com/onmetal/onmetal-api/broker/machinebroker/apiutils"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

type AggreagateOnmetalVolume struct {
	Volume       *storagev1alpha1.Volume
	AccessSecret *corev1.Secret
}

func (s *Server) convertOnmetalVolume(volume *AggreagateOnmetalVolume) (*ori.Volume, error) {
	metadata, err := apiutils.GetObjectMetadata(volume.Volume)
	if err != nil {
		return nil, err
	}

	access := volume.Volume.Status.Access
	if access == nil {
		return nil, fmt.Errorf("volume does not specify access")
	}

	var secretData map[string][]byte
	if volume.AccessSecret != nil {
		secretData = volume.AccessSecret.Data
	}

	return &ori.Volume{
		Metadata: metadata,
		Spec: &ori.VolumeSpec{
			Driver:     access.Driver,
			Handle:     access.Handle,
			Attributes: access.VolumeAttributes,
			SecretData: secretData,
		},
	}, nil
}
