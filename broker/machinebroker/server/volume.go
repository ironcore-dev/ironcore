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
	"context"
	"fmt"

	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	"github.com/onmetal/onmetal-api/utils/generic"
	"golang.org/x/exp/slices"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func onmetalMachineVolumeIndex(onmetalMachine *computev1alpha1.Machine, name string) int {
	return slices.IndexFunc(
		onmetalMachine.Spec.Volumes,
		func(volume computev1alpha1.Volume) bool {
			return volume.Name == name
		},
	)
}

func (s *Server) bindOnmetalMachineVolume(
	ctx context.Context,
	onmetalMachine *computev1alpha1.Machine,
	onmetalVolume *storagev1alpha1.Volume,
) error {
	baseOnmetalVolume := onmetalVolume.DeepCopy()
	if err := ctrl.SetControllerReference(onmetalMachine, onmetalVolume, s.cluster.Scheme()); err != nil {
		return err
	}
	onmetalVolume.Spec.ClaimRef = generic.Pointer(s.localObjectReferenceTo(onmetalMachine))
	return s.cluster.Client().Patch(ctx, onmetalVolume, client.StrategicMergeFrom(baseOnmetalVolume))
}

func (s *Server) aggregateOnmetalVolume(
	ctx context.Context,
	rd client.Reader,
	onmetalVolume *storagev1alpha1.Volume,
) (*AggregateOnmetalVolume, error) {
	access := onmetalVolume.Status.Access
	if access == nil {
		return nil, fmt.Errorf("volume does not specify access")
	}

	var onmetalVolumeAccessSecret *corev1.Secret
	if onmetalVolumeSecretRef := access.SecretRef; onmetalVolumeSecretRef != nil {
		secret := &corev1.Secret{}
		secretKey := client.ObjectKey{Namespace: s.cluster.Namespace(), Name: onmetalVolumeSecretRef.Name}
		if err := rd.Get(ctx, secretKey, secret); err != nil {
			return nil, fmt.Errorf("error access secret %s: %w", onmetalVolumeSecretRef.Name, err)
		}

		onmetalVolumeAccessSecret = secret
	}

	return &AggregateOnmetalVolume{
		Volume:       onmetalVolume,
		AccessSecret: onmetalVolumeAccessSecret,
	}, nil
}
