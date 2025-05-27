// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	"github.com/ironcore-dev/ironcore/poollet/machinepoollet/controllers/events"
	"github.com/ironcore-dev/ironcore/utils/claimmanager"
	utilslices "github.com/ironcore-dev/ironcore/utils/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
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

func (r *MachineReconciler) prepareRemoteIRIVolume(
	ctx context.Context,
	machine *computev1alpha1.Machine,
	machineVolume *computev1alpha1.Volume,
	volume *storagev1alpha1.Volume,
) (*iri.Volume, bool, error) {
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

	var encryptionData map[string][]byte
	if encryption := volume.Spec.Encryption; encryption != nil {
		secret := &corev1.Secret{}
		secretKey := client.ObjectKey{Namespace: volume.Namespace, Name: encryption.SecretRef.Name}
		if err := r.Get(ctx, secretKey, secret); err != nil {
			if !apierrors.IsNotFound(err) {
				return nil, false, fmt.Errorf("error getting volume encryption secret %s: %w", secretKey.Name, err)
			}

			r.Eventf(machine, corev1.EventTypeNormal, events.VolumeNotReady,
				"Volume %s encryption secret %s not found",
				volume.Name,
				secretKey.Name,
			)
			return nil, false, nil
		}

		encryptionData = secret.Data
	}

	effectiveSize := volume.Spec.Resources.Storage().Value()
	if resources := volume.Status.Resources; resources != nil {
		effectiveSize = resources.Storage().Value()
	}

	return &iri.Volume{
		Name:   machineVolume.Name,
		Device: *machineVolume.Device,
		Connection: &iri.VolumeConnection{
			Driver:                access.Driver,
			Handle:                access.Handle,
			Attributes:            access.VolumeAttributes,
			SecretData:            secretData,
			EncryptionData:        encryptionData,
			EffectiveStorageBytes: effectiveSize,
		},
	}, true, nil
}

func (r *MachineReconciler) prepareEmptyDiskIRIVolume(machineVolume *computev1alpha1.Volume) *iri.Volume {
	var sizeBytes int64
	if sizeLimit := machineVolume.EmptyDisk.SizeLimit; sizeLimit != nil {
		sizeBytes = sizeLimit.Value()
	}
	return &iri.Volume{
		Name:   machineVolume.Name,
		Device: *machineVolume.Device,
		EmptyDisk: &iri.EmptyDisk{
			SizeBytes: sizeBytes,
		},
	}
}

func (r *MachineReconciler) prepareIRIVolumes(
	ctx context.Context,
	machine *computev1alpha1.Machine,
	volumes []storagev1alpha1.Volume,
) ([]*iri.Volume, bool, error) {
	var (
		volumeNameToMachineVolume = r.volumeNameToMachineVolume(machine)
		iriVolumes                []*iri.Volume
		errs                      []error
	)
	for _, volume := range volumes {
		machineVolume := volumeNameToMachineVolume[volume.Name]
		iriVolume, ok, err := r.prepareRemoteIRIVolume(ctx, machine, &machineVolume, &volume)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if !ok {
			continue
		}

		iriVolumes = append(iriVolumes, iriVolume)
	}
	if err := errors.Join(errs...); err != nil {
		return nil, false, err
	}

	for _, machineVolume := range machine.Spec.Volumes {
		if machineVolume.EmptyDisk == nil {
			continue
		}

		iriVolume := r.prepareEmptyDiskIRIVolume(&machineVolume)
		iriVolumes = append(iriVolumes, iriVolume)
	}

	if len(iriVolumes) != len(machine.Spec.Volumes) {
		expectedVolumeNames := utilslices.ToSetFunc(machine.Spec.Volumes, func(v computev1alpha1.Volume) string { return v.Name })
		actualVolumeNames := utilslices.ToSetFunc(iriVolumes, (*iri.Volume).GetName)
		missingVolumeNames := sets.List(expectedVolumeNames.Difference(actualVolumeNames))
		r.Eventf(machine, corev1.EventTypeNormal, events.VolumeNotReady, "Machine volumes are not ready: %v", missingVolumeNames)
		return iriVolumes, false, nil
	}
	return iriVolumes, true, nil
}

func (r *MachineReconciler) getExistingIRIVolumesForMachine(
	ctx context.Context,
	log logr.Logger,
	iriMachine *iri.Machine,
	desiredIRIVolumes []*iri.Volume,
) ([]*iri.Volume, error) {
	var (
		iriVolumes              []*iri.Volume
		desiredIRIVolumesByName = utilslices.ToMapByKey(desiredIRIVolumes, (*iri.Volume).GetName)
		errs                    []error
	)

	for _, iriVolume := range iriMachine.Spec.Volumes {
		log := log.WithValues("Volume", iriVolume.Name)

		desiredIRIVolume, ok := desiredIRIVolumesByName[iriVolume.Name]
		if ok && proto.Equal(desiredIRIVolume, iriVolume) {
			log.V(1).Info("Existing IRI volume is up-to-date")
			iriVolumes = append(iriVolumes, iriVolume)
			continue
		}

		log.V(1).Info("Detaching outdated IRI volume")
		_, err := r.MachineRuntime.DetachVolume(ctx, &iri.DetachVolumeRequest{
			MachineId: iriMachine.Metadata.Id,
			Name:      iriVolume.Name,
		})
		if err != nil {
			if status.Code(err) != codes.NotFound {
				errs = append(errs, fmt.Errorf("[volume %s] %w", iriVolume.Name, err))
				continue
			}
		}
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return iriVolumes, nil
}

func (r *MachineReconciler) getNewIRIVolumesForMachine(
	ctx context.Context,
	log logr.Logger,
	iriMachine *iri.Machine,
	desiredIRIVolumes, existingIRIVolumes []*iri.Volume,
) ([]*iri.Volume, error) {
	var (
		desiredNewIRIVolumes = FindNewIRIVolumes(desiredIRIVolumes, existingIRIVolumes)
		iriVolumes           []*iri.Volume
		errs                 []error
	)
	for _, newIRIVolume := range desiredNewIRIVolumes {
		log := log.WithValues("Volume", newIRIVolume.Name)
		log.V(1).Info("Attaching new volume")
		if _, err := r.MachineRuntime.AttachVolume(ctx, &iri.AttachVolumeRequest{
			MachineId: iriMachine.Metadata.Id,
			Volume:    newIRIVolume,
		}); err != nil {
			errs = append(errs, fmt.Errorf("[volume %s] %w", newIRIVolume.Name, err))
			continue
		}

		iriVolumes = append(iriVolumes, newIRIVolume)
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return iriVolumes, nil
}

func (r *MachineReconciler) updateIRIVolumes(
	ctx context.Context,
	log logr.Logger,
	machine *computev1alpha1.Machine,
	iriMachine *iri.Machine,
	volumes []storagev1alpha1.Volume,
) error {
	desiredIRIVolumes, _, err := r.prepareIRIVolumes(ctx, machine, volumes)
	if err != nil {
		return fmt.Errorf("error preparing iri volumes: %w", err)
	}

	extistingIRIVolumes, err := r.getExistingIRIVolumesForMachine(ctx, log, iriMachine, desiredIRIVolumes)
	if err != nil {
		return fmt.Errorf("error getting existing iri volumes for machine: %w", err)
	}

	_, err = r.getNewIRIVolumesForMachine(ctx, log, iriMachine, desiredIRIVolumes, extistingIRIVolumes)
	if err != nil {
		return fmt.Errorf("error getting new iri volumes for machine: %w", err)
	}

	return nil
}

func (r *MachineReconciler) getVolumeStatusesForMachine(
	machine *computev1alpha1.Machine,
	iriMachine *iri.Machine,
	now metav1.Time,
) ([]computev1alpha1.VolumeStatus, error) {
	var (
		iriVolumeStatusByName        = utilslices.ToMapByKey(iriMachine.Status.Volumes, (*iri.VolumeStatus).GetName)
		existingVolumeStatusesByName = utilslices.ToMapByKey(machine.Status.Volumes, func(status computev1alpha1.VolumeStatus) string { return status.Name })
		volumeStatuses               []computev1alpha1.VolumeStatus
		errs                         []error
	)

	for _, machineVolume := range machine.Spec.Volumes {
		var (
			iriVolumeStatus, ok = iriVolumeStatusByName[machineVolume.Name]
			volumeStatusValues  computev1alpha1.VolumeStatus
		)
		volumeName := computev1alpha1.MachineVolumeName(machine.Name, machineVolume)
		if ok {
			var err error
			volumeStatusValues, err = r.convertIRIVolumeStatus(iriVolumeStatus, volumeName)
			if err != nil {
				return nil, fmt.Errorf("[volume %s] %w", machineVolume.Name, err)
			}
		} else {
			volumeStatusValues = computev1alpha1.VolumeStatus{
				Name:      machineVolume.Name,
				State:     computev1alpha1.VolumeStatePending,
				VolumeRef: corev1.LocalObjectReference{Name: volumeName},
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

var iriVolumeStateToVolumeState = map[iri.VolumeState]computev1alpha1.VolumeState{
	iri.VolumeState_VOLUME_ATTACHED: computev1alpha1.VolumeStateAttached,
	iri.VolumeState_VOLUME_PENDING:  computev1alpha1.VolumeStatePending,
}

func (r *MachineReconciler) convertIRIVolumeState(iriState iri.VolumeState) (computev1alpha1.VolumeState, error) {
	if res, ok := iriVolumeStateToVolumeState[iriState]; ok {
		return res, nil
	}
	return "", fmt.Errorf("unknown iri volume state %v", iriState)
}

func (r *MachineReconciler) convertIRIVolumeStatus(iriVolumeStatus *iri.VolumeStatus, volumeName string) (computev1alpha1.VolumeStatus, error) {
	state, err := r.convertIRIVolumeState(iriVolumeStatus.State)
	if err != nil {
		return computev1alpha1.VolumeStatus{}, err
	}

	return computev1alpha1.VolumeStatus{
		Name:      iriVolumeStatus.Name,
		Handle:    iriVolumeStatus.Handle,
		State:     state,
		VolumeRef: corev1.LocalObjectReference{Name: volumeName},
	}, nil
}

func (r *MachineReconciler) addVolumeStatusValues(now metav1.Time, existing, newValues *computev1alpha1.VolumeStatus) {
	if existing.State != newValues.State {
		existing.LastStateTransitionTime = &now
	}
	existing.Name = newValues.Name
	existing.State = newValues.State
	existing.Handle = newValues.Handle
	existing.VolumeRef = newValues.VolumeRef
}
