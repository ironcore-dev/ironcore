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

package controllers

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/gogo/protobuf/proto"
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	ori "github.com/ironcore-dev/ironcore/ori/apis/machine/v1alpha1"
	"github.com/ironcore-dev/ironcore/poollet/machinepoollet/controllers/events"
	"github.com/ironcore-dev/ironcore/utils/claimmanager"
	utilslices "github.com/ironcore-dev/ironcore/utils/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type volumeClaimStrategy struct {
	client.Client
}

func (s *volumeClaimStrategy) ClaimState(claimer client.Object, obj client.Object) claimmanager.ClaimState {
	volume := obj.(*storagev1alpha1.Volume)
	if claimRef := volume.Spec.ClaimRef; claimRef != nil {
		if claimRef.UID == claimer.GetUID() {
			return claimmanager.ClaimStateClaimed
		}
		return claimmanager.ClaimStateTaken
	}
	return claimmanager.ClaimStateFree
}

func (s *volumeClaimStrategy) Adopt(ctx context.Context, claimer client.Object, obj client.Object) error {
	volume := obj.(*storagev1alpha1.Volume)
	base := volume.DeepCopy()
	volume.Spec.ClaimRef = commonv1alpha1.NewLocalObjUIDRef(claimer)
	return s.Patch(ctx, volume, client.StrategicMergeFrom(base))
}

func (s *volumeClaimStrategy) Release(ctx context.Context, claimer client.Object, obj client.Object) error {
	volume := obj.(*storagev1alpha1.Volume)
	base := volume.DeepCopy()
	volume.Spec.ClaimRef = nil
	return s.Patch(ctx, volume, client.StrategicMergeFrom(base))
}

func (r *MachineReconciler) volumeNameToMachineVolume(machine *computev1alpha1.Machine) map[string]computev1alpha1.Volume {
	sel := make(map[string]computev1alpha1.Volume)
	for _, machineVolume := range machine.Spec.Volumes {
		volumeName := computev1alpha1.MachineVolumeName(machine.Name, machineVolume)
		if volumeName == "" {
			// volume name is empty on empty disk volumes.
			continue
		}
		sel[volumeName] = machineVolume
	}
	return sel
}

func (r *MachineReconciler) machineVolumeSelector(machine *computev1alpha1.Machine) claimmanager.Selector {
	names := sets.New(computev1alpha1.MachineVolumeNames(machine)...)
	return claimmanager.SelectorFunc(func(obj client.Object) bool {
		volume := obj.(*storagev1alpha1.Volume)
		return names.Has(volume.Name)
	})
}

func (r *MachineReconciler) getVolumesForMachine(ctx context.Context, machine *computev1alpha1.Machine) ([]storagev1alpha1.Volume, error) {
	volumeList := &storagev1alpha1.VolumeList{}
	if err := r.List(ctx, volumeList,
		client.InNamespace(machine.Namespace),
	); err != nil {
		return nil, fmt.Errorf("error listing volumes: %w", err)
	}

	var (
		sel      = r.machineVolumeSelector(machine)
		claimMgr = claimmanager.New(machine, sel, &volumeClaimStrategy{r.Client})
		volumes  []storagev1alpha1.Volume
		errs     []error
	)
	for _, volume := range volumeList.Items {
		ok, err := claimMgr.Claim(ctx, &volume)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if !ok {
			continue
		}

		if volume.Status.State != storagev1alpha1.VolumeStateAvailable || volume.Status.Access == nil {
			r.Eventf(machine, corev1.EventTypeNormal, events.VolumeNotReady, "Volume %s does not access information", volume.Name)
			continue
		}

		volumes = append(volumes, volume)
	}
	return volumes, errors.Join(errs...)
}

func (r *MachineReconciler) prepareRemoteORIVolume(
	ctx context.Context,
	machine *computev1alpha1.Machine,
	machineVolume *computev1alpha1.Volume,
	volume *storagev1alpha1.Volume,
) (*ori.Volume, bool, error) {
	access := volume.Status.Access
	if access == nil {
		r.Eventf(machine, corev1.EventTypeNormal, events.VolumeNotReady, "Volume %s does not report status access", volume.Name)
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
				volume.Name,
				secretKey.Name,
			)
			return nil, false, nil
		}

		secretData = secret.Data
	}

	return &ori.Volume{
		Name:   machineVolume.Name,
		Device: *machineVolume.Device,
		Connection: &ori.VolumeConnection{
			Driver:     access.Driver,
			Handle:     access.Handle,
			Attributes: access.VolumeAttributes,
			SecretData: secretData,
		},
	}, true, nil
}

func (r *MachineReconciler) prepareEmptyDiskORIVolume(machineVolume *computev1alpha1.Volume) *ori.Volume {
	var sizeBytes int64
	if sizeLimit := machineVolume.EmptyDisk.SizeLimit; sizeLimit != nil {
		sizeBytes = sizeLimit.Value()
	}
	return &ori.Volume{
		Name:   machineVolume.Name,
		Device: *machineVolume.Device,
		EmptyDisk: &ori.EmptyDisk{
			SizeBytes: sizeBytes,
		},
	}
}

func (r *MachineReconciler) prepareORIVolumes(
	ctx context.Context,
	machine *computev1alpha1.Machine,
	volumes []storagev1alpha1.Volume,
) ([]*ori.Volume, bool, error) {
	var (
		volumeNameToMachineVolume = r.volumeNameToMachineVolume(machine)
		oriVolumes                []*ori.Volume
		errs                      []error
	)
	for _, volume := range volumes {
		machineVolume := volumeNameToMachineVolume[volume.Name]
		oriVolume, ok, err := r.prepareRemoteORIVolume(ctx, machine, &machineVolume, &volume)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if !ok {
			continue
		}

		oriVolumes = append(oriVolumes, oriVolume)
	}
	if err := errors.Join(errs...); err != nil {
		return nil, false, err
	}

	for _, machineVolume := range machine.Spec.Volumes {
		if machineVolume.EmptyDisk == nil {
			continue
		}

		oriVolume := r.prepareEmptyDiskORIVolume(&machineVolume)
		oriVolumes = append(oriVolumes, oriVolume)
	}

	if len(oriVolumes) != len(machine.Spec.Volumes) {
		expectedVolumeNames := utilslices.ToSetFunc(machine.Spec.Volumes, func(v computev1alpha1.Volume) string { return v.Name })
		actualVolumeNames := utilslices.ToSetFunc(oriVolumes, (*ori.Volume).GetName)
		missingVolumeNames := sets.List(expectedVolumeNames.Difference(actualVolumeNames))
		r.Eventf(machine, corev1.EventTypeNormal, events.VolumeNotReady, "Machine volumes are not ready: %v", missingVolumeNames)
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
	volumes []storagev1alpha1.Volume,
) error {
	desiredORIVolumes, _, err := r.prepareORIVolumes(ctx, machine, volumes)
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
