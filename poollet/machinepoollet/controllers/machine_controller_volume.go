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

package controllers

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/gogo/protobuf/proto"
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"github.com/onmetal/onmetal-api/poollet/machinepoollet/controllers/events"
	utilslices "github.com/onmetal/onmetal-api/utils/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// isVolumeBoundToMachine checks if the referenced volume is bound to the machine.
//
// It is assumed that the caller verified that the machine points to the volume.
func (r *MachineReconciler) isVolumeBoundToMachine(machine *computev1alpha1.Machine, name string, volume *storagev1alpha1.Volume) bool {
	if volumePhase := volume.Status.Phase; volumePhase != storagev1alpha1.VolumePhaseBound {
		r.Eventf(machine, corev1.EventTypeNormal, events.VolumeNotReady,
			"Volume %s is in phase %s",
			volume.Name,
			volumePhase,
		)
		return false
	}

	claimRef := volume.Spec.ClaimRef
	if claimRef == nil {
		r.Eventf(machine, corev1.EventTypeNormal, events.VolumeNotReady,
			"Volume %s does not reference any claimer",
			volume.Name,
		)
		return false
	}

	if claimRef.Name != machine.Name || claimRef.UID != machine.UID {
		r.Eventf(machine, corev1.EventTypeNormal, events.VolumeNotReady,
			"Volume %s references a different claimer %s (uid %s)",
			volume.Name,
			claimRef.Name,
			claimRef.UID,
		)
		return false
	}

	for _, volumeStatus := range machine.Status.Volumes {
		if volumeStatus.Name == name {
			if volumeStatus.Phase == computev1alpha1.VolumePhaseBound {
				return true
			}

			r.Eventf(machine, corev1.EventTypeNormal, events.VolumeNotReady,
				"Machine volume status is in phase %s",
				volumeStatus.Phase,
			)
			return false
		}
	}

	r.Eventf(machine, corev1.EventTypeNormal, events.VolumeNotReady,
		"Machine does not yet report volume status",
	)
	return false
}

func (r *MachineReconciler) prepareORIVolumeConnection(
	ctx context.Context,
	machine *computev1alpha1.Machine,
	machineVolume *computev1alpha1.Volume,
) (*ori.VolumeConnection, bool, error) {
	volume := &storagev1alpha1.Volume{}
	volumeKey := client.ObjectKey{Namespace: machine.Namespace, Name: computev1alpha1.MachineVolumeName(machine.Name, *machineVolume)}
	if err := r.Get(ctx, volumeKey, volume); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, false, fmt.Errorf("error getting volume: %w", err)
		}
		r.Eventf(machine, corev1.EventTypeNormal, events.VolumeNotReady, "Volume %s not found", volumeKey.Name)
		return nil, false, nil
	}

	if state := volume.Status.State; state != storagev1alpha1.VolumeStateAvailable {
		r.Eventf(machine, corev1.EventTypeNormal, events.VolumeNotReady, "Volume %s is in state %s", volumeKey.Name, state)
		return nil, false, nil
	}

	access := volume.Status.Access
	if access == nil {
		r.Eventf(machine, corev1.EventTypeNormal, events.VolumeNotReady, "Volume %s does not report status access", volumeKey.Name)
		return nil, false, nil
	}

	if !r.isVolumeBoundToMachine(machine, machineVolume.Name, volume) {
		return nil, false, nil
	}

	var secretData map[string][]byte
	if secretRef := access.SecretRef; secretRef != nil {
		secret := &corev1.Secret{}
		secretKey := client.ObjectKey{Namespace: volume.Namespace, Name: secretRef.Name}
		if err := r.Get(ctx, secretKey, secret); err != nil {
			if !apierrors.IsNotFound(err) {
				return nil, false, fmt.Errorf("error getting volume access secret %s: %w", secretKey.Name, err)
			}

			r.Eventf(machine, corev1.EventTypeNormal, events.VolumeNotReady,
				"Volume %s access secret %s not found",
				volumeKey.Name,
				secretKey.Name,
			)
			return nil, false, nil
		}

		secretData = secret.Data
	}

	return &ori.VolumeConnection{
		Driver:     access.Driver,
		Handle:     access.Handle,
		Attributes: access.VolumeAttributes,
		SecretData: secretData,
	}, true, nil
}

func (r *MachineReconciler) prepareORIVolume(
	ctx context.Context,
	log logr.Logger,
	machine *computev1alpha1.Machine,
	machineVolume *computev1alpha1.Volume,
) (*ori.Volume, bool, error) {
	name := machineVolume.Name
	switch {
	case machineVolume.EmptyDisk != nil:
		var sizeBytes uint64
		if sizeLimit := machineVolume.EmptyDisk.SizeLimit; sizeLimit != nil {
			sizeBytes = sizeLimit.AsDec().UnscaledBig().Uint64()
		}
		return &ori.Volume{
			Name:   name,
			Device: *machineVolume.Device,
			EmptyDisk: &ori.EmptyDisk{
				SizeBytes: sizeBytes,
			},
		}, true, nil
	case machineVolume.VolumeRef != nil || machineVolume.Ephemeral != nil:
		oriVolumeConnection, ok, err := r.prepareORIVolumeConnection(ctx, machine, machineVolume)
		if err != nil || !ok {
			return nil, ok, err
		}

		oriVolumeResources, ok, err := r.prepareORIVolumeResources(ctx, machine, machineVolume)
		if err != nil || !ok {
			return nil, ok, err
		}

		return &ori.Volume{
			Name:       name,
			Device:     *machineVolume.Device,
			Connection: oriVolumeConnection,
			Resources:  oriVolumeResources,
		}, true, nil
	default:
		return nil, false, fmt.Errorf("unrecognized volume %#v", machineVolume)
	}
}

func (r *MachineReconciler) prepareORIVolumeResources(ctx context.Context, machine *computev1alpha1.Machine, machineVolume *computev1alpha1.Volume) (*ori.VolumeResources, bool, error) {
	volume := &storagev1alpha1.Volume{}
	volumeKey := client.ObjectKey{Namespace: machine.Namespace, Name: computev1alpha1.MachineVolumeName(machine.Name, *machineVolume)}

	if err := r.Get(ctx, volumeKey, volume); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, false, fmt.Errorf("error getting volume: %w", err)
		}
		r.Eventf(machine, corev1.EventTypeNormal, events.VolumeNotReady, "Volume %s not found", volumeKey.Name)
		return nil, false, err
	}

	if state := volume.Status.State; state != storagev1alpha1.VolumeStateAvailable {
		r.Eventf(machine, corev1.EventTypeNormal, events.VolumeNotReady, "Volume %s is in state %s", volumeKey.Name, state)
		return nil, false, nil
	}

	if !r.isVolumeBoundToMachine(machine, machineVolume.Name, volume) {
		return nil, false, nil
	}

	var oriVolumeResources *ori.VolumeResources
	if resources := volume.Spec.Resources; resources != nil && resources.Storage() != nil {
		oriVolumeResources = &ori.VolumeResources{
			StorageBytes: resources.Storage().AsDec().UnscaledBig().Uint64(),
		}
	}
	return oriVolumeResources, true, nil
}

func (r *MachineReconciler) prepareORIVolumes(
	ctx context.Context,
	log logr.Logger,
	machine *computev1alpha1.Machine,
) ([]*ori.Volume, bool, error) {
	var (
		oriVolumes []*ori.Volume
		errs       []error
	)

	for _, machineVolume := range machine.Spec.Volumes {
		log := log.WithValues("MachineVolume", machineVolume.Name)
		oriVolume, ok, err := r.prepareORIVolume(ctx, log, machine, &machineVolume)
		if err != nil {
			errs = append(errs, fmt.Errorf("[volume %s] %w", machineVolume.Name, err))
			continue
		}
		if ok {
			oriVolumes = append(oriVolumes, oriVolume)
			continue
		}
	}

	if len(errs) > 0 {
		return nil, false, fmt.Errorf("error(s) preparing ori volume volumes: %w", errors.Join(errs...))
	}
	if len(oriVolumes) != len(machine.Spec.Volumes) {
		return nil, false, nil
	}
	return oriVolumes, true, nil
}

func (r *MachineReconciler) getExistingORIVolumesForMachine(
	ctx context.Context,
	log logr.Logger,
	oriMachine *ori.Machine,
	desiredORIVolumes []*ori.Volume,
) ([]*ori.Volume, error) {
	var (
		oriVolumes              []*ori.Volume
		desiredORIVolumesByName = utilslices.ToMapByKey(desiredORIVolumes, (*ori.Volume).GetName)
		errs                    []error
	)

	for _, oriVolume := range oriMachine.Spec.Volumes {
		log := log.WithValues("Volume", oriVolume.Name)

		desiredORIVolume, ok := desiredORIVolumesByName[oriVolume.Name]
		if ok && proto.Equal(desiredORIVolume, oriVolume) {
			log.V(1).Info("Existing ORI volume is up-to-date")
			oriVolumes = append(oriVolumes, oriVolume)
			continue
		}

		log.V(1).Info("Detaching outdated ORI volume")
		_, err := r.MachineRuntime.DetachVolume(ctx, &ori.DetachVolumeRequest{
			MachineId: oriMachine.Metadata.Id,
			Name:      oriVolume.Name,
		})
		if err != nil {
			if status.Code(err) != codes.NotFound {
				errs = append(errs, fmt.Errorf("[volume %s] %w", oriVolume.Name, err))
				continue
			}
		}
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return oriVolumes, nil
}

func (r *MachineReconciler) getNewORIVolumesForMachine(
	ctx context.Context,
	log logr.Logger,
	oriMachine *ori.Machine,
	desiredORIVolumes, existingORIVolumes []*ori.Volume,
) ([]*ori.Volume, error) {
	var (
		desiredNewORIVolumes = FindNewORIVolumes(desiredORIVolumes, existingORIVolumes)
		oriVolumes           []*ori.Volume
		errs                 []error
	)
	for _, newORIVolume := range desiredNewORIVolumes {
		log := log.WithValues("Volume", newORIVolume.Name)
		log.V(1).Info("Attaching new volume")
		if _, err := r.MachineRuntime.AttachVolume(ctx, &ori.AttachVolumeRequest{
			MachineId: oriMachine.Metadata.Id,
			Volume:    newORIVolume,
		}); err != nil {
			errs = append(errs, fmt.Errorf("[volume %s] %w", newORIVolume.Name, err))
			continue
		}

		oriVolumes = append(oriVolumes, newORIVolume)
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return oriVolumes, nil
}

func (r *MachineReconciler) updateORIVolumes(
	ctx context.Context,
	log logr.Logger,
	machine *computev1alpha1.Machine,
	oriMachine *ori.Machine,
) error {
	desiredORIVolumes, _, err := r.prepareORIVolumes(ctx, log, machine)
	if err != nil {
		return fmt.Errorf("error preparing ori volumes: %w", err)
	}

	extistingORIVolumes, err := r.getExistingORIVolumesForMachine(ctx, log, oriMachine, desiredORIVolumes)
	if err != nil {
		return fmt.Errorf("error getting existing ori volumes for machine: %w", err)
	}

	_, err = r.getNewORIVolumesForMachine(ctx, log, oriMachine, desiredORIVolumes, extistingORIVolumes)
	if err != nil {
		return fmt.Errorf("error getting new ori volumes for machine: %w", err)
	}

	return nil
}

func (r *MachineReconciler) getVolumeStatusesForMachine(
	machine *computev1alpha1.Machine,
	oriMachine *ori.Machine,
	now metav1.Time,
) ([]computev1alpha1.VolumeStatus, error) {
	var (
		oriVolumeStatusByName        = utilslices.ToMapByKey(oriMachine.Status.Volumes, (*ori.VolumeStatus).GetName)
		existingVolumeStatusesByName = utilslices.ToMapByKey(machine.Status.Volumes, func(status computev1alpha1.VolumeStatus) string { return status.Name })
		volumeStatuses               []computev1alpha1.VolumeStatus
		errs                         []error
	)

	for _, machineVolume := range machine.Spec.Volumes {
		var (
			oriVolumeStatus, ok = oriVolumeStatusByName[machineVolume.Name]
			volumeStatusValues  computev1alpha1.VolumeStatus
		)
		if ok {
			var err error
			volumeStatusValues, err = r.convertORIVolumeStatus(oriVolumeStatus)
			if err != nil {
				return nil, fmt.Errorf("[volume %s] %w", machineVolume.Name, err)
			}
		} else {
			volumeStatusValues = computev1alpha1.VolumeStatus{
				Name:  machineVolume.Name,
				State: computev1alpha1.VolumeStatePending,
			}
		}

		volumeStatus := existingVolumeStatusesByName[machineVolume.Name]
		r.addVolumeStatusValues(now, &volumeStatus, &volumeStatusValues)
		volumeStatuses = append(volumeStatuses, volumeStatus)
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return volumeStatuses, nil
}

var oriVolumeStateToVolumeState = map[ori.VolumeState]computev1alpha1.VolumeState{
	ori.VolumeState_VOLUME_ATTACHED: computev1alpha1.VolumeStateAttached,
	ori.VolumeState_VOLUME_PENDING:  computev1alpha1.VolumeStatePending,
}

func (r *MachineReconciler) convertORIVolumeState(oriState ori.VolumeState) (computev1alpha1.VolumeState, error) {
	if res, ok := oriVolumeStateToVolumeState[oriState]; ok {
		return res, nil
	}
	return "", fmt.Errorf("unknown ori volume state %v", oriState)
}

func (r *MachineReconciler) convertORIVolumeStatus(oriVolumeStatus *ori.VolumeStatus) (computev1alpha1.VolumeStatus, error) {
	state, err := r.convertORIVolumeState(oriVolumeStatus.State)
	if err != nil {
		return computev1alpha1.VolumeStatus{}, err
	}

	return computev1alpha1.VolumeStatus{
		Name:   oriVolumeStatus.Name,
		Handle: oriVolumeStatus.Handle,
		State:  state,
	}, nil
}

func (r *MachineReconciler) addVolumeStatusValues(now metav1.Time, existing, newValues *computev1alpha1.VolumeStatus) {
	if existing.State != newValues.State {
		existing.LastStateTransitionTime = &now
	}
	existing.Name = newValues.Name
	existing.State = newValues.State
	existing.Handle = newValues.Handle
}
