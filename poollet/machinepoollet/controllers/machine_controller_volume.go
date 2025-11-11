// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	"golang.org/x/exp/maps"

	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	poolletutils "github.com/ironcore-dev/ironcore/poollet/common/utils"
	"github.com/ironcore-dev/ironcore/poollet/machinepoollet/api/v1alpha1"
	"github.com/ironcore-dev/ironcore/poollet/machinepoollet/controllers/events"
	"github.com/ironcore-dev/ironcore/utils/claimmanager"
	utilsmaps "github.com/ironcore-dev/ironcore/utils/maps"
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

func (r *MachineReconciler) iriVolumeLabels(volume *storagev1alpha1.Volume) (map[string]string, error) {
	labels := map[string]string{
		v1alpha1.VolumeUIDLabel:       string(volume.UID),
		v1alpha1.VolumeNamespaceLabel: volume.Namespace,
		v1alpha1.VolumeNameLabel:      volume.Name,
	}
	apiLabels, err := poolletutils.PrepareDownwardAPILabels(volume, r.VolumeDownwardAPILabels, v1alpha1.MachineDownwardAPIPrefix)
	if err != nil {
		return nil, err
	}
	labels = utilsmaps.AppendMap(labels, apiLabels)
	return labels, nil
}

func (r *MachineReconciler) prepareVolumeAttributes(
	volAttributes map[string]string,
	labels map[string]string,
) (map[string]string, error) {
	var attributes map[string]string
	if volAttributes != nil {
		attributes = maps.Clone(volAttributes)
	} else {
		attributes = make(map[string]string)
	}
	labelsJSON, err := json.Marshal(labels)
	if err != nil {
		return nil, fmt.Errorf("error marshaling volume labels: %w", err)
	}
	attributes[v1alpha1.VolumeLabelsAttributeKey] = string(labelsJSON)

	return attributes, nil
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
	labels, err := r.iriVolumeLabels(volume)
	if err != nil {
		return nil, false, fmt.Errorf("error preparing iri volume labels: %w", err)
	}
	attributes, err := r.prepareVolumeAttributes(access.VolumeAttributes, labels)
	if err != nil {
		return nil, false, err
	}

	return &iri.Volume{
		Name:   machineVolume.Name,
		Device: *machineVolume.Device,
		Connection: &iri.VolumeConnection{
			Driver:                access.Driver,
			Handle:                access.Handle,
			Attributes:            attributes,
			SecretData:            secretData,
			EncryptionData:        encryptionData,
			EffectiveStorageBytes: effectiveSize,
		},
	}, true, nil
}

func (r *MachineReconciler) prepareLocalDiskIRIVolume(machineVolume *computev1alpha1.Volume) *iri.Volume {
	var sizeBytes int64
	if sizeLimit := machineVolume.LocalDisk.SizeLimit; sizeLimit != nil {
		sizeBytes = sizeLimit.Value()
	}

	var imageSpec *iri.ImageSpec
	if image := machineVolume.LocalDisk.Image; image != "" {
		imageSpec = &iri.ImageSpec{
			Image: machineVolume.LocalDisk.Image,
		}
	}

	return &iri.Volume{
		Name:   machineVolume.Name,
		Device: *machineVolume.Device,
		LocalDisk: &iri.LocalDisk{
			SizeBytes: sizeBytes,
			Image:     imageSpec,
		},
	}
}

func (r *MachineReconciler) prepareIRIVolumes(
	ctx context.Context,
	log logr.Logger,
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
		if machineVolume.LocalDisk != nil {
			iriVolume := r.prepareLocalDiskIRIVolume(&machineVolume)
			iriVolumes = append(iriVolumes, iriVolume)
		}
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

		// Volume is up-to-date, keep it as is
		if ok && proto.Equal(desiredIRIVolume, iriVolume) {
			log.V(1).Info("Existing IRI volume is up-to-date")
			iriVolumes = append(iriVolumes, iriVolume)
			continue
		}

		// Volume exists but needs updates, update it in-place
		if ok && r.shouldUpdateVolume(iriVolume, desiredIRIVolume) {
			log.V(1).Info("Updating volume")
			_, err := r.MachineRuntime.UpdateVolume(ctx, &iri.UpdateVolumeRequest{
				MachineId: iriMachine.Metadata.Id,
				Volume:    desiredIRIVolume,
			})
			if err != nil {
				errs = append(errs, fmt.Errorf("[volume %s] %w", iriVolume.Name, err))
				continue
			}
			iriVolumes = append(iriVolumes, desiredIRIVolume)
			continue
		}

		// Detach volume if not desired or has other changes
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

func (r *MachineReconciler) shouldUpdateVolume(iriVolume, desiredIRIVolume *iri.Volume) bool {
	if iriVolume.Connection == nil || desiredIRIVolume.Connection == nil {
		return false
	}

	if iriVolume.Connection.Driver != desiredIRIVolume.Connection.Driver ||
		iriVolume.Connection.Handle != desiredIRIVolume.Connection.Handle {
		return false
	}

	if iriVolume.Connection.EffectiveStorageBytes != desiredIRIVolume.Connection.EffectiveStorageBytes {
		return true
	}

	// TODO: Add support for credential rotation

	return false
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
	desiredIRIVolumes, _, err := r.prepareIRIVolumes(ctx, log, machine, volumes)
	if err != nil {
		return fmt.Errorf("error preparing iri volumes: %w", err)
	}

	existingIRIVolumes, err := r.getExistingIRIVolumesForMachine(ctx, log, iriMachine, desiredIRIVolumes)
	if err != nil {
		return fmt.Errorf("error getting existing iri volumes for machine: %w", err)
	}

	_, err = r.getNewIRIVolumesForMachine(ctx, log, iriMachine, desiredIRIVolumes, existingIRIVolumes)
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

	for _, iriVolume := range iriMachine.Status.Volumes {
		var volumeStatusValues computev1alpha1.VolumeStatus
		volumeStatusValues, err := r.convertIRIVolumeStatus(iriVolume, iriVolume.Name)
		if err != nil {
			return nil, fmt.Errorf("[volume %s] %w", iriVolume.Name, err)
		}
		volumeStatus := existingVolumeStatusesByName[iriVolume.Name]
		r.addVolumeStatusValues(now, &volumeStatus, &volumeStatusValues)
		volumeStatuses = append(volumeStatuses, volumeStatus)
	}

	for _, machineVolume := range machine.Spec.Volumes {
		var (
			_, ok              = iriVolumeStatusByName[machineVolume.Name]
			volumeStatusValues computev1alpha1.VolumeStatus
		)
		volumeName := computev1alpha1.MachineVolumeName(machine.Name, machineVolume)
		if !ok {
			volumeStatusValues = computev1alpha1.VolumeStatus{
				Name:      machineVolume.Name,
				State:     computev1alpha1.VolumeStatePending,
				VolumeRef: corev1.LocalObjectReference{Name: volumeName},
			}
			volumeStatus := existingVolumeStatusesByName[machineVolume.Name]
			r.addVolumeStatusValues(now, &volumeStatus, &volumeStatusValues)
			volumeStatuses = append(volumeStatuses, volumeStatus)
		}
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
