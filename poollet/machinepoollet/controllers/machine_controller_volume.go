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
	"fmt"

	"github.com/go-logr/logr"
	"github.com/gogo/protobuf/proto"
	"github.com/onmetal/controller-utils/set"
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	machinepoolletv1alpha1 "github.com/onmetal/onmetal-api/poollet/machinepoollet/api/v1alpha1"
	"github.com/onmetal/onmetal-api/poollet/machinepoollet/controllers/events"
	utilslices "github.com/onmetal/onmetal-api/utils/slices"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *MachineReconciler) listORIVolumesByMachineUID(ctx context.Context, machineUID types.UID) ([]*ori.Volume, error) {
	res, err := r.MachineRuntime.ListVolumes(ctx, &ori.ListVolumesRequest{
		Filter: &ori.VolumeFilter{
			LabelSelector: r.machineUIDLabelSelector(machineUID),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error listing runtime volumes by machine uid %s: %w", machineUID, err)
	}
	return res.Volumes, nil
}

func (r *MachineReconciler) listORIVolumesByMachineKey(ctx context.Context, machineKey client.ObjectKey) ([]*ori.Volume, error) {
	res, err := r.MachineRuntime.ListVolumes(ctx, &ori.ListVolumesRequest{
		Filter: &ori.VolumeFilter{
			LabelSelector: r.machineKeyLabelSelector(machineKey),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error listing runtime volumes by machine key %s: %w", machineKey, err)
	}
	return res.Volumes, nil
}

func (r *MachineReconciler) oriVolumeLabels(machine *computev1alpha1.Machine, name string) map[string]string {
	lbls := r.oriMachineLabels(machine)
	lbls[machinepoolletv1alpha1.VolumeNameLabel] = name
	return lbls
}

func (r *MachineReconciler) machineVolumeFinder(machineVolumes []computev1alpha1.Volume) func(name string) *computev1alpha1.Volume {
	return func(name string) *computev1alpha1.Volume {
		return utilslices.FindRefFunc(machineVolumes, func(volume computev1alpha1.Volume) bool {
			return volume.Name == name
		})
	}
}

func (r *MachineReconciler) oriVolumeAttachmentFinder(specs []*ori.VolumeAttachment) func(name string) *ori.VolumeAttachment {
	return func(name string) *ori.VolumeAttachment {
		spec, _ := utilslices.FindFunc(specs, func(v *ori.VolumeAttachment) bool {
			return v.Name == name
		})
		return spec
	}
}

func (r *MachineReconciler) oriVolumeFinder(volumes []*ori.Volume) func(name, driver, handle string) *ori.Volume {
	return func(name, driver, handle string) *ori.Volume {
		volume, _ := utilslices.FindFunc(volumes, func(volume *ori.Volume) bool {
			actualName := volume.Metadata.Labels[machinepoolletv1alpha1.VolumeNameLabel]
			return volume.Metadata.DeletedAt == 0 &&
				actualName == name &&
				volume.Spec.Driver == driver &&
				volume.Spec.Handle == handle
		})
		return volume
	}
}

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

func (r *MachineReconciler) prepareORIVolumeSpec(
	ctx context.Context,
	machine *computev1alpha1.Machine,
	name string,
	volumeName string,
) (*ori.VolumeSpec, bool, error) {
	volume := &storagev1alpha1.Volume{}
	volumeKey := client.ObjectKey{Namespace: machine.Namespace, Name: volumeName}
	if err := r.Get(ctx, volumeKey, volume); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, false, fmt.Errorf("error getting volume: %w", err)
		}
		r.Eventf(machine, corev1.EventTypeNormal, events.VolumeNotReady, "Volume %s not found", volumeName)
		return nil, false, nil
	}

	if state := volume.Status.State; state != storagev1alpha1.VolumeStateAvailable {
		r.Eventf(machine, corev1.EventTypeNormal, events.VolumeNotReady, "Volume %s is in state %s", volumeName, state)
		return nil, false, nil
	}

	access := volume.Status.Access
	if access == nil {
		r.Eventf(machine, corev1.EventTypeNormal, events.VolumeNotReady, "Volume %s does not report status access", volumeName)
		return nil, false, nil
	}

	if !r.isVolumeBoundToMachine(machine, name, volume) {
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
				volumeName,
				secretKey.Name,
			)
			return nil, false, nil
		}

		secretData = secret.Data
	}

	return &ori.VolumeSpec{
		Driver:     access.Driver,
		Handle:     access.Handle,
		Attributes: access.VolumeAttributes,
		SecretData: secretData,
	}, true, nil
}

func (r *MachineReconciler) prepareORIVolumeAttachmentAndVolumeSpec(
	ctx context.Context,
	machine *computev1alpha1.Machine,
	volume computev1alpha1.Volume,
) (*ori.VolumeAttachment, *ori.VolumeSpec, bool, error) {
	name := volume.Name
	switch {
	case volume.EmptyDisk != nil:
		var sizeBytes uint64
		if sizeLimit := volume.EmptyDisk.SizeLimit; sizeLimit != nil {
			sizeBytes = sizeLimit.AsDec().UnscaledBig().Uint64()
		}
		return &ori.VolumeAttachment{
			Name:   name,
			Device: *volume.Device,
			EmptyDisk: &ori.EmptyDiskSpec{
				SizeBytes: sizeBytes,
			},
		}, nil, true, nil
	case volume.VolumeRef != nil || volume.Ephemeral != nil:
		volumeName := computev1alpha1.MachineVolumeName(machine.Name, volume)

		oriVolumeSpec, ok, err := r.prepareORIVolumeSpec(ctx, machine, name, volumeName)
		if err != nil || !ok {
			return nil, nil, ok, err
		}

		return &ori.VolumeAttachment{
			Name:   name,
			Device: *volume.Device,
		}, oriVolumeSpec, true, nil
	default:
		return nil, nil, false, fmt.Errorf("unrecognized volume %#v", volume)
	}
}

func (r *MachineReconciler) createORIVolume(
	ctx context.Context,
	machine *computev1alpha1.Machine,
	name string,
	spec *ori.VolumeSpec,
) (*ori.Volume, error) {
	res, err := r.MachineRuntime.CreateVolume(ctx, &ori.CreateVolumeRequest{
		Volume: &ori.Volume{
			Metadata: &ori.ObjectMetadata{
				Labels: r.oriVolumeLabels(machine, name),
			},
			Spec: spec,
		},
	})
	if err != nil {
		return nil, err
	}
	return res.Volume, nil
}

func (r *MachineReconciler) prepareORIVolumeAttachments(
	ctx context.Context,
	log logr.Logger,
	machine *computev1alpha1.Machine,
) ([]*ori.VolumeAttachment, bool, error) {
	log.V(1).Info("Listing ori volumes")
	oriVolumes, err := r.listORIVolumesByMachineUID(ctx, machine.UID)
	if err != nil {
		return nil, false, fmt.Errorf("error listing ori volumes: %w", err)
	}

	var (
		findORIVolume        = r.oriVolumeFinder(oriVolumes)
		oriVolumeAttachments []*ori.VolumeAttachment
		ok                   = true
		errs                 []error
		usedIDs              = set.New[string]()
	)

	for _, volume := range machine.Spec.Volumes {
		name := volume.Name
		oriVolumeAttachment, volumeUsedORIVolumeIDs, volumeOK, err := r.prepareORIVolumeAttachmentAndCreateVolumeIfRequired(ctx, log, machine, volume, findORIVolume)
		usedIDs.Insert(volumeUsedORIVolumeIDs...)
		if err != nil {
			errs = append(errs, fmt.Errorf("[attachment %s] %w", name, err))
			continue
		}
		if !volumeOK {
			ok = false
			continue
		}

		oriVolumeAttachments = append(oriVolumeAttachments, oriVolumeAttachment)
	}

	for _, oriVolume := range oriVolumes {
		oriVolumeID := oriVolume.Metadata.Id
		if usedIDs.Has(oriVolumeID) {
			continue
		}

		log := log.WithValues("ORIVolumeID", oriVolumeID)
		log.V(1).Info("Deleting unused ori volume")
		if _, err := r.MachineRuntime.DeleteVolume(ctx, &ori.DeleteVolumeRequest{
			VolumeId: oriVolumeID,
		}); err != nil && status.Code(err) != codes.NotFound {
			log.Error(err, "Error deleting unused ori volume")
		} else {
			log.V(1).Info("Deleted unused ori volume")
		}
	}

	if len(errs) > 0 {
		return nil, false, fmt.Errorf("error(s) preparing ori volume attachments: %v", errs)
	}
	if !ok {
		return nil, false, nil
	}
	return oriVolumeAttachments, true, nil
}

func (r *MachineReconciler) prepareORIVolumeAttachmentAndCreateVolumeIfRequired(
	ctx context.Context,
	log logr.Logger,
	machine *computev1alpha1.Machine,
	volume computev1alpha1.Volume,
	findORIVolume func(name string, driver string, handle string) *ori.Volume,
) (volumeAttachment *ori.VolumeAttachment, usedVolumeIDs []string, ok bool, err error) {
	name := volume.Name
	oriVolumeAttachment, oriVolumeSpec, volumeOK, err := r.prepareORIVolumeAttachmentAndVolumeSpec(ctx, machine, volume)
	if err != nil {
		return nil, nil, false, fmt.Errorf("error preparing ori volume attachment: %w", err)
	}
	if !volumeOK {
		return nil, nil, false, nil
	}

	if oriVolumeSpec != nil {
		oriVolume := findORIVolume(name, oriVolumeSpec.Driver, oriVolumeSpec.Handle)
		if oriVolume == nil {
			log.V(1).Info("Creating ori volume")
			v, err := r.createORIVolume(ctx, machine, name, oriVolumeSpec)
			if err != nil {
				return nil, nil, false, fmt.Errorf("error creating ori ori volume: %w", err)
			}

			oriVolume = v
		}

		oriVolumeID := oriVolume.Metadata.Id
		usedVolumeIDs = append(usedVolumeIDs, oriVolumeID)
		oriVolumeAttachment.VolumeId = oriVolumeID
	}

	return oriVolumeAttachment, usedVolumeIDs, true, nil
}

func (r *MachineReconciler) updateORIVolume(
	ctx context.Context,
	log logr.Logger,
	machine *computev1alpha1.Machine,
	volume computev1alpha1.Volume,
	findORIVolume func(name, driver, handle string) *ori.Volume,
	oriMachine *ori.Machine,
	existingVolumeAttachment *ori.VolumeAttachment,
) (usedVolumeIDs []string, err error) {
	name := volume.Name
	machineID := oriMachine.Metadata.Id

	addExistingVolumeIDIfPresent := func() {
		if existingVolumeAttachment != nil {
			if volumeID := existingVolumeAttachment.VolumeId; volumeID != "" {
				usedVolumeIDs = append(usedVolumeIDs, volumeID)
			}
		}
	}

	oriVolumeAttachment, oriVolumeSpec, volumeOK, err := r.prepareORIVolumeAttachmentAndVolumeSpec(ctx, machine, volume)
	if err != nil {
		addExistingVolumeIDIfPresent()
		return usedVolumeIDs, fmt.Errorf("error preparing ori volume attachment: %w", err)
	}
	if !volumeOK {
		if existingVolumeAttachment != nil {
			log.V(1).Info("Deleting outdated ori volume attachment")
			if _, err := r.MachineRuntime.DeleteVolumeAttachment(ctx, &ori.DeleteVolumeAttachmentRequest{
				MachineId: machineID,
				Name:      name,
			}); err != nil && status.Code(err) != codes.NotFound {
				addExistingVolumeIDIfPresent()
				return usedVolumeIDs, fmt.Errorf("error deleting outdated ori volume attachment: %w", err)
			}
		}
		return nil, nil
	}

	if oriVolumeSpec != nil {
		oriVolume := findORIVolume(name, oriVolumeSpec.Driver, oriVolumeSpec.Handle)
		if oriVolume == nil {
			v, err := r.createORIVolume(ctx, machine, name, oriVolumeSpec)
			if err != nil {
				addExistingVolumeIDIfPresent()
				return usedVolumeIDs, fmt.Errorf("error creating ori volume: %w", err)
			}

			oriVolume = v
		}

		oriVolumeID := oriVolume.Metadata.Id
		usedVolumeIDs = append(usedVolumeIDs, oriVolumeID)
		oriVolumeAttachment.VolumeId = oriVolumeID
	}

	if existingVolumeAttachment != nil {
		if proto.Equal(existingVolumeAttachment, oriVolumeAttachment) {
			log.V(1).Info("Existing ori volume attachment is up-to-date")
			return usedVolumeIDs, nil
		}

		log.V(1).Info("Existing ori volume attachment is outdated, deleting")
		if _, err := r.MachineRuntime.DeleteVolumeAttachment(ctx, &ori.DeleteVolumeAttachmentRequest{
			MachineId: machineID,
			Name:      name,
		}); err != nil && status.Code(err) != codes.NotFound {
			addExistingVolumeIDIfPresent()
			return usedVolumeIDs, fmt.Errorf("error deleting outdate ori volume attachment: %w", err)
		}
	}

	log.V(1).Info("Creating new ori volume attachment")
	if _, err := r.MachineRuntime.CreateVolumeAttachment(ctx, &ori.CreateVolumeAttachmentRequest{
		MachineId: machineID,
		Volume:    oriVolumeAttachment,
	}); err != nil {
		return usedVolumeIDs, fmt.Errorf("error creating ori volume attachment: %w", err)
	}

	return usedVolumeIDs, nil
}

func (r *MachineReconciler) updateORIVolumes(ctx context.Context, log logr.Logger, machine *computev1alpha1.Machine, oriMachine *ori.Machine) error {
	machineID := oriMachine.Metadata.Id

	log.V(1).Info("Listing ori volumes")
	oriVolumes, err := r.listORIVolumesByMachineUID(ctx, machine.UID)
	if err != nil {
		return fmt.Errorf("error listing ori volumes: %w", err)
	}

	var (
		errs                    []error
		findVolume              = r.machineVolumeFinder(machine.Spec.Volumes)
		findORIVolume           = r.oriVolumeFinder(oriVolumes)
		findORIVolumeAttachment = r.oriVolumeAttachmentFinder(oriMachine.Spec.Volumes)
		usedORIVolumeIDs        = set.New[string]()
	)

	for _, oriVolume := range oriMachine.Spec.Volumes {
		name := oriVolume.Name
		if volume := findVolume(name); volume != nil {
			continue
		}

		log.V(1).Info("Deleting outdated ori volume attachment")
		if _, err := r.MachineRuntime.DeleteVolumeAttachment(ctx, &ori.DeleteVolumeAttachmentRequest{
			MachineId: machineID,
			Name:      name,
		}); err != nil && status.Code(err) != codes.NotFound {
			if volumeID := oriVolume.VolumeId; volumeID != "" {
				usedORIVolumeIDs.Insert(volumeID)
			}
			errs = append(errs, fmt.Errorf("[attachment %s] error deleting outdated ori volume attachment: %w", name, err))
		}
	}

	for _, volume := range machine.Spec.Volumes {
		name := volume.Name
		existingORIVolumeAttachment := findORIVolumeAttachment(name)
		volumeUsedVolumeIDs, err := r.updateORIVolume(ctx, log, machine, volume, findORIVolume, oriMachine, existingORIVolumeAttachment)
		usedORIVolumeIDs.Insert(volumeUsedVolumeIDs...)
		if err != nil {
			errs = append(errs, fmt.Errorf("[attachment %s] %w", name, err))
			continue
		}
	}

	for _, oriVolume := range oriVolumes {
		volumeID := oriVolume.Metadata.Id
		if !usedORIVolumeIDs.Has(volumeID) {
			log := log.WithValues("VolumeID", volumeID)
			log.V(1).Info("Deleting unused ori volume")
			if _, err := r.MachineRuntime.DeleteVolume(ctx, &ori.DeleteVolumeRequest{
				VolumeId: volumeID,
			}); err != nil && status.Code(err) != codes.NotFound {
				log.Error(err, "Error deleting unused ori volume")
			} else {
				log.V(1).Info("Deleted unused ori volume")
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("error(s) updating ori volumes: %v", errs)
	}
	return nil
}

func (r *MachineReconciler) updateVolumeStates(machine *computev1alpha1.Machine, oriMachine *ori.Machine, now metav1.Time) error {
	seenNames := set.New[string]()
	for _, oriVolumeAttachmentStatus := range oriMachine.Status.Volumes {
		seenNames.Insert(oriVolumeAttachmentStatus.Name)
		newVolumeStatus, err := r.convertORIVolumeAttachmentStatus(oriVolumeAttachmentStatus)
		if err != nil {
			return fmt.Errorf("error converting ori volume attachment %s status: %w", oriVolumeAttachmentStatus.Name, err)
		}

		idx := slices.IndexFunc(
			machine.Status.Volumes,
			func(status computev1alpha1.VolumeStatus) bool {
				return status.Name == oriVolumeAttachmentStatus.Name
			},
		)
		if idx < 0 {
			newVolumeStatus.LastStateTransitionTime = &now
			machine.Status.Volumes = append(machine.Status.Volumes, newVolumeStatus)
		} else {
			volumeStatus := &machine.Status.Volumes[idx]
			volumeStatus.Handle = newVolumeStatus.Handle
			lastStateTransitionTime := volumeStatus.LastStateTransitionTime
			if volumeStatus.State != newVolumeStatus.State {
				lastStateTransitionTime = &now
			}
			volumeStatus.LastStateTransitionTime = lastStateTransitionTime
			volumeStatus.State = newVolumeStatus.State
		}
	}

	for i := range machine.Status.Volumes {
		volumeStatus := &machine.Status.Volumes[i]
		if seenNames.Has(volumeStatus.Name) {
			continue
		}

		newState := computev1alpha1.VolumeStateDetached
		if volumeStatus.State != newState {
			volumeStatus.LastStateTransitionTime = &now
		}
		volumeStatus.State = newState
	}
	return nil
}

var oriVolumeAttachmentStateToVolumeState = map[ori.VolumeAttachmentState]computev1alpha1.VolumeState{
	ori.VolumeAttachmentState_VOLUME_ATTACHMENT_ATTACHED: computev1alpha1.VolumeStateAttached,
	ori.VolumeAttachmentState_VOLUME_ATTACHMENT_DETACHED: computev1alpha1.VolumeStateDetached,
	ori.VolumeAttachmentState_VOLUME_ATTACHMENT_PENDING:  computev1alpha1.VolumeStatePending,
}

func (r *MachineReconciler) convertORIVolumeAttachmentState(oriState ori.VolumeAttachmentState) (computev1alpha1.VolumeState, error) {
	if res, ok := oriVolumeAttachmentStateToVolumeState[oriState]; ok {
		return res, nil
	}
	return "", fmt.Errorf("unknown ori volume attachment state %v", oriState)
}

func (r *MachineReconciler) convertORIVolumeAttachmentStatus(oriVolumeStatus *ori.VolumeAttachmentStatus) (computev1alpha1.VolumeStatus, error) {
	state, err := r.convertORIVolumeAttachmentState(oriVolumeStatus.State)
	if err != nil {
		return computev1alpha1.VolumeStatus{}, err
	}

	return computev1alpha1.VolumeStatus{
		Name:   oriVolumeStatus.Name,
		Handle: oriVolumeStatus.VolumeHandle,
		State:  state,
	}, nil
}

func (r *MachineReconciler) deleteVolumes(ctx context.Context, log logr.Logger, volumes []*ori.Volume) (bool, error) {
	var (
		errs                 []error
		deletingORIVolumeIDs []string
	)
	for _, volume := range volumes {
		volumeID := volume.Metadata.Id
		log := log.WithValues("VolumeID", volumeID)

		log.V(1).Info("Deleting ori volume")
		if _, err := r.MachineRuntime.DeleteVolume(ctx, &ori.DeleteVolumeRequest{
			VolumeId: volumeID,
		}); err != nil {
			if status.Code(err) != codes.NotFound {
				errs = append(errs, fmt.Errorf("error deleting ori volume %s: %w", volumeID, err))
				continue
			}

			deletingORIVolumeIDs = append(deletingORIVolumeIDs, volumeID)
		}
	}

	switch {
	case len(errs) > 0:
		return false, fmt.Errorf("error(s) deleting ori volumes: %v", errs)
	case len(deletingORIVolumeIDs) > 0:
		log.V(1).Info("ORI Volumes are deleting", "DeletingORIVolumeIDs", deletingORIVolumeIDs)
		return false, nil
	default:
		log.V(1).Info("All ori volumes are gone")
		return true, nil
	}
}

func (r *MachineReconciler) deleteVolumesByMachineUID(ctx context.Context, log logr.Logger, machineUID types.UID) (bool, error) {
	log.V(1).Info("Listing ori volumes by machine uid")
	volumes, err := r.listORIVolumesByMachineUID(ctx, machineUID)
	if err != nil {
		return false, fmt.Errorf("error listing ori volumes by machine uid: %w", err)
	}

	return r.deleteVolumes(ctx, log, volumes)
}

func (r *MachineReconciler) deleteVolumesByMachineKey(ctx context.Context, log logr.Logger, machineKey client.ObjectKey) (bool, error) {
	log.V(1).Info("Listing ori volumes by machine key")
	volumes, err := r.listORIVolumesByMachineKey(ctx, machineKey)
	if err != nil {
		return false, fmt.Errorf("error listing ori volumes by machine key: %w", err)
	}

	return r.deleteVolumes(ctx, log, volumes)
}
