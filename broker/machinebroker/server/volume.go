// Copyright 2022 IronCore authors
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

	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	"github.com/ironcore-dev/ironcore/utils/generic"
	"golang.org/x/exp/slices"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ironcoreMachineVolumeIndex(ironcoreMachine *computev1alpha1.Machine, name string) int {
	return slices.IndexFunc(
		ironcoreMachine.Spec.Volumes,
		func(volume computev1alpha1.Volume) bool {
			return volume.Name == name
		},
	)
}

func (s *Server) bindIronCoreMachineVolume(
	ctx context.Context,
	ironcoreMachine *computev1alpha1.Machine,
	ironcoreVolume *storagev1alpha1.Volume,
) error {
	baseIronCoreVolume := ironcoreVolume.DeepCopy()
	if err := ctrl.SetControllerReference(ironcoreMachine, ironcoreVolume, s.cluster.Scheme()); err != nil {
		return err
	}
	ironcoreVolume.Spec.ClaimRef = generic.Pointer(s.localObjectReferenceTo(ironcoreMachine))
	return s.cluster.Client().Patch(ctx, ironcoreVolume, client.StrategicMergeFrom(baseIronCoreVolume))
}

func (s *Server) aggregateIronCoreVolume(
	ctx context.Context,
	rd client.Reader,
	ironcoreVolume *storagev1alpha1.Volume,
) (*AggregateIronCoreVolume, error) {
	access := ironcoreVolume.Status.Access
	if access == nil {
		return nil, fmt.Errorf("volume does not specify access")
	}

	var ironcoreVolumeAccessSecret *corev1.Secret
	if ironcoreVolumeSecretRef := access.SecretRef; ironcoreVolumeSecretRef != nil {
		secret := &corev1.Secret{}
		secretKey := client.ObjectKey{Namespace: s.cluster.Namespace(), Name: ironcoreVolumeSecretRef.Name}
		if err := rd.Get(ctx, secretKey, secret); err != nil {
			return nil, fmt.Errorf("error access secret %s: %w", ironcoreVolumeSecretRef.Name, err)
		}

		ironcoreVolumeAccessSecret = secret
	}

	return &AggregateIronCoreVolume{
		Volume:       ironcoreVolume,
		AccessSecret: ironcoreVolumeAccessSecret,
	}, nil
}
